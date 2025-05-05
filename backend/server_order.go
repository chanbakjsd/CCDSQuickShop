package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/chanbakjsd/CCDSQuickShop/backend/db"
	"github.com/stripe/stripe-go/v81"
)

type OrderResponse struct {
	Orders []Order `json:"orders"`
}

type Order struct {
	OrderID          string      `json:"id"`
	Name             string      `json:"name"`
	Email            string      `json:"email"`
	MatricNumber     string      `json:"matricNumber"`
	PaymentReference string      `json:"paymentRef"`
	SalePeriod       string      `json:"salePeriod"`
	PaymentTime      *time.Time  `json:"paymentTime"`
	CollectionTime   *time.Time  `json:"collectionTime"`
	Cancelled        bool        `json:"cancelled"`
	Coupon           *Coupon     `json:"coupon"`
	Items            []OrderItem `json:"items"`
}

type OrderItem struct {
	ProductID string `json:"id"`
	Name      string `json:"name"`
	Variant   string `json:"variant"`
	ImageURL  string `json:"imageURL"`
	Amount    int    `json:"amount"`
	UnitPrice int    `json:"unitPrice"`
}

func (s *Server) OrderLookup(w http.ResponseWriter, req *http.Request) {
	includeCancelled := req.URL.Query().Get("include_cancelled") != ""
	allowFromItem := req.URL.Query().Get("from_item") != ""
	if (includeCancelled || allowFromItem) && !s.authCheck(w, req) {
		return
	}
	ctx := req.Context()
	orderID := req.PathValue("id")
	dbOrders, err := s.Queries.LookupOrder(ctx, db.LookupOrderParams{
		ID:               orderID,
		IncludeCancelled: includeCancelled,
	})
	if (err == nil || errors.Is(err, sql.ErrNoRows)) && allowFromItem && strings.Contains(orderID, ", ") {
		split := strings.SplitN(orderID, ", ", 2)
		orders, dbErr := s.Queries.LookupOrderFromItem(ctx, db.LookupOrderFromItemParams{
			ProductName: split[0],
			Variant:     split[1],
		})
		for _, order := range orders {
			dbOrders = append(dbOrders, db.LookupOrderRow(order))
		}
		err = dbErr
	}
	switch {
	case errors.Is(err, sql.ErrNoRows):
		http.Error(w, "Invalid order ID", http.StatusNotFound)
		return
	case err != nil:
		slog.Error("error looking up order", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	orders := make([]Order, 0, len(dbOrders))
	for _, dbOrder := range dbOrders {
		dbOrderItems, err := s.Queries.ListOrderItems(ctx, dbOrder.OrderID)
		if err != nil {
			slog.Error("error looking up order items", "err", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		var coupon *Coupon
		if dbOrder.CouponID.Valid {
			dbCoupon, err := s.Queries.CouponByID(ctx, dbOrder.CouponID.Int64)
			if err != nil {
				slog.Error("error looking up coupon", "err", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			c := s.dbCouponToCoupon(dbCoupon, false)
			coupon = &c
		}
		orderItems := make([]OrderItem, 0, len(dbOrderItems))
		for _, item := range dbOrderItems {
			orderItems = append(orderItems, OrderItem{
				ProductID: item.ProductID,
				Name:      item.ProductName,
				Variant:   item.Variant,
				ImageURL:  item.ImageUrl,
				Amount:    int(item.Amount),
				UnitPrice: int(item.UnitPrice),
			})
		}
		emailSplit := strings.SplitN(dbOrder.Email, "@", 2)
		emailSplit[0] = censorBack(emailSplit[0], 3, 10, ' ')
		order := Order{
			OrderID:          dbOrder.OrderID,
			Name:             censorBack(dbOrder.Name, 4, 10, ' '),
			Email:            strings.Join(emailSplit, "@"),
			MatricNumber:     censorFront(dbOrder.MatricNumber, 4, 10, ' '),
			PaymentReference: censorFront(dbOrder.PaymentReference.String, 8, 10, ' '),
			SalePeriod:       dbOrder.AdminName,
			Cancelled:        dbOrder.Cancelled,
			Coupon:           coupon,
			Items:            orderItems,
		}
		if dbOrder.PaymentTime.Valid {
			order.PaymentTime = &dbOrder.PaymentTime.Time
		}
		if dbOrder.CollectionTime.Valid {
			order.CollectionTime = &dbOrder.CollectionTime.Time
		}
		orders = append(orders, order)
	}
	if err := json.NewEncoder(w).Encode(OrderResponse{
		Orders: orders,
	}); err != nil {
		slog.Error("error writing order response", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (s *Server) OrderCollect(w http.ResponseWriter, req *http.Request) {
	if !s.authCheck(w, req) {
		return
	}
	ctx := req.Context()
	orderID := req.PathValue("id")
	err := s.Queries.UpdateCollectionTime(ctx, db.UpdateCollectionTimeParams{
		CollectionTime: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
		OrderID: orderID,
	})
	switch {
	case errors.Is(err, sql.ErrNoRows):
		http.Error(w, "Invalid order ID", http.StatusNotFound)
		return
	case err != nil:
		slog.Error("error marking order as collected", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) OrderCancel(w http.ResponseWriter, req *http.Request) {
	if !s.authCheck(w, req) {
		return
	}
	ctx := req.Context()
	orderID := req.PathValue("id")
	paymentRef, err := s.Queries.UpdateCancelled(ctx, db.UpdateCancelledParams{
		Cancelled: true,
		OrderID:   orderID,
	})
	switch {
	case errors.Is(err, sql.ErrNoRows):
		http.Error(w, "Invalid order ID", http.StatusNotFound)
		return
	case err != nil:
		slog.Error("error marking order as cancelled", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if _, err := s.Stripe.CheckoutSessions.Expire(paymentRef.String, &stripe.CheckoutSessionExpireParams{}); err != nil {
		slog.Error("error expiring checkout sessions on Stripe", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type OrderSummaryEntry struct {
	Name    string `json:"name"`
	Variant string `json:"variant"`
	Count   int    `json:"count"`
}

type OrderSummaryResponse struct {
	Unfulfilled           []OrderSummaryEntry `json:"unfulfilled"`
	OrderIDSamples        []string            `json:"order_id_samples"`
	UnfulfilledOrderCount int                 `json:"unfulfilled_order_count"`
	FulfilledOrderCount   int                 `json:"fulfilled_order_count"`
}

func (s *Server) OrderSummary(w http.ResponseWriter, req *http.Request) {
	if !s.authCheck(w, req) {
		return
	}
	ctx := req.Context()
	showCollected := req.URL.Query().Get("show_collected") != ""
	salePeriod, ok := s.resolveSalePeriod(w, req, req.PathValue("sale_id"))
	if !ok {
		return
	}
	summary, err := s.Queries.OrderSummary(ctx, db.OrderSummaryParams{
		SalePeriod:        salePeriod,
		ShowOnlyCollected: !showCollected,
	})
	switch {
	case errors.Is(err, sql.ErrNoRows):
		summary = []db.OrderSummaryRow{}
	case err != nil:
		slog.Error("error fetching order summary", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	orderIDs, err := s.Queries.UnfulfilledOrderIDs(ctx, db.UnfulfilledOrderIDsParams{
		MaxCount:   10,
		SalePeriod: salePeriod,
	})
	switch {
	case errors.Is(err, sql.ErrNoRows):
	case err != nil:
		slog.Error("error fetching unfulfilled order IDs", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if orderIDs == nil {
		orderIDs = []string{}
	}
	entries := make([]OrderSummaryEntry, 0, len(summary))
	for _, v := range summary {
		entries = append(entries, OrderSummaryEntry{
			Name:    v.ProductName,
			Variant: v.Variant,
			Count:   int(v.Sum.Float64),
		})
	}
	unfulfilledCount := 0
	fulfilledCount := 0
	orderCount, err := s.Queries.OrderNumberStats(ctx, salePeriod)
	switch {
	case errors.Is(err, sql.ErrNoRows):
	case err != nil:
		slog.Error("error fetching order number statistics", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	for _, count := range orderCount {
		if count.Uncollected {
			unfulfilledCount = int(count.OrderCount)
		} else {
			fulfilledCount = int(count.OrderCount)
		}
	}
	if err := json.NewEncoder(w).Encode(OrderSummaryResponse{
		Unfulfilled:           entries,
		OrderIDSamples:        orderIDs,
		UnfulfilledOrderCount: unfulfilledCount,
		FulfilledOrderCount:   fulfilledCount,
	}); err != nil {
		slog.Error("error writing order summary response", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

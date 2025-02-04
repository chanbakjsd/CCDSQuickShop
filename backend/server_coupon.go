package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/chanbakjsd/CCDSQuickShop/backend/shop"
)

type CouponsResponse struct {
	Coupons []Coupon `json:"coupons"`
}

type Coupon struct {
	Requirements []json.RawMessage `json:"requirements"`
	CouponCode   string            `json:"couponCode"`
	Discount     json.RawMessage   `json:"discount"`
}

func (s *Server) Coupons(w http.ResponseWriter, req *http.Request) {
	if !s.closureCheck(w, req) {
		return
	}
	dbCoupons, err := s.Queries.ListPublicCoupons(req.Context())
	switch {
	case errors.Is(err, context.Canceled):
		// Cancelled by user. Do nothing.
		return
	case err != nil:
		slog.Error("error fetching coupons", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	coupons := make([]Coupon, 0, len(dbCoupons))
	for _, coupon := range dbCoupons {
		coupons = append(coupons, dbCouponToCoupon(coupon))
	}
	if err := json.NewEncoder(w).Encode(CouponsResponse{
		Coupons: coupons,
	}); err != nil {
		slog.Error("error writing coupons response", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (s *Server) CouponLookup(w http.ResponseWriter, req *http.Request) {
	if !s.closureCheck(w, req) {
		return
	}
	couponCode := req.PathValue("id")
	ctx := req.Context()
	dbCoupon, err := s.Queries.CouponEnabledByCode(ctx, couponCode)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		http.Error(w, "Invalid coupon ID", http.StatusNotFound)
		return
	case err != nil:
		slog.Error("error looking up coupons", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	coupon := dbCouponToCoupon(dbCoupon)
	if err := json.NewEncoder(w).Encode(coupon); err != nil {
		slog.Error("error writing coupon lookup response", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func dbCouponToCoupon(dbCoupon shop.Coupon) Coupon {
	requirements := make([]json.RawMessage, 0)
	if dbCoupon.MinPurchaseQuantity.Valid {
		requirements = append(requirements, json.RawMessage(fmt.Sprintf(`{"type":"purchase_count","amount":%d}`, dbCoupon.MinPurchaseQuantity.Int64)))
	}
	return Coupon{
		Requirements: requirements,
		CouponCode:   dbCoupon.CouponCode,
		Discount:     json.RawMessage(fmt.Sprintf(`{"type":"percentage","amount":%d}`, dbCoupon.DiscountPercentage)),
	}
}

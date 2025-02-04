package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/chanbakjsd/CCDSQuickShop/backend/shop"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/webhook"
)

type CheckoutRequest struct {
	Name         string     `json:"name"`
	MatricNumber string     `json:"matricNumber"`
	Email        string     `json:"email"`
	Items        []CartItem `json:"items"`
	Coupon       *string    `json:"coupon"`
}

type CheckoutResponse struct {
	CheckoutURL string `json:"checkoutURL"`
}

type CartItem struct {
	ID      string            `json:"id"`
	Variant []CartItemVariant `json:"variant"`
	Amount  int               `json:"amount"`
}

type CartItemVariant struct {
	Type   string `json:"type"`
	Option string `json:"option"`
}

var (
	matricRegex   = regexp.MustCompile(`^[UG]\d{7}[A-Z]$`)
	ntuEmailRegex = regexp.MustCompile(`^[A-Za-z\d]+@(e\.)?ntu\.edu\.sg$`)
)

func (s *Server) CheckoutComplete(w http.ResponseWriter, req *http.Request) {
	sessionID := req.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "Session ID not provided", http.StatusBadRequest)
		return
	}
	ctx := req.Context()
	orderID, err := s.checkAndFulfill(ctx, sessionID)
	if err != nil {
		slog.Error("failed to check and fulfill checkout", "err", err)
		http.Error(w, "Failed to complete checkout. We will still process your order if your payment was successful.", http.StatusBadRequest)
		return
	}
	http.Redirect(w, req, s.Config.FrontendURL+"/orders/"+orderID, http.StatusTemporaryRedirect)
}

func (s *Server) StripeWebhook(w http.ResponseWriter, req *http.Request) {
	const MaxBodyBytes = int64(65536)
	req.Body = http.MaxBytesReader(w, req.Body, MaxBodyBytes)
	payload, err := io.ReadAll(req.Body)
	if err != nil {
		slog.Error("error reading request body for Stripe webhook", "err", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	event, err := webhook.ConstructEvent(payload, req.Header.Get("Stripe-Signature"), s.Config.StripeWebhookSecret)
	if err != nil {
		slog.Error("error verifying webhook signature", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	switch event.Type {
	case "checkout.session.completed":
		fallthrough
	case "checkout.session.expired":
		sessionID, ok := event.Data.Object["id"].(string)
		if !ok {
			slog.Error("error reading ID of checkout session", "obj", event.Data.Object)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if _, err := s.checkAndFulfill(req.Context(), sessionID); err != nil {
			slog.Error("failed to check and fulfill checkout", "sessionID", sessionID, "err", err)
			http.Error(w, "failed to validate Stripe checkout", http.StatusBadRequest)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) checkAndFulfill(ctx context.Context, sessionID string) (string, error) {
	// We would ideally use sql.NullString but sqlc does not play well with our query so we just use a non-sensical value instead.
	couponCode := "INVALID_STRIPE_COUPON_CODE"
	if s.Stripe == nil {
		slog.Warn("skipping Stripe validation as Stripe is not configured", "session_id", sessionID)
	} else {
		session, err := s.Stripe.CheckoutSessions.Get(sessionID, &stripe.CheckoutSessionParams{
			Expand: []*string{stripe.String("total_details.breakdown")},
		})
		if err != nil {
			return "", fmt.Errorf("failed to fetch checkout session: %w", err)
		}
		if session.Status == "expired" {
			slog.Debug("expiring checkout session", "session_id", sessionID)
			orderID, err := s.Queries.ExpireCheckout(ctx, sql.NullString{
				String: sessionID,
				Valid:  true,
			})
			if err != nil {
				return "", fmt.Errorf("error expiring checkout: %w", err)
			}
			return orderID, nil
		}
		if session.PaymentStatus == "unpaid" {
			return "", fmt.Errorf("payment status of checkout session is unpaid, expiry time %d", session.ExpiresAt)
		}
		if len(session.TotalDetails.Breakdown.Discounts) > 0 {
			couponCode = session.TotalDetails.Breakdown.Discounts[0].Discount.Coupon.ID
		}
	}
	orderID, err := s.Queries.CompleteCheckout(ctx, shop.CompleteCheckoutParams{
		PaymentReference: sql.NullString{
			String: sessionID,
			Valid:  true,
		},
		CouponStripeID: couponCode,
		PaymentTime: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
	})
	if err != nil {
		return "", fmt.Errorf("error updating payment time: %w", err)
	}
	return orderID, nil
}

func (s *Server) Checkout(w http.ResponseWriter, req *http.Request) {
	if !s.closureCheck(w, req) {
		return
	}
	ctx := req.Context()
	var checkoutReq CheckoutRequest
	if err := json.NewDecoder(req.Body).Decode(&checkoutReq); err != nil {
		slog.Error("error parsing request", "err", err)
		http.Error(w, "Invalid Body", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(checkoutReq.Name) == "" {
		http.Error(w, "Invalid Name", http.StatusBadRequest)
		return
	}
	if !matricRegex.MatchString(checkoutReq.MatricNumber) {
		http.Error(w, "Invalid Matric Number", http.StatusBadRequest)
		return
	}
	if !ntuEmailRegex.MatchString(checkoutReq.Email) {
		http.Error(w, "Invalid Email", http.StatusBadRequest)
		return
	}
	if len(checkoutReq.Items) == 0 {
		http.Error(w, "At least one item is required", http.StatusBadRequest)
		return
	}
	dbProducts, err := s.Queries.ListProducts(ctx, false)
	if err != nil {
		slog.Error("error fetching products", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	products, err := dbProductsToProducts(dbProducts, false)
	if err != nil {
		slog.Error("error parsing products", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	// Validate the order.
	var couponID *int64
	var couponStripeID *string
	if checkoutReq.Coupon != nil {
		coupon, err := s.Queries.CouponEnabledByCode(ctx, *checkoutReq.Coupon)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			http.Error(w, "Invalid coupon code", http.StatusBadRequest)
			return
		case err != nil:
			slog.Error("error fetching coupon codes", "err", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		couponID = &coupon.CouponID
		couponStripeID = &coupon.StripeID
	}
	items, err := constructOrder(checkoutReq, products)
	if err != nil {
		slog.Error("error constructing order", "err", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var nullCouponID sql.NullInt64
	if checkoutReq.Coupon != nil {
		nullCouponID = sql.NullInt64{
			Int64: *couponID,
			Valid: true,
		}
	}
	// Write to database.
	for i := 0; i < 5; i++ {
		orderID := randomOrderID()
		order := shop.CreateOrderParams{
			OrderID:      orderID,
			Name:         checkoutReq.Name,
			MatricNumber: checkoutReq.MatricNumber,
			Email:        checkoutReq.Email,
			CouponID:     nullCouponID,
		}
		tx, err := s.DB.Begin()
		if err != nil {
			slog.Error("error creating transaction for order", "err", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		queries := s.Queries.WithTx(tx)
		if err := queries.CreateOrder(ctx, order); err != nil {
			// Try a different order ID.
			slog.Error("error creating order", "err", err)
			continue
		}
		for _, item := range items {
			if err := queries.CreateOrderItem(ctx, shop.CreateOrderItemParams{
				OrderID:     orderID,
				ProductID:   item.ProductID,
				ProductName: item.ProductName,
				UnitPrice:   item.UnitPrice,
				Amount:      item.Amount,
				ImageUrl:    item.ImageUrl,
				Variant:     item.Variant,
			}); err != nil {
				slog.Error("error creating order item", "err", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		}
		if err := tx.Commit(); err != nil {
			slog.Error("error commiting order", "err", err)
			continue
		}
		var redirectURL string
		var paymentRef string
		if s.Stripe == nil {
			paymentRef = "nonstripe_mock_" + randomOrderID()
			slog.Warn("skipping checkout session creation as Stripe is not configured")
			redirectURL = s.Config.FrontendURL + "/api/v0/checkout/complete?session_id=" + paymentRef
		} else {
			checkoutSession, err := s.createStripeCheckoutSession(orderID, checkoutReq.Email, items, couponStripeID)
			if err != nil {
				slog.Error("error creating checkout session", "err", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			redirectURL = checkoutSession.URL
			paymentRef = checkoutSession.ID
		}
		if err := s.Queries.AssociateOrder(ctx, shop.AssociateOrderParams{
			PaymentReference: sql.NullString{
				String: paymentRef,
				Valid:  true,
			},
			OrderID: orderID,
		}); err != nil {
			slog.Error("error associating order", "err", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if err := json.NewEncoder(w).Encode(CheckoutResponse{
			CheckoutURL: redirectURL,
		}); err != nil {
			slog.Error("error writing checkout response", "err", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	slog.Error("failing checkout due to too many failures")
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}

func constructOrder(req CheckoutRequest, products []Product) ([]shop.OrderItem, error) {
	orderItems := make([]shop.OrderItem, 0, len(req.Items))
	for _, v := range req.Items {
		var product *Product
		for _, p := range products {
			if v.ID == p.ID {
				product = &p
				break
			}
		}
		if product == nil {
			return nil, fmt.Errorf("Invalid product ID %q", v.ID)
		}
		if len(product.Variants) != len(v.Variant) {
			return nil, fmt.Errorf("Variants for product ID %q is invalid", product.ID)
		}
		if v.Amount <= 0 || v.Amount > 100 {
			return nil, fmt.Errorf("Amount must be between 1-100 (was %d for product ID %q)", v.Amount, v.ID)
		}
		price := product.BasePrice
		// Check variants.
		variantText := make([]string, 0, len(product.Variants))
		for _, variant := range product.Variants {
			var chosen *string
			var validOptions []ProductVariantOptions
			for _, w := range v.Variant {
				if variant.Type == w.Type {
					chosen = &w.Option
					validOptions = variant.Options
					break
				}
			}
			if chosen == nil {
				return nil, fmt.Errorf("Variant %q was not chosen for product ID %q", variant.Type, v.ID)
			}
			var additionalPrice *int
			for _, w := range validOptions {
				if w.Text == *chosen {
					additionalPrice = &w.AdditionalPrice
					break
				}
			}
			if additionalPrice == nil {
				return nil, fmt.Errorf("Variant %q has option %q for product ID %q which is invalid", variant.Type, *chosen, v.ID)
			}
			price += *additionalPrice
			variantText = append(variantText, *chosen)
		}
		imageURL := product.DefaultImageURL
		bestMatch := 0
		for _, urlCandidate := range product.ImageURLs {
			mismatch := false
			match := 0
			for i := 0; i < len(variantText); i++ {
				switch {
				case urlCandidate.SelectedOptions[i] == nil:
					continue
				case *urlCandidate.SelectedOptions[i] == variantText[i]:
					match++
					continue
				default:
					mismatch = true
				}
				if mismatch {
					break
				}
			}
			if mismatch || match <= bestMatch {
				continue
			}
			bestMatch = match
			imageURL = urlCandidate.URL
		}
		orderItems = append(orderItems, shop.OrderItem{
			ProductID:   product.ID,
			ProductName: product.Name,
			UnitPrice:   int64(price),
			Amount:      int64(v.Amount),
			ImageUrl:    imageURL,
			Variant:     strings.Join(variantText, ", "),
		})
	}
	return orderItems, nil
}

func (s *Server) createStripeCheckoutSession(orderID string, email string, items []shop.OrderItem, couponID *string) (*stripe.CheckoutSession, error) {
	checkoutLineItems := make([]*stripe.CheckoutSessionLineItemParams, 0, len(items))
	for _, v := range items {
		var imageData []*string
		if v.ImageUrl != "" {
			imageData = append(imageData, stripe.String(v.ImageUrl))
		}
		checkoutLineItems = append(checkoutLineItems, &stripe.CheckoutSessionLineItemParams{
			PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
				Currency:   stripe.String("sgd"),
				UnitAmount: stripe.Int64(v.UnitPrice),
				ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
					Name:        &v.ProductName,
					Images:      imageData,
					Description: stripe.String(v.Variant),
				},
			},
			Quantity: &v.Amount,
		})
	}
	var discount []*stripe.CheckoutSessionDiscountParams
	if couponID != nil {
		discount = []*stripe.CheckoutSessionDiscountParams{
			{
				Coupon: couponID,
			},
		}
	}
	checkoutParams := &stripe.CheckoutSessionParams{
		Mode:              stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL:        stripe.String(s.Config.FrontendURL + "/api/v0/checkout/complete?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:         stripe.String(s.Config.FrontendURL),
		LineItems:         checkoutLineItems,
		Discounts:         discount,
		ClientReferenceID: stripe.String(orderID),
		CustomerEmail:     stripe.String(email),
		PaymentIntentData: &stripe.CheckoutSessionPaymentIntentDataParams{
			Description: stripe.String("Your order ID is " + orderID + "."),
		},
	}
	return s.Stripe.CheckoutSessions.New(checkoutParams)
}

// Alphabets that are easily disambiguated.
var alphabet = "CDEFHJKMNPRTVWXY"

func randomOrderID() string {
	a := alphabet[rand.IntN(len(alphabet))]
	b := alphabet[rand.IntN(len(alphabet))]
	n := rand.IntN(10000)
	return fmt.Sprintf("%c%c%04d", a, b, n)
}

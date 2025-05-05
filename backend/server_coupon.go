package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/chanbakjsd/CCDSQuickShop/backend/db"
	"github.com/stripe/stripe-go/v81"
)

type CouponsResponse struct {
	Coupons []Coupon `json:"coupons"`
}

type Coupon struct {
	Requirements []json.RawMessage `json:"requirements"`
	CouponCode   string            `json:"couponCode"`
	Discount     json.RawMessage   `json:"discount"`

	// Admin fields.
	ID         *int64  `json:"id,omitempty"`
	StripeID   *string `json:"stripe_id,omitempty"`
	Enabled    *bool   `json:"enabled,omitempty"`
	Public     *bool   `json:"public,omitempty"`
	StripeDesc *string `json:"stripe_desc,omitempty"`
}

func (s *Server) Coupons(w http.ResponseWriter, req *http.Request) {
	includeDisabled := req.URL.Query().Get("include_disabled") != ""
	if !includeDisabled && !s.closureCheck(w, req) {
		return
	}
	if includeDisabled && !s.authCheck(w, req) {
		return
	}
	var dbCoupons []db.Coupon
	var err error
	if includeDisabled {
		dbCoupons, err = s.Queries.ListCoupons(req.Context())
	} else {
		dbCoupons, err = s.Queries.ListPublicCoupons(req.Context())
	}
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
		coupons = append(coupons, s.dbCouponToCoupon(coupon, includeDisabled))
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
	coupon := s.dbCouponToCoupon(dbCoupon, false)
	if err := json.NewEncoder(w).Encode(coupon); err != nil {
		slog.Error("error writing coupon lookup response", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

type couponRequirement struct {
	Type   string `json:"type"`
	Amount int    `json:"amount"`
	Value  string `json:"value"`
}

type couponDiscount struct {
	Type   string `json:"type"`
	Amount int    `json:"amount"`
}

func (s *Server) SaveCoupon(w http.ResponseWriter, req *http.Request) {
	if !s.authCheck(w, req) {
		return
	}
	ctx := req.Context()
	var coupon Coupon
	if err := json.NewDecoder(req.Body).Decode(&coupon); err != nil {
		slog.Error("error parsing request", "err", err)
		http.Error(w, "Invalid Body", http.StatusBadRequest)
		return
	}
	var minPurchaseQuantity sql.NullInt64
	var email sql.NullString
	for _, marshalledReq := range coupon.Requirements {
		var req couponRequirement
		if err := json.Unmarshal(marshalledReq, &req); err != nil {
			slog.Error("error parsing request: coupon requirement is invalid", "err", err)
			http.Error(w, "Invalid Body", http.StatusBadRequest)
			return
		}
		switch req.Type {
		case "purchase_count":
			if !minPurchaseQuantity.Valid || int64(req.Amount) > minPurchaseQuantity.Int64 {
				minPurchaseQuantity = sql.NullInt64{
					Valid: true,
					Int64: int64(req.Amount),
				}
			}
		case "email":
			if email.Valid {
				slog.Error("error parsing request: multiple emails provided")
				http.Error(w, "Invalid Body", http.StatusBadRequest)
				return
			}
			email = sql.NullString{
				Valid:  true,
				String: req.Value,
			}
		default:
			slog.Error("error parsing request: coupon requirement is invalid", "type", req.Type)
			http.Error(w, "Invalid Coupon Requirement", http.StatusBadRequest)
			return
		}
	}
	var discount couponDiscount
	var discountPercentage int
	if err := json.Unmarshal(coupon.Discount, &discount); err != nil {
		slog.Error("error parsing request: coupon discount is invalid", "err", err)
		http.Error(w, "Invalid Body", http.StatusBadRequest)
		return
	}
	switch discount.Type {
	case "percentage":
		discountPercentage = discount.Amount
	default:
		slog.Error("error parsing request: coupon discount is invalid", "type", discount.Type)
		http.Error(w, "Invalid Coupon Discount", http.StatusBadRequest)
		return
	}
	couponEnabled := coupon.Enabled != nil && *coupon.Enabled
	couponPublic := coupon.Public != nil && *coupon.Public
	var stripeID string
	if couponEnabled {
		if coupon.StripeID != nil {
			stripeID = *coupon.StripeID
		}
		if coupon.StripeDesc == nil {
			http.Error(w, "Invalid Body: Missing Stripe Desc", http.StatusBadRequest)
			return
		}
		couponID, err := s.upsertStripeCoupon(stripeID, *coupon.StripeDesc, discountPercentage)
		if err != nil {
			slog.Error("error upserting Stripe coupon", "err", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		coupon.StripeID = &couponID
		stripeID = couponID
	} else if coupon.StripeID != nil {
		stripeID = *coupon.StripeID
	}
	var sqlErr error
	switch coupon.ID {
	case nil:
		var newID int64
		newID, sqlErr = s.Queries.CreateCoupon(ctx, db.CreateCouponParams{
			StripeID:            stripeID,
			CouponCode:          coupon.CouponCode,
			MinPurchaseQuantity: minPurchaseQuantity,
			EmailMatch:          email,
			DiscountPercentage:  int64(discountPercentage),
			Enabled:             couponEnabled,
			Public:              couponPublic,
		})
		coupon.ID = &newID
	default:
		sqlErr = s.Queries.UpdateCoupon(ctx, db.UpdateCouponParams{
			CouponID:            *coupon.ID,
			StripeID:            stripeID,
			CouponCode:          coupon.CouponCode,
			MinPurchaseQuantity: minPurchaseQuantity,
			EmailMatch:          email,
			DiscountPercentage:  int64(discountPercentage),
			Enabled:             couponEnabled,
			Public:              couponPublic,
		})
	}
	switch {
	case errors.Is(sqlErr, sql.ErrNoRows):
		http.Error(w, "Invalid Coupon ID", http.StatusBadRequest)
		return
	case sqlErr != nil:
		slog.Error("error updating coupon", "err", sqlErr)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(coupon); err != nil {
		slog.Error("error writing update coupon response", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

var descCache = make(map[int64]*string)

func (s *Server) dbCouponToCoupon(dbCoupon db.Coupon, includeSensitiveFields bool) Coupon {
	requirements := make([]json.RawMessage, 0)
	if dbCoupon.MinPurchaseQuantity.Valid {
		requirements = append(requirements, json.RawMessage(fmt.Sprintf(`{"type":"purchase_count","amount":%d}`, dbCoupon.MinPurchaseQuantity.Int64)))
	}
	if dbCoupon.EmailMatch.Valid {
		requirements = append(requirements, json.RawMessage(fmt.Sprintf(`{"type":"email","value":"%s"}`, dbCoupon.EmailMatch.String)))
	}
	coupon := Coupon{
		Requirements: requirements,
		CouponCode:   dbCoupon.CouponCode,
		Discount:     json.RawMessage(fmt.Sprintf(`{"type":"percentage","amount":%d}`, dbCoupon.DiscountPercentage)),
	}
	if includeSensitiveFields {
		coupon.ID = &dbCoupon.CouponID
		coupon.StripeID = &dbCoupon.StripeID
		coupon.Enabled = &dbCoupon.Enabled
		coupon.Public = &dbCoupon.Public
		if s.Stripe == nil || dbCoupon.StripeID == "" {
			return coupon
		}
		if _, ok := descCache[dbCoupon.CouponID]; !ok {
			stripeCoupon, err := s.Stripe.Coupons.Get(dbCoupon.StripeID, nil)
			if err != nil {
				slog.Warn("error fetching Stripe coupon", "stripe_id", dbCoupon.StripeID, "err", err)
				return coupon
			}
			descCache[dbCoupon.CouponID] = &stripeCoupon.Name
		}
		coupon.StripeDesc = descCache[dbCoupon.CouponID]
	}
	return coupon
}

// upsertStripeCoupon attempts to update the given Stripe coupon but creates a
// new one if it is not possible to do so.
// It returns the Stripe coupon ID with the given attribute.
func (s *Server) upsertStripeCoupon(couponID string, name string, discountPercentage int) (string, error) {
	if s.Stripe == nil {
		slog.Warn("not updating actual Stripe coupon %q: Stripe not configured", "coupon_id", couponID)
		return "mock_stripe_coupon_" + randomOrderID(), nil
	}
	couponParam := &stripe.CouponParams{
		Name:       &name,
		PercentOff: stripe.Float64(float64(discountPercentage)),
	}
	tryUpdate := func() (ok bool) {
		if couponID == "" {
			return false
		}
		coupon, err := s.Stripe.Coupons.Get(couponID, nil)
		if err != nil {
			slog.Warn("error fetching Stripe coupon, falling back to creating new coupon", "old_coupon_id", couponID, "err", err)
			return false
		}
		if int(coupon.PercentOff) != discountPercentage {
			slog.Debug(
				"falling back to creating new coupon, discount percentage changed",
				"old_coupon_id", couponID,
				"old_percentage", coupon.PercentOff,
				"new_percentage", discountPercentage,
			)
			return false
		}
		if coupon.Name == name {
			slog.Debug("skipping update, name is equal", "coupon_id", couponID)
			return true
		}
		_, err = s.Stripe.Coupons.Update(couponID, &stripe.CouponParams{
			Name: &name,
		})
		if err != nil {
			slog.Warn("error updating coupon, falling back to creating new coupon", "old_coupon_id", couponID, "err", err)
			return false
		}
		return true
	}
	if tryUpdate() {
		return couponID, nil
	}
	coupon, err := s.Stripe.Coupons.New(couponParam)
	if err != nil {
		return "", fmt.Errorf("error creating Stripe coupon (%q): %w", name, err)
	}
	return coupon.ID, nil
}

package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/chanbakjsd/CCDSQuickShop/backend/db"
)

type ProductsResponse struct {
	Products []Product `json:"products"`
}

type Product struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	BasePrice       int               `json:"basePrice"`
	Variants        []ProductVariant  `json:"variants"`
	DefaultImageURL string            `json:"defaultImageURL"`
	ImageURLs       []ProductImageURL `json:"imageURLs"`
	Enabled         *bool             `json:"enabled,omitempty"`
	SalePeriod      int               `json:"salePeriod"`
}

type ProductVariant struct {
	Type     string                  `json:"type"`
	ChartURL *string                 `json:"chart_url,omitempty"`
	Options  []ProductVariantOptions `json:"options"`
}

type ProductVariantOptions struct {
	Text            string `json:"text"`
	AdditionalPrice int    `json:"additionalPrice"`
}

type ProductImageURL struct {
	SelectedOptions []*string `json:"selectedOptions"`
	URL             string    `json:"url"`
}

func (s *Server) Products(w http.ResponseWriter, req *http.Request) {
	includeDisabled := req.URL.Query().Get("include_disabled") != ""
	requestOK := false
	if includeDisabled {
		// Admin request.
		requestOK = s.authCheck(w, req)
	} else {
		// Normal user request: Check whether store is closed.
		requestOK = s.closureCheck(w, req)
	}
	if !requestOK {
		return
	}
	salePeriod, ok := s.resolveSalePeriod(w, req, req.PathValue("sale_id"))
	if !ok {
		return
	}
	dbProducts, err := s.Queries.ListProducts(req.Context(), db.ListProductsParams{
		IncludeDisabled: includeDisabled,
		SalePeriod:      salePeriod,
	})
	switch {
	case errors.Is(err, context.Canceled):
		// Cancelled by user. Do nothing.
		return
	case err != nil:
		slog.Error("error fetching products", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	products, err := dbProductsToProducts(dbProducts, includeDisabled)
	if err != nil {
		slog.Error("error parsing products", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(ProductsResponse{
		Products: products,
	}); err != nil {
		slog.Error("error writing products response", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (s *Server) SaveProduct(w http.ResponseWriter, req *http.Request) {
	if !s.authCheck(w, req) {
		return
	}
	ctx := req.Context()
	var product Product
	if err := json.NewDecoder(req.Body).Decode(&product); err != nil {
		slog.Error("error parsing request", "err", err)
		http.Error(w, "Invalid Body", http.StatusBadRequest)
		return
	}
	if product.Enabled == nil {
		slog.Error("expected product enabled to be provided")
		http.Error(w, "Invalid Body", http.StatusBadRequest)
		return
	}
	salePeriod, ok := s.resolveSalePeriod(w, req, req.PathValue("sale_id"))
	if !ok {
		return
	}
	productVariants, err := json.Marshal(product.Variants)
	if err != nil {
		slog.Error("error marshalling product variant", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
	imageURLs, err := json.Marshal(product.ImageURLs)
	if err != nil {
		slog.Error("error marshalling image URLs", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
	var sqlErr error
	switch product.ID {
	case "":
		// Create new product.
		var newID int64
		newID, sqlErr = s.Queries.CreateProduct(ctx, db.CreateProductParams{
			Name:             product.Name,
			BasePrice:        int64(product.BasePrice),
			DefaultImageUrl:  product.DefaultImageURL,
			Variants:         string(productVariants),
			VariantImageUrls: string(imageURLs),
			Enabled:          *product.Enabled,
			SalePeriod:       salePeriod,
		})
		product.ID = strconv.Itoa(int(newID))
	default:
		// Update existing ID.
		id, err := strconv.Atoi(product.ID)
		if err != nil {
			http.Error(w, "Invalid product ID", http.StatusBadRequest)
			return
		}
		sqlErr = s.Queries.UpdateProduct(ctx, db.UpdateProductParams{
			ProductID:        int64(id),
			Name:             product.Name,
			BasePrice:        int64(product.BasePrice),
			DefaultImageUrl:  product.DefaultImageURL,
			Variants:         string(productVariants),
			VariantImageUrls: string(imageURLs),
			Enabled:          *product.Enabled,
			SalePeriod:       salePeriod,
		})
	}
	switch {
	case errors.Is(sqlErr, sql.ErrNoRows):
		http.Error(w, "Invalid Product ID", http.StatusBadRequest)
		return
	case sqlErr != nil:
		slog.Error("error updating product", "err", sqlErr)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(product); err != nil {
		slog.Error("error writing update product response", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func dbProductsToProducts(dbProducts []db.Product, includeDisabled bool) ([]Product, error) {
	products := make([]Product, 0, len(dbProducts))
	for _, p := range dbProducts {
		var variants []ProductVariant
		var imageURLs []ProductImageURL
		if err := json.Unmarshal([]byte(p.Variants), &variants); err != nil {
			return nil, fmt.Errorf("error unmarshalling products: %w", err)
		}
		if err := json.Unmarshal([]byte(p.VariantImageUrls), &imageURLs); err != nil {
			return nil, fmt.Errorf("error unmarshalling image URLs: %w", err)
		}
		product := Product{
			ID:              strconv.Itoa(int(p.ProductID)),
			Name:            p.Name,
			BasePrice:       int(p.BasePrice),
			Variants:        variants,
			DefaultImageURL: p.DefaultImageUrl,
			ImageURLs:       imageURLs,
			SalePeriod:      int(p.SalePeriod),
		}
		if includeDisabled {
			product.Enabled = &p.Enabled
		}
		products = append(products, product)
	}
	return products, nil
}

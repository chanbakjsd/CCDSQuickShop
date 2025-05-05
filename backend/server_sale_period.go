package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/chanbakjsd/CCDSQuickShop/backend/db"
)

type SalePeriodsResponse struct {
	Periods []SalePeriod `json:"periods"`
}

type SalePeriod struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	StartTime time.Time `json:"start_time"`
}

func (s *Server) SalePeriods(w http.ResponseWriter, req *http.Request) {
	if !s.authCheck(w, req) {
		return
	}
	dbSalePeriods, err := s.Queries.ListSalePeriods(req.Context())
	if err != nil {
		slog.Error("error fetching sale periods", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	salePeriods := make([]SalePeriod, 0, len(dbSalePeriods))
	for _, v := range dbSalePeriods {
		salePeriods = append(salePeriods, SalePeriod{
			ID:        strconv.Itoa(int(v.ID)),
			Name:      v.AdminName,
			StartTime: v.StartTime,
		})
	}
	if err := json.NewEncoder(w).Encode(SalePeriodsResponse{
		Periods: salePeriods,
	}); err != nil {
		slog.Error("error writing sale periods response", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (s *Server) SaveSalePeriod(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	if !s.authCheck(w, req) {
		return
	}
	var salePeriod SalePeriod
	if err := json.NewDecoder(req.Body).Decode(&salePeriod); err != nil {
		slog.Error("error parsing request", "err", err)
		http.Error(w, "Invalid Body", http.StatusBadRequest)
		return
	}
	var sqlErr error
	switch salePeriod.ID {
	case "":
		// Create new product.
		var newID int64
		newID, sqlErr = s.Queries.CreateSalePeriod(ctx, db.CreateSalePeriodParams{
			AdminName: salePeriod.Name,
			StartTime: salePeriod.StartTime,
		})
		salePeriod.ID = strconv.Itoa(int(newID))
	default:
		// Update existing ID.
		id, err := strconv.Atoi(salePeriod.ID)
		if err != nil {
			http.Error(w, "Invalid product ID", http.StatusBadRequest)
			return
		}
		sqlErr = s.Queries.UpdateSalePeriod(ctx, db.UpdateSalePeriodParams{
			ID:        int64(id),
			AdminName: salePeriod.Name,
			StartTime: salePeriod.StartTime,
		})
	}
	switch {
	case errors.Is(sqlErr, sql.ErrNoRows):
		http.Error(w, "Invalid Product ID", http.StatusBadRequest)
		return
	case sqlErr != nil:
		slog.Error("error updating sale period", "err", sqlErr)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(salePeriod); err != nil {
		slog.Error("error writing update sale period response", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

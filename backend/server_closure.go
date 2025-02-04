package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/chanbakjsd/CCDSQuickShop/backend/shop"
)

type StoreClosureError struct {
	Type            string `json:"type"`
	EndTime         *int   `json:"end_time"`
	Message         string `json:"message"`
	AllowOrderCheck bool   `json:"show_order_check"`
}

func (s *Server) closureCheck(w http.ResponseWriter, req *http.Request) (ok bool) {
	closure, err := s.Queries.StoreClosureCurrent(req.Context(), time.Now().UTC())
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return true
	case err != nil:
		slog.Error("failed to check store closure", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return false
	}
	var endTime *int
	actualEndTime := time.Until(closure.EndTime)
	if actualEndTime < 24*time.Hour {
		secondsLeft := int(actualEndTime / time.Second)
		endTime = &secondsLeft
	}
	if err := json.NewEncoder(w).Encode(StoreClosureError{
		Type:            "store_closure",
		EndTime:         endTime,
		Message:         closure.UserMessage,
		AllowOrderCheck: closure.AllowOrderCheck,
	}); err != nil {
		slog.Error("error writing store closure response", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	return false
}

type StoreClosureResponse struct {
	Closures []StoreClosure `json:"closures"`
}

type StoreClosure struct {
	ID              string    `json:"id"`
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	Message         string    `json:"message"`
	AllowOrderCheck bool      `json:"show_order_check"`
}

func (s *Server) StoreClosures(w http.ResponseWriter, req *http.Request) {
	if !s.authCheck(w, req) {
		return
	}
	dbStoreClosures, err := s.Queries.ListStoreClosures(req.Context())
	if err != nil {
		slog.Error("error fetching store closures", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	closures := make([]StoreClosure, 0, len(dbStoreClosures))
	for _, v := range dbStoreClosures {
		closures = append(closures, StoreClosure{
			ID:              strconv.Itoa(int(v.ID)),
			StartTime:       v.StartTime,
			EndTime:         v.EndTime,
			Message:         v.UserMessage,
			AllowOrderCheck: v.AllowOrderCheck,
		})
	}
	if err := json.NewEncoder(w).Encode(StoreClosureResponse{
		Closures: closures,
	}); err != nil {
		slog.Error("error writing store closures response", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (s *Server) SaveStoreClosure(w http.ResponseWriter, req *http.Request) {
	if !s.authCheck(w, req) {
		return
	}
	ctx := req.Context()
	var closure StoreClosure
	if err := json.NewDecoder(req.Body).Decode(&closure); err != nil {
		slog.Error("error parsing request", "err", err)
		http.Error(w, "Invalid Body", http.StatusBadRequest)
		return
	}
	var sqlErr error
	switch closure.ID {
	case "":
		// Create new store closure.
		var newID int64
		newID, sqlErr = s.Queries.CreateStoreClosure(ctx, shop.CreateStoreClosureParams{
			StartTime:       closure.StartTime,
			EndTime:         closure.EndTime,
			UserMessage:     closure.Message,
			AllowOrderCheck: closure.AllowOrderCheck,
		})
		closure.ID = strconv.Itoa(int(newID))
	default:
		// Update existing ID.
		id, err := strconv.Atoi(closure.ID)
		if err != nil {
			http.Error(w, "Invalid closure ID", http.StatusBadRequest)
			return
		}
		sqlErr = s.Queries.UpdateStoreClosure(ctx, shop.UpdateStoreClosureParams{
			ID:              int64(id),
			StartTime:       closure.StartTime,
			EndTime:         closure.EndTime,
			UserMessage:     closure.Message,
			AllowOrderCheck: closure.AllowOrderCheck,
		})
	}
	switch {
	case errors.Is(sqlErr, sql.ErrNoRows):
		http.Error(w, "Invalid closure ID", http.StatusBadRequest)
		return
	case sqlErr != nil:
		slog.Error("error updating store closure", "err", sqlErr)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(closure); err != nil {
		slog.Error("error writing update closure response", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (s *Server) DeleteStoreClosure(w http.ResponseWriter, req *http.Request) {
	if !s.authCheck(w, req) {
		return
	}
	closureID := req.PathValue("id")
	id, err := strconv.Atoi(closureID)
	if err != nil {
		http.Error(w, "Invalid closure ID", http.StatusBadRequest)
		return
	}
	if err := s.Queries.DeleteStoreClosure(req.Context(), int64(id)); err != nil {
		http.Error(w, "Invalid closure ID", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

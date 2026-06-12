package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"mi-tech/internal/service"
)

// SyncHandler is a thin HTTP adapter for Shopify sync operations.
type SyncHandler struct {
	syncService *service.SyncService
}

func NewSyncHandler(syncService *service.SyncService) *SyncHandler {
	return &SyncHandler{
		syncService: syncService,
	}
}

// SyncRequest represents the body for sync operation.
type SyncRequest struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

// SyncOrders handles POST /api/shopify/sync.
// @Summary Sync Shopify orders
// @Description Fetch and update orders from Shopify for a specific date range.
// @Tags sync
// @Security Bearer
// @Accept json
// @Produce json
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Param body body SyncRequest false "Date range in body"
// @Success 200 {object} map[string]interface{}
// @Router /shopify/sync [post]
func (h *SyncHandler) SyncOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var startTime, endTime *time.Time

	// Try to get dates from query params
	if s := r.URL.Query().Get("start_date"); s != "" {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			http.Error(w, "Invalid start_date format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		startTime = &t
	}
	if e := r.URL.Query().Get("end_date"); e != "" {
		t, err := time.Parse("2006-01-02", e)
		if err != nil {
			http.Error(w, "Invalid end_date format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		// Set end of day for end_date using helper
		et := endOfDay(t)
		endTime = &et
	}

	// Also check body for JSON (preferred for POST)
	if r.Body != nil && r.ContentLength > 0 {
		var body SyncRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}
		if body.StartDate != "" {
			t, err := time.Parse("2006-01-02", body.StartDate)
			if err != nil {
				http.Error(w, "Invalid start_date format in body. Use YYYY-MM-DD", http.StatusBadRequest)
				return
			}
			startTime = &t
		}
		if body.EndDate != "" {
			t, err := time.Parse("2006-01-02", body.EndDate)
			if err != nil {
				http.Error(w, "Invalid end_date format in body. Use YYYY-MM-DD", http.StatusBadRequest)
				return
			}
			et := endOfDay(t)
			endTime = &et
		}
	}

	if startTime != nil && endTime != nil && startTime.After(*endTime) {
		http.Error(w, "start_date cannot be after end_date", http.StatusBadRequest)
		return
	}

	count, err := h.syncService.Sync(startTime, endTime)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Sync completed successfully",
		"count":   count,
	})
}

// ResetOrders handles POST /api/shopify/reset.
// @Summary Reset and sync orders
// @Description Wipe local orders and re-sync everything from Shopify
// @Tags sync
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /shopify/reset [post]
func (h *SyncHandler) ResetOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 1. Reset AND Sync: Wipe everything and then re-fetch from Shopify
	count, err := h.syncService.ResetAndSync()
	if err != nil {
		http.Error(w, "Database reset and sync failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Database wiped and re-sync triggered successfully.",
		"count":   count,
	})
}

// endOfDay returns the latest nanosecond of the given date.
func endOfDay(t time.Time) time.Time {
	return t.AddDate(0, 0, 1).Add(-time.Nanosecond)
}

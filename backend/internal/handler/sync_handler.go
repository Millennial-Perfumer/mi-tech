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

// NewSyncHandler creates a new SyncHandler.
func NewSyncHandler(syncService *service.SyncService) *SyncHandler {
	return &SyncHandler{syncService: syncService}
}

// SyncOrders handles POST /api/shopify/sync.
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
		// Set end of day for end_date: increment by 1 day and subtract 1 nanosecond
		et := t.AddDate(0, 0, 1).Add(-time.Nanosecond)
		endTime = &et
	}

	// Also check body for JSON (preferred for POST)
	var body struct {
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
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
			et := t.AddDate(0, 0, 1).Add(-time.Nanosecond)
			endTime = &et
		}
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
func (h *SyncHandler) ResetOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	count, err := h.syncService.ResetAndSync()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Reset and Sync completed successfully",
		"count":   count,
	})
}

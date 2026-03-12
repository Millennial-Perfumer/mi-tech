package handler

import (
	"encoding/json"
	"net/http"

	"shopify-gst-app/internal/repository"
)

// SettingsHandler handles settings API requests.
type SettingsHandler struct {
	settingsRepo *repository.SettingsRepository
}

// NewSettingsHandler creates a new SettingsHandler.
func NewSettingsHandler(settingsRepo *repository.SettingsRepository) *SettingsHandler {
	return &SettingsHandler{settingsRepo: settingsRepo}
}

// GetDateRange returns the persisted date range.
func (h *SettingsHandler) GetDateRange(w http.ResponseWriter, r *http.Request) {
	startDate, endDate, err := h.settingsRepo.GetDateRange()
	if err != nil {
		http.Error(w, "Failed to get date range", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"start_date": startDate,
		"end_date":   endDate,
	})
}

// SetDateRange persists the date range.
func (h *SettingsHandler) SetDateRange(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.settingsRepo.SetDateRange(body.StartDate, body.EndDate); err != nil {
		http.Error(w, "Failed to save date range", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Date range saved",
	})
}

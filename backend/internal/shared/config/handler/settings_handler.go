package handler

import (
	"encoding/json"
	"net/http"

	"mi-tech/internal/shared/config/entity"
	"mi-tech/internal/shared/config/repository"
)

// SettingsHandler handles settings API requests.
type SettingsHandler struct {
	settingsRepo repository.SettingsRepository
}

// NewSettingsHandler creates a new SettingsHandler.
func NewSettingsHandler(settingsRepo repository.SettingsRepository) *SettingsHandler {
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

// GetAllSettings returns all key-value pairs from app_settings.
func (h *SettingsHandler) GetAllSettings(w http.ResponseWriter, r *http.Request) {
	var settings []entity.AppSetting
	if err := h.settingsRepo.GetAll(&settings); err != nil {
		http.Error(w, "Failed to get settings", http.StatusInternalServerError)
		return
	}

	res := make(map[string]string)
	for _, s := range settings {
		res[s.Key] = s.Value
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"settings": res,
	})
}

// UpdateSetting updates a single setting value.
func (h *SettingsHandler) UpdateSetting(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if body.Key == "" {
		http.Error(w, "Key is required", http.StatusBadRequest)
		return
	}

	if err := h.settingsRepo.Set(body.Key, body.Value); err != nil {
		http.Error(w, "Failed to update setting", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Setting updated",
	})
}

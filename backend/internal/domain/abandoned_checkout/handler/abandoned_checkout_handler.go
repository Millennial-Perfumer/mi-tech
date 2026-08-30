package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"mi-tech/internal/domain/abandoned_checkout/service"
	"mi-tech/internal/shared/config"
)

// AbandonedCheckoutHandler is the HTTP controller for abandoned checkout operations.
type AbandonedCheckoutHandler struct {
	acService service.AbandonedCheckoutService
}

// NewAbandonedCheckoutHandler creates a new AbandonedCheckoutHandler.
func NewAbandonedCheckoutHandler(acService service.AbandonedCheckoutService) *AbandonedCheckoutHandler {
	return &AbandonedCheckoutHandler{
		acService: acService,
	}
}

// GetAbandonedCheckouts handles GET and DELETE /api/abandoned-checkouts
func (h *AbandonedCheckoutHandler) GetAbandonedCheckouts(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodDelete {
		h.DeleteCheckout(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	search := r.URL.Query().Get("search")
	status := r.URL.Query().Get("status")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	checkouts, total, err := h.acService.ListCheckouts(r.Context(), config.StoreIDShopify, page, limit, search, status, startDate, endDate)
	if err != nil {
		http.Error(w, "Failed to retrieve checkouts: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"checkouts":   checkouts,
		"total_count": total,
		"page":        page,
		"limit":       limit,
	})
}

// RecoverCheckout handles POST /api/abandoned-checkouts/recover
func (h *AbandonedCheckoutHandler) RecoverCheckout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	var id int
	var err error

	if idStr == "" {
		// Try reading from request body
		var body struct {
			ID int `json:"id"`
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1048576)
		if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
			id = body.ID
		}
	} else {
		id, err = strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid checkout ID", http.StatusBadRequest)
			return
		}
	}

	if id == 0 {
		http.Error(w, "Missing or invalid checkout ID", http.StatusBadRequest)
		return
	}

	err = h.acService.TriggerManualRecovery(r.Context(), config.StoreIDShopify, id)
	if err != nil {
		http.Error(w, "Failed to send manual recovery: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Manual WhatsApp recovery triggered successfully",
	})
}

// GetAbandonedCheckoutAnalytics handles GET /api/abandoned-checkouts/analytics
func (h *AbandonedCheckoutHandler) GetAbandonedCheckoutAnalytics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	analytics, err := h.acService.GetAnalytics(r.Context(), config.StoreIDShopify, startDate, endDate)
	if err != nil {
		http.Error(w, "Failed to retrieve analytics: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"analytics": analytics,
	})
}

// DeleteCheckout handles DELETE /api/abandoned-checkouts
func (h *AbandonedCheckoutHandler) DeleteCheckout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid checkout ID", http.StatusBadRequest)
		return
	}

	err = h.acService.DeleteCheckout(r.Context(), config.StoreIDShopify, id)
	if err != nil {
		http.Error(w, "Failed to delete checkout: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Checkout deleted successfully",
	})
}

// UpdateCheckoutStatus handles PUT /api/abandoned-checkouts/status
func (h *AbandonedCheckoutHandler) UpdateCheckoutStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID             int    `json:"id"`
		RecoveryStatus string `json:"recovery_status"`
		Completed      bool   `json:"completed"`
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ID <= 0 {
		http.Error(w, "Missing or invalid checkout ID", http.StatusBadRequest)
		return
	}

	err := h.acService.UpdateCheckoutStatus(r.Context(), config.StoreIDShopify, req.ID, req.RecoveryStatus, req.Completed)
	if err != nil {
		http.Error(w, "Failed to update status: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Checkout status updated successfully",
	})
}

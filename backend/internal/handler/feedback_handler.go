package handler

import (
	"encoding/json"
	"fmt"
	"mi-tech/internal/entity"
	"mi-tech/internal/service"
	"net/http"
)

type FeedbackHandler struct {
	orderService *service.OrderService
}

func NewFeedbackHandler(orderService *service.OrderService) *FeedbackHandler {
	return &FeedbackHandler{
		orderService: orderService,
	}
}

// SubmitFeedback handles POST /api/feedback/submit
func (h *FeedbackHandler) SubmitFeedback(w http.ResponseWriter, r *http.Request) {
	// Enable CORS for public endpoint
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		OrderID int64  `json:"order_id"`
		Rating  int    `json:"rating"`
		Message string `json:"message"`
		Phone   string `json:"phone"` // Optional validation
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.OrderID == 0 || req.Rating < 1 || req.Rating > 5 {
		http.Error(w, "Missing required fields or invalid rating", http.StatusBadRequest)
		return
	}

	// 1. Save feedback
	feedback := entity.CustomerFeedback{
		OrderID:       req.OrderID,
		CustomerPhone: req.Phone,
		Rating:        req.Rating,
		Message:       req.Message,
	}

	if err := h.orderService.SaveCustomerFeedback(feedback); err != nil {
		http.Error(w, "Failed to save feedback", http.StatusInternalServerError)
		return
	}

	// 2. Update order feedback status to 'completed' (Status ID: 3)
	if err := h.orderService.UpdateFeedbackStatus(req.OrderID, 3); err != nil {
		// Log error but don't fail the request since feedback is saved
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Feedback received. Thank you!",
	})
}

// GetFeedback handles GET /api/feedback (Admin only)
func (h *FeedbackHandler) GetFeedback(w http.ResponseWriter, r *http.Request) {
	feedbacks, err := h.orderService.GetCustomerFeedback()
	if err != nil {
		http.Error(w, "Failed to fetch feedback", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"feedbacks": feedbacks,
	})
}

// ValidateFeedback handles GET /api/feedback/validate?o=order_id&p=phone
func (h *FeedbackHandler) ValidateFeedback(w http.ResponseWriter, r *http.Request) {
	// Enable CORS for public endpoint
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	orderIDStr := r.URL.Query().Get("o")
	phone := r.URL.Query().Get("p")

	if orderIDStr == "" || phone == "" {
		http.Error(w, "Missing order_id or phone", http.StatusBadRequest)
		return
	}

	var orderID int64
	_, err := fmt.Sscanf(orderIDStr, "%d", &orderID)
	if err != nil {
		http.Error(w, "Invalid order_id format", http.StatusBadRequest)
		return
	}

	valid, err := h.orderService.ValidateFeedback(orderID, phone)
	if err != nil || !valid {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Unauthorized access to this feedback request.",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Validated",
	})
}

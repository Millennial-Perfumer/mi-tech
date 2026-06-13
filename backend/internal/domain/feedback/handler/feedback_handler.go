package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	communicationRepoPkg "mi-tech/internal/domain/communication/repository"
	communicationServicePkg "mi-tech/internal/domain/communication/service"
	"mi-tech/internal/domain/feedback/dto"
	"mi-tech/internal/domain/feedback/entity"
	feedbackService "mi-tech/internal/domain/feedback/service"
	"mi-tech/internal/shared/config"
)

type FeedbackHandler struct {
	service          *feedbackService.FeedbackService
	settingsProvider *config.SettingsProvider
	mappingService   *communicationServicePkg.WebhookMappingService
	templatesRepo    communicationRepoPkg.TemplatesRepository
}

func NewFeedbackHandler(
	service *feedbackService.FeedbackService,
	settingsProvider *config.SettingsProvider,
	mappingService *communicationServicePkg.WebhookMappingService,
	templatesRepo communicationRepoPkg.TemplatesRepository,
) *FeedbackHandler {
	return &FeedbackHandler{
		service:          service,
		settingsProvider: settingsProvider,
		mappingService:   mappingService,
		templatesRepo:    templatesRepo,
	}
}

// SubmitFeedback handles POST /api/feedback/submit
func (h *FeedbackHandler) SubmitFeedback(w http.ResponseWriter, r *http.Request) {
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
		Phone   string `json:"phone"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.OrderID == 0 || req.Rating < 1 || req.Rating > 5 {
		http.Error(w, "Missing required fields or invalid rating", http.StatusBadRequest)
		return
	}

	feedback := entity.CustomerFeedback{
		OrderID:       req.OrderID,
		CustomerPhone: req.Phone,
		Rating:        req.Rating,
		Message:       req.Message,
	}

	if err := h.service.SaveCustomerFeedback(feedback); err != nil {
		http.Error(w, "Failed to save feedback", http.StatusInternalServerError)
		return
	}

	if err := h.service.UpdateFeedbackStatus(req.OrderID, 3); err != nil {
		log.Printf("Warning: Failed to update order feedback status to completed: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Feedback received. Thank you!",
	})
}

// GetFeedback handles GET /api/feedback (Admin only)
func (h *FeedbackHandler) GetFeedback(w http.ResponseWriter, r *http.Request) {
	feedbacks, err := h.service.GetCustomerFeedback()
	if err != nil {
		http.Error(w, "Failed to fetch feedback", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"feedback": feedbacks,
	})
}

// ValidateFeedback handles GET /api/feedback/validate?o=order_id&p=phone
func (h *FeedbackHandler) ValidateFeedback(w http.ResponseWriter, r *http.Request) {
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

	order, err := h.service.GetOrderEntity(orderID)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Feedback request not found.",
		})
		return
	}

	valid, err := h.service.ValidateFeedback(orderID, phone)
	if err != nil || !valid {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Unauthorized access to this feedback request.",
		})
		return
	}

	if order.FeedbackSentAt == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Feedback request has not been sent yet.",
		})
		return
	}

	expiryMins := h.settingsProvider.GetFeedbackExpiryMinutes()
	expiryWindow := time.Duration(expiryMins) * time.Minute
	if time.Since(*order.FeedbackSentAt) > expiryWindow {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("This feedback link has expired (%d-minute limit). Please request a new one.", expiryMins),
		})
		return
	}

	alreadySubmitted := false
	if order.FeedbackStatusID != nil && *order.FeedbackStatusID == 3 {
		alreadySubmitted = true
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":           true,
		"message":           "Validated",
		"already_submitted": alreadySubmitted,
	})
}

// ScanFeedbackCandidates handles GET /api/feedback/scan
func (h *FeedbackHandler) ScanFeedbackCandidates(w http.ResponseWriter, r *http.Request) {
	delayMins := h.settingsProvider.GetFeedbackAutomationDelayMinutes()
	orders, err := h.service.GetOrdersForFeedback(delayMins)
	if err != nil {
		http.Error(w, "Failed to scan for candidates", http.StatusInternalServerError)
		return
	}

	var results []dto.FeedbackScanResult
	for _, order := range orders {
		if order.DeliveredAt == nil {
			log.Printf("DEBUG: Skipping order %d, delivered_at is nil", order.ID)
			continue
		}

		// Keep importing helpers from shared/utils if we need them, but order.CustomerPhone is *string
		// So we dereference it using our local logic or import package
		var phone string
		if order.CustomerPhone != nil {
			phone = *order.CustomerPhone
		}

		if phone == "" {
			log.Printf("DEBUG: Skipping order %d, customer_phone is empty", order.ID)
			continue
		}

		var customerName string
		if order.CustomerName != nil {
			customerName = *order.CustomerName
		}

		results = append(results, dto.FeedbackScanResult{
			ID:            order.ID,
			OrderNumber:   order.OrderNumber,
			CustomerName:  customerName,
			CustomerPhone: phone,
			DeliveredAt:   *order.DeliveredAt,
			FeedbackURL:   h.mappingService.GenerateFeedbackURL(order),
		})
	}

	log.Printf("DEBUG: Found %d eligible feedback orders", len(results))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"orders":  results,
	})
}

// BulkSendFeedbackRequests handles POST /api/feedback/bulk-send
func (h *FeedbackHandler) BulkSendFeedbackRequests(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OrderIDs []int64 `json:"order_ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if len(req.OrderIDs) == 0 {
		http.Error(w, "No orders selected", http.StatusBadRequest)
		return
	}

	templateName := h.settingsProvider.Get("feedback_whatsapp_template_name")
	if templateName == "" {
		http.Error(w, "Feedback template name not configured in Settings", http.StatusBadRequest)
		return
	}

	template, err := h.templatesRepo.GetTemplateByName(config.StoreIDShopify, templateName)
	if err != nil {
		http.Error(w, "Error fetching feedback template", http.StatusInternalServerError)
		return
	}
	if template == nil {
		http.Error(w, fmt.Sprintf("Template '%s' not found", templateName), http.StatusBadRequest)
		return
	}

	successCount := 0
	errorCount := 0

	var wg sync.WaitGroup
	var mu sync.Mutex
	sem := make(chan struct{}, 5)

	for _, id := range req.OrderIDs {
		wg.Add(1)
		go func(orderID int64) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			order, err := h.service.GetOrderEntity(orderID)
			if err != nil {
				log.Printf("Bulk Send Error: Failed to fetch order %d: %v", orderID, err)
				mu.Lock()
				errorCount++
				mu.Unlock()
				return
			}

			storeID := config.StoreIDShopify
			err = h.mappingService.ExecuteManualSend(storeID, template.ID, order)
			if err != nil {
				log.Printf("Bulk Send Error: Failed to send feedback for order %d: %v", orderID, err)
				mu.Lock()
				errorCount++
				mu.Unlock()
				return
			}

			_ = h.service.UpdateFeedbackStatus(order.ID, 2)

			mu.Lock()
			successCount++
			mu.Unlock()
		}(id)
	}

	wg.Wait()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"sent":    successCount,
		"failed":  errorCount,
	})
}

// GetConfigStatus checks if the feedback system is fully configured
func (h *FeedbackHandler) GetConfigStatus(w http.ResponseWriter, r *http.Request) {
	templateName := strings.TrimSpace(h.settingsProvider.Get("feedback_whatsapp_template_name"))
	baseURL := strings.TrimSpace(h.settingsProvider.Get("feedback_base_url"))

	var templateFound bool
	var actualStoreID string
	if templateName != "" {
		t, _ := h.templatesRepo.GetTemplateByName(config.StoreIDShopify, templateName)
		if t != nil {
			templateFound = true
			actualStoreID = config.StoreIDShopify
		}
	}

	missing := []string{}
	if templateName == "" {
		missing = append(missing, "template_name")
	} else if !templateFound {
		missing = append(missing, "template_not_found")
	}

	if baseURL == "" {
		missing = append(missing, "base_url")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":       true,
		"is_configured": len(missing) == 0,
		"missing_items": missing,
		"config": map[string]string{
			"template_name": templateName,
			"base_url":      baseURL,
			"store_id":      actualStoreID,
		},
	})
}

// UpdateFeedbackAdminComment handles PUT /api/orders/feedback/comment.
func (h *FeedbackHandler) UpdateFeedbackAdminComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodOptions {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Missing feedback id", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid feedback id", http.StatusBadRequest)
		return
	}

	var reqBody struct {
		AdminComment string `json:"admin_comment"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateFeedbackAdminComment(id, reqBody.AdminComment); err != nil {
		log.Printf("Error updating admin comment for feedback %d: %v", id, err)
		http.Error(w, "Failed to update admin comment", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Admin comment updated successfully",
	})
}

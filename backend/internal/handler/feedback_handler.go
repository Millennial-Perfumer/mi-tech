package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"mi-tech/internal/automation/whatsapp"
	"mi-tech/internal/config"
	"mi-tech/internal/dto"
	"mi-tech/internal/entity"
	"mi-tech/internal/service"
	"net/http"
	"strings"
	"sync"
	"time"
)

type FeedbackHandler struct {
	orderService     *service.OrderService
	settingsProvider *config.SettingsProvider
	mappingService   *whatsapp.WebhookMappingService
	templatesRepo    *whatsapp.TemplatesRepository
}

func NewFeedbackHandler(orderService *service.OrderService, settingsProvider *config.SettingsProvider, mappingService *whatsapp.WebhookMappingService, templatesRepo *whatsapp.TemplatesRepository) *FeedbackHandler {
	return &FeedbackHandler{
		orderService:     orderService,
		settingsProvider: settingsProvider,
		mappingService:   mappingService,
		templatesRepo:    templatesRepo,
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
		"success":  true,
		"feedback": feedbacks,
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

	order, err := h.orderService.GetOrderEntity(orderID)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Feedback request not found.",
		})
		return
	}

	// 2. Check phone match
	valid, err := h.orderService.ValidateFeedback(orderID, phone)
	if err != nil || !valid {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Unauthorized access to this feedback request.",
		})
		return
	}

	// 3. Check expiry
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
	orders, err := h.orderService.GetOrdersForFeedback(delayMins)
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

		customerPhone := entity.DerefStr(order.CustomerPhone)
		if customerPhone == "" {
			log.Printf("DEBUG: Skipping order %d, customer_phone is empty", order.ID)
			continue
		}

		results = append(results, dto.FeedbackScanResult{
			ID:            order.ID,
			OrderNumber:   order.OrderNumber,
			CustomerName:  entity.DerefStr(order.CustomerName),
			CustomerPhone: customerPhone,
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

	// Fetch explicit feedback template name from settings
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
	// Use a semaphore to limit concurrency to 5 total parallel sends
	sem := make(chan struct{}, 5)

	for _, id := range req.OrderIDs {
		wg.Add(1)
		go func(orderID int64) {
			defer wg.Done()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			order, err := h.orderService.GetOrderEntity(orderID)
			if err != nil {
				log.Printf("Bulk Send Error: Failed to fetch order %d: %v", orderID, err)
				mu.Lock()
				errorCount++
				mu.Unlock()
				return
			}

			// Explicitly use standardized store ID
			storeID := config.StoreIDShopify

			// Trigger explicit feedback send
			err = h.mappingService.ExecuteManualSend(storeID, template.ID, order)
			if err != nil {
				log.Printf("Bulk Send Error: Failed to send feedback for order %d: %v", orderID, err)
				mu.Lock()
				errorCount++
				mu.Unlock()
				return
			}

			// Update feedback status to 'sent' (2)
			_ = h.orderService.UpdateFeedbackStatus(order.ID, 2)

			mu.Lock()
			successCount++
			mu.Unlock()
		}(id)
	}

	// Wait for all goroutines to finish
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

	// Verify template exists
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

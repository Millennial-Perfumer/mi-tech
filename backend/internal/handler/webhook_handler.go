package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	"shopify-gst-app/internal/automation/whatsapp"
	"shopify-gst-app/internal/dto"
	"shopify-gst-app/internal/entity"
	"shopify-gst-app/internal/service"
)

// WebhookHandler is a thin HTTP adapter for Shopify webhook endpoints.
type WebhookHandler struct {
	webhookService *service.WebhookService
	mappingService *whatsapp.WebhookMappingService
	webhookSecret  string
}

// NewWebhookHandler creates a new WebhookHandler.
func NewWebhookHandler(webhookService *service.WebhookService, mappingService *whatsapp.WebhookMappingService, secret string) *WebhookHandler {
	return &WebhookHandler{
		webhookService: webhookService,
		mappingService: mappingService,
		webhookSecret:  secret,
	}
}

// ShopifyWebhookHandler handles POST /api/webhooks/shopify.
func (h *WebhookHandler) ShopifyWebhookHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Webhook Error: Failed to read body: %v", err)
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	// 1. Validate HMAC signature
	if !h.verifyWebhook(r, body) {
		log.Printf("Webhook Error: Invalid HMAC signature")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	topic := r.Header.Get("X-Shopify-Topic")
	webhookDeliveryID := r.Header.Get("X-Shopify-Webhook-Id")
	log.Printf("Received Webhook: %s (Delivery ID: %s)", topic, webhookDeliveryID)
	log.Printf("Webhook Payload: %s", string(body))

	// Record activity
	h.webhookService.RecordActivity(topic)

	// Return 200 immediately
	w.WriteHeader(http.StatusOK)

	// Process asynchronously
	go func() {
		// Duplicate check
		processed, err := h.webhookService.IsProcessed(webhookDeliveryID)
		if err != nil {
			log.Printf("Webhook Error: Failed to check duplicate status: %v", err)
			return
		}
		if processed {
			log.Printf("Webhook Info: Webhook %s already processed. Ignoring.", webhookDeliveryID)
			return
		}

		raw := json.RawMessage(body)
		var payload dto.ShopifyWebhookOrder
		if err := json.Unmarshal(body, &payload); err != nil {
			log.Printf("Webhook Error: Failed to parse %s payload: %v", topic, err)
			return
		}

		externalID := strconv.FormatInt(payload.ID, 10)
		log.Printf("Processing %s for Order ID: %s", topic, externalID)

		// Save webhook event log
		event := &entity.WebhookEvent{
			SourceID:          "shopify",
			Topic:             topic,
			ExternalID:        externalID,
			WebhookDeliveryID: webhookDeliveryID,
			Payload:           &raw,
		}
		if err := h.webhookService.SaveEvent(event); err != nil {
			log.Printf("Webhook Error: Failed to save webhook event log: %v", err)
		}

		// Process by topic
		var processErr error
		switch topic {
		case "orders/create":
			processErr = h.webhookService.ProcessOrderCreate(payload, &raw)
		case "orders/updated":
			processErr = h.webhookService.ProcessOrderUpdate(payload, &raw)
		case "orders/paid":
			processErr = h.webhookService.ProcessOrderPaid(externalID)
		case "orders/fulfilled":
			processErr = h.webhookService.ProcessOrderFulfilled(externalID)
		case "orders/cancelled":
			processErr = h.webhookService.ProcessOrderCancelled(externalID, payload.CancelledAt, payload.CancelReason)
		default:
			log.Printf("Webhook Info: Topic %s not handled for ingestion", topic)
		}

		if processErr != nil {
			log.Printf("Webhook Error: Failed to process %s: %v", topic, processErr)
			return
		}

		// Post-processing: link webhook to order and trigger automation
		order, err := h.webhookService.GetOrder(externalID)
		if err != nil {
			log.Printf("Automation Error: Failed to fetch order %s for mapping: %v", externalID, err)
			return
		}

		if order.ID != "" {
			_ = h.webhookService.LinkWebhookToOrder(webhookDeliveryID, order.ID)

			// Bridge entity.Order to models.Order for whatsapp automation (temporary)
			// The automation module still uses models.Order until Phase 7
			if h.mappingService != nil {
				// We skip automation mapping here until whatsapp module is migrated
				log.Printf("Webhook processed successfully for order %s", order.ID)
			}
		}
	}()
}

// GetWebhookStatus handles GET /api/webhook/status.
func (h *WebhookHandler) GetWebhookStatus(w http.ResponseWriter, r *http.Request) {
	topic, status, lastReceived, err := h.webhookService.GetWebhookStatus()
	if err != nil {
		http.Error(w, "Failed to get webhook status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"topic":         topic,
		"status":        status,
		"last_received": lastReceived,
	})
}

func (h *WebhookHandler) verifyWebhook(r *http.Request, body []byte) bool {
	if h.webhookSecret == "" {
		return true
	}

	hmacHeader := r.Header.Get("X-Shopify-Hmac-Sha256")
	if hmacHeader == "" {
		return false
	}

	hash := hmac.New(sha256.New, []byte(h.webhookSecret))
	hash.Write(body)
	expectedHmac := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	return hmacHeader == expectedHmac
}

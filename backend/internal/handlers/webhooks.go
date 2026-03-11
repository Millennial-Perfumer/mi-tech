package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"shopify-gst-app/internal/automation/whatsapp"
	"shopify-gst-app/internal/models"
	"shopify-gst-app/internal/orders"
	"strconv"
)

type WebhooksHandler struct {
	ordersService  *orders.Service
	mappingService *whatsapp.WebhookMappingService
	webhookSecret  string
	db             *sql.DB
}

func NewWebhooksHandler(orders *orders.Service, mapping *whatsapp.WebhookMappingService, secret string, db *sql.DB) *WebhooksHandler {
	return &WebhooksHandler{
		ordersService:  orders,
		mappingService: mapping,
		webhookSecret:  secret,
		db:             db,
	}
}

func (h *WebhooksHandler) VerifyWebhook(r *http.Request, body []byte) bool {
	if h.webhookSecret == "" {
		return true // Skip validation if secret is not configured
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

func (h *WebhooksHandler) ShopifyWebhookHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Webhook Error: Failed to read body: %v", err)
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	// 1. Validate Secret
	if !h.VerifyWebhook(r, body) {
		log.Printf("Webhook Error: Invalid HMAC signature")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	topic := r.Header.Get("X-Shopify-Topic")
	webhookDeliveryID := r.Header.Get("X-Shopify-Webhook-Id")
	log.Printf("Received Webhook: %s (Delivery ID: %s)", topic, webhookDeliveryID)
	log.Printf("Webhook Payload: %s", string(body))

	// Record activity
	_, _ = h.db.Exec(`
		UPDATE webhook_status 
		SET topic = $1, status = 'active', last_received = CURRENT_TIMESTAMP 
		WHERE id = 1
	`, topic)

	// Return 200 immediately
	w.WriteHeader(http.StatusOK)

	// Process asynchronously
	// Process asynchronously
	go func() {
		// Duplicate Protection Log Check
		processed, err := h.ordersService.IsWebhookProcessed(webhookDeliveryID)
		if err != nil {
			log.Printf("Webhook Error: Failed to check duplicate status: %v", err)
			return
		}
		if processed {
			log.Printf("Webhook Info: Webhook %s already processed. Ignoring.", webhookDeliveryID)
			return
		}

		raw := json.RawMessage(body)
		var payload orders.ShopifyWebhookOrder
		if err := json.Unmarshal(body, &payload); err != nil {
			log.Printf("Webhook Error: Failed to parse %s payload: %v", topic, err)
			return
		}

		externalID := strconv.FormatInt(payload.ID, 10)
		log.Printf("Processing %s for Order ID: %s", topic, externalID)

		// Record Webhook Event Initially
		event := &models.WebhookEvent{
			SourceID:          "shopify",
			Topic:             topic,
			ExternalID:        externalID,
			WebhookDeliveryID: webhookDeliveryID,
			Payload:           &raw,
		}
		if err := h.ordersService.SaveWebhookEvent(event); err != nil {
			log.Printf("Webhook Error: Failed to save webhook event log: %v", err)
			// Proceeding anyway but logging
		}

		var processErr error
		switch topic {
		case "orders/create":
			processErr = h.ordersService.CreateOrderFromWebhook(payload, &raw)
		case "orders/updated":
			processErr = h.ordersService.UpdateOrderFromWebhook(payload, &raw)
		case "orders/paid":
			processErr = h.ordersService.UpdatePaymentStatus(externalID, "PAID")
		case "orders/fulfilled":
			processErr = h.ordersService.UpdateFulfillmentStatus(externalID, "FULFILLED")
		case "orders/cancelled":
			processErr = h.ordersService.CancelOrder(externalID, payload.CancelledAt, payload.CancelReason)
		default:
			log.Printf("Webhook Info: Topic %s not handled for ingestion", topic)
		}

		if processErr != nil {
			log.Printf("Webhook Error: Failed to process %s: %v", topic, processErr)
			return
		}

		// Post-processing order linkage for events and automation mapped
		order, err := h.ordersService.GetOrder(externalID)
		if err != nil {
			log.Printf("Automation Error: Failed to fetch order %s for mapping: %v", externalID, err)
			return
		}

		if order.ID != "" {
			// Link the processed event to the found/created internal Order
			_ = h.ordersService.LinkWebhookToOrder(webhookDeliveryID, order.ID)

			if err := h.mappingService.ExecuteMapping("1", topic, order); err != nil {
				log.Printf("Automation Error: Failed to execute mapping for topic %s: %v", topic, err)
			}
		}
	}()
}

func (h *WebhooksHandler) GetWebhookStatus(w http.ResponseWriter, r *http.Request) {
	var topic, status, lastReceived string
	err := h.db.QueryRow("SELECT topic, status, last_received FROM webhook_status WHERE id = 1").Scan(&topic, &status, &lastReceived)
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

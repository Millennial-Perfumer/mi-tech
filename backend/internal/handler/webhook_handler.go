package handler

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"mi-tech/internal/automation/whatsapp"
	"mi-tech/internal/config"
	"mi-tech/internal/dto"
	"mi-tech/internal/entity"
	"mi-tech/internal/service"
)

// WebhookHandler is a thin HTTP adapter for Shopify webhook endpoints.
type WebhookHandler struct {
	webhookService *service.WebhookService
	mappingService *whatsapp.WebhookMappingService
	settings       *config.SettingsProvider
}

// NewWebhookHandler creates a new WebhookHandler.
func NewWebhookHandler(webhookService *service.WebhookService, mappingService *whatsapp.WebhookMappingService, settings *config.SettingsProvider) *WebhookHandler {
	return &WebhookHandler{
		webhookService: webhookService,
		mappingService: mappingService,
		settings:       settings,
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
		var externalID string
		var automationTopic string
		var processErr error

		// Use Background context for async processing
		ctx := context.Background()

		// Process by topic using appropriate payload type
		if strings.HasPrefix(topic, "orders/") {
			var payload dto.ShopifyWebhookOrder
			if err := json.Unmarshal(body, &payload); err != nil {
				log.Printf("Webhook Error: Failed to parse %s payload: %v", topic, err)
				return
			}
			externalID = strconv.FormatInt(payload.ID, 10)
			automationTopic = topic

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
				processErr = h.webhookService.ProcessOrderUpdate(payload, &raw)
			}
		} else if strings.HasPrefix(topic, "fulfillments/") {
			var payload dto.ShopifyWebhookFulfillment
			if err := json.Unmarshal(body, &payload); err != nil {
				log.Printf("Webhook Error: Failed to parse %s payload: %v", topic, err)
				return
			}
			externalID = strconv.FormatInt(payload.OrderID, 10)

			switch topic {
			case "fulfillments/create":
				automationTopic = "orders/assigned"
				processErr = h.webhookService.ProcessFulfillmentCreate(ctx, payload)
			case "fulfillments/update":
				automationTopic = "fulfillments/update"
				processErr = h.webhookService.ProcessFulfillmentUpdate(ctx, payload)
			}
		} else {
			log.Printf("Webhook Info: Topic %s not handled for ingestion", topic)
			return
		}

		if processErr != nil {
			log.Printf("Webhook Error: Failed to process %s: %v", topic, processErr)
		}

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

		// Post-processing: link webhook to order and trigger automation
		order, err := h.webhookService.GetOrder(externalID)
		if err != nil {
			log.Printf("Automation Error: Failed to fetch order %s from DB for mapping: %v", externalID, err)
			return
		}

		if order.ID == 0 {
			log.Printf("Automation Warning: Order %s not found in DB after ingestion. Skipping automation.", externalID)
			return
		}

		_ = h.webhookService.LinkWebhookToOrder(webhookDeliveryID, order.ID)

		if h.mappingService != nil && automationTopic != "" {

			log.Printf("Automation Info: Proceeding to trigger mapping for topic %s and Order %d", automationTopic, order.ID)
			// Re-verify granular status ONLY for fulfillment updates.
			// Generic orders/updated will retain its topic to allow manual edit notifications (invoices).
			if topic == "fulfillments/update" {
				deliveryStatus := ""
				if order.DeliveryStatus != nil {
					deliveryStatus = strings.ToLower(*order.DeliveryStatus)
				}

				switch deliveryStatus {
				case "delivered":
					automationTopic = "orders/delivered"
				case "out for delivery", "out_for_delivery":
					automationTopic = "orders/out_for_delivery"
				case "picked up", "in transit", "in_transit", "picked_up":
					automationTopic = "orders/fulfilled"
				case "confirmed":
					log.Printf("Automation Info: Fulfillment status is 'confirmed' (Order %d). Skipping as assignment is handled by fulfillments/create.", order.ID)
					return
				}
			}

			// Guard 1: Skip ghost updates following order creation (30s window)
			if automationTopic == "orders/updated" && time.Since(order.CreatedAt).Seconds() < 30 {
				log.Printf("Automation Skip: Topic %s ignored for order %d (Created %v ago). Filtering ghost update.", automationTopic, order.ID, time.Since(order.CreatedAt))
				return
			}

			// Guard 2: Skip generic updates following a specific status change/fulfillment (15s window)
			// Shopify sends redundant generic 'orders/updated' right after more specific ones.
			// We check the delivery status to see if it was RECENTLY changed by a specific hook.
			if automationTopic == "orders/updated" && order.DeliveryStatus != nil && time.Since(order.UpdatedAt).Seconds() < 5 {
				log.Printf("Automation Info: Topic %s detected for order %d. This might be a side-effect, but we will allow it if it is a manual edit.", automationTopic, order.ID)
			}

			err = h.mappingService.ExecuteMapping("1", automationTopic, order)
			if err != nil {
				log.Printf("Automation Error: Failed to execute mapping for order %d, topic %s: %v", order.ID, automationTopic, err)
			} else {
				log.Printf("Automation Success: Triggered %s for order %d", automationTopic, order.ID)
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
	secret := h.settings.GetShopifyWebhookSecret()
	if secret == "" {
		return true
	}

	hmacHeader := r.Header.Get("X-Shopify-Hmac-Sha256")
	if hmacHeader == "" {
		return false
	}

	hash := hmac.New(sha256.New, []byte(secret))
	hash.Write(body)
	expectedHmac := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	return hmacHeader == expectedHmac
}

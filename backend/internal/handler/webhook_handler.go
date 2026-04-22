package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
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
// @Summary Shopify Webhook Ingestion
// @Description Secure endpoint for Shopify to push real-time order, fulfillment, and customer updates.
// @Tags webhooks
// @Accept json
// @Param X-Shopify-Topic header string true "Webhook Topic (e.g., orders/create)"
// @Param X-Shopify-Hmac-Sha256 header string true "HMAC Signature"
// @Success 200 {string} string "OK"
// @Router /webhooks/shopify [post]
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

	// Return 200 immediately
	w.WriteHeader(http.StatusOK)

	// Process asynchronously
	go func() {
		log.Printf("Webhook background processing started for delivery %s", webhookDeliveryID)
		h.webhookService.RecordActivity(topic)
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

		// Process by topic using appropriate payload type
		if strings.HasPrefix(topic, "orders/") {
			var payload dto.ShopifyWebhookOrder
			if err := json.Unmarshal(body, &payload); err != nil {
				log.Printf("Webhook Error: Failed to parse %s payload: %v", topic, err)
				return
			}
			externalID = strconv.FormatInt(payload.ID, 10)
			automationTopic = topic
			log.Printf("Webhook Processing: Handling topic %s for External ID %s", topic, externalID)

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
			log.Printf("Webhook Processing: Handling topic %s for Order External ID %s", topic, externalID)

			switch topic {
			case "fulfillments/create":
				automationTopic = "orders/assigned"
				processErr = h.webhookService.ProcessFulfillmentCreate(payload)
			case "fulfillments/update":
				// Granular status mapping for Automation
				if payload.ShipmentStatus != nil && *payload.ShipmentStatus == "in_transit" {
					automationTopic = "orders/dispatched"
					log.Printf("Webhook Processing: Fulfillment In Transit -> Using topic: %s", automationTopic)
				} else {
					automationTopic = "fulfillments/update"
				}
				processErr = h.webhookService.ProcessFulfillmentUpdate(payload)
			}
		} else if strings.HasPrefix(topic, "customers/") {
			var payload dto.ShopifyWebhookCustomer
			if err := json.Unmarshal(body, &payload); err != nil {
				log.Printf("Webhook Error: Failed to parse %s payload: %v", topic, err)
				return
			}
			externalID = strconv.FormatInt(payload.ID, 10)
			log.Printf("Webhook Processing: Handling topic %s for Customer External ID %s", topic, externalID)

			switch topic {
			case "customers/create", "customers/update":
				processErr = h.webhookService.ProcessCustomerCreateUpdate(payload, &raw)
			case "customers/delete":
				processErr = h.webhookService.ProcessCustomerDelete(externalID)
			}
		} else {
			log.Printf("Webhook Info: Topic %s not handled for ingestion", topic)
			return
		}

		if processErr != nil {
			log.Printf("Webhook Error: Failed to process %s: %v", topic, processErr)
		} else {
			log.Printf("Webhook Success: Completed processing for topic %s and ID %s", topic, externalID)
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

		// Post-processing: link webhook to order and trigger automation (Only for Order/Fulfillment topics)
		if strings.HasPrefix(topic, "orders/") || strings.HasPrefix(topic, "fulfillments/") {
			order, err := h.webhookService.GetOrder(externalID)
			if err != nil {
				log.Printf("Automation Error: Failed to fetch order %s from DB for mapping: %v", externalID, err)
				return
			}

			if order.ID == 0 {
				log.Printf("Automation Warning: Order %s not found in DB after ingestion (Topic: %s). Skipping automation.", externalID, topic)
				return
			}

			_ = h.webhookService.LinkWebhookToOrder(webhookDeliveryID, order.ID)

			if h.mappingService != nil && automationTopic != "" {
				log.Printf("Automation Trace: Topic=%s, AutomationTopic=%s, OrderID=%d, CurrentDeliveryStatus=%v", topic, automationTopic, order.ID, entity.DerefStr(order.DeliveryStatus))

				// Re-verify granular status ONLY for fulfillment updates.
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
						automationTopic = "orders/dispatched"
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
				if automationTopic == "orders/updated" && order.DeliveryStatus != nil && time.Since(order.UpdatedAt).Seconds() < 5 {
					log.Printf("Automation Info: Topic %s detected for order %d. This might be a side-effect, but we will allow it if it is a manual edit.", automationTopic, order.ID)
				}

				err = h.mappingService.ExecuteMapping(config.StoreIDShopify, automationTopic, order)
				if err != nil {
					log.Printf("Automation Error: Failed to execute mapping for order %d, topic %s: %v", order.ID, automationTopic, err)
				} else {
					log.Printf("Automation Success: Triggered %s for order %d", automationTopic, order.ID)
				}
			}
		}
	}()
}

// GetWebhookStatus handles GET /api/webhook/status.
// @Summary Webhook Health Status
// @Description Check the status of the last received webhook and its processing state.
// @Tags webhooks
// @Security Bearer
// @Produce json
// @Success 200 {object} map[string]string
// @Router /webhook/status [get]
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
		log.Printf("Webhook Error: No shopify_webhook_secret configured. Rejecting request.")
		return false
	}

	hmacHeader := r.Header.Get("X-Shopify-Hmac-Sha256")
	if hmacHeader == "" {
		log.Printf("Webhook Error: Missing X-Shopify-Hmac-Sha256 header")
		return false
	}

	hash := hmac.New(sha256.New, []byte(secret))
	hash.Write(body)
	expectedHmac := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	// Security: Use constant-time comparison to prevent timing attacks
	isMatch := subtle.ConstantTimeCompare([]byte(hmacHeader), []byte(expectedHmac)) == 1
	if !isMatch {
		log.Printf("Webhook HMAC Mismatch!")
		log.Printf("  Body Length: %d", len(body))
	}

	return isMatch
}

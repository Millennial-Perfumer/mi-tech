package service

import (
	"encoding/json"
	"log"

	"shopify-gst-app/internal/dto"
	"shopify-gst-app/internal/entity"
	"shopify-gst-app/internal/mapper"
	"shopify-gst-app/internal/repository"
)

// WebhookService handles the business logic for processing Shopify webhooks.
type WebhookService struct {
	orderService      *OrderService
	webhookEventRepo  repository.WebhookEventRepository
	webhookStatusRepo repository.WebhookStatusRepository
}

// NewWebhookService creates a new WebhookService.
func NewWebhookService(orderService *OrderService, webhookEventRepo repository.WebhookEventRepository, webhookStatusRepo repository.WebhookStatusRepository) *WebhookService {
	return &WebhookService{
		orderService:      orderService,
		webhookEventRepo:  webhookEventRepo,
		webhookStatusRepo: webhookStatusRepo,
	}
}

// RecordActivity updates the webhook status to show activity.
func (s *WebhookService) RecordActivity(topic string) {
	if err := s.webhookStatusRepo.UpdateActivity(topic); err != nil {
		log.Printf("Webhook Error: Failed to update activity: %v", err)
	}
}

// IsProcessed checks if a webhook delivery has already been processed.
func (s *WebhookService) IsProcessed(deliveryID string) (bool, error) {
	return s.webhookEventRepo.IsProcessed(deliveryID)
}

// SaveEvent logs a webhook event.
func (s *WebhookService) SaveEvent(event *entity.WebhookEvent) error {
	return s.webhookEventRepo.Save(event)
}

// ProcessOrderCreate handles the orders/create webhook topic.
func (s *WebhookService) ProcessOrderCreate(payload dto.ShopifyWebhookOrder, raw *json.RawMessage) error {
	order := mapper.WebhookOrderToEntity(payload, raw)
	log.Printf("Processing orders/create for Order ID: %s", order.ID)
	return s.orderService.UpsertOrder(order)
}

// ProcessOrderUpdate handles the orders/updated webhook topic.
func (s *WebhookService) ProcessOrderUpdate(payload dto.ShopifyWebhookOrder, raw *json.RawMessage) error {
	order := mapper.WebhookOrderToEntity(payload, raw)
	log.Printf("Processing orders/updated for Order ID: %s", order.ID)
	return s.orderService.UpsertOrder(order)
}

// ProcessOrderPaid handles the orders/paid webhook topic.
func (s *WebhookService) ProcessOrderPaid(externalOrderID string) error {
	log.Printf("Processing orders/paid for Order ID: %s", externalOrderID)
	return s.orderService.UpdatePaymentStatus(externalOrderID, "PAID")
}

// ProcessOrderFulfilled handles the orders/fulfilled webhook topic.
func (s *WebhookService) ProcessOrderFulfilled(externalOrderID string) error {
	log.Printf("Processing orders/fulfilled for Order ID: %s", externalOrderID)
	return s.orderService.UpdateFulfillmentStatus(externalOrderID, "FULFILLED")
}

// ProcessOrderCancelled handles the orders/cancelled webhook topic.
func (s *WebhookService) ProcessOrderCancelled(externalOrderID string, cancelledAt *string, reason string) error {
	log.Printf("Processing orders/cancelled for Order ID: %s", externalOrderID)
	return s.orderService.CancelOrder(externalOrderID, cancelledAt, reason)
}

// GetOrder retrieves an order by external ID (used for post-processing webhook linkage).
func (s *WebhookService) GetOrder(externalID string) (entity.Order, error) {
	return s.orderService.GetOrderByExternalID(externalID)
}

// LinkWebhookToOrder links a processed webhook event to an internal order.
func (s *WebhookService) LinkWebhookToOrder(deliveryID string, orderID string) error {
	return s.webhookEventRepo.LinkToOrder(deliveryID, orderID)
}

// GetWebhookStatus retrieves the current webhook status for the API.
func (s *WebhookService) GetWebhookStatus() (topic, status, lastReceived string, err error) {
	return s.webhookStatusRepo.Get()
}

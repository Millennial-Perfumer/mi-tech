package service

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"mi-tech/internal/client/shopify"
	"mi-tech/internal/dto"
	"mi-tech/internal/entity"
	"mi-tech/internal/mapper"
	"mi-tech/internal/repository"
)

// WebhookService handles the business logic for processing Shopify webhooks.
type WebhookService struct {
	orderService      *OrderService
	shopifyClient     *shopify.Client
	webhookEventRepo  repository.WebhookEventRepository
	webhookStatusRepo repository.WebhookStatusRepository
}

// NewWebhookService creates a new WebhookService.
func NewWebhookService(orderService *OrderService, shopifyClient *shopify.Client, webhookEventRepo repository.WebhookEventRepository, webhookStatusRepo repository.WebhookStatusRepository) *WebhookService {
	return &WebhookService{
		orderService:      orderService,
		shopifyClient:     shopifyClient,
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
	log.Printf("Processing orders/create for Order ID: %d", order.ID)
	return s.orderService.UpsertOrder(order)
}

// ProcessOrderUpdate handles the orders/updated webhook topic.
func (s *WebhookService) ProcessOrderUpdate(payload dto.ShopifyWebhookOrder, raw *json.RawMessage) error {
	order := mapper.WebhookOrderToEntity(payload, raw)
	log.Printf("Processing orders/updated for Order ID: %d", order.ID)
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

// ProcessFulfillmentCreate handles fulfillments/create webhook.
func (s *WebhookService) ProcessFulfillmentCreate(f dto.ShopifyWebhookFulfillment) error {
	extOrderID := strconv.FormatInt(f.OrderID, 10)
	log.Printf("Processing fulfillments/create for Order ID: %s. Refreshing from Shopify API...", extOrderID)
	
	err := s.RefreshOrder(extOrderID)
	if err != nil {
		log.Printf("Webhook Warning: Failed to refresh order %s from API: %v. Falling back to payload update.", extOrderID, err)
		// Fallback: update tracking info from payload
		return s.orderService.UpdateTrackingInfo(extOrderID, f.TrackingNumber, f.TrackingCompany, f.TrackingUrl, "")
	}
	return nil
}

// ProcessFulfillmentUpdate handles fulfillments/update webhook.
func (s *WebhookService) ProcessFulfillmentUpdate(f dto.ShopifyWebhookFulfillment) error {
	extOrderID := strconv.FormatInt(f.OrderID, 10)
	log.Printf("Processing fulfillments/update for Order ID: %s. Refreshing from Shopify API...", extOrderID)
	
	err := s.RefreshOrder(extOrderID)
	if err != nil {
		log.Printf("Webhook Warning: Failed to refresh order %s from API: %v. Falling back to payload update.", extOrderID, err)
		// Fallback: update tracking info from payload
		return s.orderService.UpdateTrackingInfo(extOrderID, f.TrackingNumber, f.TrackingCompany, f.TrackingUrl, "")
	}
	return nil
}

func (s *WebhookService) RefreshOrder(externalOrderID string) error {
	log.Printf("Refreshing order %s from Shopify API...", externalOrderID)
	so, err := s.shopifyClient.FetchOrderByID(externalOrderID)
	if err != nil {
		return err
	}
	if so == nil {
		return fmt.Errorf("order %s not found in Shopify", externalOrderID)
	}

	order := mapper.GraphQLOrderToEntity(*so)
	order.LineItems = mapper.GraphQLLineItemsToEntities(order.ID, so.LineItems)
	
	return s.orderService.UpsertOrder(order)
}

// GetOrder retrieves an order by external ID (used for post-processing webhook linkage).
func (s *WebhookService) GetOrder(externalID string) (entity.Order, error) {
	return s.orderService.GetOrderByExternalID(externalID)
}

// LinkWebhookToOrder links a processed webhook event to an internal order.
func (s *WebhookService) LinkWebhookToOrder(deliveryID string, orderID int64) error {
	return s.webhookEventRepo.LinkToOrder(deliveryID, orderID)
}

// GetWebhookStatus retrieves the current webhook status for the API.
func (s *WebhookService) GetWebhookStatus() (topic, status, lastReceived string, err error) {
	return s.webhookStatusRepo.Get()
}

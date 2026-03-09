package orders

import (
	"log"
	"shopify-gst-app/internal/models"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateOrderFromWebhook(payload ShopifyWebhookOrder) error {
	order := MapWebhookToOrder(payload)
	log.Printf("Processing orders/create for Order ID: %s", order.ID)
	return s.repo.UpsertOrder(order)
}

func (s *Service) UpdateOrderFromWebhook(payload ShopifyWebhookOrder) error {
	order := MapWebhookToOrder(payload)
	log.Printf("Processing orders/updated for Order ID: %s", order.ID)
	return s.repo.UpsertOrder(order)
}

func (s *Service) UpdatePaymentStatus(shopifyOrderID string, status string) error {
	log.Printf("Processing orders/paid for Order ID: %s", shopifyOrderID)
	// Fulfillment status remains the same during a paid event unless otherwise specified
	return s.repo.UpdateStatus(shopifyOrderID, status, "")
}

func (s *Service) UpdateFulfillmentStatus(shopifyOrderID string, status string) error {
	log.Printf("Processing orders/fulfilled for Order ID: %s", shopifyOrderID)
	// Financial status remains the same during a fulfilled event
	return s.repo.UpdateStatus(shopifyOrderID, "", status)
}

func (s *Service) CancelOrder(shopifyOrderID string, cancelledAt *string, reason string) error {
	log.Printf("Processing orders/cancelled for Order ID: %s", shopifyOrderID)
	return s.repo.CancelOrder(shopifyOrderID, cancelledAt, reason)
}

func (s *Service) GetOrder(id string) (models.Order, error) {
	return s.repo.GetOrder(id)
}

package repository

import (
	"mi-tech/internal/domain/webhook/entity"
)

// WebhookEventRepository defines data access for the webhook_events table.
type WebhookEventRepository interface {
	Save(event *entity.WebhookEvent) error
	IsProcessed(deliveryID string) (bool, error)
	LinkToOrder(deliveryID string, orderID int64) error
}

// WebhookStatusRepository defines data access for the webhook_status table.
type WebhookStatusRepository interface {
	Get() (topic string, status string, lastReceived string, err error)
	UpdateActivity(topic string) error
}

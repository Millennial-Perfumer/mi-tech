package models

import (
	"encoding/json"
	"time"
)

type WebhookEvent struct {
	ID                int              `json:"id"`
	SourceID          string           `json:"source_id"`
	OrderID           *string          `json:"order_id"` // Pointer because it can be null initially
	Topic             string           `json:"topic"`
	ExternalID        string           `json:"external_id"`
	WebhookDeliveryID string           `json:"webhook_delivery_id"`
	Payload           *json.RawMessage `json:"payload"`
	Processed         bool             `json:"processed"`
	CreatedAt         time.Time        `json:"created_at"`
}

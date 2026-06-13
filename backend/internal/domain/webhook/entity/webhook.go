package entity

import (
	"encoding/json"
	"time"
)

// WebhookEvent represents a row in the "webhook_events" table.
type WebhookEvent struct {
	ID                int              `gorm:"column:id;primaryKey;autoIncrement"`
	SourceID          string           `gorm:"column:source_id"`
	OrderID           *int64           `gorm:"column:order_id"`
	Topic             string           `gorm:"column:topic"`
	ExternalID        string           `gorm:"column:external_id"`
	WebhookDeliveryID string           `gorm:"column:webhook_delivery_id"`
	Payload           *json.RawMessage `gorm:"column:payload;type:jsonb"`
	Processed         bool             `gorm:"column:processed"`
	CreatedAt         time.Time        `gorm:"column:created_at"`
}

// TableName matches the DB table name for WebhookEvent.
func (WebhookEvent) TableName() string { return "webhook_events" }

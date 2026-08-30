package entity

import (
	"encoding/json"
	"time"
)

// OrderEvent is an immutable record of a meaningful order change. The current
// order row remains the fast, current-state snapshot; this table preserves the
// values that existed before and after each change.
type OrderEvent struct {
	ID                int64            `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	OrderID           *int64           `gorm:"column:order_id" json:"order_id,omitempty"`
	ExternalOrderID   *string          `gorm:"column:external_order_id" json:"external_order_id,omitempty"`
	EventType         string           `gorm:"column:event_type" json:"event_type"`
	Source            string           `gorm:"column:source" json:"source"`
	ActorType         string           `gorm:"column:actor_type" json:"actor_type"`
	ActorID           *string          `gorm:"column:actor_id" json:"actor_id,omitempty"`
	BeforeData        *json.RawMessage `gorm:"column:before_data;type:jsonb" json:"before_data,omitempty"`
	AfterData         *json.RawMessage `gorm:"column:after_data;type:jsonb" json:"after_data,omitempty"`
	DiffData          *json.RawMessage `gorm:"column:diff_data;type:jsonb" json:"diff_data,omitempty"`
	WebhookDeliveryID *string          `gorm:"column:webhook_delivery_id" json:"webhook_delivery_id,omitempty"`
	RequestID         *string          `gorm:"column:request_id" json:"request_id,omitempty"`
	OccurredAt        time.Time        `gorm:"column:occurred_at" json:"occurred_at"`
	CreatedAt         time.Time        `gorm:"column:created_at" json:"created_at"`
}

func (OrderEvent) TableName() string { return "order_events" }

// CustomerEvent is an immutable record of a customer profile or identity
// change. OrderID is populated when the change was caused by an order sync.
type CustomerEvent struct {
	ID            int64            `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	CustomerID    *int64           `gorm:"column:customer_id" json:"customer_id,omitempty"`
	OrderID       *int64           `gorm:"column:order_id" json:"order_id,omitempty"`
	CustomerPhone *string          `gorm:"column:customer_phone" json:"customer_phone,omitempty"`
	EventType     string           `gorm:"column:event_type" json:"event_type"`
	Source        string           `gorm:"column:source" json:"source"`
	ActorType     string           `gorm:"column:actor_type" json:"actor_type"`
	ActorID       *string          `gorm:"column:actor_id" json:"actor_id,omitempty"`
	BeforeData    *json.RawMessage `gorm:"column:before_data;type:jsonb" json:"before_data,omitempty"`
	AfterData     *json.RawMessage `gorm:"column:after_data;type:jsonb" json:"after_data,omitempty"`
	DiffData      *json.RawMessage `gorm:"column:diff_data;type:jsonb" json:"diff_data,omitempty"`
	RequestID     *string          `gorm:"column:request_id" json:"request_id,omitempty"`
	OccurredAt    time.Time        `gorm:"column:occurred_at" json:"occurred_at"`
	CreatedAt     time.Time        `gorm:"column:created_at" json:"created_at"`
}

func (CustomerEvent) TableName() string { return "customer_events" }

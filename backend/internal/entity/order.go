package entity

import (
	"encoding/json"
	"time"
)

// Order represents a row in the "orders" table.
type Order struct {
	ID                string           `gorm:"column:id;primaryKey"`
	SourceID          string           `gorm:"column:source_id"`
	ExternalOrderID   string           `gorm:"column:external_order_id"`
	StoreID           *string          `gorm:"column:store_id"`
	OrderNumber       string           `gorm:"column:order_number"`
	TotalPrice        float64          `gorm:"column:total_price"`
	SubtotalPrice     *float64         `gorm:"column:subtotal_price"`
	TotalTax          *float64         `gorm:"column:total_tax"`
	Currency          *string          `gorm:"column:currency"`
	FinancialStatus   *string          `gorm:"column:financial_status"`
	FulfillmentStatus *string          `gorm:"column:fulfillment_status"`
	DeliveryStatus    *string          `gorm:"column:delivery_status"`
	TrackingNumber    *string          `gorm:"column:tracking_number"`
	ShippingCompany   *string          `gorm:"column:shipping_company"`
	TrackingUrl       *string          `gorm:"column:tracking_url"`
	Status            *string          `gorm:"column:status"`
	CreatedAt         time.Time        `gorm:"column:created_at"`
	UpdatedAt         time.Time        `gorm:"column:updated_at"`
	CancelledAt       *time.Time       `gorm:"column:cancelled_at"`
	CancelReason      *string          `gorm:"column:cancel_reason"`
	CustomerName      *string          `gorm:"column:customer_name"`
	CustomerFirstName *string          `gorm:"column:customer_first_name"`
	CustomerLastName  *string          `gorm:"column:customer_last_name"`
	CustomerEmail     *string          `gorm:"column:customer_email"`
	CustomerPhone     *string          `gorm:"column:customer_phone"`
	CustomerCity      *string          `gorm:"column:customer_city"`
	CustomerState     *string          `gorm:"column:customer_state"`
	CustomerCountry   *string          `gorm:"column:customer_country"`
	CustomerAddress1  *string          `gorm:"column:customer_address1"`
	CustomerAddress2  *string          `gorm:"column:customer_address2"`
	CustomerZip       *string          `gorm:"column:customer_zip"`
	RawPayload        *json.RawMessage `gorm:"column:raw_payload;type:jsonb"`
	LineItems         []LineItem       `gorm:"foreignKey:OrderID"`
}

func (Order) TableName() string { return "orders" }

// LineItem represents a row in the "order_line_items" table.
type LineItem struct {
	ID        string   `gorm:"column:id;primaryKey"`
	OrderID   string   `gorm:"column:order_id"`
	ProductID *string  `gorm:"column:product_id"`
	VariantID *string  `gorm:"column:variant_id"`
	Title     *string  `gorm:"column:title"`
	SKU       *string  `gorm:"column:sku"`
	HSCode    *string  `gorm:"column:hs_code"`
	Quantity  int      `gorm:"column:quantity"`
	Price     float64  `gorm:"column:price"`
	Discount  float64  `gorm:"column:discount"`
}

func (LineItem) TableName() string { return "order_line_items" }

// WebhookEvent represents a row in the "webhook_events" table.
type WebhookEvent struct {
	ID                int              `gorm:"column:id;primaryKey;autoIncrement"`
	SourceID          string           `gorm:"column:source_id"`
	OrderID           *string          `gorm:"column:order_id"`
	Topic             string           `gorm:"column:topic"`
	ExternalID        string           `gorm:"column:external_id"`
	WebhookDeliveryID string           `gorm:"column:webhook_delivery_id"`
	Payload           *json.RawMessage `gorm:"column:payload;type:jsonb"`
	Processed         bool             `gorm:"column:processed"`
	CreatedAt         time.Time        `gorm:"column:created_at"`
}

func (WebhookEvent) TableName() string { return "webhook_events" }

// Source represents a row in the "sources" table.
type Source struct {
	ID        string    `gorm:"column:id;primaryKey"`
	Name      string    `gorm:"column:name"`
	Enabled   bool      `gorm:"column:enabled"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (Source) TableName() string { return "sources" }

// --- Helper functions to create pointers ---

func StrPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func Float64Ptr(f float64) *float64 {
	return &f
}

func DerefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func DerefFloat64(f *float64) float64 {
	if f == nil {
		return 0
	}
	return *f
}

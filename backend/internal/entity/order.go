package entity

import (
	"encoding/json"
	"time"
)

// Order represents a row in the "orders" table.
type Order struct {
	ID                 int64            `gorm:"column:id;primaryKey;autoIncrement"`
	SourceID           string           `gorm:"column:source_id"`
	ExternalOrderID    string           `gorm:"column:external_order_id"`
	StoreID            *string          `gorm:"column:store_id"`
	OrderNumber        string           `gorm:"column:order_number"`
	TotalPrice         float64          `gorm:"column:total_price"`
	SubtotalPrice      *float64         `gorm:"column:subtotal_price"`
	TotalTax           *float64         `gorm:"column:total_tax"`
	Currency           *string          `gorm:"column:currency"`
	FinancialStatus    *string          `gorm:"column:financial_status"`
	FulfillmentStatus  *string          `gorm:"column:fulfillment_status"`
	DeliveryStatus     *string          `gorm:"column:delivery_status"`
	TrackingNumber     *string          `gorm:"column:tracking_number"`
	ShippingCompany    *string          `gorm:"column:shipping_company"`
	TrackingUrl        *string          `gorm:"column:tracking_url"`
	Status             *string          `gorm:"column:status"`
	CreatedAt          time.Time        `gorm:"column:created_at"`
	UpdatedAt          time.Time        `gorm:"column:updated_at"`
	CancelledAt        *time.Time       `gorm:"column:cancelled_at"`
	CancelReason       *string          `gorm:"column:cancel_reason"`
	CustomerName       *string          `gorm:"column:customer_name"`
	CustomerFirstName  *string          `gorm:"column:customer_first_name"`
	CustomerLastName   *string          `gorm:"column:customer_last_name"`
	CustomerEmail      *string          `gorm:"column:customer_email"`
	CustomerPhone      *string          `gorm:"column:customer_phone"`
	CustomerCity       *string          `gorm:"column:customer_city"`
	CustomerState      *string          `gorm:"column:customer_state"`
	CustomerCountry    *string          `gorm:"column:customer_country"`
	CustomerAddress1   *string          `gorm:"column:customer_address1"`
	CustomerAddress2   *string          `gorm:"column:customer_address2"`
	CustomerZip        *string          `gorm:"column:customer_zip"`
	CustomerExternalID *string          `gorm:"column:customer_external_id"`
	TotalDiscount      float64          `gorm:"column:total_discount"`
	RawPayload         *json.RawMessage `gorm:"column:raw_payload;type:jsonb"`
	LineItems          []LineItem       `gorm:"foreignKey:OrderID"`

	// Feedback System integration
	DeliveredAt       *time.Time `gorm:"column:delivered_at"`
	FeedbackStatusID  *int       `gorm:"column:feedback_status_id"`
	FeedbackSentAt    *time.Time `gorm:"column:feedback_sent_at"`
	InventoryDeducted bool       `gorm:"column:inventory_deducted;default:false"`
	SkipInventorySync bool       `gorm:"-"` // Virtual field to bypass repository-level inventory logic
}

func (Order) TableName() string { return "orders" }

// FeedbackStatus represents a state in the feedback lifecycle
type FeedbackStatus struct {
	ID   int    `gorm:"column:id;primaryKey"`
	Name string `gorm:"column:name"`
}

func (FeedbackStatus) TableName() string { return "feedback_statuses" }

// CustomerFeedback stores the actual rating and message from a customer
type CustomerFeedback struct {
	ID            int       `gorm:"column:id;primaryKey;autoIncrement"`
	OrderID       int64     `gorm:"column:order_id"`
	CustomerPhone string    `gorm:"column:customer_phone"`
	Rating        int       `gorm:"column:rating"`
	Message       string    `gorm:"column:message"`
	AdminComment  *string   `gorm:"column:admin_comment"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

func (CustomerFeedback) TableName() string { return "customer_feedback" }

// LineItem represents a row in the "order_line_items" table.
type LineItem struct {
	ID            string  `gorm:"column:id;primaryKey"`
	OrderID       int64   `gorm:"column:order_id"`
	ProductID     *string `gorm:"column:product_id"`
	VariantID     *string `gorm:"column:variant_id"`
	Title         *string `gorm:"column:title"`
	SKU           *string `gorm:"column:sku"`
	HSCode        *string `gorm:"column:hs_code"`
	Quantity      int     `gorm:"column:quantity"`
	Price         float64 `gorm:"column:price"`
	Discount      float64 `gorm:"column:discount"`
	OrderDiscount float64 `gorm:"column:order_discount"`
}

func (LineItem) TableName() string { return "order_line_items" }

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

func (WebhookEvent) TableName() string { return "webhook_events" }

// Source represents a row in the "sources" table.
type Source struct {
	ID        string    `gorm:"column:id;primaryKey" json:"id"`
	Name      string    `gorm:"column:name" json:"name"`
	Enabled   bool      `gorm:"column:enabled" json:"enabled"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
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

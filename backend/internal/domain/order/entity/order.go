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
}

func (Order) TableName() string { return "orders" }

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

// Source represents a row in the "sources" table.
type Source struct {
	ID        string    `gorm:"column:id;primaryKey" json:"id"`
	Name      string    `gorm:"column:name" json:"name"`
	Enabled   bool      `gorm:"column:enabled" json:"enabled"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

func (Source) TableName() string { return "sources" }

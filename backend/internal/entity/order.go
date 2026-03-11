package entity

import (
	"database/sql"
	"encoding/json"
	"time"
)

// Order represents a row in the "orders" table.
// All fields map 1:1 to database columns.
type Order struct {
	ID                string
	SourceID          string
	ExternalOrderID   string
	StoreID           sql.NullString
	OrderNumber       string
	TotalPrice        float64
	SubtotalPrice     sql.NullFloat64
	TotalTax          sql.NullFloat64
	Currency          sql.NullString
	FinancialStatus   sql.NullString
	FulfillmentStatus sql.NullString
	DeliveryStatus    sql.NullString
	TrackingNumber    sql.NullString
	ShippingCompany   sql.NullString
	TrackingUrl       sql.NullString
	Status            sql.NullString
	CreatedAt         time.Time
	UpdatedAt         time.Time
	CancelledAt       sql.NullTime
	CancelReason      sql.NullString
	CustomerName      sql.NullString
	CustomerFirstName sql.NullString
	CustomerLastName  sql.NullString
	CustomerEmail     sql.NullString
	CustomerPhone     sql.NullString
	CustomerCity      sql.NullString
	CustomerState     sql.NullString
	CustomerCountry   sql.NullString
	CustomerAddress1  sql.NullString
	CustomerAddress2  sql.NullString
	CustomerZip       sql.NullString
	RawPayload        *json.RawMessage
	LineItems         []LineItem
}

// LineItem represents a row in the "order_line_items" table.
type LineItem struct {
	ID        string
	OrderID   string
	ProductID sql.NullString
	VariantID sql.NullString
	Title     sql.NullString
	SKU       sql.NullString
	HSCode    sql.NullString
	Quantity  int
	Price     float64
	Discount  float64
}

// WebhookEvent represents a row in the "webhook_events" table.
type WebhookEvent struct {
	ID                int
	SourceID          string
	OrderID           sql.NullString
	Topic             string
	ExternalID        string
	WebhookDeliveryID string
	Payload           *json.RawMessage
	Processed         bool
	CreatedAt         time.Time
}

// Source represents a row in the "sources" table.
type Source struct {
	ID        string
	Name      string
	Enabled   bool
	CreatedAt time.Time
}

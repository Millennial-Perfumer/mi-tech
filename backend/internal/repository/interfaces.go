package repository

import (
	"shopify-gst-app/internal/entity"
)

// OrderFilter holds query parameters for listing orders.
type OrderFilter struct {
	StartDate string
	EndDate   string
	Page      int
	Limit     int
}

// OrderRepository defines all data access operations for the orders table.
type OrderRepository interface {
	// List retrieves paginated orders with optional date filters.
	List(filter OrderFilter) ([]entity.Order, int, error)

	// GetByID retrieves a single order by internal ID.
	GetByID(id string) (entity.Order, error)

	// GetByExternalID retrieves a single order by external (Shopify) ID.
	GetByExternalID(externalID string) (entity.Order, error)

	// Upsert inserts or updates an order using ON CONFLICT on (source_id, external_order_id).
	Upsert(order entity.Order) error

	// UpsertBatch inserts or updates orders within a single transaction.
	UpsertBatch(orders []entity.Order) error

	// UpdateStatus updates the financial and fulfillment status of an order.
	UpdateStatus(externalOrderID string, financialStatus, fulfillmentStatus string) error

	// UpdateOrderStatus updates the status field of an order by internal ID.
	UpdateOrderStatus(id string, status string) (int64, error)

	// CancelOrder marks an order as cancelled.
	CancelOrder(externalOrderID string, cancelledAt *string, reason string) error

	// TruncateAll removes all orders (and cascades to line items and webhook events).
	TruncateAll() error
}

// LineItemRepository defines all data access operations for the order_line_items table.
type LineItemRepository interface {
	// GetByOrderID retrieves all line items for a given order.
	GetByOrderID(orderID string) ([]entity.LineItem, error)

	// UpsertBatch inserts or updates a batch of line items within a transaction.
	UpsertBatch(tx interface{}, orderID string, items []entity.LineItem) error

	// DeleteByOrderID removes all line items for a given order.
	DeleteByOrderID(tx interface{}, orderID string) error
}

// WebhookEventRepository defines data access for the webhook_events table.
type WebhookEventRepository interface {
	// Save inserts a new webhook event log entry.
	Save(event *entity.WebhookEvent) error

	// IsProcessed checks if a webhook delivery ID already exists.
	IsProcessed(deliveryID string) (bool, error)

	// LinkToOrder updates a webhook event to reference an internal order ID.
	LinkToOrder(deliveryID string, orderID string) error
}

// WebhookStatusRepository defines data access for the webhook_status table.
type WebhookStatusRepository interface {
	// Get retrieves the current webhook status.
	Get() (topic string, status string, lastReceived string, err error)

	// UpdateActivity updates the webhook status to active with the given topic.
	UpdateActivity(topic string) error
}

// MetricsRepository defines data access for dashboard metric queries.
type MetricsRepository interface {
	// GetDashboardMetrics returns aggregated order metrics for a date range.
	GetDashboardMetrics(startDate, endDate string) (totalRevenue float64, totalOrders, cancelledOrders, fulfilledOrders, unfulfilledOrders int, err error)
}

// ReportRepository defines data access for GST report queries.
type ReportRepository interface {
	// GetGSTSummary returns aggregate GST figures for a date range.
	GetGSTSummary(startDate, endDate string) (totalOrders, cancelledOrders, fulfilledOrders, unfulfilledOrders, paidOrders int, totalRevenue, totalTaxable, totalTax float64, err error)

	// GetStateSummary returns per-state revenue/tax breakdown for a date range.
	GetStateSummary(startDate, endDate string) (results []StateSummaryResult, err error)

	// GetHSNSummary returns per-HSN revenue/tax breakdown for a date range.
	GetHSNSummary(startDate, endDate string) (results []HSNSummaryResult, err error)

	// GetDocumentsIssued returns the min/max order numbers and totals for a date range.
	GetDocumentsIssued(startDate, endDate string) (minOrder, maxOrder *int64, total, cancelled int, err error)

	// GetTaxByState returns per-state tax totals for GST split calculation.
	GetTaxByState(startDate, endDate string) (results []StateTaxResult, err error)
}

// --- Result structs used by ReportRepository ---

// StateSummaryResult holds a single row from the state-wise report query.
type StateSummaryResult struct {
	State        string
	Orders       int
	TaxableValue float64
	TotalGST     float64
	Revenue      float64
}

// HSNSummaryResult holds a single row from the HSN-wise report query.
type HSNSummaryResult struct {
	HSNCode      string
	State        string
	ProductCount int
	QtySold      int
	TaxableValue float64
	TotalGST     float64
	Revenue      float64
}

// StateTaxResult holds a single row from the per-state tax query.
type StateTaxResult struct {
	State string
	Tax   float64
}

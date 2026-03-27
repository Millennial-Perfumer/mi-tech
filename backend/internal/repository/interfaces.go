package repository

import (
	"mi-tech/internal/entity"

	"gorm.io/gorm"
)

// OrderFilter holds query parameters for listing orders.
type OrderFilter struct {
	StartDate         string
	EndDate           string
	Page              int
	Limit             int
	Search            string
	Source            string
	FinancialStatus   string
	FulfillmentStatus string
	SortBy            string
	SortOrder         string
}

// OrderRepository defines all data access operations for the orders table.
type OrderRepository interface {
	List(filter OrderFilter) ([]entity.Order, int, error)
	GetByFlexibleID(id string) (entity.Order, error)
	GetByID(id int64) (entity.Order, error)
	GetByExternalID(externalID string) (entity.Order, error)
	Upsert(order entity.Order) error
	UpsertBatch(orders []entity.Order) error
	UpdateStatus(externalOrderID string, financialStatus, fulfillmentStatus string) error
	UpdateOrderStatus(id int64, status string) (int64, error)
	CancelOrder(externalOrderID string, cancelledAt *string, reason string) error
	UpdateTrackingInfo(externalOrderID string, trackingNumber, shippingCompany, trackingUrl, deliveryStatus string) error
	UpdateOrderDetails(id int64, order entity.Order) error
	ListSources() ([]entity.Source, error)
	TruncateAll() error
}

// LineItemRepository defines all data access operations for the order_line_items table.
type LineItemRepository interface {
	GetByOrderID(orderID int64) ([]entity.LineItem, error)
	UpsertBatch(tx *gorm.DB, orderID int64, items []entity.LineItem) error
	DeleteByOrderID(tx *gorm.DB, orderID int64) error
}

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

// MetricsRepository defines data access for dashboard metric queries.
type MetricsRepository interface {
	GetDashboardMetrics(startDate, endDate string) (totalRevenue, cgst, sgst, igst float64, totalOrders, cancelledOrders, fulfilledOrders, unfulfilledOrders int, err error)
}

// ReportRepository defines data access for GST report queries.
type ReportRepository interface {
	GetGSTSummary(startDate, endDate string) (totalOrders, cancelledOrders, fulfilledOrders, unfulfilledOrders, paidOrders int, totalRevenue, totalTaxable, totalTax float64, err error)
	GetStateSummary(startDate, endDate string) (results []StateSummaryResult, err error)
	GetHSNSummary(startDate, endDate string) (results []HSNSummaryResult, err error)
	GetDocumentsIssued(startDate, endDate string) (minOrder, maxOrder *int64, total, cancelled int, err error)
	GetTaxByState(startDate, endDate string) (results []StateTaxResult, err error)
}

// --- Result structs used by ReportRepository ---

type StateSummaryResult struct {
	State        string
	Orders       int
	TaxableValue float64
	TotalGST     float64
	Revenue      float64
}

type HSNSummaryResult struct {
	HSNCode      string
	State        string
	ProductCount int
	QtySold      int
	TaxableValue float64
	TotalGST     float64
	Revenue      float64
}

type StateTaxResult struct {
	State string
	Tax   float64
}

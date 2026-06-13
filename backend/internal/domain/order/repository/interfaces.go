package repository

import (
	"mi-tech/internal/domain/order/entity"

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
	Status            string
}

// OrderRepository defines all data access operations for the orders table.
type OrderRepository interface {
	List(filter OrderFilter) ([]entity.Order, int, error)
	GetByFlexibleID(id string) (entity.Order, error)
	GetByID(id int64) (entity.Order, error)
	GetByExternalID(externalID string) (entity.Order, error)
	Upsert(order entity.Order) ([]int, error)
	UpsertBatch(orders []entity.Order) ([]int, error)
	UpdateStatus(externalOrderID string, financialStatus, fulfillmentStatus string) error
	UpdateFinancialStatus(id int64, status string) error
	UpdateOrderStatus(id int64, status string) (int64, error)
	CancelOrder(externalOrderID string, cancelledAt *string, reason string) error
	UpdateTrackingInfo(externalOrderID string, trackingNumber, shippingCompany, trackingUrl, deliveryStatus string) error
	UpdateOrderDetails(id int64, order entity.Order) error
	ListSources() ([]entity.Source, error)
	GetCustomerStats(phone string) (totalOrders int, totalSpent float64, err error)
	GetCustomersStats(phones []string) (map[string]struct {
		Count int
		Sum   float64
	}, error)
	TruncateAll() error

	// Feedback & Delivery System
	MarkAsDelivered(id int64) error
	GetOrdersForFeedback(delayMinutes int) ([]entity.Order, error)
	GetByIDAndPhone(id int64, phone string) (entity.Order, error)
	UpdateFeedbackStatus(id int64, statusID int) error
	GetNextPOSSequence(terminalCode string) (string, error)
}

// LineItemRepository defines all data access operations for the order_line_items table.
type LineItemRepository interface {
	GetByOrderID(orderID int64) ([]entity.LineItem, error)
	UpsertBatch(tx *gorm.DB, orderID int64, items []entity.LineItem) error
	DeleteByOrderID(tx *gorm.DB, orderID int64) error
}

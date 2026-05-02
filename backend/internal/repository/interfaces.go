package repository

import (
	"mi-tech/internal/dto"
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

// PlannerFilter holds query parameters for tasks and analytics.
type PlannerFilter struct {
	BoardID  uint
	SprintID *uint
	Status   string
	Priority string
	Search   string
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
	UpdateOrderStatus(id int64, status string) (int64, error)
	CancelOrder(externalOrderID string, cancelledAt *string, reason string) error
	UpdateTrackingInfo(externalOrderID string, trackingNumber, shippingCompany, trackingUrl, deliveryStatus string) error
	UpdateOrderDetails(id int64, order entity.Order) error
	ListSources() ([]entity.Source, error)
	GetCustomerStats(phone string) (totalOrders int, totalSpent float64, err error)
	GetCustomersStats(phones []string) (map[string]struct{ Count int; Sum float64 }, error)
	TruncateAll() error

	// Feedback & Delivery System
	MarkAsDelivered(id int64) error
	GetOrdersForFeedback(delayMinutes int) ([]entity.Order, error)
	GetByIDAndPhone(id int64, phone string) (entity.Order, error)
	UpdateFeedbackStatus(id int64, statusID int) error
	SaveCustomerFeedback(feedback entity.CustomerFeedback) error
	GetCustomerFeedback() ([]dto.FeedbackResponse, error)
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
	GetGSTSummary(startDate, endDate string) (GSTSummaryResult, error)
	GetStateSummary(startDate, endDate string) (results []StateSummaryResult, err error)
	GetHSNSummary(startDate, endDate string) (results []HSNSummaryResult, err error)
	GetDocumentsIssued(startDate, endDate string) (minOrder, maxOrder *int64, total, cancelled int, err error)
}

// --- Result structs used by ReportRepository ---

type GSTSummaryResult struct {
	TotalOrders       int
	CancelledOrders   int
	FulfilledOrders   int
	UnfulfilledOrders int
	PaidOrders        int
	TotalRevenue      float64
	TotalTaxable      float64
	TotalTax          float64
	CGST              float64
	SGST              float64
	IGST              float64
}

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

// SocialRepository defines all data access for social media persistence and history.
type SocialRepository interface {
	UpsertAccount(account entity.SocialAccount) error
	GetAccount(platform string) (entity.SocialAccount, error)
	UpsertPost(post entity.SocialPost) error
	ListPosts(platform string, limit int) ([]entity.SocialPost, error)
	UpsertMetricSnapshot(metric entity.SocialMetricHistory) error
	GetHistoricalMetrics(platform string, postID string, days int) ([]entity.SocialMetricHistory, error)
	GetPlatformSummary(platform string, startDate, endDate string) (map[string]interface{}, error)
}

// PlannerRepository defines all data access operations for the planner module.
type PlannerRepository interface {
	// Boards
	ListBoards() ([]entity.PlannerBoard, error)
	GetBoardByID(id uint) (entity.PlannerBoard, error)
	CreateBoard(board *entity.PlannerBoard) error

	// Columns
	ListColumns(boardID uint) ([]entity.PlannerColumn, error)
	UpdateColumnOrder(columns []entity.PlannerColumn) error

	// Sprints
	ListSprints(status string) ([]entity.PlannerSprint, error)
	GetSprintByID(id uint) (entity.PlannerSprint, error)
	CreateSprint(sprint *entity.PlannerSprint) error
	UpdateSprint(sprint *entity.PlannerSprint) error
	DeleteSprint(id uint) error
	UpdateSprintStatus(id uint, status string) error

	// Tasks
	ListTasks(filter PlannerFilter) ([]entity.PlannerTask, error)
	GetTaskByID(id uint) (entity.PlannerTask, error)
	CreateTask(task *entity.PlannerTask) error
	UpdateTask(task *entity.PlannerTask) error
	DeleteTask(id uint) error
	MoveTask(taskID uint, toColumnID uint, newOrder int) error
	
	// Analytics
	GetSprintVelocity(sprintID uint) (int, error)
	GetTaskLeadTime(taskID uint) (float64, error)
	GetNextTicketNumber() (string, error)
}
// InventoryRepository defines all data access for the inventory hub and SKU mappings.
type InventoryRepository interface {
	// Items
	ListItems(search string) ([]entity.InventoryItem, error)
	GetItemByID(id int) (entity.InventoryItem, error)
	CreateItem(item *entity.InventoryItem) error
	UpdateItem(item *entity.InventoryItem) error
	AdjustStock(id int, delta int) error
	UpdateStockCount(id int, val int) error
	GetMaxMISKU() (string, error) // For auto-generation

	// Mappings
	ListMappings() ([]entity.InventoryMapping, error)
	CreateMapping(mapping *entity.InventoryMapping) error

	// Logs
	LogAdjustment(log *entity.InventoryLog) error
	GetLogsByItemID(itemID int) ([]entity.InventoryLog, error)

	// Utilities
	DeleteAll() error
	BulkCreateItem(items []entity.InventoryItem) error
	GetItemByPlatformSKU(platform, externalSKU string) (entity.InventoryItem, error)
}

// OilInventoryRepository defines all data access for raw material oil stock.
type OilInventoryRepository interface {
	List(search string) ([]entity.OilInventory, error)
	GetByID(id int) (entity.OilInventory, error)
	Create(item *entity.OilInventory) error
	Update(item *entity.OilInventory) error
	Delete(id int) error
}

// SupplierRepository defines all data access for vendors.
type SupplierRepository interface {
	List(search string) ([]entity.Supplier, error)
	GetByID(id int) (entity.Supplier, error)
	Create(supplier *entity.Supplier) error
	Update(supplier *entity.Supplier) error
	Delete(id int) error
}

// PurchaseOrderRepository defines all data access for raw material purchases.
type PurchaseOrderRepository interface {
	List() ([]entity.PurchaseOrder, error)
	Create(po *entity.PurchaseOrder) error
	Delete(id int) error
}

// ManufacturingRepository defines all data access for production logs.
type ManufacturingRepository interface {
	List() ([]entity.ManufacturingRecord, error)
	Create(record *entity.ManufacturingRecord) error
	Update(record *entity.ManufacturingRecord) error
	Delete(id int) error
}

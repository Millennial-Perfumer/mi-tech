package mocks

import (
	"mi-tech/internal/dto"
	"mi-tech/internal/entity"
	"mi-tech/internal/repository"

	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) List(filter repository.OrderFilter) ([]entity.Order, int, error) {
	args := m.Called(filter)
	return args.Get(0).([]entity.Order), args.Int(1), args.Error(2)
}

func (m *MockOrderRepository) GetByFlexibleID(id string) (entity.Order, error) {
	args := m.Called(id)
	return args.Get(0).(entity.Order), args.Error(1)
}

func (m *MockOrderRepository) GetByID(id int64) (entity.Order, error) {
	args := m.Called(id)
	return args.Get(0).(entity.Order), args.Error(1)
}

func (m *MockOrderRepository) GetByExternalID(externalID string) (entity.Order, error) {
	args := m.Called(externalID)
	return args.Get(0).(entity.Order), args.Error(1)
}

func (m *MockOrderRepository) Upsert(order entity.Order) ([]int, error) {
	args := m.Called(order)
	return args.Get(0).([]int), args.Error(1)
}

func (m *MockOrderRepository) UpsertBatch(orders []entity.Order) ([]int, error) {
	args := m.Called(orders)
	return args.Get(0).([]int), args.Error(1)
}

func (m *MockOrderRepository) UpdateStatus(externalOrderID string, financialStatus, fulfillmentStatus string) error {
	args := m.Called(externalOrderID, financialStatus, fulfillmentStatus)
	return args.Error(0)
}

func (m *MockOrderRepository) UpdateOrderStatus(id int64, status string) (int64, error) {
	args := m.Called(id, status)
	return int64(args.Int(0)), args.Error(1)
}

func (m *MockOrderRepository) UpdateFinancialStatus(id int64, status string) error {
	args := m.Called(id, status)
	return args.Error(0)
}

func (m *MockOrderRepository) CancelOrder(externalOrderID string, cancelledAt *string, reason string) error {
	args := m.Called(externalOrderID, cancelledAt, reason)
	return args.Error(0)
}

func (m *MockOrderRepository) UpdateTrackingInfo(externalOrderID string, trackingNumber, shippingCompany, trackingUrl, deliveryStatus string) error {
	args := m.Called(externalOrderID, trackingNumber, shippingCompany, trackingUrl, deliveryStatus)
	return args.Error(0)
}

func (m *MockOrderRepository) UpdateOrderDetails(id int64, order entity.Order) error {
	args := m.Called(id, order)
	return args.Error(0)
}

func (m *MockOrderRepository) ListSources() ([]entity.Source, error) {
	args := m.Called()
	return args.Get(0).([]entity.Source), args.Error(1)
}

func (m *MockOrderRepository) GetCustomerStats(phone string) (totalOrders int, totalSpent float64, err error) {
	args := m.Called(phone)
	return args.Int(0), args.Get(1).(float64), args.Error(2)
}

func (m *MockOrderRepository) GetCustomersStats(phones []string) (map[string]struct{ Count int; Sum float64 }, error) {
	args := m.Called(phones)
	return args.Get(0).(map[string]struct{ Count int; Sum float64 }), args.Error(1)
}

func (m *MockOrderRepository) TruncateAll() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockOrderRepository) MarkAsDelivered(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockOrderRepository) GetOrdersForFeedback(delayMinutes int) ([]entity.Order, error) {
	args := m.Called(delayMinutes)
	return args.Get(0).([]entity.Order), args.Error(1)
}

func (m *MockOrderRepository) UpdateFeedbackStatus(id int64, statusID int) error {
	args := m.Called(id, statusID)
	return args.Error(0)
}

func (m *MockOrderRepository) GetByIDAndPhone(id int64, phone string) (entity.Order, error) {
	args := m.Called(id, phone)
	return args.Get(0).(entity.Order), args.Error(1)
}

func (m *MockOrderRepository) SaveCustomerFeedback(feedback entity.CustomerFeedback) error {
	args := m.Called(feedback)
	return args.Error(0)
}

func (m *MockOrderRepository) UpdateFeedbackAdminComment(id int, comment string) error {
	args := m.Called(id, comment)
	return args.Error(0)
}

func (m *MockOrderRepository) GetCustomerFeedback() ([]dto.FeedbackResponse, error) {
	args := m.Called()
	return args.Get(0).([]dto.FeedbackResponse), args.Error(1)
}

func (m *MockOrderRepository) GetNextPOSSequence(terminalCode string) (string, error) {
	args := m.Called(terminalCode)
	return args.String(0), args.Error(1)
}


type MockLineItemRepository struct {
	mock.Mock
}

func (m *MockLineItemRepository) GetByOrderID(orderID int64) ([]entity.LineItem, error) {
	args := m.Called(orderID)
	return args.Get(0).([]entity.LineItem), args.Error(1)
}

func (m *MockLineItemRepository) UpsertBatch(tx *gorm.DB, orderID int64, items []entity.LineItem) error {
	args := m.Called(tx, orderID, items)
	return args.Error(0)
}

func (m *MockLineItemRepository) DeleteByOrderID(tx *gorm.DB, orderID int64) error {
	args := m.Called(tx, orderID)
	return args.Error(0)
}
type MockInventoryRepository struct {
	mock.Mock
}

func (m *MockInventoryRepository) WithTx(tx *gorm.DB) repository.InventoryRepository {
	args := m.Called(tx)
	return args.Get(0).(repository.InventoryRepository)
}

func (m *MockInventoryRepository) ListItems(search string) ([]entity.InventoryItem, error) {
	args := m.Called(search)
	return args.Get(0).([]entity.InventoryItem), args.Error(1)
}

func (m *MockInventoryRepository) GetItemByID(id int) (entity.InventoryItem, error) {
	args := m.Called(id)
	return args.Get(0).(entity.InventoryItem), args.Error(1)
}

func (m *MockInventoryRepository) CreateItem(item *entity.InventoryItem) error {
	args := m.Called(item)
	return args.Error(0)
}

func (m *MockInventoryRepository) UpdateItem(item *entity.InventoryItem) error {
	args := m.Called(item)
	return args.Error(0)
}

func (m *MockInventoryRepository) AdjustStock(id int, delta int) error {
	args := m.Called(id, delta)
	return args.Error(0)
}

func (m *MockInventoryRepository) UpdateStockCount(id int, val int) error {
	args := m.Called(id, val)
	return args.Error(0)
}

func (m *MockInventoryRepository) GetMaxMISKU() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockInventoryRepository) ListMappings() ([]entity.InventoryMapping, error) {
	args := m.Called()
	return args.Get(0).([]entity.InventoryMapping), args.Error(1)
}

func (m *MockInventoryRepository) CreateMapping(mapping *entity.InventoryMapping) error {
	args := m.Called(mapping)
	return args.Error(0)
}

func (m *MockInventoryRepository) DeleteMapping(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockInventoryRepository) DeleteAll() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockInventoryRepository) BulkCreateItem(items []entity.InventoryItem) error {
	args := m.Called(items)
	return args.Error(0)
}

func (m *MockInventoryRepository) LogAdjustment(l *entity.InventoryLog) error {
	args := m.Called(l)
	return args.Error(0)
}

func (m *MockInventoryRepository) GetLogsByItemID(itemID int) ([]entity.InventoryLog, error) {
	args := m.Called(itemID)
	return args.Get(0).([]entity.InventoryLog), args.Error(1)
}

func (m *MockInventoryRepository) GetLogsByExternalOrderID(externalOrderID string) ([]entity.InventoryLog, error) {
	args := m.Called(externalOrderID)
	return args.Get(0).([]entity.InventoryLog), args.Error(1)
}

func (m *MockInventoryRepository) GetItemByPlatformSKU(platform, externalSKU string) (entity.InventoryItem, error) {
	args := m.Called(platform, externalSKU)
	return args.Get(0).(entity.InventoryItem), args.Error(1)
}

type MockMetricsRepository struct {
	mock.Mock
}

func (m *MockMetricsRepository) GetDashboardMetrics(startDate, endDate string, sourceIDs []string) (dto.DashboardMetrics, error) {
	args := m.Called(startDate, endDate, sourceIDs)
	return args.Get(0).(dto.DashboardMetrics), args.Error(1)
}

func (m *MockMetricsRepository) GetTopProducts(startDate, endDate string, sourceIDs []string, limit int) ([]dto.TopProductRow, error) {
	args := m.Called(startDate, endDate, sourceIDs, limit)
	return args.Get(0).([]dto.TopProductRow), args.Error(1)
}

func (m *MockMetricsRepository) GetRevenueTrend(startDate, endDate string, sourceIDs []string) ([]dto.RevenueTrendRow, error) {
	args := m.Called(startDate, endDate, sourceIDs)
	return args.Get(0).([]dto.RevenueTrendRow), args.Error(1)
}

func (m *MockMetricsRepository) GetGeoDistribution(startDate, endDate string, sourceIDs []string, limit int) ([]dto.GeoDistributionRow, error) {
	args := m.Called(startDate, endDate, sourceIDs, limit)
	return args.Get(0).([]dto.GeoDistributionRow), args.Error(1)
}

type MockReportRepository struct {
	mock.Mock
}

func (m *MockReportRepository) GetGSTSummary(startDate, endDate string) (repository.GSTSummaryResult, error) {
	args := m.Called(startDate, endDate)
	return args.Get(0).(repository.GSTSummaryResult), args.Error(1)
}

func (m *MockReportRepository) GetStateSummary(startDate, endDate string) ([]repository.StateSummaryResult, error) {
	args := m.Called(startDate, endDate)
	return args.Get(0).([]repository.StateSummaryResult), args.Error(1)
}

func (m *MockReportRepository) GetHSNSummary(startDate, endDate string) ([]repository.HSNSummaryResult, error) {
	args := m.Called(startDate, endDate)
	return args.Get(0).([]repository.HSNSummaryResult), args.Error(1)
}

func (m *MockReportRepository) GetDocumentsIssued(startDate, endDate string) (*int64, *int64, int, int, error) {
	args := m.Called(startDate, endDate)
	return args.Get(0).(*int64), args.Get(1).(*int64), args.Int(2), args.Int(3), args.Error(4)
}

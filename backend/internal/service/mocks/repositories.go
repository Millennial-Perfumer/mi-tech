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

func (m *MockOrderRepository) Upsert(order entity.Order) error {
	args := m.Called(order)
	return args.Error(0)
}

func (m *MockOrderRepository) UpsertBatch(orders []entity.Order) error {
	args := m.Called(orders)
	return args.Error(0)
}

func (m *MockOrderRepository) UpdateStatus(externalOrderID string, financialStatus, fulfillmentStatus string) error {
	args := m.Called(externalOrderID, financialStatus, fulfillmentStatus)
	return args.Error(0)
}

func (m *MockOrderRepository) UpdateOrderStatus(id int64, status string) (int64, error) {
	args := m.Called(id, status)
	return int64(args.Int(0)), args.Error(1)
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

func (m *MockOrderRepository) GetCustomerFeedback() ([]dto.FeedbackResponse, error) {
	args := m.Called()
	return args.Get(0).([]dto.FeedbackResponse), args.Error(1)
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

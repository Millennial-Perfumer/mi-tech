package test

import (
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
	"mi-tech/internal/domain/inventory/entity"
	"mi-tech/internal/domain/inventory/repository"
)

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

func (m *MockInventoryRepository) GetItemsByIDs(ids []int) ([]entity.InventoryItem, error) {
	args := m.Called(ids)
	return args.Get(0).([]entity.InventoryItem), args.Error(1)
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

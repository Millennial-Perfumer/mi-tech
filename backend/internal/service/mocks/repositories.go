package mocks

import (
	inventoryEntity "mi-tech/internal/domain/inventory/entity"
	inventoryRepo "mi-tech/internal/domain/inventory/repository"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type MockInventoryRepository struct {
	mock.Mock
}

func (m *MockInventoryRepository) WithTx(tx *gorm.DB) inventoryRepo.InventoryRepository {
	args := m.Called(tx)
	return args.Get(0).(inventoryRepo.InventoryRepository)
}

func (m *MockInventoryRepository) ListItems(search string) ([]inventoryEntity.InventoryItem, error) {
	args := m.Called(search)
	return args.Get(0).([]inventoryEntity.InventoryItem), args.Error(1)
}

func (m *MockInventoryRepository) GetItemByID(id int) (inventoryEntity.InventoryItem, error) {
	args := m.Called(id)
	return args.Get(0).(inventoryEntity.InventoryItem), args.Error(1)
}

func (m *MockInventoryRepository) GetItemsByIDs(ids []int) ([]inventoryEntity.InventoryItem, error) {
	args := m.Called(ids)
	return args.Get(0).([]inventoryEntity.InventoryItem), args.Error(1)
}

func (m *MockInventoryRepository) CreateItem(item *inventoryEntity.InventoryItem) error {
	args := m.Called(item)
	return args.Error(0)
}

func (m *MockInventoryRepository) UpdateItem(item *inventoryEntity.InventoryItem) error {
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

func (m *MockInventoryRepository) ListMappings() ([]inventoryEntity.InventoryMapping, error) {
	args := m.Called()
	return args.Get(0).([]inventoryEntity.InventoryMapping), args.Error(1)
}

func (m *MockInventoryRepository) CreateMapping(mapping *inventoryEntity.InventoryMapping) error {
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

func (m *MockInventoryRepository) BulkCreateItem(items []inventoryEntity.InventoryItem) error {
	args := m.Called(items)
	return args.Error(0)
}

func (m *MockInventoryRepository) LogAdjustment(l *inventoryEntity.InventoryLog) error {
	args := m.Called(l)
	return args.Error(0)
}

func (m *MockInventoryRepository) GetLogsByItemID(itemID int) ([]inventoryEntity.InventoryLog, error) {
	args := m.Called(itemID)
	return args.Get(0).([]inventoryEntity.InventoryLog), args.Error(1)
}

func (m *MockInventoryRepository) GetLogsByExternalOrderID(externalOrderID string) ([]inventoryEntity.InventoryLog, error) {
	args := m.Called(externalOrderID)
	return args.Get(0).([]inventoryEntity.InventoryLog), args.Error(1)
}

func (m *MockInventoryRepository) GetItemByPlatformSKU(platform, externalSKU string) (inventoryEntity.InventoryItem, error) {
	args := m.Called(platform, externalSKU)
	return args.Get(0).(inventoryEntity.InventoryItem), args.Error(1)
}

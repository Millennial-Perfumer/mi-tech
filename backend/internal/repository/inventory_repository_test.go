package repository

import (
	"testing"

	"mi-tech/internal/entity"
	"mi-tech/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type InventoryRepositoryTestSuite struct {
	suite.Suite
	db   *gorm.DB
	repo InventoryRepository
}

func (s *InventoryRepositoryTestSuite) SetupSuite() {
	db, err := testutil.SetupTestDB()
	if err != nil {
		s.T().Skip("Skipping InventoryRepository tests: database not available")
	}
	s.db = db
	s.repo = NewInventoryRepository(db)
}

func (s *InventoryRepositoryTestSuite) TearDownSuite() {
	if s.db != nil {
		testutil.CleanupTestDB(s.db)
	}
}

func (s *InventoryRepositoryTestSuite) SetupTest() {
	s.db.Exec("TRUNCATE TABLE inventory_logs RESTART IDENTITY CASCADE")
	s.db.Exec("TRUNCATE TABLE inventory_mappings RESTART IDENTITY CASCADE")
	s.db.Exec("TRUNCATE TABLE inventory_items RESTART IDENTITY CASCADE")
}

func (s *InventoryRepositoryTestSuite) TestSingleMappingPerPlatformAndProduct() {
	// 1. Create a master inventory item
	item := &entity.InventoryItem{
		MISKU:        "mi-01",
		Title:        "Guidance Perfume",
		CurrentStock: 10,
	}
	err := s.repo.CreateItem(item)
	assert.NoError(s.T(), err)

	// 2. Map Amazon SKU SJ-M75W-0PIW to this item
	var1 := "default"
	mapping1 := &entity.InventoryMapping{
		InventoryItemID:   item.ID,
		Platform:          "amazon",
		ExternalSKU:       "SJ-M75W-0PIW",
		ExternalVariantID: &var1,
	}
	err = s.repo.CreateMapping(mapping1)
	assert.NoError(s.T(), err)

	// Verify the row exists in database
	var count int64
	s.db.Model(&entity.InventoryMapping{}).Count(&count)
	assert.Equal(s.T(), int64(1), count)

	// 3. User updates the SKU to a new one (QY-53V9-YDLO)
	mapping2 := &entity.InventoryMapping{
		InventoryItemID:   item.ID,
		Platform:          "amazon",
		ExternalSKU:       "QY-53V9-YDLO",
		ExternalVariantID: &var1,
	}
	err = s.repo.CreateMapping(mapping2)
	assert.NoError(s.T(), err)

	// Verify the database has EXACTLY one amazon mapping for this item, and it is the new one
	s.db.Model(&entity.InventoryMapping{}).Count(&count)
	assert.Equal(s.T(), int64(1), count)

	var fetched entity.InventoryMapping
	err = s.db.First(&fetched).Error
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "QY-53V9-YDLO", fetched.ExternalSKU)
	assert.Equal(s.T(), item.ID, fetched.InventoryItemID)
}

func (s *InventoryRepositoryTestSuite) TestReassignmentOfSKUToAnotherItem() {
	// 1. Create two inventory items
	item1 := &entity.InventoryItem{
		MISKU:        "mi-01",
		Title:        "Perfume One",
		CurrentStock: 10,
	}
	err := s.repo.CreateItem(item1)
	assert.NoError(s.T(), err)

	item2 := &entity.InventoryItem{
		MISKU:        "mi-02",
		Title:        "Perfume Two",
		CurrentStock: 5,
	}
	err = s.repo.CreateItem(item2)
	assert.NoError(s.T(), err)

	// 2. Map Shopify SKU SKU-1 to item1
	var1 := "default"
	mapping1 := &entity.InventoryMapping{
		InventoryItemID:   item1.ID,
		Platform:          "shopify",
		ExternalSKU:       "SKU-1",
		ExternalVariantID: &var1,
	}
	err = s.repo.CreateMapping(mapping1)
	assert.NoError(s.T(), err)

	// 3. Reassign the SAME Shopify SKU (SKU-1) to item2
	mapping2 := &entity.InventoryMapping{
		InventoryItemID:   item2.ID,
		Platform:          "shopify",
		ExternalSKU:       "SKU-1",
		ExternalVariantID: &var1,
	}
	err = s.repo.CreateMapping(mapping2)
	assert.NoError(s.T(), err)

	// Verify that SKU-1 is now mapped ONLY to item2
	var count int64
	s.db.Model(&entity.InventoryMapping{}).Count(&count)
	assert.Equal(s.T(), int64(1), count)

	var fetched entity.InventoryMapping
	err = s.db.First(&fetched).Error
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), item2.ID, fetched.InventoryItemID)
	assert.Equal(s.T(), "SKU-1", fetched.ExternalSKU)
}

func TestInventoryRepositorySuite(t *testing.T) {
	suite.Run(t, new(InventoryRepositoryTestSuite))
}

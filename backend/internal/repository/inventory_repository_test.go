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

func (s *InventoryRepositoryTestSuite) TestCreateAndUpsertMapping() {
	// 1. Create inventory items
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

	// 2. Create a mapping for item1
	variant1 := "var-shopify-1"
	mapping := &entity.InventoryMapping{
		InventoryItemID:   item1.ID,
		Platform:          "shopify",
		ExternalSKU:       "SHOPIFY-SKU-1",
		ExternalVariantID: &variant1,
	}
	err = s.repo.CreateMapping(mapping)
	assert.NoError(s.T(), err)

	// Verify it exists in db
	var count int64
	s.db.Model(&entity.InventoryMapping{}).Count(&count)
	assert.Equal(s.T(), int64(1), count)

	// 3. Upsert mapping for the same platform and SKU but mapping it to item2
	variant2 := "var-shopify-2"
	dupMapping := &entity.InventoryMapping{
		InventoryItemID:   item2.ID,
		Platform:          "shopify",
		ExternalSKU:       "SHOPIFY-SKU-1",
		ExternalVariantID: &variant2,
	}
	// This should run without error using our ON CONFLICT upsert clause
	err = s.repo.CreateMapping(dupMapping)
	assert.NoError(s.T(), err)

	// Count should still be 1 (upserted, not duplicated)
	s.db.Model(&entity.InventoryMapping{}).Count(&count)
	assert.Equal(s.T(), int64(1), count)

	// Verify the mapping was updated to point to item2 and variant2
	var updated entity.InventoryMapping
	err = s.db.First(&updated).Error
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), item2.ID, updated.InventoryItemID)
	assert.Equal(s.T(), "var-shopify-2", *updated.ExternalVariantID)
}

func TestInventoryRepositorySuite(t *testing.T) {
	suite.Run(t, new(InventoryRepositoryTestSuite))
}

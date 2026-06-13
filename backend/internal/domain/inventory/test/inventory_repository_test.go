package test

import (
	"testing"

	"mi-tech/internal/domain/inventory/entity"
	"mi-tech/internal/domain/inventory/repository"
	"mi-tech/internal/shared/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type InventoryRepositoryTestSuite struct {
	suite.Suite
	db   *gorm.DB
	repo repository.InventoryRepository
}

func (s *InventoryRepositoryTestSuite) SetupSuite() {
	db, err := testutil.SetupTestDB()
	if err != nil {
		s.T().Skip("Skipping InventoryRepository tests: database not available")
	}
	s.db = db
	s.repo = repository.NewInventoryRepository(db)
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

func (s *InventoryRepositoryTestSuite) TestBulkCreateItem() {
	var1 := "gid://shopify/InventoryItem/12345"
	items := []entity.InventoryItem{
		{
			MISKU:        "mi-999",
			Title:        "Test Product 1",
			CurrentStock: 5,
			Mappings: []entity.InventoryMapping{
				{
					Platform:          "shopify",
					ExternalSKU:       "TEST-SKU-1",
					ExternalVariantID: &var1,
				},
			},
		},
		{
			MISKU:        "mi-998",
			Title:        "Test Product 2",
			CurrentStock: 10,
			Mappings: []entity.InventoryMapping{
				{
					Platform:          "shopify",
					ExternalSKU:       "TEST-SKU-2",
					ExternalVariantID: &var1,
				},
			},
		},
	}

	err := s.repo.BulkCreateItem(items)
	assert.NoError(s.T(), err)

	// Verify items are created
	var count int64
	s.db.Model(&entity.InventoryItem{}).Count(&count)
	assert.Equal(s.T(), int64(2), count)

	// Verify mappings are created and linked correctly
	var mappings []entity.InventoryMapping
	err = s.db.Find(&mappings).Error
	assert.NoError(s.T(), err)
	assert.Len(s.T(), mappings, 2)
	assert.NotZero(s.T(), mappings[0].InventoryItemID)
	assert.NotZero(s.T(), mappings[1].InventoryItemID)
}

func TestInventoryRepositorySuite(t *testing.T) {
	suite.Run(t, new(InventoryRepositoryTestSuite))
}

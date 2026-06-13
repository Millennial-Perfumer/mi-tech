package test

import (
	"testing"
	"time"

	"mi-tech/internal/domain/production/entity"
	"mi-tech/internal/domain/production/repository"
	"mi-tech/internal/shared/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type ProductionRepositoryTestSuite struct {
	suite.Suite
	db       *gorm.DB
	supplier repository.SupplierRepository
	oil      repository.OilInventoryRepository
	po       repository.PurchaseOrderRepository
	mfg      repository.ManufacturingRepository
}

func (s *ProductionRepositoryTestSuite) SetupSuite() {
	db, err := testutil.SetupTestDB()
	if err != nil {
		s.T().Skip("Skipping ProductionRepository tests: database not available")
	}
	s.db = db
	s.supplier = repository.NewSupplierRepository(db)
	s.oil = repository.NewOilInventoryRepository(db)
	s.po = repository.NewPurchaseOrderRepository(db)
	s.mfg = repository.NewManufacturingRepository(db)
}

func (s *ProductionRepositoryTestSuite) TearDownSuite() {
	if s.db != nil {
		testutil.CleanupTestDB(s.db)
	}
}

func (s *ProductionRepositoryTestSuite) SetupTest() {
	s.db.Exec("TRUNCATE TABLE manufacturing_products RESTART IDENTITY CASCADE")
	s.db.Exec("TRUNCATE TABLE manufacturing_oils RESTART IDENTITY CASCADE")
	s.db.Exec("TRUNCATE TABLE manufacturing_records RESTART IDENTITY CASCADE")
	s.db.Exec("TRUNCATE TABLE purchase_orders RESTART IDENTITY CASCADE")
	s.db.Exec("TRUNCATE TABLE oil_inventory RESTART IDENTITY CASCADE")
	s.db.Exec("TRUNCATE TABLE suppliers RESTART IDENTITY CASCADE")
}

func (s *ProductionRepositoryTestSuite) TestSupplierOperations() {
	sup := &entity.Supplier{
		Name:        "Test Supplier",
		ContactInfo: "info@test.com",
	}

	err := s.supplier.Create(sup)
	assert.NoError(s.T(), err)
	assert.NotZero(s.T(), sup.ID)

	fetched, err := s.supplier.GetByID(sup.ID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "Test Supplier", fetched.Name)

	sup.Name = "Updated Supplier"
	err = s.supplier.Update(sup)
	assert.NoError(s.T(), err)

	list, err := s.supplier.List("Updated")
	assert.NoError(s.T(), err)
	assert.Len(s.T(), list, 1)

	err = s.supplier.Delete(sup.ID)
	assert.NoError(s.T(), err)
}

func (s *ProductionRepositoryTestSuite) TestOilInventoryAndPurchaseOrder() {
	sup := &entity.Supplier{
		Name: "PO Supplier",
	}
	err := s.supplier.Create(sup)
	assert.NoError(s.T(), err)

	oil := &entity.OilInventory{
		Name:       "Rose Oil",
		SupplierID: &sup.ID,
	}
	err = s.oil.Create(oil)
	assert.NoError(s.T(), err)

	po := &entity.PurchaseOrder{
		OilInventoryID: oil.ID,
		SupplierID:     sup.ID,
		QuantityGrams:  1000,
		UnitPricePerKg: 500,
		TotalPrice:     500,
		PurchaseDate:   time.Now(),
	}
	err = s.po.Create(po)
	assert.NoError(s.T(), err)

	list, err := s.po.List()
	assert.NoError(s.T(), err)
	assert.Len(s.T(), list, 1)
}

func TestProductionRepositorySuite(t *testing.T) {
	suite.Run(t, new(ProductionRepositoryTestSuite))
}

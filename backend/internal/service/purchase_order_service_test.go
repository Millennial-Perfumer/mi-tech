package service

import (
	"testing"
	"time"

	"mi-tech/internal/entity"
	"mi-tech/internal/repository"
	"mi-tech/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPurchaseOrderService_BulkCreate(t *testing.T) {
	db, err := testutil.SetupTestDB()
	if err != nil {
		t.Skip("DB not available")
	}
	defer testutil.CleanupTestDB(db)

	// 0. Setup dependencies
	poRepo := repository.NewPurchaseOrderRepository(db)
	oilRepo := repository.NewOilInventoryRepository(db)
	supplierRepo := repository.NewSupplierRepository(db)
	svc := NewPurchaseOrderService(db, poRepo, oilRepo)

	// 1. Create a supplier and some oils
	supplier := &entity.Supplier{Name: "Bulk Supplier"}
	require.NoError(t, supplierRepo.Create(supplier))

	oil1 := &entity.OilInventory{Name: "Oil 1", GramsLeft: entity.Float64Ptr(100)}
	oil2 := &entity.OilInventory{Name: "Oil 2", GramsLeft: entity.Float64Ptr(200)}
	require.NoError(t, oilRepo.Create(oil1))
	require.NoError(t, oilRepo.Create(oil2))

	// 2. Prepare bulk POs
	pos := []entity.PurchaseOrder{
		{
			OilInventoryID: oil1.ID,
			SupplierID:     supplier.ID,
			QuantityGrams:  50,
			UnitPricePerKg: 1000,
			PurchaseDate:   time.Now().Add(-2 * time.Hour),
		},
		{
			OilInventoryID: oil1.ID,
			SupplierID:     supplier.ID,
			QuantityGrams:  30,
			UnitPricePerKg: 1100, // Latest price for oil1
			PurchaseDate:   time.Now().Add(-1 * time.Hour),
		},
		{
			OilInventoryID: oil2.ID,
			SupplierID:     supplier.ID,
			QuantityGrams:  100,
			UnitPricePerKg: 2000,
			PurchaseDate:   time.Now(),
		},
	}

	// 3. Execute BulkCreate
	err = svc.BulkCreate(pos)
	require.NoError(t, err)

	// 4. Verify POs created
	listedPOs, err := svc.List()
	require.NoError(t, err)
	assert.Len(t, listedPOs, 3)

	// 5. Verify Oil 1 Stock and Metadata
	updatedOil1, err := oilRepo.GetByID(oil1.ID)
	require.NoError(t, err)
	assert.Equal(t, 180.0, *updatedOil1.GramsLeft) // 100 + 50 + 30
	assert.Equal(t, 1100.0, *updatedOil1.PurchasePricePerKg)
	assert.Equal(t, supplier.ID, *updatedOil1.SupplierID)

	// 6. Verify Oil 2 Stock
	updatedOil2, err := oilRepo.GetByID(oil2.ID)
	require.NoError(t, err)
	assert.Equal(t, 300.0, *updatedOil2.GramsLeft) // 200 + 100
	assert.Equal(t, 2000.0, *updatedOil2.PurchasePricePerKg)
}

func TestPurchaseOrderService_UpdateAndStockRevert(t *testing.T) {
	db, err := testutil.SetupTestDB()
	if err != nil {
		t.Skip("DB not available")
	}
	defer testutil.CleanupTestDB(db)

	poRepo := repository.NewPurchaseOrderRepository(db)
	oilRepo := repository.NewOilInventoryRepository(db)
	supplierRepo := repository.NewSupplierRepository(db)
	svc := NewPurchaseOrderService(db, poRepo, oilRepo)

	supplier := &entity.Supplier{Name: "Supplier"}
	require.NoError(t, supplierRepo.Create(supplier))

	oil := &entity.OilInventory{Name: "Oil", GramsLeft: entity.Float64Ptr(100)}
	require.NoError(t, oilRepo.Create(oil))

	po := &entity.PurchaseOrder{
		OilInventoryID: oil.ID,
		SupplierID:     supplier.ID,
		QuantityGrams:  50,
		UnitPricePerKg: 1000,
	}
	require.NoError(t, svc.Create(po))

	// Verify initial stock
	o, _ := oilRepo.GetByID(oil.ID)
	assert.Equal(t, 150.0, *o.GramsLeft)

	// Update PO quantity
	po.QuantityGrams = 100
	require.NoError(t, svc.Update(po))

	// Verify updated stock (150 - 50 + 100 = 200)
	o, _ = oilRepo.GetByID(oil.ID)
	assert.Equal(t, 200.0, *o.GramsLeft)

	// Delete PO
	require.NoError(t, svc.Delete(po.ID))

	// Verify reverted stock (200 - 100 = 100)
	o, _ = oilRepo.GetByID(oil.ID)
	assert.Equal(t, 100.0, *o.GramsLeft)
}

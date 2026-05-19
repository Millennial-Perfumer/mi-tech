package service

import (
	"fmt"
	"time"
	"mi-tech/internal/entity"
	"mi-tech/internal/repository"
	"gorm.io/gorm"
)

type PurchaseOrderService struct {
	db      *gorm.DB
	poRepo  repository.PurchaseOrderRepository
	oilRepo repository.OilInventoryRepository
}

func NewPurchaseOrderService(db *gorm.DB, poRepo repository.PurchaseOrderRepository, oilRepo repository.OilInventoryRepository) *PurchaseOrderService {
	return &PurchaseOrderService{
		db:      db,
		poRepo:  poRepo,
		oilRepo: oilRepo,
	}
}

func (s *PurchaseOrderService) List() ([]entity.PurchaseOrder, error) {
	return s.poRepo.List()
}

func (s *PurchaseOrderService) Create(po *entity.PurchaseOrder) error {
	// 0. Ensure date is set
	if po.PurchaseDate.IsZero() {
		po.PurchaseDate = time.Now()
	}

	// 1. Create the PO record
	if err := s.poRepo.Create(po); err != nil {
		return err
	}

	// 2. Automatically update Oil Stock
	oil, err := s.oilRepo.GetByID(po.OilInventoryID)
	if err != nil {
		return fmt.Errorf("failed to find oil stock: %w", err)
	}

	// Update grams and price
	newGrams := 0.0
	if oil.GramsLeft != nil {
		newGrams = *oil.GramsLeft
	}
	newGrams += po.QuantityGrams
	oil.GramsLeft = &newGrams

	// Update price per kg to the latest PO price
	oil.PurchasePricePerKg = &po.UnitPricePerKg
	
	// Ensure supplier is set if not already or if changed
	oil.SupplierID = &po.SupplierID

	return s.oilRepo.Update(&oil)
}

// BulkCreate handles multiple purchase orders in a single optimized transaction.
// Optimization: Batch inserts POs and aggregates oil stock updates to eliminate N+1 queries.
// Expected Impact: Reduces database roundtrips from O(3N) to O(M+1) (where M is unique oils).
func (s *PurchaseOrderService) BulkCreate(pos []entity.PurchaseOrder) error {
	if len(pos) == 0 {
		return nil
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		for i := range pos {
			if pos[i].PurchaseDate.IsZero() {
				pos[i].PurchaseDate = now
			}
		}

		// 1. Batch Create POs in a single roundtrip
		if err := s.poRepo.WithTx(tx).BulkCreate(pos); err != nil {
			return err
		}

		// 2. Aggregate Stock Adjustments and Metadata per Oil
		type oilUpdate struct {
			Delta          float64
			LatestPrice    float64
			LatestSupplier int
			LatestDate     time.Time
		}
		oilUpdates := make(map[int]*oilUpdate)

		for i := range pos {
			po := &pos[i]
			update, exists := oilUpdates[po.OilInventoryID]
			if !exists {
				update = &oilUpdate{}
				oilUpdates[po.OilInventoryID] = update
			}

			update.Delta += po.QuantityGrams
			// Track latest metadata to update the oil record
			if po.PurchaseDate.After(update.LatestDate) || update.LatestDate.IsZero() {
				update.LatestDate = po.PurchaseDate
				update.LatestPrice = po.UnitPricePerKg
				update.LatestSupplier = po.SupplierID
			}
		}

		// 3. Apply aggregated updates to Oil Inventory
		for oilID, upd := range oilUpdates {
			// Atomic stock update + metadata update in one query per oil
			if err := tx.Model(&entity.OilInventory{}).
				Where("id = ?", oilID).
				Updates(map[string]interface{}{
					"grams_left":            gorm.Expr("COALESCE(grams_left, 0) + ?", upd.Delta),
					"purchase_price_per_kg": upd.LatestPrice,
					"supplier_id":           upd.LatestSupplier,
					"updated_at":            now,
				}).Error; err != nil {
				return fmt.Errorf("failed to update oil stock and metadata for ID %d: %w", oilID, err)
			}
		}

		return nil
	})
}

func (s *PurchaseOrderService) Update(po *entity.PurchaseOrder) error {
	if po.ID == 0 {
		return fmt.Errorf("id is required for update")
	}

	// 1. Get old record to calculate delta
	oldPO, err := s.poRepo.GetByID(po.ID)
	if err != nil {
		return fmt.Errorf("failed to find existing purchase order: %w", err)
	}

	// 2. Ensure date is set
	if po.PurchaseDate.IsZero() {
		po.PurchaseDate = time.Now()
	}

	// 3. Adjust Oil Stock
	if oldPO.OilInventoryID != po.OilInventoryID {
		// Oil changed: Revert old, add new
		if err := s.adjustOilStock(oldPO.OilInventoryID, -oldPO.QuantityGrams); err != nil {
			return err
		}
		if err := s.adjustOilStock(po.OilInventoryID, po.QuantityGrams); err != nil {
			return err
		}
	} else {
		// Same oil: Adjust by difference
		delta := po.QuantityGrams - oldPO.QuantityGrams
		if delta != 0 {
			if err := s.adjustOilStock(po.OilInventoryID, delta); err != nil {
				return err
			}
		}
	}

	// 4. Always update latest price and supplier on the oil record
	oil, err := s.oilRepo.GetByID(po.OilInventoryID)
	if err != nil {
		return fmt.Errorf("failed to find oil for metadata update: %w", err)
	}
	oil.PurchasePricePerKg = &po.UnitPricePerKg
	oil.SupplierID = &po.SupplierID
	if err := s.oilRepo.Update(&oil); err != nil {
		return fmt.Errorf("failed to update oil metadata: %w", err)
	}

	return s.poRepo.Update(po)
}

func (s *PurchaseOrderService) Delete(id int) error {
	// 1. Get the record to revert stock
	po, err := s.poRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to find purchase order: %w", err)
	}

	// 2. Revert stock — block deletion if this fails
	if err := s.adjustOilStock(po.OilInventoryID, -po.QuantityGrams); err != nil {
		return fmt.Errorf("failed to revert oil stock: %w", err)
	}

	return s.poRepo.Delete(id)
}

// Helper to adjust stock
func (s *PurchaseOrderService) adjustOilStock(oilID int, delta float64) error {
	oil, err := s.oilRepo.GetByID(oilID)
	if err != nil {
		return fmt.Errorf("failed to find oil stock: %w", err)
	}

	currentGrams := 0.0
	if oil.GramsLeft != nil {
		currentGrams = *oil.GramsLeft
	}
	newGrams := currentGrams + delta
	if newGrams < 0 {
		newGrams = 0
	}
	oil.GramsLeft = &newGrams

	return s.oilRepo.Update(&oil)
}

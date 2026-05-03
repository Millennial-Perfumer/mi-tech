package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"
	"mi-tech/internal/entity"
	"mi-tech/internal/repository"
)

type ManufacturingService struct {
	db           *gorm.DB
	mfgRepo      repository.ManufacturingRepository
	oilRepo      repository.OilInventoryRepository
	orchestrator *SyncOrchestrator
}

func NewManufacturingService(db *gorm.DB, mfgRepo repository.ManufacturingRepository, oilRepo repository.OilInventoryRepository, orchestrator *SyncOrchestrator) *ManufacturingService {
	return &ManufacturingService{
		db:           db,
		mfgRepo:      mfgRepo,
		oilRepo:      oilRepo,
		orchestrator: orchestrator,
	}
}

func (s *ManufacturingService) List() ([]entity.ManufacturingRecord, error) {
	return s.mfgRepo.List()
}

func (s *ManufacturingService) Create(ctx context.Context, record *entity.ManufacturingRecord) error {
	// 0. Ensure date is set
	if record.ManufacturingDate.IsZero() {
		record.ManufacturingDate = time.Now()
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// 1. Create the manufacturing record
		if err := s.mfgRepo.WithTx(tx).Create(record); err != nil {
			return err
		}

		// 2. Apply stock changes (internal only)
		// Collect products that need syncing
		var productsToSync []int
		for _, mo := range record.Oils {
			if !mo.DeductInventory {
				continue
			}
			if err := s.adjustOilStockWithTx(tx, mo.OilInventoryID, -mo.QuantityGrams); err != nil {
				return err
			}
		}

		for _, mp := range record.Products {
			if !mp.AddStock {
				continue
			}
			reason := fmt.Sprintf("Production Batch #%d", record.ID)
			// Adjust internal stock only
			if err := s.orchestrator.WithTx(tx).AdjustStockInternal(ctx, mp.InventoryItemID, mp.QuantityProduced, "internal", reason, nil); err != nil {
				return err
			}
			productsToSync = append(productsToSync, mp.InventoryItemID)
		}

		// After transaction commit, we trigger external sync
		defer func() {
			for _, itemID := range productsToSync {
				s.orchestrator.GlobalSync(ctx, itemID, "internal")
			}
		}()

		return nil
	})
}

func (s *ManufacturingService) Update(ctx context.Context, record *entity.ManufacturingRecord) error {
	var productsToSync []int

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 1. Get old record state (using tx)
		oldRecord, err := s.mfgRepo.WithTx(tx).GetByID(record.ID)
		if err != nil {
			return fmt.Errorf("failed to find existing manufacturing record: %w", err)
		}

		// 2. Compute oil inventory deltas
		oldOilMap := make(map[int]oilState)
		for _, mo := range oldRecord.Oils {
			oldOilMap[mo.OilInventoryID] = oilState{grams: mo.QuantityGrams, deduct: mo.DeductInventory}
		}
		for _, mo := range record.Oils {
			old, exists := oldOilMap[mo.OilInventoryID]
			if !exists {
				if mo.DeductInventory {
					if err := s.adjustOilStockWithTx(tx, mo.OilInventoryID, -mo.QuantityGrams); err != nil {
						return err
					}
				}
			} else {
				oldDeducted := 0.0
				if old.deduct {
					oldDeducted = old.grams
				}
				newDeducted := 0.0
				if mo.DeductInventory {
					newDeducted = mo.QuantityGrams
				}
				delta := oldDeducted - newDeducted
				if delta != 0 {
					if err := s.adjustOilStockWithTx(tx, mo.OilInventoryID, delta); err != nil {
						return err
					}
				}
				delete(oldOilMap, mo.OilInventoryID)
			}
		}
		for oilID, old := range oldOilMap {
			if old.deduct {
				if err := s.adjustOilStockWithTx(tx, oilID, old.grams); err != nil {
					return err
				}
			}
		}

		// 3. Compute product inventory deltas
		oldProdMap := make(map[int]prodState)
		for _, mp := range oldRecord.Products {
			oldProdMap[mp.InventoryItemID] = prodState{qty: mp.QuantityProduced, addStock: mp.AddStock}
		}
		for _, mp := range record.Products {
			old, exists := oldProdMap[mp.InventoryItemID]
			if !exists {
				if mp.AddStock {
					reason := fmt.Sprintf("Production Batch #%d", record.ID)
					if err := s.orchestrator.WithTx(tx).AdjustStockInternal(ctx, mp.InventoryItemID, mp.QuantityProduced, "internal", reason, nil); err != nil {
						return err
					}
					productsToSync = append(productsToSync, mp.InventoryItemID)
				}
			} else {
				oldAdded := 0
				if old.addStock {
					oldAdded = old.qty
				}
				newAdded := 0
				if mp.AddStock {
					newAdded = mp.QuantityProduced
				}
				delta := newAdded - oldAdded
				if delta != 0 {
					reason := fmt.Sprintf("Production Batch #%d Adjustment", record.ID)
					if err := s.orchestrator.WithTx(tx).AdjustStockInternal(ctx, mp.InventoryItemID, delta, "internal", reason, nil); err != nil {
						return err
					}
					productsToSync = append(productsToSync, mp.InventoryItemID)
				}
				delete(oldProdMap, mp.InventoryItemID)
			}
		}
		for itemID, old := range oldProdMap {
			if old.addStock {
				reason := fmt.Sprintf("Production Batch #%d Product Removed", record.ID)
				if err := s.orchestrator.WithTx(tx).AdjustStockInternal(ctx, itemID, -old.qty, "internal", reason, nil); err != nil {
					return err
				}
				productsToSync = append(productsToSync, itemID)
			}
		}

		// 4. Update record in DB
		return s.mfgRepo.WithTx(tx).Update(record)
	})

	if err == nil {
		// External Sync after commit
		for _, itemID := range productsToSync {
			s.orchestrator.GlobalSync(ctx, itemID, "internal")
		}
	}
	return err
}

// Helper types for delta computation
type oilState struct {
	grams  float64
	deduct bool
}

type prodState struct {
	qty      int
	addStock bool
}


func (s *ManufacturingService) Delete(ctx context.Context, id int) error {
	var productsToSync []int

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 1. Get record to revert (using tx)
		record, err := s.mfgRepo.WithTx(tx).GetByID(id)
		if err != nil {
			return fmt.Errorf("failed to find record for deletion: %w", err)
		}

		// 2. Revert stock changes (internal only)
		// Add back Fragrance Oils
		for _, mo := range record.Oils {
			if !mo.DeductInventory {
				continue
			}
			if err := s.adjustOilStockWithTx(tx, mo.OilInventoryID, mo.QuantityGrams); err != nil {
				return err
			}
		}

		// Deduct Finished Product Stock
		for _, mp := range record.Products {
			if !mp.AddStock {
				continue
			}
			reason := fmt.Sprintf("Production Batch #%d Deleted", record.ID)
			if err := s.orchestrator.WithTx(tx).AdjustStockInternal(ctx, mp.InventoryItemID, -mp.QuantityProduced, "internal", reason, nil); err != nil {
				return err
			}
			productsToSync = append(productsToSync, mp.InventoryItemID)
		}

		// 3. Delete from DB
		return s.mfgRepo.WithTx(tx).Delete(id)
	})

	if err == nil {
		// External Sync after commit
		for _, itemID := range productsToSync {
			s.orchestrator.GlobalSync(ctx, itemID, "internal")
		}
	}
	return err
}

func (s *ManufacturingService) adjustOilStockWithTx(tx *gorm.DB, oilID int, delta float64) error {
	oil, err := s.oilRepo.WithTx(tx).GetByID(oilID)
	if err != nil {
		return fmt.Errorf("failed to find oil stock ID %d: %w", oilID, err)
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
	if err := s.oilRepo.WithTx(tx).Update(&oil); err != nil {
		return fmt.Errorf("failed to update oil stock for %s: %w", oil.Name, err)
	}
	return nil
}

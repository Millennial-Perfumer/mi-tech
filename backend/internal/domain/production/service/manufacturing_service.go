package service

import (
	"context"
	"fmt"
	"time"

	"mi-tech/internal/domain/production/entity"
	"mi-tech/internal/domain/production/repository"
	syncServicePkg "mi-tech/internal/domain/sync/service"

	"gorm.io/gorm"
)

type ManufacturingService struct {
	db           *gorm.DB
	mfgRepo      repository.ManufacturingRepository
	oilRepo      repository.OilInventoryRepository
	orchestrator *syncServicePkg.SyncOrchestrator
}

func NewManufacturingService(db *gorm.DB, mfgRepo repository.ManufacturingRepository, oilRepo repository.OilInventoryRepository, orchestrator *syncServicePkg.SyncOrchestrator) *ManufacturingService {
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
	var productsToSync []int

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 1. Create the manufacturing record
		if err := s.mfgRepo.WithTx(tx).Create(record); err != nil {
			return err
		}

		// 2. Apply stock changes (internal only)
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

		return nil
	})

	if err == nil {
		// External Sync after commit
		for _, itemID := range productsToSync {
			s.orchestrator.GlobalSync(ctx, itemID, "internal")
		}
	}

	return err
}

func (s *ManufacturingService) Update(ctx context.Context, record *entity.ManufacturingRecord) error {
	var productsToSync []int

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 1. Get old record state (using tx)
		oldRecord, err := s.mfgRepo.WithTx(tx).GetByID(record.ID)
		if err != nil {
			return fmt.Errorf("failed to find existing manufacturing record: %w", err)
		}

		// 2. Aggregate Oil Inventory State
		oldOils := make(map[int]float64)
		for _, mo := range oldRecord.Oils {
			if mo.DeductInventory {
				oldOils[mo.OilInventoryID] += mo.QuantityGrams
			}
		}

		newOils := make(map[int]float64)
		for _, mo := range record.Oils {
			if mo.DeductInventory {
				newOils[mo.OilInventoryID] += mo.QuantityGrams
			}
		}

		// Apply Oil deltas
		allOilIDs := make(map[int]bool)
		for id := range oldOils {
			allOilIDs[id] = true
		}
		for id := range newOils {
			allOilIDs[id] = true
		}

		for id := range allOilIDs {
			oldGrams := oldOils[id]
			newGrams := newOils[id]
			delta := oldGrams - newGrams // Positive delta means we return stock, negative means we deduct more
			if delta != 0 {
				if err := s.adjustOilStockWithTx(tx, id, delta); err != nil {
					return err
				}
			}
		}

		// 3. Aggregate Product Inventory State
		oldProds := make(map[int]int)
		for _, mp := range oldRecord.Products {
			if mp.AddStock {
				oldProds[mp.InventoryItemID] += mp.QuantityProduced
			}
		}

		newProds := make(map[int]int)
		for _, mp := range record.Products {
			if mp.AddStock {
				newProds[mp.InventoryItemID] += mp.QuantityProduced
			}
		}

		// Apply Product deltas
		allProdIDs := make(map[int]bool)
		for id := range oldProds {
			allProdIDs[id] = true
		}
		for id := range newProds {
			allProdIDs[id] = true
		}

		for id := range allProdIDs {
			oldQty := oldProds[id]
			newQty := newProds[id]
			delta := newQty - oldQty
			if delta != 0 {
				reason := fmt.Sprintf("Production Batch #%d Adjustment", record.ID)
				if err := s.orchestrator.WithTx(tx).AdjustStockInternal(ctx, id, delta, "internal", reason, nil); err != nil {
					return err
				}
				productsToSync = append(productsToSync, id)
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

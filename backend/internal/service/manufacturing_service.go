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
	mfgRepo      repository.ManufacturingRepository
	oilRepo      repository.OilInventoryRepository
	orchestrator *SyncOrchestrator
}

func NewManufacturingService(mfgRepo repository.ManufacturingRepository, oilRepo repository.OilInventoryRepository, orchestrator *SyncOrchestrator) *ManufacturingService {
	return &ManufacturingService{
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

	// 1. Create the manufacturing record
	if err := s.mfgRepo.Create(record); err != nil {
		return err
	}

	// 2. Apply stock changes
	return s.applyInventoryChanges(ctx, record)
}

func (s *ManufacturingService) Update(ctx context.Context, record *entity.ManufacturingRecord) error {
	// 1. Get old record to revert
	oldRecord, err := s.mfgRepo.GetByID(record.ID)
	if err != nil {
		return fmt.Errorf("failed to find existing manufacturing record: %w", err)
	}

	// 2. Revert old stock changes
	if err := s.revertInventoryChanges(ctx, oldRecord); err != nil {
		return fmt.Errorf("failed to revert old stock changes: %w", err)
	}

	// 3. Apply new stock changes (and update record in DB)
	if err := s.applyInventoryChanges(ctx, record); err != nil {
		return fmt.Errorf("failed to apply new stock changes: %w", err)
	}

	// 4. Update record associations
	return s.mfgRepo.Update(record)
}

func (s *ManufacturingService) Delete(ctx context.Context, id int) error {
	// 1. Get record to revert
	record, err := s.mfgRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to find record for deletion: %w", err)
	}

	// 2. Revert stock changes
	if err := s.revertInventoryChanges(ctx, record); err != nil {
		slog.Error("Failed to revert stock changes during deletion", "error", err)
		// Continue deletion anyway? Or block?
		// Usually we want to keep inventory accurate, so we might want to block.
		return err
	}

	// 3. Delete from DB
	return s.mfgRepo.Delete(id)
}

func (s *ManufacturingService) applyInventoryChanges(ctx context.Context, record *entity.ManufacturingRecord) error {
	// Deduct Fragrance Oils from Stock
	for _, mo := range record.Oils {
		if !mo.DeductInventory {
			continue
		}
		if err := s.adjustOilStock(mo.OilInventoryID, -mo.QuantityGrams); err != nil {
			return err
		}
	}

	// Update Finished Product Stock
	for _, mp := range record.Products {
		if !mp.AddStock {
			continue
		}
		reason := fmt.Sprintf("Production Batch #%d", record.ID)
		if err := s.orchestrator.AdjustStock(ctx, mp.InventoryItemID, mp.QuantityProduced, "internal", reason, nil); err != nil {
			slog.Error("Failed to sync manufacturing stock update", "product_id", mp.InventoryItemID, "error", err)
		}
	}
	return nil
}

func (s *ManufacturingService) revertInventoryChanges(ctx context.Context, record *entity.ManufacturingRecord) error {
	// Add back Fragrance Oils
	for _, mo := range record.Oils {
		if !mo.DeductInventory {
			continue
		}
		if err := s.adjustOilStock(mo.OilInventoryID, mo.QuantityGrams); err != nil {
			return err
		}
	}

	// Deduct Finished Product Stock
	for _, mp := range record.Products {
		if !mp.AddStock {
			continue
		}
		reason := fmt.Sprintf("Production Batch #%d Adjustment/Delete", record.ID)
		if err := s.orchestrator.AdjustStock(ctx, mp.InventoryItemID, -mp.QuantityProduced, "internal", reason, nil); err != nil {
			slog.Error("Failed to revert manufacturing stock update", "product_id", mp.InventoryItemID, "error", err)
		}
	}
	return nil
}

func (s *ManufacturingService) adjustOilStock(oilID int, delta float64) error {
	oil, err := s.oilRepo.GetByID(oilID)
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
	if err := s.oilRepo.Update(&oil); err != nil {
		return fmt.Errorf("failed to update oil stock for %s: %w", oil.Name, err)
	}
	return nil
}

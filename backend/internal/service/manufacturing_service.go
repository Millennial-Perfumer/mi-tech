package service

import (
	"fmt"
	"mi-tech/internal/entity"
	"mi-tech/internal/repository"
)

type ManufacturingService struct {
	mfgRepo repository.ManufacturingRepository
	oilRepo repository.OilInventoryRepository
	invRepo repository.InventoryRepository
}

func NewManufacturingService(mfgRepo repository.ManufacturingRepository, oilRepo repository.OilInventoryRepository, invRepo repository.InventoryRepository) *ManufacturingService {
	return &ManufacturingService{
		mfgRepo: mfgRepo,
		oilRepo: oilRepo,
		invRepo: invRepo,
	}
}

func (s *ManufacturingService) List() ([]entity.ManufacturingRecord, error) {
	return s.mfgRepo.List()
}

func (s *ManufacturingService) Create(record *entity.ManufacturingRecord) error {
	// 1. Create the manufacturing record
	if err := s.mfgRepo.Create(record); err != nil {
		return err
	}

	// 2. Deduct Fragrance Oils from Stock
	for _, mo := range record.Oils {
		if !mo.DeductInventory {
			continue
		}
		oil, err := s.oilRepo.GetByID(mo.OilInventoryID)
		if err != nil {
			return fmt.Errorf("failed to find oil stock for oil ID %d: %w", mo.OilInventoryID, err)
		}

		if oil.GramsLeft != nil {
			newGrams := *oil.GramsLeft - mo.QuantityGrams
			if newGrams < 0 {
				newGrams = 0 // Or return error if negative stock not allowed
			}
			oil.GramsLeft = &newGrams
			if err := s.oilRepo.Update(&oil); err != nil {
				return fmt.Errorf("failed to deduct oil stock for %s: %w", oil.Name, err)
			}
		}
	}

	// 3. Update Finished Product Stock (Warehouse Authority)
	for _, mp := range record.Products {
		if !mp.AddStock {
			continue
		}
		product, err := s.invRepo.GetItemByID(mp.InventoryItemID)
		if err != nil {
			continue // Log error and continue with other products
		}
		
		product.CurrentStock += mp.QuantityProduced
		if err := s.invRepo.UpdateItem(&product); err != nil {
			return fmt.Errorf("failed to update product stock for SKU %s: %w", product.MISKU, err)
		}
	}

	return nil
}

func (s *ManufacturingService) Update(record *entity.ManufacturingRecord) error {
	// For now, we only update metadata. 
	// Adjusting inventory based on composition changes is complex 
	// and should probably be handled by delete/re-create or specific adjustment logic.
	return s.mfgRepo.Update(record)
}

func (s *ManufacturingService) Delete(id int) error {
	return s.mfgRepo.Delete(id)
}

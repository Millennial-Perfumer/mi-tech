package service

import (
	"fmt"
	"mi-tech/internal/entity"
	"mi-tech/internal/repository"
)

type PurchaseOrderService struct {
	poRepo  repository.PurchaseOrderRepository
	oilRepo repository.OilInventoryRepository
}

func NewPurchaseOrderService(poRepo repository.PurchaseOrderRepository, oilRepo repository.OilInventoryRepository) *PurchaseOrderService {
	return &PurchaseOrderService{
		poRepo:  poRepo,
		oilRepo: oilRepo,
	}
}

func (s *PurchaseOrderService) List() ([]entity.PurchaseOrder, error) {
	return s.poRepo.List()
}

func (s *PurchaseOrderService) Create(po *entity.PurchaseOrder) error {
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

func (s *PurchaseOrderService) Delete(id int) error {
	return s.poRepo.Delete(id)
}

package repository

import (
	"mi-tech/internal/entity"
	"gorm.io/gorm"
)

type pgPurchaseOrderRepository struct {
	db *gorm.DB
}

func NewPurchaseOrderRepository(db *gorm.DB) PurchaseOrderRepository {
	return &pgPurchaseOrderRepository{db: db}
}

func (r *pgPurchaseOrderRepository) List() ([]entity.PurchaseOrder, error) {
	var pos []entity.PurchaseOrder
	err := r.db.Preload("OilInventory").Preload("Supplier").Order("purchase_date desc").Find(&pos).Error
	return pos, err
}

func (r *pgPurchaseOrderRepository) Create(po *entity.PurchaseOrder) error {
	return r.db.Create(po).Error
}

func (r *pgPurchaseOrderRepository) Delete(id int) error {
	return r.db.Delete(&entity.PurchaseOrder{}, id).Error
}

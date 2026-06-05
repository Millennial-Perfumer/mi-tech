package repository

import (
	"gorm.io/gorm"
	"mi-tech/internal/entity"
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

func (r *pgPurchaseOrderRepository) GetByID(id int) (*entity.PurchaseOrder, error) {
	var po entity.PurchaseOrder
	err := r.db.First(&po, id).Error
	return &po, err
}

func (r *pgPurchaseOrderRepository) Create(po *entity.PurchaseOrder) error {
	return r.db.Create(po).Error
}

func (r *pgPurchaseOrderRepository) Update(po *entity.PurchaseOrder) error {
	return r.db.Save(po).Error
}

func (r *pgPurchaseOrderRepository) Delete(id int) error {
	return r.db.Delete(&entity.PurchaseOrder{}, id).Error
}

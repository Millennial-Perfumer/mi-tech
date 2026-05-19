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

func (r *pgPurchaseOrderRepository) WithTx(tx *gorm.DB) PurchaseOrderRepository {
	if tx == nil {
		return r
	}
	return &pgPurchaseOrderRepository{db: tx}
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

func (r *pgPurchaseOrderRepository) BulkCreate(pos []entity.PurchaseOrder) error {
	if len(pos) == 0 {
		return nil
	}
	return r.db.Create(&pos).Error
}

func (r *pgPurchaseOrderRepository) Update(po *entity.PurchaseOrder) error {
	return r.db.Save(po).Error
}

func (r *pgPurchaseOrderRepository) Delete(id int) error {
	return r.db.Delete(&entity.PurchaseOrder{}, id).Error
}

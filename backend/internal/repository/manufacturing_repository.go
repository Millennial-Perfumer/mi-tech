package repository

import (
	"mi-tech/internal/entity"
	"gorm.io/gorm"
)

type pgManufacturingRepository struct {
	db *gorm.DB
}

func NewManufacturingRepository(db *gorm.DB) ManufacturingRepository {
	return &pgManufacturingRepository{db: db}
}

func (r *pgManufacturingRepository) List() ([]entity.ManufacturingRecord, error) {
	var records []entity.ManufacturingRecord
	err := r.db.Preload("Oils.OilInventory").Preload("Products.InventoryItem").Order("manufacturing_date desc").Find(&records).Error
	return records, err
}

func (r *pgManufacturingRepository) Create(record *entity.ManufacturingRecord) error {
	return r.db.Create(record).Error
}

func (r *pgManufacturingRepository) Delete(id int) error {
	return r.db.Delete(&entity.ManufacturingRecord{}, id).Error
}

package repository

import (
	"mi-tech/internal/domain/production/entity"

	"gorm.io/gorm"
)

// ManufacturingRepository defines all data access for production logs.
type ManufacturingRepository interface {
	WithTx(tx *gorm.DB) ManufacturingRepository
	List() ([]entity.ManufacturingRecord, error)
	GetByID(id int) (*entity.ManufacturingRecord, error)
	Create(record *entity.ManufacturingRecord) error
	Update(record *entity.ManufacturingRecord) error
	Delete(id int) error
}

type pgManufacturingRepository struct {
	db *gorm.DB
}

func NewManufacturingRepository(db *gorm.DB) ManufacturingRepository {
	return &pgManufacturingRepository{db: db}
}

func (r *pgManufacturingRepository) WithTx(tx *gorm.DB) ManufacturingRepository {
	if tx == nil {
		return r
	}
	return &pgManufacturingRepository{db: tx}
}

func (r *pgManufacturingRepository) List() ([]entity.ManufacturingRecord, error) {
	var records []entity.ManufacturingRecord
	err := r.db.Preload("Oils.OilInventory").Preload("Products.InventoryItem").Order("manufacturing_date desc").Find(&records).Error
	return records, err
}

func (r *pgManufacturingRepository) GetByID(id int) (*entity.ManufacturingRecord, error) {
	var record entity.ManufacturingRecord
	err := r.db.Preload("Oils").Preload("Products").First(&record, id).Error
	return &record, err
}

func (r *pgManufacturingRepository) Create(record *entity.ManufacturingRecord) error {
	return r.db.Create(record).Error
}

func (r *pgManufacturingRepository) Update(record *entity.ManufacturingRecord) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Clear existing associations to prevent duplication when IDs are not tracked in frontend
		if err := tx.Where("manufacturing_record_id = ?", record.ID).Delete(&entity.ManufacturingOil{}).Error; err != nil {
			return err
		}
		if err := tx.Where("manufacturing_record_id = ?", record.ID).Delete(&entity.ManufacturingProduct{}).Error; err != nil {
			return err
		}
		// Save with new associations
		return tx.Save(record).Error
	})
}

func (r *pgManufacturingRepository) Delete(id int) error {
	return r.db.Delete(&entity.ManufacturingRecord{}, id).Error
}

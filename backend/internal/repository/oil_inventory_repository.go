package repository

import (
	"mi-tech/internal/entity"

	"gorm.io/gorm"
)

type gormOilInventoryRepository struct {
	db *gorm.DB
}

func NewOilInventoryRepository(db *gorm.DB) OilInventoryRepository {
	return &gormOilInventoryRepository{db: db}
}

func (r *gormOilInventoryRepository) List(search string) ([]entity.OilInventory, error) {
	var items []entity.OilInventory
	query := r.db.Preload("InventoryItem").Preload("Supplier")
	if search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}
	err := query.Find(&items).Error
	return items, err
}

func (r *gormOilInventoryRepository) GetByID(id int) (entity.OilInventory, error) {
	var item entity.OilInventory
	err := r.db.Preload("InventoryItem").Preload("Supplier").First(&item, id).Error
	return item, err
}

func (r *gormOilInventoryRepository) Create(item *entity.OilInventory) error {
	return r.db.Create(item).Error
}

func (r *gormOilInventoryRepository) Update(item *entity.OilInventory) error {
	return r.db.Save(item).Error
}

func (r *gormOilInventoryRepository) Delete(id int) error {
	return r.db.Delete(&entity.OilInventory{}, id).Error
}

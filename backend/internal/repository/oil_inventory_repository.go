package repository

import (
	"mi-tech/internal/entity"

	"gorm.io/gorm"
)

type pgOilInventoryRepository struct {
	db *gorm.DB
}

func NewOilInventoryRepository(db *gorm.DB) OilInventoryRepository {
	return &pgOilInventoryRepository{db: db}
}

func (r *pgOilInventoryRepository) WithTx(tx *gorm.DB) OilInventoryRepository {
	if tx == nil {
		return r
	}
	return &pgOilInventoryRepository{db: tx}
}

func (r *pgOilInventoryRepository) List(search string) ([]entity.OilInventory, error) {
	var items []entity.OilInventory
	query := r.db.Preload("InventoryItem").Preload("Supplier")
	if search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}
	err := query.Find(&items).Error
	return items, err
}

func (r *pgOilInventoryRepository) GetByID(id int) (entity.OilInventory, error) {
	var item entity.OilInventory
	err := r.db.Preload("InventoryItem").Preload("Supplier").First(&item, id).Error
	return item, err
}

func (r *pgOilInventoryRepository) Create(item *entity.OilInventory) error {
	return r.db.Create(item).Error
}

func (r *pgOilInventoryRepository) Update(item *entity.OilInventory) error {
	return r.db.Save(item).Error
}

func (r *pgOilInventoryRepository) Delete(id int) error {
	return r.db.Delete(&entity.OilInventory{}, id).Error
}

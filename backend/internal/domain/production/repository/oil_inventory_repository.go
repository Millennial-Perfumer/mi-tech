package repository

import (
	"gorm.io/gorm"
	"mi-tech/internal/domain/production/entity"
)

// OilInventoryRepository defines all data access for raw material oil stock.
type OilInventoryRepository interface {
	WithTx(tx *gorm.DB) OilInventoryRepository
	List(search string) ([]entity.OilInventory, error)
	GetByID(id int) (entity.OilInventory, error)
	Create(item *entity.OilInventory) error
	Update(item *entity.OilInventory) error
	Delete(id int) error
	BulkDelete(ids []int) error
}

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

func (r *pgOilInventoryRepository) BulkDelete(ids []int) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.Delete(&entity.OilInventory{}, ids).Error
}

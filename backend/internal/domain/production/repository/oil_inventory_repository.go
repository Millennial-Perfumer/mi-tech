package repository

import (
	"gorm.io/gorm"
	"mi-tech/internal/domain/production/entity"
)

// OilInventoryRepository defines all data access for raw material oil stock.
type OilInventoryRepository interface {
	WithTx(tx *gorm.DB) OilInventoryRepository
	List(search string) ([]entity.OilInventory, error)
	ListPage(search, sort string, page, limit int) ([]entity.OilInventory, int64, error)
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

func (r *pgOilInventoryRepository) ListPage(search, sort string, page, limit int) ([]entity.OilInventory, int64, error) {
	var items []entity.OilInventory
	query := r.db.Model(&entity.OilInventory{})

	if search != "" {
		searchTerm := "%" + search + "%"
		query = query.Where(
			"oil_inventories.name ILIKE ? OR oil_inventories.inventory_item_id IN (SELECT id FROM inventory_items WHERE title ILIKE ? OR mi_sku ILIKE ?) OR oil_inventories.supplier_id IN (SELECT id FROM suppliers WHERE name ILIKE ?)",
			searchTerm, searchTerm, searchTerm, searchTerm,
		)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderBy := map[string]string{
		"name-asc":                   "oil_inventories.name ASC",
		"name-desc":                  "oil_inventories.name DESC",
		"inventory_item.mi_sku-asc":  "inventory_items.mi_sku ASC",
		"inventory_item.mi_sku-desc": "inventory_items.mi_sku DESC",
		"inventory_item.title-asc":   "inventory_items.title ASC",
		"inventory_item.title-desc":  "inventory_items.title DESC",
		"supplier.name-asc":          "suppliers.name ASC",
		"supplier.name-desc":         "suppliers.name DESC",
		"purchase_price_per_kg-asc":  "oil_inventories.purchase_price_per_kg ASC",
		"purchase_price_per_kg-desc": "oil_inventories.purchase_price_per_kg DESC",
		"grams_left-asc":             "oil_inventories.grams_left ASC",
		"grams_left-desc":            "oil_inventories.grams_left DESC",
	}[sort]
	if orderBy == "" {
		orderBy = "oil_inventories.name ASC"
	}

	query = query.Joins("LEFT JOIN inventory_items ON inventory_items.id = oil_inventories.inventory_item_id").Joins("LEFT JOIN suppliers ON suppliers.id = oil_inventories.supplier_id")
	err := query.Preload("InventoryItem").Preload("Supplier").Order(orderBy).Offset((page - 1) * limit).Limit(limit).Find(&items).Error
	return items, total, err
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

package repository

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"mi-tech/internal/domain/inventory/entity"
)

// InventoryRepository defines all data access for the inventory hub and SKU mappings.
type InventoryRepository interface {
	WithTx(tx *gorm.DB) InventoryRepository
	// Items
	ListItems(search string) ([]entity.InventoryItem, error)
	ListItemsPage(search, sort string, page, limit int) ([]entity.InventoryItem, int64, error)
	GetItemByID(id int) (entity.InventoryItem, error)
	GetItemsByIDs(ids []int) ([]entity.InventoryItem, error)
	CreateItem(item *entity.InventoryItem) error
	UpdateItem(item *entity.InventoryItem) error
	AdjustStock(id int, delta int) error
	UpdateStockCount(id int, val int) error
	GetMaxMISKU() (string, error) // For auto-generation

	// Mappings
	ListMappings() ([]entity.InventoryMapping, error)
	CreateMapping(mapping *entity.InventoryMapping) error
	DeleteMapping(id int) error

	// Logs
	LogAdjustment(log *entity.InventoryLog) error
	GetLogsByItemID(itemID int) ([]entity.InventoryLog, error)
	GetLogsByExternalOrderID(externalOrderID string) ([]entity.InventoryLog, error)

	// Utilities
	DeleteAll() error
	BulkCreateItem(items []entity.InventoryItem) error
	BulkUpdatePrices(items []entity.InventoryItem) error
	GetItemByPlatformSKU(platform, externalSKU string) (entity.InventoryItem, error)
	GetItemsBySKUs(skus []string) ([]entity.InventoryItem, error)
}

type gormInventoryRepository struct {
	db *gorm.DB
}

func NewInventoryRepository(db *gorm.DB) InventoryRepository {
	return &gormInventoryRepository{db: db}
}

func (r *gormInventoryRepository) WithTx(tx *gorm.DB) InventoryRepository {
	if tx == nil {
		return r
	}
	return &gormInventoryRepository{db: tx}
}

func (r *gormInventoryRepository) ListItems(search string) ([]entity.InventoryItem, error) {
	var items []entity.InventoryItem
	query := r.db.Preload("Mappings")
	if search != "" {
		searchTerm := "%" + search + "%"
		query = query.Where("mi_sku ILIKE ? OR title ILIKE ?", searchTerm, searchTerm)
	}
	err := query.Order("mi_sku ASC").Find(&items).Error
	return items, err
}

func (r *gormInventoryRepository) ListItemsPage(search, sort string, page, limit int) ([]entity.InventoryItem, int64, error) {
	var items []entity.InventoryItem
	query := r.db.Model(&entity.InventoryItem{})

	if search != "" {
		searchTerm := "%" + search + "%"
		query = query.Where(
			"mi_sku ILIKE ? OR title ILIKE ? OR id IN (SELECT inventory_item_id FROM inventory_mappings WHERE external_sku ILIKE ?)",
			searchTerm, searchTerm, searchTerm,
		)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderBy := map[string]string{
		"mi-sku-asc":  "mi_sku ASC",
		"mi-sku-desc": "mi_sku DESC",
		"name-asc":    "title ASC",
		"stock-desc":  "current_stock DESC",
		"stock-asc":   "current_stock ASC",
	}[sort]
	if orderBy == "" {
		orderBy = "mi_sku ASC"
	}

	err := query.Preload("Mappings").Order(orderBy).Offset((page - 1) * limit).Limit(limit).Find(&items).Error
	return items, total, err
}

func (r *gormInventoryRepository) GetItemByID(id int) (entity.InventoryItem, error) {
	var item entity.InventoryItem
	err := r.db.Preload("Mappings").First(&item, id).Error
	return item, err
}

func (r *gormInventoryRepository) GetItemsByIDs(ids []int) ([]entity.InventoryItem, error) {
	var items []entity.InventoryItem
	if len(ids) == 0 {
		return items, nil
	}
	err := r.db.Preload("Mappings").Where("id IN ?", ids).Find(&items).Error
	return items, err
}

func (r *gormInventoryRepository) CreateItem(item *entity.InventoryItem) error {
	return r.db.Create(item).Error
}

func (r *gormInventoryRepository) UpdateItem(item *entity.InventoryItem) error {
	return r.db.Save(item).Error
}

func (r *gormInventoryRepository) AdjustStock(id int, delta int) error {
	return r.db.Model(&entity.InventoryItem{}).
		Where("id = ?", id).
		Update("current_stock", gorm.Expr("GREATEST(current_stock + ?, 0)", delta)).Error
}

func (r *gormInventoryRepository) UpdateStockCount(id int, val int) error {
	return r.db.Model(&entity.InventoryItem{}).
		Where("id = ?", id).
		Update("current_stock", val).Error
}

func (r *gormInventoryRepository) GetMaxMISKU() (string, error) {
	var sku string
	// Find the highest mi-XX using regex or simply by order since the format is fixed
	// We use the raw order to get the lexicographically largest SKU
	err := r.db.Model(&entity.InventoryItem{}).
		Where("mi_sku LIKE 'mi-%'").
		Order("mi_sku DESC").
		Limit(1).
		Pluck("mi_sku", &sku).Error

	if err == gorm.ErrRecordNotFound {
		return "", nil
	}
	return sku, err
}

func (r *gormInventoryRepository) ListMappings() ([]entity.InventoryMapping, error) {
	var mappings []entity.InventoryMapping
	err := r.db.Find(&mappings).Error
	return mappings, err
}

func (r *gormInventoryRepository) CreateMapping(mapping *entity.InventoryMapping) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Delete any existing mappings for this product and platform to maintain the one-SKU-per-platform rule
		if err := tx.Where("inventory_item_id = ? AND platform = ?", mapping.InventoryItemID, mapping.Platform).Delete(&entity.InventoryMapping{}).Error; err != nil {
			return err
		}

		// 2. Delete any existing mappings for this platform and SKU to prevent unique index collisions on re-assignments
		if err := tx.Where("platform = ? AND external_sku = ?", mapping.Platform, mapping.ExternalSKU).Delete(&entity.InventoryMapping{}).Error; err != nil {
			return err
		}

		// 3. Create the new mapping
		return tx.Create(mapping).Error
	})
}

func (r *gormInventoryRepository) DeleteMapping(id int) error {
	return r.db.Delete(&entity.InventoryMapping{}, id).Error
}

func (r *gormInventoryRepository) DeleteAll() error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("TRUNCATE TABLE inventory_logs RESTART IDENTITY").Error; err != nil {
			return err
		}
		if err := tx.Exec("TRUNCATE TABLE inventory_mappings RESTART IDENTITY").Error; err != nil {
			return err
		}
		return tx.Exec("TRUNCATE TABLE inventory_items RESTART IDENTITY CASCADE").Error
	})
}

func (r *gormInventoryRepository) BulkCreateItem(items []entity.InventoryItem) error {
	return r.db.Create(&items).Error
}

func (r *gormInventoryRepository) BulkUpdatePrices(items []entity.InventoryItem) error {
	if len(items) == 0 {
		return nil
	}
	return r.db.Omit(clause.Associations).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"price"}),
	}).Create(&items).Error
}

func (r *gormInventoryRepository) GetItemsBySKUs(skus []string) ([]entity.InventoryItem, error) {
	var items []entity.InventoryItem
	if len(skus) == 0 {
		return items, nil
	}
	err := r.db.Preload("Mappings").
		Where("mi_sku IN ? OR id IN (SELECT inventory_item_id FROM inventory_mappings WHERE external_sku IN ?)", skus, skus).
		Find(&items).Error
	return items, err
}

func (r *gormInventoryRepository) LogAdjustment(l *entity.InventoryLog) error {
	return r.db.Create(l).Error
}

func (r *gormInventoryRepository) GetLogsByItemID(itemID int) ([]entity.InventoryLog, error) {
	var logs []entity.InventoryLog
	err := r.db.Where("inventory_item_id = ?", itemID).Order("created_at DESC").Find(&logs).Error
	return logs, err
}

func (r *gormInventoryRepository) GetLogsByExternalOrderID(externalOrderID string) ([]entity.InventoryLog, error) {
	var logs []entity.InventoryLog
	err := r.db.Where("external_order_id = ?", externalOrderID).Order("created_at DESC").Find(&logs).Error
	return logs, err
}

func (r *gormInventoryRepository) GetItemByPlatformSKU(platform, externalSKU string) (entity.InventoryItem, error) {
	var item entity.InventoryItem
	err := r.db.Preload("Mappings").
		Joins("JOIN inventory_mappings ON inventory_mappings.inventory_item_id = inventory_items.id").
		Where("inventory_mappings.platform = ? AND inventory_mappings.external_sku = ?", platform, externalSKU).
		First(&item).Error
	return item, err
}

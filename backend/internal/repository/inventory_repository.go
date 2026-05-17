package repository

import (
	"mi-tech/internal/entity"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

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


func (r *gormInventoryRepository) GetItemByID(id int) (entity.InventoryItem, error) {
	var item entity.InventoryItem
	err := r.db.Preload("Mappings").First(&item, id).Error
	return item, err
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
	// Use OnConflict to handle case where SKU might already be mapped
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "platform"}, {Name: "external_sku"}},
		DoUpdates: clause.AssignmentColumns([]string{"inventory_item_id", "external_variant_id"}),
	}).Create(mapping).Error
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

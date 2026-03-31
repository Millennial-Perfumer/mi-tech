package repository

import (
	"fmt"

	"mi-tech/internal/entity"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// gormLineItemRepository is the GORM implementation of LineItemRepository.
type gormLineItemRepository struct {
	db *gorm.DB
}

// NewLineItemRepository creates a new GORM-backed LineItemRepository.
func NewLineItemRepository(db *gorm.DB) LineItemRepository {
	return &gormLineItemRepository{db: db}
}

func (r *gormLineItemRepository) GetByOrderID(orderID int64) ([]entity.LineItem, error) {
	var items []entity.LineItem
	if err := r.db.Where("order_id = ?", orderID).Find(&items).Error; err != nil {
		return nil, fmt.Errorf("failed to query line items: %w", err)
	}
	return items, nil
}

func (r *gormLineItemRepository) UpsertBatch(tx *gorm.DB, orderID int64, items []entity.LineItem) error {
	if len(items) == 0 {
		return nil
	}

	// Optimization: Batch create line items with conflict handling in a single database roundtrip.
	// We iterate by index to update the original slice elements with the orderID.
	for i := range items {
		items[i].OrderID = orderID
	}

	if err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"title", "sku", "hs_code", "quantity", "price", "discount", "order_discount"}),
	}).Create(&items).Error; err != nil {
		return fmt.Errorf("failed to batch upsert line items for order %d: %w", orderID, err)
	}
	return nil
}

func (r *gormLineItemRepository) DeleteByOrderID(tx *gorm.DB, orderID int64) error {
	return tx.Where("order_id = ?", orderID).Delete(&entity.LineItem{}).Error
}

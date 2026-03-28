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

// UpsertBatch performs a batch upsert of line items for a specific order.
// Optimization: Replaces iterative inserts with a single GORM batch create (O(1) roundtrip).
func (r *gormLineItemRepository) UpsertBatch(tx *gorm.DB, orderID int64, items []entity.LineItem) error {
	if len(items) == 0 {
		return nil
	}

	// Set the correct OrderID for all items in the slice
	for i := range items {
		items[i].OrderID = orderID
	}

	if err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"title", "sku", "hs_code", "quantity", "price", "discount"}),
	}).Create(&items).Error; err != nil {
		return fmt.Errorf("failed to batch upsert line items: %w", err)
	}
	return nil
}

func (r *gormLineItemRepository) DeleteByOrderID(tx *gorm.DB, orderID int64) error {
	return tx.Where("order_id = ?", orderID).Delete(&entity.LineItem{}).Error
}

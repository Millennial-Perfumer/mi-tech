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

func (r *gormLineItemRepository) GetByOrderID(orderID string) ([]entity.LineItem, error) {
	var items []entity.LineItem
	if err := r.db.Where("order_id = ?", orderID).Find(&items).Error; err != nil {
		return nil, fmt.Errorf("failed to query line items: %w", err)
	}
	return items, nil
}

func (r *gormLineItemRepository) UpsertBatch(tx *gorm.DB, orderID string, items []entity.LineItem) error {
	for _, item := range items {
		item.OrderID = orderID
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"title", "sku", "hs_code", "quantity", "price", "discount"}),
		}).Create(&item).Error; err != nil {
			return fmt.Errorf("failed to upsert line item %s: %w", item.ID, err)
		}
	}
	return nil
}

func (r *gormLineItemRepository) DeleteByOrderID(tx *gorm.DB, orderID string) error {
	return tx.Where("order_id = ?", orderID).Delete(&entity.LineItem{}).Error
}

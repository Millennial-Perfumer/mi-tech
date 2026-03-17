package repository

import (
	"fmt"

	"mi-tech/internal/entity"

	"gorm.io/gorm"
)

// gormWebhookEventRepository is the GORM implementation of WebhookEventRepository.
type gormWebhookEventRepository struct {
	db *gorm.DB
}

// NewWebhookEventRepository creates a new GORM-backed WebhookEventRepository.
func NewWebhookEventRepository(db *gorm.DB) WebhookEventRepository {
	return &gormWebhookEventRepository{db: db}
}

func (r *gormWebhookEventRepository) Save(event *entity.WebhookEvent) error {
	if err := r.db.Create(event).Error; err != nil {
		return fmt.Errorf("failed to save webhook event: %w", err)
	}
	return nil
}

func (r *gormWebhookEventRepository) IsProcessed(deliveryID string) (bool, error) {
	var count int64
	err := r.db.Model(&entity.WebhookEvent{}).Where("webhook_delivery_id = ?", deliveryID).Count(&count).Error
	return count > 0, err
}

func (r *gormWebhookEventRepository) LinkToOrder(deliveryID string, orderID int64) error {
	result := r.db.Model(&entity.WebhookEvent{}).
		Where("webhook_delivery_id = ?", deliveryID).
		Updates(map[string]interface{}{
			"order_id":  orderID,
			"processed": true,
		})
	if result.Error != nil {
		return fmt.Errorf("failed to link webhook to order: %w", result.Error)
	}
	return nil
}

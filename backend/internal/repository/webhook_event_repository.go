package repository

import (
	"database/sql"
	"fmt"

	"shopify-gst-app/internal/entity"
)

// pgWebhookEventRepository is the PostgreSQL implementation of WebhookEventRepository.
type pgWebhookEventRepository struct {
	db *sql.DB
}

// NewWebhookEventRepository creates a new PostgreSQL-backed WebhookEventRepository.
func NewWebhookEventRepository(db *sql.DB) WebhookEventRepository {
	return &pgWebhookEventRepository{db: db}
}

func (r *pgWebhookEventRepository) Save(event *entity.WebhookEvent) error {
	query := `
		INSERT INTO webhook_events (source_id, topic, external_id, webhook_delivery_id, payload)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`
	return r.db.QueryRow(query,
		event.SourceID, event.Topic, event.ExternalID, event.WebhookDeliveryID, event.Payload,
	).Scan(&event.ID, &event.CreatedAt)
}

func (r *pgWebhookEventRepository) IsProcessed(deliveryID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM webhook_events WHERE webhook_delivery_id = $1)`
	err := r.db.QueryRow(query, deliveryID).Scan(&exists)
	return exists, err
}

func (r *pgWebhookEventRepository) LinkToOrder(deliveryID string, orderID string) error {
	query := `UPDATE webhook_events SET order_id = $1, processed = true WHERE webhook_delivery_id = $2`
	_, err := r.db.Exec(query, orderID, deliveryID)
	if err != nil {
		return fmt.Errorf("failed to link webhook to order: %w", err)
	}
	return nil
}

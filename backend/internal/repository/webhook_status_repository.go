package repository

import (
	"database/sql"
)

// pgWebhookStatusRepository is the PostgreSQL implementation of WebhookStatusRepository.
type pgWebhookStatusRepository struct {
	db *sql.DB
}

// NewWebhookStatusRepository creates a new PostgreSQL-backed WebhookStatusRepository.
func NewWebhookStatusRepository(db *sql.DB) WebhookStatusRepository {
	return &pgWebhookStatusRepository{db: db}
}

func (r *pgWebhookStatusRepository) Get() (topic string, status string, lastReceived string, err error) {
	err = r.db.QueryRow("SELECT topic, status, last_received FROM webhook_status WHERE id = 1").
		Scan(&topic, &status, &lastReceived)
	return
}

func (r *pgWebhookStatusRepository) UpdateActivity(topic string) error {
	_, err := r.db.Exec(`
		UPDATE webhook_status 
		SET topic = $1, status = 'active', last_received = CURRENT_TIMESTAMP 
		WHERE id = 1
	`, topic)
	return err
}

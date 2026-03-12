package repository

import (
	"gorm.io/gorm"
)

// gormWebhookStatusRepository is the GORM implementation of WebhookStatusRepository.
type gormWebhookStatusRepository struct {
	db *gorm.DB
}

// NewWebhookStatusRepository creates a new GORM-backed WebhookStatusRepository.
func NewWebhookStatusRepository(db *gorm.DB) WebhookStatusRepository {
	return &gormWebhookStatusRepository{db: db}
}

type webhookStatusRow struct {
	Topic        string
	Status       string
	LastReceived string `gorm:"column:last_received"`
}

func (webhookStatusRow) TableName() string { return "webhook_status" }

func (r *gormWebhookStatusRepository) Get() (topic string, status string, lastReceived string, err error) {
	var row webhookStatusRow
	err = r.db.Where("id = ?", 1).First(&row).Error
	return row.Topic, row.Status, row.LastReceived, err
}

func (r *gormWebhookStatusRepository) UpdateActivity(topic string) error {
	return r.db.Exec(
		"UPDATE webhook_status SET topic = ?, status = 'active', last_received = CURRENT_TIMESTAMP WHERE id = 1",
		topic,
	).Error
}

package repository

import (
	"strings"
	"time"

	"mi-tech/internal/shared/config/entity"

	"gorm.io/gorm"
)

// gormConfigsRepository is the GORM implementation of ConfigsRepository.
type gormConfigsRepository struct {
	db *gorm.DB
}

// NewConfigsRepository creates a new ConfigsRepository.
func NewConfigsRepository(db *gorm.DB) ConfigsRepository {
	return &gormConfigsRepository{db: db}
}

// GetAll returns all configs. Secret values are masked.
func (r *gormConfigsRepository) GetAll() ([]entity.AppConfig, error) {
	var configs []entity.AppConfig
	if err := r.db.Order("category, sort_order").Find(&configs).Error; err != nil {
		return nil, err
	}

	// Mask secret values
	for i := range configs {
		if configs[i].IsSecret && configs[i].Value != "" {
			configs[i].Value = maskValue(configs[i].Value)
		}
	}
	return configs, nil
}

// GetAllRevealed returns all configs with values unmasked.
func (r *gormConfigsRepository) GetAllRevealed() ([]entity.AppConfig, error) {
	var configs []entity.AppConfig
	if err := r.db.Order("category, sort_order").Find(&configs).Error; err != nil {
		return nil, err
	}
	return configs, nil
}

// Get retrieves a single config value by key (always unmasked for internal use).
func (r *gormConfigsRepository) Get(key string) (string, error) {
	var config entity.AppConfig
	err := r.db.Where("key = ?", key).First(&config).Error
	if err != nil {
		return "", err
	}
	return config.Value, nil
}

// Set upserts a config value.
func (r *gormConfigsRepository) Set(key, value string) error {
	return r.db.Exec(
		`INSERT INTO app_configs (key, value, updated_at) VALUES (?, ?, ?)
		 ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at`,
		key, value, time.Now(),
	).Error
}

// maskValue shows only the last 4 characters behind dots.
func maskValue(val string) string {
	if len(val) <= 4 {
		return strings.Repeat("•", len(val))
	}
	return strings.Repeat("•", 8) + val[len(val)-4:]
}

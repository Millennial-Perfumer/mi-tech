package repository

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

// AppConfig represents a row in the app_configs table.
type AppConfig struct {
	Key       string    `gorm:"column:key;primaryKey" json:"key"`
	Value     string    `gorm:"column:value" json:"value"`
	IsSecret  bool      `gorm:"column:is_secret" json:"is_secret"`
	Label     string    `gorm:"column:label" json:"label"`
	Category  string    `gorm:"column:category" json:"category"`
	SortOrder int       `gorm:"column:sort_order" json:"sort_order"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (AppConfig) TableName() string { return "app_configs" }

// ConfigsRepository handles app_configs CRUD.
type ConfigsRepository struct {
	db *gorm.DB
}

// NewConfigsRepository creates a new ConfigsRepository.
func NewConfigsRepository(db *gorm.DB) *ConfigsRepository {
	return &ConfigsRepository{db: db}
}

// GetAll returns all configs. Secret values are masked.
func (r *ConfigsRepository) GetAll() ([]AppConfig, error) {
	var configs []AppConfig
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
func (r *ConfigsRepository) GetAllRevealed() ([]AppConfig, error) {
	var configs []AppConfig
	if err := r.db.Order("category, sort_order").Find(&configs).Error; err != nil {
		return nil, err
	}
	return configs, nil
}

// Get retrieves a single config value by key (always unmasked for internal use).
func (r *ConfigsRepository) Get(key string) (string, error) {
	var config AppConfig
	err := r.db.Where("key = ?", key).First(&config).Error
	if err != nil {
		return "", err
	}
	return config.Value, nil
}

// Set upserts a config value.
func (r *ConfigsRepository) Set(key, value string) error {
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

package repository

import (
	"time"

	"mi-tech/internal/shared/config/entity"

	"gorm.io/gorm"
)

// gormSettingsRepository is the GORM implementation of SettingsRepository.
type gormSettingsRepository struct {
	db *gorm.DB
}

// NewSettingsRepository creates a new SettingsRepository.
func NewSettingsRepository(db *gorm.DB) SettingsRepository {
	return &gormSettingsRepository{db: db}
}

// Get retrieves a setting value by key.
func (r *gormSettingsRepository) Get(key string) (string, error) {
	var setting entity.AppSetting
	err := r.db.Where("key = ?", key).First(&setting).Error
	if err != nil {
		return "", err
	}
	return setting.Value, nil
}

// Set upserts a setting value.
func (r *gormSettingsRepository) Set(key, value string) error {
	return r.db.Exec(
		"INSERT INTO app_settings (key, value, updated_at) VALUES (?, ?, ?) ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at",
		key, value, time.Now(),
	).Error
}

// GetDateRange returns the persisted date range.
func (r *gormSettingsRepository) GetDateRange() (startDate, endDate string, err error) {
	startDate, _ = r.Get("date_range_start")
	endDate, _ = r.Get("date_range_end")
	return startDate, endDate, nil
}

// SetDateRange persists the date range.
func (r *gormSettingsRepository) SetDateRange(startDate, endDate string) error {
	if err := r.Set("date_range_start", startDate); err != nil {
		return err
	}
	return r.Set("date_range_end", endDate)
}

// GetAll retrieves all settings.
func (r *gormSettingsRepository) GetAll(settings *[]entity.AppSetting) error {
	return r.db.Find(settings).Error
}

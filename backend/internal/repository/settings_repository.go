package repository

import (
	"time"

	"gorm.io/gorm"
)

// AppSetting represents a row in the app_settings table.
type AppSetting struct {
	Key       string    `gorm:"column:key;primaryKey"`
	Value     string    `gorm:"column:value"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (AppSetting) TableName() string { return "app_settings" }

// SettingsRepository handles app_settings CRUD.
type SettingsRepository struct {
	db *gorm.DB
}

// NewSettingsRepository creates a new SettingsRepository.
func NewSettingsRepository(db *gorm.DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

// Get retrieves a setting value by key.
func (r *SettingsRepository) Get(key string) (string, error) {
	var setting AppSetting
	err := r.db.Where("key = ?", key).First(&setting).Error
	if err != nil {
		return "", err
	}
	return setting.Value, nil
}

// Set upserts a setting value.
func (r *SettingsRepository) Set(key, value string) error {
	return r.db.Exec(
		"INSERT INTO app_settings (key, value, updated_at) VALUES (?, ?, ?) ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at",
		key, value, time.Now(),
	).Error
}

// GetDateRange returns the persisted date range.
func (r *SettingsRepository) GetDateRange() (startDate, endDate string, err error) {
	startDate, _ = r.Get("date_range_start")
	endDate, _ = r.Get("date_range_end")
	return startDate, endDate, nil
}

// SetDateRange persists the date range.
func (r *SettingsRepository) SetDateRange(startDate, endDate string) error {
	if err := r.Set("date_range_start", startDate); err != nil {
		return err
	}
	return r.Set("date_range_end", endDate)
}

package repository

import (
	"mi-tech/internal/shared/config/entity"
)

// ConfigsRepository defines database operations for configuration keys.
type ConfigsRepository interface {
	GetAll() ([]entity.AppConfig, error)
	GetAllRevealed() ([]entity.AppConfig, error)
	Get(key string) (string, error)
	Set(key, value string) error
}

// SettingsRepository defines database operations for application settings.
type SettingsRepository interface {
	Get(key string) (string, error)
	Set(key, value string) error
	GetDateRange() (startDate, endDate string, err error)
	SetDateRange(startDate, endDate string) error
	GetAll(settings *[]entity.AppSetting) error
}

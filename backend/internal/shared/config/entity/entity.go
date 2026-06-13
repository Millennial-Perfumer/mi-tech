package entity

import (
	"time"
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

// TableName matches the DB table name for AppConfig.
func (AppConfig) TableName() string { return "app_configs" }

// AppSetting represents a row in the app_settings table.
type AppSetting struct {
	Key       string    `gorm:"column:key;primaryKey"`
	Value     string    `gorm:"column:value"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

// TableName matches the DB table name for AppSetting.
func (AppSetting) TableName() string { return "app_settings" }

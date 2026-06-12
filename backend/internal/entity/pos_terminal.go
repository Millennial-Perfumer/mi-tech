package entity

import "time"

// POSTerminal represents a point-of-sale terminal used to generate manual orders.
type POSTerminal struct {
	ID           int       `gorm:"primaryKey" json:"id"`
	Code         string    `gorm:"column:code;uniqueIndex" json:"code"`
	Name         string    `gorm:"column:name" json:"name"`
	NextSequence int       `gorm:"column:next_sequence" json:"next_sequence"`
	IsActive     bool      `gorm:"column:is_active" json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}

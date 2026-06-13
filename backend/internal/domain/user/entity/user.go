package entity

import "time"

// User represents a user in the system.
type User struct {
	ID               uint       `gorm:"primaryKey" json:"id"`
	Username         string     `gorm:"uniqueIndex;not null" json:"username"`
	PasswordHash     string     `gorm:"not null" json:"-"`
	Role             string     `gorm:"not null;default:'read'" json:"role"`
	PhoneNumber      string     `json:"phone_number"`
	TwoFactorEnabled bool       `gorm:"default:true" json:"two_factor_enabled"`
	OTPCode          string     `json:"-"`
	OTPExpiry        *time.Time `json:"-"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

package entity

import (
	"time"
)

// B2BPaymentTerm represents a configurable payment term option
type B2BPaymentTerm struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"column:name;uniqueIndex" json:"name"`
	DueDays   int       `gorm:"column:due_days" json:"due_days"`
	CreatedAt time.Time `gorm:"column:created_at;default:NOW()" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;default:NOW()" json:"updated_at"`
}

func (B2BPaymentTerm) TableName() string {
	return "b2b_payment_terms"
}

package entity

import (
	"time"
)

// GSTPeriod represents a monthly GST filing lock boundary
type GSTPeriod struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Month     int       `gorm:"column:month" json:"month"`
	Year      int       `gorm:"column:year" json:"year"`
	Status    string    `gorm:"column:status;default:OPEN" json:"status"` // 'OPEN', 'LOCKED'
	CreatedAt time.Time `gorm:"column:created_at;default:NOW()" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;default:NOW()" json:"updated_at"`
}

func (GSTPeriod) TableName() string {
	return "gst_periods"
}

// B2BFinancialAuditLog logs all financial changes
type B2BFinancialAuditLog struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Action      string    `gorm:"column:action" json:"action"`
	EntityType  string    `gorm:"column:entity_type" json:"entity_type"` // INVOICE, CREDIT_NOTE, DEBIT_NOTE, PAYMENT
	EntityID    int64     `gorm:"column:entity_id" json:"entity_id"`
	UserID      string    `gorm:"column:user_id" json:"user_id"`
	Description string    `gorm:"column:description" json:"description"`
	OldValue    string    `gorm:"column:old_value" json:"old_value"`
	NewValue    string    `gorm:"column:new_value" json:"new_value"`
	CreatedAt   time.Time `gorm:"column:created_at;default:NOW()" json:"created_at"`
}

func (B2BFinancialAuditLog) TableName() string {
	return "b2b_financial_audit_logs"
}

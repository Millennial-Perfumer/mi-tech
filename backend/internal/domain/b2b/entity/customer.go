package entity

import (
	"time"
)

// B2BCustomer represents a client registered for B2B billing
type B2BCustomer struct {
	ID              int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	LegalName       string    `gorm:"column:legal_name" json:"legal_name"`
	TradeName       *string   `gorm:"column:trade_name" json:"trade_name"`
	GSTIN           string    `gorm:"column:gstin;uniqueIndex" json:"gstin"`
	PAN             *string   `gorm:"column:pan" json:"pan"`
	Email           *string   `gorm:"column:email" json:"email"`
	Phone           *string   `gorm:"column:phone" json:"phone"`
	BillingAddress  string    `gorm:"column:billing_address" json:"billing_address"`
	ShippingAddress *string   `gorm:"column:shipping_address" json:"shipping_address"`
	State           string    `gorm:"column:state" json:"state"`
	StateCode       string    `gorm:"column:state_code" json:"state_code"`
	Notes           *string   `gorm:"column:notes" json:"notes"`
	CreatedAt       time.Time `gorm:"column:created_at;default:NOW()" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at;default:NOW()" json:"updated_at"`
}

func (B2BCustomer) TableName() string {
	return "b2b_customers"
}

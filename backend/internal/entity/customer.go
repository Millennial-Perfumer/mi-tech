package entity

import (
	"time"

	"gorm.io/gorm"
)

// Customer represents a row in the "customers" table.
type Customer struct {
	ID           int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	PhoneNumber  string    `gorm:"column:phone_number;unique;not null" json:"phone_number"`
	FirstName    *string   `gorm:"column:first_name" json:"first_name"`
	LastName     *string   `gorm:"column:last_name" json:"last_name"`
	Email        *string   `gorm:"column:email" json:"email"`
	Address1     *string   `gorm:"column:address1" json:"address1"`
	Address2     *string   `gorm:"column:address2" json:"address2"`
	City         *string   `gorm:"column:city" json:"city"`
	State        *string   `gorm:"column:state" json:"state"`
	Country      *string   `gorm:"column:country" json:"country"`
	ZipCode      *string   `gorm:"column:zip_code" json:"zip_code"`
	TotalOrders  int       `gorm:"column:total_orders;default:0" json:"total_orders"`
	TotalSpent   float64   `gorm:"column:total_spent;default:0" json:"total_spent"`
	SourceID     string    `gorm:"column:source_id" json:"source_id"`
	ExternalID   *string   `gorm:"column:external_id" json:"external_id"`
	CreatedAt    time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Customer) TableName() string { return "customers" }

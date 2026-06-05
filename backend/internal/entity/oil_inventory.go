package entity

import (
	"time"
)

// Supplier represents a vendor providing raw materials.
type Supplier struct {
	ID          int       `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"column:name" json:"name"`
	ContactInfo string    `gorm:"column:contact_info" json:"contact_info"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	OilsCount   int       `gorm:"-" json:"oils_count"`
}

func (Supplier) TableName() string {
	return "suppliers"
}

// OilInventory represents the stock level and pricing of perfume oils.
type OilInventory struct {
	ID                 int       `gorm:"primaryKey" json:"id"`
	Name               string    `gorm:"column:name" json:"name"`
	InventoryItemID    *int      `gorm:"column:inventory_item_id" json:"inventory_item_id"`
	PurchasePricePerKg *float64  `gorm:"column:purchase_price_per_kg" json:"purchase_price_per_kg"`
	GramsLeft          *float64  `gorm:"column:grams_left" json:"grams_left"`
	SupplierID         *int      `gorm:"column:supplier_id" json:"supplier_id"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`

	// Relationships
	InventoryItem *InventoryItem `gorm:"foreignKey:InventoryItemID" json:"inventory_item,omitempty"`
	Supplier      *Supplier      `gorm:"foreignKey:SupplierID" json:"supplier,omitempty"`
}

func (OilInventory) TableName() string {
	return "oil_inventory"
}

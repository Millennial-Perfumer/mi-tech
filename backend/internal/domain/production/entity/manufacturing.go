package entity

import (
	"time"

	inventoryEntity "mi-tech/internal/domain/inventory/entity"
)

// PurchaseOrder represents a record of buying raw materials.
type PurchaseOrder struct {
	ID             int       `gorm:"primaryKey" json:"id"`
	OilInventoryID int       `gorm:"column:oil_inventory_id" json:"oil_inventory_id"`
	SupplierID     int       `gorm:"column:supplier_id" json:"supplier_id"`
	QuantityGrams  float64   `gorm:"column:quantity_grams" json:"quantity_grams"`
	UnitPricePerKg float64   `gorm:"column:unit_price_per_kg" json:"unit_price_per_kg"`
	TotalPrice     float64   `gorm:"column:total_price" json:"total_price"`
	PurchaseDate   time.Time `gorm:"column:purchase_date" json:"purchase_date"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// Relationships
	OilInventory *OilInventory `gorm:"foreignKey:OilInventoryID" json:"oil_inventory,omitempty"`
	Supplier     *Supplier     `gorm:"foreignKey:SupplierID" json:"supplier,omitempty"`
}

func (PurchaseOrder) TableName() string {
	return "purchase_orders"
}

// ManufacturingRecord represents a session where raw materials are converted to products.
type ManufacturingRecord struct {
	ID                int       `gorm:"primaryKey" json:"id"`
	ManufacturingDate time.Time `gorm:"column:manufacturing_date" json:"manufacturing_date"`
	Notes             string    `gorm:"column:notes" json:"notes"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`

	// Relationships
	Oils     []ManufacturingOil     `gorm:"foreignKey:ManufacturingRecordID" json:"oils"`
	Products []ManufacturingProduct `gorm:"foreignKey:ManufacturingRecordID" json:"products"`
}

func (ManufacturingRecord) TableName() string {
	return "manufacturing_records"
}

// ManufacturingOil links a manufacturing session to the fragrance oils used.
type ManufacturingOil struct {
	ID                    int       `gorm:"primaryKey" json:"id"`
	ManufacturingRecordID int       `gorm:"column:manufacturing_record_id" json:"manufacturing_record_id"`
	OilInventoryID        int       `gorm:"column:oil_inventory_id" json:"oil_inventory_id"`
	QuantityGrams         float64   `gorm:"column:quantity_grams" json:"quantity_grams"`
	DeductInventory       bool      `gorm:"column:deduct_inventory" json:"deduct_inventory"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`

	// Relationships
	OilInventory *OilInventory `gorm:"foreignKey:OilInventoryID" json:"oil_inventory,omitempty"`
}

func (ManufacturingOil) TableName() string {
	return "manufacturing_oils"
}

// ManufacturingProduct links a manufacturing session to the products produced.
type ManufacturingProduct struct {
	ID                    int       `gorm:"primaryKey" json:"id"`
	ManufacturingRecordID int       `gorm:"column:manufacturing_record_id" json:"manufacturing_record_id"`
	InventoryItemID       int       `gorm:"column:inventory_item_id" json:"inventory_item_id"`
	QuantityProduced      int       `gorm:"column:quantity_produced" json:"quantity_produced"`
	AddStock              bool      `gorm:"column:add_stock" json:"add_stock"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`

	// Relationships
	InventoryItem *inventoryEntity.InventoryItem `gorm:"foreignKey:InventoryItemID" json:"inventory_item,omitempty"`
}

func (ManufacturingProduct) TableName() string {
	return "manufacturing_products"
}

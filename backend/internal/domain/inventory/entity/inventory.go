package entity

import (
	"time"
)

// InventoryItem represents a physical product in the warehouse (Warehouse Authority).
type InventoryItem struct {
	ID            int       `gorm:"primaryKey" json:"id"`
	MISKU         string    `gorm:"column:mi_sku;uniqueIndex" json:"mi_sku"` // The canonical mi-XX SKU
	Title         string    `gorm:"column:title" json:"title"`
	Description   *string   `gorm:"column:description" json:"description"`
	Specification *string   `gorm:"column:specification" json:"specification"`
	CurrentStock  int       `gorm:"column:current_stock" json:"current_stock"`
	Price         float64   `gorm:"column:price" json:"price"`
	HSNCode       string    `gorm:"column:hsn_code" json:"hsn_code"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// Relationships
	Mappings []InventoryMapping `gorm:"foreignKey:InventoryItemID" json:"mappings,omitempty"`
	Logs     []InventoryLog     `gorm:"foreignKey:InventoryItemID" json:"logs,omitempty"`
}

// InventoryMapping maps external platform SKUs to our internal warehouse SKUs.
type InventoryMapping struct {
	ID                int       `gorm:"primaryKey" json:"id"`
	InventoryItemID   int       `gorm:"column:inventory_item_id" json:"inventory_item_id"`
	Platform          string    `gorm:"column:platform" json:"platform"` // 'shopify', 'amazon'
	ExternalSKU       string    `gorm:"column:external_sku;uniqueIndex:idx_platform_sku" json:"external_sku"`
	ExternalVariantID *string   `gorm:"column:external_variant_id" json:"external_variant_id"`
	CreatedAt         time.Time `json:"created_at"`
}

// InventoryLog tracks a specific change in stock levels.
type InventoryLog struct {
	ID              int       `gorm:"primaryKey" json:"id"`
	InventoryItemID int       `gorm:"column:inventory_item_id" json:"inventory_item_id"`
	OrderID         *int64    `gorm:"column:order_id" json:"order_id,omitempty"`
	CustomerID      *int64    `gorm:"column:customer_id" json:"customer_id,omitempty"`
	Delta           int       `gorm:"column:delta" json:"delta"`
	Reason          string    `gorm:"column:reason" json:"reason"` // sale, cancellation, return, lost, manual, correction
	Platform        string    `gorm:"column:platform" json:"platform"`
	ExternalOrderID *string   `gorm:"column:external_order_id" json:"external_order_id"`
	StockBefore     *int      `gorm:"column:stock_before" json:"stock_before,omitempty"`
	StockAfter      *int      `gorm:"column:stock_after" json:"stock_after,omitempty"`
	ActorType       *string   `gorm:"column:actor_type" json:"actor_type,omitempty"`
	ActorID         *string   `gorm:"column:actor_id" json:"actor_id,omitempty"`
	RequestID       *string   `gorm:"column:request_id" json:"request_id,omitempty"`
	IdempotencyKey  *string   `gorm:"column:idempotency_key" json:"idempotency_key,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

func (InventoryItem) TableName() string {
	return "inventory_items"
}

func (InventoryMapping) TableName() string {
	return "inventory_mappings"
}

func (InventoryLog) TableName() string {
	return "inventory_logs"
}

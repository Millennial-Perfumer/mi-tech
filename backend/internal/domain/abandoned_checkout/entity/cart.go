package entity

import (
	"encoding/json"
	"time"
)

type ShopifyCart struct {
	ID        int              `json:"id" gorm:"primaryKey;autoIncrement"`
	StoreID   string           `json:"store_id" gorm:"column:store_id"`
	CartToken string           `json:"cart_token" gorm:"column:cart_token;unique"`
	LineItems *json.RawMessage `json:"line_items" gorm:"column:line_items;type:jsonb"`
	CreatedAt time.Time        `json:"created_at" gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time        `json:"updated_at" gorm:"column:updated_at;default:CURRENT_TIMESTAMP"`
}

func (ShopifyCart) TableName() string {
	return "shopify_carts"
}

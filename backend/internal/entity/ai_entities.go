package entity

import (
	"encoding/json"
	"time"
)

type AIConversation struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	UserID    int64     `json:"user_id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AIMessage struct {
	ID             int64            `json:"id" gorm:"primaryKey"`
	ConversationID int64            `json:"conversation_id"`
	Role           string           `json:"role"`
	Content        string           `json:"content"`
	Metadata       *json.RawMessage `json:"metadata,omitempty"`
	CreatedAt      time.Time        `json:"created_at"`
}

// AI Aggregation DTOs for tool calling results

type AIRevenueSummary struct {
	TotalRevenue float64 `json:"total_revenue"`
	TotalOrders  int     `json:"total_orders"`
	StartDate    string  `json:"start_date"`
	EndDate      string  `json:"end_date"`
}

type AIChannelRevenue struct {
	Channel string  `json:"channel"`
	Revenue float64 `json:"revenue"`
	Orders  int     `json:"orders"`
}

type AIStateRevenue struct {
	State   string  `json:"state"`
	Revenue float64 `json:"revenue"`
	Orders  int     `json:"orders"`
}

type AIDailyRevenue struct {
	Date    string  `json:"date"`
	Revenue float64 `json:"revenue"`
}

type AIProductRank struct {
	SKU     string  `json:"sku"`
	Name    string  `json:"name"`
	QtySold int     `json:"qty_sold"`
	Revenue float64 `json:"revenue"`
}

type AIProductStats struct {
	SKU            string  `json:"sku"`
	Name           string  `json:"name"`
	CurrentStock   int     `json:"current_stock"`
	AverageDaily   float64 `json:"average_daily_sales"`
	TotalSold      int     `json:"total_sold"`
	InventoryValue float64 `json:"inventory_value"`
}

type AICustomerSegments struct {
	TotalCustomers int            `json:"total_customers"`
	NewCustomers   int            `json:"new_customers"` // last 30 days
	RepeatRate     float64        `json:"repeat_rate"`
	ChannelSplit   map[string]int `json:"channel_split"`
}

type AITopCustomer struct {
	Name       string  `json:"name"`
	Phone      string  `json:"phone"`
	TotalSpend float64 `json:"total_spend"`
	OrderCount int     `json:"order_count"`
}

type AIInventoryStatus struct {
	SKU           string `json:"sku"`
	Name          string `json:"name"`
	Stock         int    `json:"stock"`
	Specification string `json:"specification"`
	Status        string `json:"status"` // "In Stock", "Low Stock", "Out of Stock"
}

type AIBusinessSnapshot struct {
	MTDRevenue    float64 `json:"mtd_revenue"`
	MTDOrders     int     `json:"mtd_orders"`
	TodayRevenue  float64 `json:"today_revenue"`
	TodayOrders   int     `json:"today_orders"`
	LowStockCount int     `json:"low_stock_count"`
	PendingOrders int     `json:"pending_orders"`
}

type AIMemory struct {
	ID        int64            `json:"id" gorm:"primaryKey"`
	Key       string           `json:"key" gorm:"unique;not null"`
	Content   string           `json:"content" gorm:"not null"`
	Category  string           `json:"category" gorm:"default:'general'"`
	Metadata  *json.RawMessage `json:"metadata,omitempty" gorm:"type:jsonb"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

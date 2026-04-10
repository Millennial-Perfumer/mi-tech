package dto

import "time"

// FeedbackResponse represents the enriched data for the Customer Sentiment dashboard
type FeedbackResponse struct {
	ID           int       `gorm:"column:id" json:"id"`
	OrderID      int64     `gorm:"column:order_id" json:"order_id"`
	OrderNumber  string    `gorm:"column:order_number" json:"order_number"`
	CustomerName string    `gorm:"column:customer_name" json:"customer_name"`
	Rating       int       `gorm:"column:rating" json:"rating"`
	Comment      string    `gorm:"column:comment" json:"comment"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
}

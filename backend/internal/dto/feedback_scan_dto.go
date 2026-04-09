package dto

import "time"

// FeedbackScanResult represents an order ready for manual feedback triggering
type FeedbackScanResult struct {
	ID            int64     `json:"id"`
	OrderNumber   string    `json:"order_number"`
	CustomerName  string    `json:"customer_name"`
	CustomerPhone string    `json:"customer_phone"`
	DeliveredAt   time.Time `json:"delivered_at"`
	FeedbackURL   string    `json:"feedback_url"`
}

package dto

import "time"

// FeedbackResponse represents the enriched data for the Customer Sentiment dashboard
type FeedbackResponse struct {
	ID           int       `gorm:"column:id" json:"id"`
	OrderID      int64     `gorm:"column:order_id" json:"order_id"`
	OrderNumber  string    `gorm:"column:order_number" json:"order_number"`
	CustomerName  string    `gorm:"column:customer_name" json:"customer_name"`
	CustomerPhone string    `gorm:"column:customer_phone" json:"customer_phone"`
	Rating       int       `gorm:"column:rating" json:"rating"`
	Comment      string    `gorm:"column:comment" json:"message"`
	AdminComment    *string    `gorm:"column:admin_comment" json:"admin_comment"`
	JudgeMePosted           bool       `gorm:"column:judgeme_posted" json:"judgeme_posted"`
	JudgeMePostedAt         *time.Time `gorm:"column:judgeme_posted_at" json:"judgeme_posted_at"`
	GoogleReviewRequested   bool       `gorm:"column:google_review_requested" json:"google_review_requested"`
	GoogleReviewRequestedAt *time.Time `gorm:"column:google_review_requested_at" json:"google_review_requested_at,omitempty"`
	CreatedAt               time.Time  `gorm:"column:created_at" json:"created_at"`
}

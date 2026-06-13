package entity

import "time"

// FeedbackStatus represents a state in the feedback lifecycle
type FeedbackStatus struct {
	ID   int    `gorm:"column:id;primaryKey"`
	Name string `gorm:"column:name"`
}

func (FeedbackStatus) TableName() string { return "feedback_statuses" }

// CustomerFeedback stores the actual rating and message from a customer
type CustomerFeedback struct {
	ID            int       `gorm:"column:id;primaryKey;autoIncrement"`
	OrderID       int64     `gorm:"column:order_id"`
	CustomerPhone string    `gorm:"column:customer_phone"`
	Rating        int       `gorm:"column:rating"`
	Message       string    `gorm:"column:message"`
	AdminComment  *string   `gorm:"column:admin_comment"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

func (CustomerFeedback) TableName() string { return "customer_feedback" }

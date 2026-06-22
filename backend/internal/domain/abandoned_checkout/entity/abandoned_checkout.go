package entity

import (
	"encoding/json"
	"time"
)

type AbandonedCheckout struct {
	ID                    int              `json:"id" gorm:"primaryKey;autoIncrement"`
	StoreID               string           `json:"store_id" gorm:"column:store_id"`
	CheckoutID            string           `json:"checkout_id" gorm:"column:checkout_id"`
	CheckoutToken         string           `json:"checkout_token" gorm:"column:checkout_token;unique"`
	CartToken             string           `json:"cart_token" gorm:"column:cart_token"`
	Email                 string           `json:"email" gorm:"column:email"`
	Phone                 string           `json:"phone" gorm:"column:phone"`
	CustomerName          string           `json:"customer_name" gorm:"column:customer_name"`
	CheckoutURL           string           `json:"checkout_url" gorm:"column:checkout_url"`
	LineItems             *json.RawMessage `json:"line_items" gorm:"column:line_items;type:jsonb"`
	TotalPrice            float64          `json:"total_price" gorm:"column:total_price"`
	Currency              string           `json:"currency" gorm:"column:currency"`
	Completed             bool             `json:"completed" gorm:"column:completed;default:false"`
	CompletedAt           *time.Time       `json:"completed_at" gorm:"column:completed_at"`
	OrderID               string           `json:"order_id" gorm:"column:order_id"`
	RecoveryStatus        string           `json:"recovery_status" gorm:"column:recovery_status;default:PENDING"` // PENDING, PROCESSING, SENT, FAILED, CANCELLED
	RecoveryAttempts      int              `json:"recovery_attempts" gorm:"column:recovery_attempts;default:0"`
	RecoveryMessageSentAt *time.Time       `json:"recovery_message_sent_at" gorm:"column:recovery_message_sent_at"`
	LastError             string           `json:"last_error" gorm:"column:last_error"`
	MarketingConsent      bool             `json:"marketing_consent" gorm:"column:marketing_consent;default:false"`
	SMSConsent            bool             `json:"sms_consent" gorm:"column:sms_consent;default:false"`
	AbandonedAt           time.Time        `json:"abandoned_at" gorm:"column:abandoned_at;default:CURRENT_TIMESTAMP"`
	CreatedAt             time.Time        `json:"created_at" gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt             time.Time        `json:"updated_at" gorm:"column:updated_at;default:CURRENT_TIMESTAMP"`
}

package entity

import (
	"encoding/json"
	"time"
)

type Conversation struct {
	ID            int       `json:"id"`
	PhoneNumber   string    `json:"phone_number"`
	ContactName   string    `json:"contact_name"`
	LastMessage   string    `json:"last_message"`
	LastMessageAt time.Time `json:"last_message_at"`
	Mode          string    `json:"mode"`
	ActiveTaskID  *int      `json:"active_task_id,omitempty"`
	Priority      string    `json:"priority"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type ChatMessage struct {
	ID             int             `json:"id"`
	ConversationID int             `json:"conversation_id"`
	MessageID      string          `json:"message_id"`
	Text           string          `json:"text"`
	Type           string          `json:"type"`
	Direction      string          `json:"direction"`
	SenderRole     string          `json:"sender_role"`
	Status         string          `json:"status"`
	IsIssue        bool            `json:"is_issue"`
	Priority       string          `json:"priority"`
	SentAt         time.Time       `json:"sent_at"`
	Metadata       json.RawMessage `json:"metadata,omitempty"`
}

type AutomationMessage struct {
	ID           int        `json:"id"`
	StoreID      string     `json:"store_id"`
	TemplateID   int        `json:"template_id"`
	TemplateName string     `json:"template_name"`
	OrderID      int64      `json:"order_id"`
	OrderNumber  string     `json:"order_number"`
	CustomerName string     `json:"customer_name"`
	PhoneNumber  string     `json:"phone_number"`
	MessageID    string     `json:"message_id"`
	Status       string     `json:"status"`
	SentAt       time.Time  `json:"sent_at"`
	DeliveredAt  *time.Time `json:"delivered_at"`
	ReadAt       *time.Time `json:"read_at"`
	ErrorMessage string     `json:"error_message"`
}

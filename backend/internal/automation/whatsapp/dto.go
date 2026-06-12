package whatsapp

import (
	"encoding/json"
	"time"
)

type TemplateHeader struct {
	Type     string          `json:"type"` // text, image, video, document, location, none
	Text     string          `json:"text,omitempty"`
	Sample   json.RawMessage `json:"sample,omitempty"`
	Location json.RawMessage `json:"location,omitempty"`
}

type TemplateButton struct {
	Type        string `json:"type"` // custom, visit_website, call_whatsapp, call_phone, flow, copy_code
	Text        string `json:"text"`
	Payload     string `json:"payload,omitempty"`
	URL         string `json:"url,omitempty"`
	PhoneNumber string `json:"phoneNumber,omitempty"` // Frontend camelCase
	FlowID      string `json:"flowID,omitempty"`      // Frontend camelCase
	FlowName    string `json:"flowName,omitempty"`    // Frontend camelCase
	OfferCode   string `json:"offerCode,omitempty"`   // Frontend camelCase
}

type CreateTemplateRequest struct {
	Name     string           `json:"name"`
	Language string           `json:"language"`
	Category string           `json:"category"`
	Header   *TemplateHeader  `json:"header,omitempty"`
	Body     string           `json:"body"`
	Footer   string           `json:"footer,omitempty"`
	Buttons  []TemplateButton `json:"buttons,omitempty"`
	Examples string           `json:"examples,omitempty"`
}

type Trigger struct {
	ID             int       `json:"id"`
	StoreID        string    `json:"store_id"`
	WebhookTopic   string    `json:"webhook_topic"`
	TemplateID     int       `json:"template_id"`
	Enabled        bool      `json:"enabled"`
	TemplateName   string    `json:"template_name,omitempty"`   // Joined field
	TemplateBody   string    `json:"template_body,omitempty"`   // Joined field
	TemplateStatus string    `json:"template_status,omitempty"` // Joined field
	CreatedAt      time.Time `json:"created_at"`
}

type AutomationEvent struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Topic       string    `json:"topic"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type AutomationTemplate struct {
	ID               int              `json:"id"`
	StoreID          string           `json:"store_id"`
	TemplateName     string           `json:"template_name"`
	Language         string           `json:"language"`
	Category         string           `json:"category"`
	Body             string           `json:"body"`
	Header           *json.RawMessage `json:"header,omitempty"`
	Footer           *string          `json:"footer,omitempty"`
	Buttons          *json.RawMessage `json:"buttons,omitempty"`
	Status           string           `json:"status"` // PENDING, APPROVED, REJECTED
	MetaTemplateID   string           `json:"meta_template_id"`
	VariableMappings *json.RawMessage `json:"variable_mappings,omitempty"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
	SentCount        int              `json:"sent_count"`
	DeliveredCount   int              `json:"delivered_count"`
	ReadCount        int              `json:"read_count"`
}

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

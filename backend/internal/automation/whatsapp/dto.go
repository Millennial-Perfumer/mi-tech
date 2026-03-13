package whatsapp

import (
	"encoding/json"
	"time"
)

type TemplateHeader struct {
	Type     string          `json:"type"` // image, video, document, location, none
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
	ID           int       `json:"id"`
	StoreID      string    `json:"store_id"`
	WebhookTopic string    `json:"webhook_topic"`
	TemplateID   int       `json:"template_id"`
	Enabled      bool      `json:"enabled"`
	TemplateName string    `json:"template_name,omitempty"` // Joined field
	CreatedAt    time.Time `json:"created_at"`
}

type AutomationTemplate struct {
	ID             int              `json:"id"`
	StoreID        string           `json:"store_id"`
	TemplateName   string           `json:"template_name"`
	Language       string           `json:"language"`
	Category       string           `json:"category"`
	Body           string           `json:"body"`
	Header         *json.RawMessage `json:"header,omitempty"`
	Footer         *string          `json:"footer,omitempty"`
	Buttons        *json.RawMessage `json:"buttons,omitempty"`
	Status         string           `json:"status"` // PENDING, APPROVED, REJECTED
	MetaTemplateID string           `json:"meta_template_id"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
}

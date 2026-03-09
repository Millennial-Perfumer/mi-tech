package whatsapp

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"shopify-gst-app/internal/config"
	"strings"
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

type TemplatesService struct {
	repo       *TemplatesRepository
	metaClient *MetaClient
	cfg        *config.Config
}

func NewTemplatesService(repo *TemplatesRepository, cfg *config.Config) *TemplatesService {
	metaClient := NewMetaClient(cfg.WhatsAppAccessToken, cfg.WhatsAppPhoneNumberID, cfg.WhatsAppWABAID)
	return &TemplatesService{
		repo:       repo,
		metaClient: metaClient,
		cfg:        cfg,
	}
}

func (s *TemplatesService) CreateTemplate(storeID string, req CreateTemplateRequest) (int, error) {
	// 1. Map to Meta Components
	components := s.mapToMetaComponents(req)

	// 2. Send to Meta
	metaReq := TemplateRequest{
		Name:       req.Name,
		Category:   req.Category,
		Language:   req.Language,
		Components: components,
	}

	metaID, err := s.metaClient.CreateTemplate(metaReq)
	if err != nil {
		return 0, fmt.Errorf("failed to create template in meta: %w", err)
	}

	// 3. Prepare local storage
	headerJSON, _ := json.Marshal(req.Header)
	buttonsJSON, _ := json.Marshal(req.Buttons)

	headerPtr := json.RawMessage(headerJSON)
	buttonsPtr := json.RawMessage(buttonsJSON)

	template := AutomationTemplate{
		StoreID:        storeID,
		TemplateName:   req.Name,
		Language:       req.Language,
		Category:       req.Category,
		Body:           req.Body,
		Header:         &headerPtr,
		Footer:         sql.NullString{String: req.Footer, Valid: req.Footer != ""},
		Buttons:        &buttonsPtr,
		Status:         "PENDING",
		MetaTemplateID: metaID,
	}

	return s.repo.SaveTemplate(template)
}

func (s *TemplatesService) mapToMetaComponents(req CreateTemplateRequest) []map[string]interface{} {
	bodyComponent := map[string]interface{}{
		"type": "BODY",
		"text": req.Body,
	}

	// Add examples if variables exist
	if req.Examples != "" {
		exampleValues := []string{}
		for _, v := range strings.Split(req.Examples, ",") {
			exampleValues = append(exampleValues, strings.TrimSpace(v))
		}
		if len(exampleValues) > 0 {
			bodyComponent["example"] = map[string]interface{}{
				"body_text": [][]string{exampleValues},
			}
		}
	}

	components := []map[string]interface{}{bodyComponent}

	if req.Header != nil && req.Header.Type != "none" {
		header := map[string]interface{}{
			"type":   "HEADER",
			"format": req.Header.Type, // IMAGE, VIDEO, DOCUMENT, LOCATION
		}
		if len(req.Header.Sample) > 0 && string(req.Header.Sample) != "null" {
			// Location is special
			if req.Header.Type == "LOCATION" {
				var locData map[string]interface{}
				// Check if location is in Sample or dedicated field
				json.Unmarshal(req.Header.Sample, &locData)
				if locData == nil {
					json.Unmarshal(req.Header.Location, &locData)
				}
				header["location"] = locData
			} else {
				// Media samples
				header["example"] = map[string]interface{}{
					"header_handle": []string{string(bytes.ReplaceAll(req.Header.Sample, []byte("\""), []byte("")))},
				}
			}
		}
		components = append(components, header)
	}

	if req.Footer != "" {
		components = append(components, map[string]interface{}{
			"type": "FOOTER",
			"text": req.Footer,
		})
	}

	if len(req.Buttons) > 0 {
		metaButtons := []map[string]interface{}{}
		for _, b := range req.Buttons {
			btn := map[string]interface{}{
				"type": b.Type,
				"text": b.Text,
			}
			// Mapping types to Meta specifics
			switch b.Type {
			case "custom":
				btn["type"] = "QUICK_REPLY"
				if b.Payload != "" {
					btn["payload"] = b.Payload
				}
			case "visit_website":
				btn["type"] = "URL"
				btn["url"] = b.URL
			case "call_phone":
				btn["type"] = "PHONE_NUMBER"
				btn["phone_number"] = b.PhoneNumber
			case "flow":
				btn["type"] = "FLOW"
				btn["flow_id"] = b.FlowID
				btn["flow_action"] = "navigate" // default
				btn["navigate_screen"] = "START"
			case "copy_code":
				btn["type"] = "COPY_CODE"
				btn["example"] = b.OfferCode
			}
			metaButtons = append(metaButtons, btn)
		}
		components = append(components, map[string]interface{}{
			"type":    "BUTTONS",
			"buttons": metaButtons,
		})
	}

	return components
}

func (s *TemplatesService) GetTriggers(storeID string) ([]Trigger, error) {
	return s.repo.GetTriggers(storeID)
}

func (s *TemplatesService) SyncStatus(storeID string) error {
	templates, err := s.repo.GetTemplates(storeID)
	if err != nil {
		return err
	}

	for _, t := range templates {
		status, err := s.metaClient.GetTemplateStatus(t.TemplateName)
		if err != nil {
			continue // Skip failed syncs
		}

		if status != t.Status {
			s.repo.UpdateStatus(t.TemplateName, status)
		}
	}

	return nil
}

func (s *TemplatesService) GetTemplates(storeID string) ([]AutomationTemplate, error) {
	return s.repo.GetTemplates(storeID)
}

func (s *TemplatesService) CreateTrigger(storeID, topic string, templateID int) error {
	trigger := Trigger{
		StoreID:      storeID,
		WebhookTopic: topic,
		TemplateID:   templateID,
		Enabled:      true,
	}
	return s.repo.SaveTrigger(trigger)
}

func (s *TemplatesService) UpdateTemplate(storeID string, id int, req CreateTemplateRequest) error {
	log.Printf("Service: UpdateTemplate called for ID: %d", id)
	// 1. Get current template
	t, err := s.repo.GetTemplateByID(id)
	if err != nil {
		return err
	}
	if t == nil || t.StoreID != storeID {
		return fmt.Errorf("template not found")
	}

	// 2. Map to Meta Components
	components := s.mapToMetaComponents(req)

	// 3. Resubmit to Meta
	if t.MetaTemplateID != "" {
		err = s.metaClient.UpdateTemplate(t.MetaTemplateID, components)
	} else {
		metaReq := TemplateRequest{
			Name:       t.TemplateName,
			Category:   t.Category,
			Language:   t.Language,
			Components: components,
		}
		_, err = s.metaClient.CreateTemplate(metaReq)
	}

	if err != nil {
		return fmt.Errorf("meta resubmission failed: %w", err)
	}

	// 4. Update local DB
	headerJSON, _ := json.Marshal(req.Header)
	buttonsJSON, _ := json.Marshal(req.Buttons)
	headerPtr := json.RawMessage(headerJSON)
	buttonsPtr := json.RawMessage(buttonsJSON)

	t.Body = req.Body
	t.Header = &headerPtr
	t.Footer = sql.NullString{String: req.Footer, Valid: req.Footer != ""}
	t.Buttons = &buttonsPtr
	t.Status = "PENDING"
	return s.repo.UpdateTemplate(*t)
}

func (s *TemplatesService) DeleteTemplate(id int, storeID string) error {
	// 1. Get template to find the name for Meta
	t, err := s.repo.GetTemplateByID(id)
	if err != nil {
		return err
	}
	if t == nil {
		return nil // Already gone
	}

	// 2. Delete from Meta
	err = s.metaClient.DeleteTemplate(t.TemplateName)
	if err != nil {
		log.Printf("Warning: Failed to delete template from Meta: %v", err)
		// We might still want to delete locally if it's missing on Meta
	}

	// 3. Delete triggers
	err = s.repo.DeleteTriggersByTemplateID(id, storeID)
	if err != nil {
		return err
	}

	// 4. Delete template
	return s.repo.DeleteTemplate(id, storeID)
}

func (s *TemplatesService) UpdateTrigger(id int, storeID string, enabled bool) error {
	return s.repo.UpdateTrigger(id, storeID, enabled)
}

func (s *TemplatesService) DeleteTrigger(id int, storeID string) error {
	return s.repo.DeleteTrigger(id, storeID)
}

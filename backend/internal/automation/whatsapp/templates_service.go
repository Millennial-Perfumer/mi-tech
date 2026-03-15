package whatsapp

import (
	"encoding/json"
	"fmt"
	"log"
	"mi-tech/internal/config"
	"strings"
)

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
	components, err := s.mapToMetaComponents(req)
	if err != nil {
		return 0, fmt.Errorf("failed to map components: %w", err)
	}

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
		Footer:         &req.Footer,
		Buttons:        &buttonsPtr,
		Status:         "PENDING",
		MetaTemplateID: metaID,
	}

	return s.repo.SaveTemplate(template)
}

func (s *TemplatesService) UploadMediaBytes(body []byte, mimeType string) (string, error) {
	appID := s.cfg.WhatsAppAppID
	if appID == "" {
		return "", fmt.Errorf("WHATSAPP_APP_ID is required to upload media")
	}
	return s.metaClient.UploadMediaFromBytes(appID, body, mimeType)
}

func (s *TemplatesService) mapToMetaComponents(req CreateTemplateRequest) ([]map[string]interface{}, error) {
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
			"format": strings.ToUpper(req.Header.Type), // IMAGE, VIDEO, DOCUMENT, LOCATION
		}
		if len(req.Header.Sample) > 0 && string(req.Header.Sample) != "null" {
			var sampleStr string
			if err := json.Unmarshal(req.Header.Sample, &sampleStr); err != nil {
				// Fallback to old behavior if not a quoted JSON string
				sampleStr = strings.Trim(string(req.Header.Sample), "\"")
			}
			
			// Sanitization: Ensure no newlines or spaces in handle
			sampleStr = strings.TrimSpace(sampleStr)
			if idx := strings.IndexAny(sampleStr, "\n\r"); idx != -1 {
				// If concatenated, take only the first one
				sampleStr = strings.TrimSpace(sampleStr[:idx])
			}

			// Auto-convert Google Drive share links to direct download links
			if strings.Contains(sampleStr, "drive.google.com") {
				if strings.Contains(sampleStr, "/file/d/") {
					parts := strings.Split(sampleStr, "/file/d/")
					if len(parts) > 1 {
						id := strings.Split(parts[1], "/")[0]
						id = strings.Split(id, "?")[0]
						sampleStr = fmt.Sprintf("https://drive.google.com/uc?export=download&id=%s", id)
					}
				}
			}

			if req.Header.Type == "LOCATION" {
				var locData map[string]interface{}
				json.Unmarshal(req.Header.Sample, &locData)
				if locData == nil {
					json.Unmarshal(req.Header.Location, &locData)
				}
				header["location"] = locData
			} else {
				// Media samples - check if it's a URL or a handle
				example := map[string]interface{}{}
				if strings.HasPrefix(sampleStr, "http") {
					// We must provide a header_handle for template creation.
					// Upload the URL to Meta first using the Resumable Upload API.
					appID := s.cfg.WhatsAppAppID
					if appID == "" {
						return nil, fmt.Errorf("WHATSAPP_APP_ID is required to upload media examples")
					}
					log.Printf("Meta Automation: Uploading media sample to Meta. AppID: %s, URL: %s", appID, sampleStr)
					handle, err := s.metaClient.UploadMediaFromURL(appID, sampleStr)
					if err != nil {
						return nil, fmt.Errorf("failed to upload media sample to Meta: %w", err)
					}
					example["header_handle"] = []string{handle}
				} else {
					example["header_handle"] = []string{sampleStr}
				}
				header["example"] = example
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
				url := b.URL
				if url != "" && !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
					url = "https://" + url
				}
				btn["url"] = url
				if strings.Contains(url, "{{1}}") {
					btn["example"] = []string{"https://www.delhivery.com/track/package/123456"}
				}
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

	return components, nil
}

func (s *TemplatesService) GetTriggers(storeID string) ([]Trigger, error) {
	return s.repo.GetTriggers(storeID)
}

func (s *TemplatesService) SyncStatus(storeID string) error {
	templates, err := s.repo.GetTemplates(storeID, "", "")
	if err != nil {
		return err
	}

	for _, t := range templates {
		remote, err := s.metaClient.GetRemoteTemplateByName(t.TemplateName)
		if err != nil || remote == nil {
			continue // Skip failed syncs
		}

		if remote.Status != t.Status {
			s.repo.UpdateStatus(t.TemplateName, remote.Status)
		}
	}

	return nil
}

func (s *TemplatesService) GetTemplates(storeID string, startDate, endDate string) ([]AutomationTemplate, error) {
	return s.repo.GetTemplates(storeID, startDate, endDate)
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
	components, err := s.mapToMetaComponents(req)
	if err != nil {
		return fmt.Errorf("failed to map components: %w", err)
	}

	// 3. Resubmit to Meta
	if t.MetaTemplateID == "" {
		log.Printf("Service: MetaTemplateID missing locally for %s. Attempting to resolve from Meta...", t.TemplateName)
		remote, err := s.metaClient.GetRemoteTemplateByName(t.TemplateName)
		if err != nil {
			return fmt.Errorf("failed to resolve template from meta: %w", err)
		}
		if remote != nil {
			log.Printf("Service: Resolved MetaTemplateID for %s: %s", t.TemplateName, remote.ID)
			t.MetaTemplateID = remote.ID
			// Update locally immediately so we don't have to fetch again
			_ = s.repo.UpdateTemplate(*t)
		}
	}

	if t.MetaTemplateID != "" {
		log.Printf("Service: Updating existing Meta template %s (ID: %s)", t.TemplateName, t.MetaTemplateID)
		err = s.metaClient.UpdateTemplate(t.MetaTemplateID, components)
	} else {
		log.Printf("Service: Template %s not found on Meta. Creating new one.", t.TemplateName)
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
	t.Footer = &req.Footer
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

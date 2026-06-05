package whatsapp

import (
	"encoding/json"
	"fmt"
	"log"
	"mi-tech/internal/config"
	"strings"
	"time"
)

type TemplatesService struct {
	repo       *TemplatesRepository
	metaClient *MetaClient
	settings   *config.SettingsProvider
}

func NewTemplatesService(repo *TemplatesRepository, settings *config.SettingsProvider) *TemplatesService {
	metaClient := NewMetaClient(settings)
	return &TemplatesService{
		repo:       repo,
		metaClient: metaClient,
		settings:   settings,
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
	appID := s.settings.GetWhatsAppAppID()
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
					appID := s.settings.GetWhatsAppAppID()
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
			case "visit_website":
				btn["type"] = "URL"
				url := b.URL
				if url != "" && !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
					url = "https://" + url
				}
				btn["url"] = url
				if strings.Contains(url, "{{1}}") {
					btn["example"] = []string{"https://www.example.com/track/123456"}
				}
			case "call_phone":
				btn["type"] = "PHONE_NUMBER"
				btn["phone_number"] = b.PhoneNumber
			case "flow":
				btn["type"] = "FLOW"
				btn["flow_id"] = b.FlowID
				btn["flow_action"] = "navigate" // default
				btn["navigate_screen"] = "START"
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
	templates, err := s.repo.GetTemplates(storeID, nil, nil)
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

func (s *TemplatesService) GetTemplates(storeID string, startDate, endDate *time.Time) ([]AutomationTemplate, error) {
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

func (s *TemplatesService) UpdateTemplateMappings(storeID string, id int, mappings *json.RawMessage) error {
	log.Printf("Service: UpdateTemplateMappings called for ID: %d", id)
	t, err := s.repo.GetTemplateByID(id)
	if err != nil {
		return err
	}
	if t == nil || t.StoreID != storeID {
		return fmt.Errorf("template not found")
	}

	t.VariableMappings = mappings
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
func (s *TemplatesService) FetchRemoteTemplate(templateName string) (*CreateTemplateRequest, error) {
	remote, err := s.metaClient.GetRemoteTemplateByName(templateName)
	if err != nil {
		return nil, err
	}
	if remote == nil {
		return nil, fmt.Errorf("template %s not found in Meta", templateName)
	}

	req := &CreateTemplateRequest{
		Name:     remote.Name,
		Category: remote.Category,
		Language: remote.Language,
	}

	for _, comp := range remote.Components {
		compType, _ := comp["type"].(string)
		switch compType {
		case "HEADER":
			format, _ := comp["format"].(string)
			req.Header = &TemplateHeader{
				Type: strings.ToLower(format),
			}
			// Meta doesn't return the sample handle in GET, but might return examples
			if example, ok := comp["example"].(map[string]interface{}); ok {
				if handles, ok := example["header_handle"].([]interface{}); ok && len(handles) > 0 {
					handle, _ := handles[0].(string)
					sampleJSON, _ := json.Marshal(handle)
					req.Header.Sample = sampleJSON
				}
			}
		case "BODY":
			req.Body, _ = comp["text"].(string)
			if example, ok := comp["example"].(map[string]interface{}); ok {
				if bodyTexts, ok := example["body_text"].([]interface{}); ok && len(bodyTexts) > 0 {
					if firstRow, ok := bodyTexts[0].([]interface{}); ok {
						var samples []string
						for _, s := range firstRow {
							samples = append(samples, fmt.Sprint(s))
						}
						req.Examples = strings.Join(samples, ", ")
					}
				}
			}
		case "FOOTER":
			req.Footer, _ = comp["text"].(string)
		case "BUTTONS":
			if btns, ok := comp["buttons"].([]interface{}); ok {
				for _, b := range btns {
					btnMap, _ := b.(map[string]interface{})
					bType, _ := btnMap["type"].(string)
					bText, _ := btnMap["text"].(string)

					newBtn := TemplateButton{
						Text: bText,
					}

					switch bType {
					case "QUICK_REPLY":
						newBtn.Type = "custom"
					case "URL":
						newBtn.Type = "visit_website"
						newBtn.URL, _ = btnMap["url"].(string)
					case "PHONE_NUMBER":
						newBtn.Type = "call_phone"
						newBtn.PhoneNumber, _ = btnMap["phone_number"].(string)
					case "FLOW":
						newBtn.Type = "flow"
						newBtn.FlowID, _ = btnMap["flow_id"].(string)
					}
					req.Buttons = append(req.Buttons, newBtn)
				}
			}
		}
	}

	return req, nil
}

func (s *TemplatesService) SyncAllTemplates(storeID string) error {
	remoteTemplates, err := s.metaClient.GetAllRemoteTemplates()
	if err != nil {
		return fmt.Errorf("failed to fetch templates from Meta: %w", err)
	}

	existingTemplates, err := s.repo.GetTemplates(storeID, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to fetch local templates: %w", err)
	}

	remoteNames := make(map[string]bool)
	for _, remote := range remoteTemplates {
		remoteNames[remote.Name] = true

		localTpl := AutomationTemplate{
			StoreID:        storeID,
			TemplateName:   remote.Name,
			Language:       remote.Language,
			Category:       remote.Category,
			Status:         remote.Status,
			MetaTemplateID: remote.ID,
		}

		var header *TemplateHeader
		var buttons []TemplateButton
		var footer string

		for _, comp := range remote.Components {
			compType, _ := comp["type"].(string)
			switch compType {
			case "HEADER":
				format, _ := comp["format"].(string)
				header = &TemplateHeader{Type: strings.ToLower(format)}
				if text, ok := comp["text"].(string); ok {
					header.Text = text
				}
				if example, ok := comp["example"].(map[string]interface{}); ok {
					if handles, ok := example["header_handle"].([]interface{}); ok && len(handles) > 0 {
						handle, _ := handles[0].(string)
						sampleJSON, _ := json.Marshal(handle)
						header.Sample = sampleJSON
					} else if texts, ok := example["header_text"].([]interface{}); ok && len(texts) > 0 {
						// For text headers, the example is often in header_text
						t, _ := texts[0].(string)
						sampleJSON, _ := json.Marshal(t)
						header.Sample = sampleJSON
					}
				}
			case "BODY":
				localTpl.Body, _ = comp["text"].(string)
			case "FOOTER":
				footer, _ = comp["text"].(string)
			case "BUTTONS":
				if btns, ok := comp["buttons"].([]interface{}); ok {
					for _, b := range btns {
						btnMap, _ := b.(map[string]interface{})
						bType, _ := btnMap["type"].(string)
						bText, _ := btnMap["text"].(string)

						newBtn := TemplateButton{Text: bText}
						switch bType {
						case "QUICK_REPLY":
							newBtn.Type = "custom"
						case "URL":
							newBtn.Type = "visit_website"
							newBtn.URL, _ = btnMap["url"].(string)
						case "PHONE_NUMBER":
							newBtn.Type = "call_phone"
							newBtn.PhoneNumber, _ = btnMap["phone_number"].(string)
						case "FLOW":
							newBtn.Type = "flow"
							newBtn.FlowID, _ = btnMap["flow_id"].(string)
						}
						buttons = append(buttons, newBtn)
					}
				}
			}
		}

		if header != nil {
			hb, _ := json.Marshal(header)
			rawH := json.RawMessage(hb)
			localTpl.Header = &rawH
		}
		if footer != "" {
			localTpl.Footer = &footer
		}
		if len(buttons) > 0 {
			bb, _ := json.Marshal(buttons)
			rawB := json.RawMessage(bb)
			localTpl.Buttons = &rawB
		}

		_, err = s.repo.UpsertMetaTemplate(localTpl)
		if err != nil {
			log.Printf("Failed to upsert template %s: %v", remote.Name, err)
		}
	}

	for _, local := range existingTemplates {
		if !remoteNames[local.TemplateName] && local.Status != "ARCHIVED" {
			log.Printf("Template %s not found in Meta, marking as ARCHIVED", local.TemplateName)
			s.repo.UpdateStatus(local.TemplateName, "ARCHIVED")
		}
	}

	return nil
}

func (s *TemplatesService) SyncSingleTemplate(storeID, templateName string) error {
	remote, err := s.metaClient.GetRemoteTemplateByName(templateName)
	if err != nil {
		return fmt.Errorf("failed to fetch template %s: %w", templateName, err)
	}
	if remote == nil {
		return fmt.Errorf("template %s not found", templateName)
	}

	localTpl := AutomationTemplate{
		StoreID:      storeID,
		TemplateName: remote.Name,
		Category:     remote.Category,
		Language:     remote.Language,
		Status:       remote.Status,
	}

	var header *TemplateHeader
	var footer string
	var buttons []TemplateButton

	for _, comp := range remote.Components {
		cType, _ := comp["type"].(string)
		switch cType {
		case "BODY":
			if text, ok := comp["text"].(string); ok {
				localTpl.Body = text
			}
		case "HEADER":
			format, _ := comp["format"].(string)
			header = &TemplateHeader{Type: strings.ToLower(format)}
			if text, ok := comp["text"].(string); ok {
				header.Text = text
			}
			if example, ok := comp["example"].(map[string]interface{}); ok {
				if handles, ok := example["header_handle"].([]interface{}); ok && len(handles) > 0 {
					handle, _ := handles[0].(string)
					sampleJSON, _ := json.Marshal(handle)
					header.Sample = sampleJSON
				} else if texts, ok := example["header_text"].([]interface{}); ok && len(texts) > 0 {
					t, _ := texts[0].(string)
					sampleJSON, _ := json.Marshal(t)
					header.Sample = sampleJSON
				}
			}
		case "FOOTER":
			footer, _ = comp["text"].(string)
		case "BUTTONS":
			if btns, ok := comp["buttons"].([]interface{}); ok {
				for _, b := range btns {
					btnMap, _ := b.(map[string]interface{})
					bType, _ := btnMap["type"].(string)
					bText, _ := btnMap["text"].(string)

					newBtn := TemplateButton{
						Text: bText,
					}

					switch bType {
					case "QUICK_REPLY":
						newBtn.Type = "custom"
					case "URL":
						newBtn.Type = "visit_website"
						newBtn.URL, _ = btnMap["url"].(string)
					case "PHONE_NUMBER":
						newBtn.Type = "call_phone"
						newBtn.PhoneNumber, _ = btnMap["phone_number"].(string)
					case "FLOW":
						newBtn.Type = "flow"
						newBtn.FlowID, _ = btnMap["flow_id"].(string)
					case "COPY_CODE":
						newBtn.Type = "copy_code"
					}
					buttons = append(buttons, newBtn)
				}
			}
		}
	}

	if header != nil {
		hb, _ := json.Marshal(header)
		rawH := json.RawMessage(hb)
		localTpl.Header = &rawH
	}
	if footer != "" {
		localTpl.Footer = &footer
	}
	if len(buttons) > 0 {
		bb, _ := json.Marshal(buttons)
		rawB := json.RawMessage(bb)
		localTpl.Buttons = &rawB
	}

	_, err = s.repo.UpsertMetaTemplate(localTpl)
	return err
}

func (s *TemplatesService) GetEvents() ([]AutomationEvent, error) {
	return s.repo.GetEvents()
}

func (s *TemplatesService) SaveEvent(e AutomationEvent) error {
	return s.repo.SaveEvent(e)
}

func (s *TemplatesService) DeleteEvent(id int) error {
	return s.repo.DeleteEvent(id)
}

package whatsapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"mi-tech/internal/entity"
	"mi-tech/internal/repository"
	"mi-tech/internal/service"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var phoneRegex = regexp.MustCompile(`[^0-9]`)

func sanitizePhoneNumber(phone string) string {
	if phone == "555-555-SHIP" {
		return "916383173716"
	}
	// Remove all non-digit characters
	sanitized := phoneRegex.ReplaceAllString(phone, "")
	// Ensure it doesn't have leading zeros if they were part of a +00 prefix
	return strings.TrimLeft(sanitized, "0")
}

type WebhookMappingService struct {
	templatesRepo   *TemplatesRepository
	messagesService *MessagesService
	invoiceService  *service.InvoiceService
	settingsRepo    *repository.SettingsRepository
}

func NewWebhookMappingService(tRepo *TemplatesRepository, mService *MessagesService, iService *service.InvoiceService, sRepo *repository.SettingsRepository) *WebhookMappingService {
	return &WebhookMappingService{
		templatesRepo:   tRepo,
		messagesService: mService,
		invoiceService:  iService,
		settingsRepo:    sRepo,
	}
}

func (s *WebhookMappingService) ExecuteMapping(storeID, topic string, order entity.Order) error {
	log.Printf("Automation Started: Executing mapping for Order %d (%s), Topic: %s", order.ID, order.OrderNumber, topic)

	// 1. Find matching trigger
	trigger, err := s.templatesRepo.GetTriggerByTopic(storeID, topic)
	if err != nil {
		log.Printf("Automation Error: Database error fetching trigger for topic %s: %v", topic, err)
		return fmt.Errorf("error fetching trigger: %w", err)
	}
	if trigger == nil {
		log.Printf("Automation Skip: No enabled trigger found for topic %s (Store: %s). Check triggers table.", topic, storeID)
		return nil
	}

	template, err := s.templatesRepo.GetTemplateByID(trigger.TemplateID)
	if err != nil {
		log.Printf("Automation Error: Database error fetching template %d: %v", trigger.TemplateID, err)
		return fmt.Errorf("error fetching template: %w", err)
	}
	if template == nil {
		log.Printf("Automation Error: Template %d not found for trigger on topic %s", trigger.TemplateID, topic)
		return fmt.Errorf("template not found")
	}

	log.Printf("Automation Progress: Found template: %s (ID: %d). Proceeding to execute.", template.TemplateName, template.ID)
	return s.executeWithTemplate(storeID, template, order, topic)
}

func (s *WebhookMappingService) ExecuteManualSend(storeID string, templateID int, order entity.Order) error {
	log.Printf("Automation Start: Executing manual send for Order %d (%s), Template ID: %d", order.ID, order.OrderNumber, templateID)

	// Fetch template
	template, err := s.templatesRepo.GetTemplateByID(templateID)
	if err != nil {
		return fmt.Errorf("error fetching template: %w", err)
	}
	if template == nil {
		return fmt.Errorf("template not found: %d", templateID)
	}

	return s.executeWithTemplate(storeID, template, order, "manual")
}

func resolveVariable(field string, order entity.Order) string {
	switch field {
	case "customer_name":
		name := entity.DerefStr(order.CustomerFirstName)
		if name == "" {
			name = entity.DerefStr(order.CustomerName)
		}
		if name == "" {
			return "Customer"
		}
		return name
	case "order_id":
		return strings.TrimPrefix(order.OrderNumber, "#")
	case "order_total":
		return fmt.Sprintf("%.2f", order.TotalPrice)
	case "tracking_link":
		return entity.DerefStr(order.TrackingUrl)
	case "tracking_number":
		return entity.DerefStr(order.TrackingNumber)
	case "shipping_company":
		return entity.DerefStr(order.ShippingCompany)
	case "customer_city":
		return entity.DerefStr(order.CustomerCity)
	case "customer_state":
		return entity.DerefStr(order.CustomerState)
	case "customer_country":
		return entity.DerefStr(order.CustomerCountry)
	case "customer_zip":
		return entity.DerefStr(order.CustomerZip)
	case "customer_address1":
		return entity.DerefStr(order.CustomerAddress1)
	case "customer_address2":
		return entity.DerefStr(order.CustomerAddress2)
	case "customer_email":
		return entity.DerefStr(order.CustomerEmail)
	case "customer_phone":
		return entity.DerefStr(order.CustomerPhone)
	case "internal_order_id":
		return fmt.Sprintf("%d", order.ID)
	default:
		return "" // Unknown or empty mapping yields empty string (will fail if Meta requires it, which is correct behavior for unmapped vars)
	}
}

func (s *WebhookMappingService) executeWithTemplate(storeID string, template *AutomationTemplate, order entity.Order, topic string) error {
	// Deduplication Check (only for automated topics)
	if topic != "manual" {
		// Only allow multiple sends for final tracking-only updates if needed.
		// For all other cases (creation, cancellation), if we already sent this template for this order, skip.
		allowMultiple := topic == "orders/out_for_delivery" || topic == "orders/delivered"

		// Use a time window for "Assigned" or "Dispatched" to allow for cancel+reassign by humans
		// while blocking near-simultaneous redundant webhooks from Shopify.
		var since time.Time
		if topic == "orders/assigned" || topic == "orders/fulfilled" || topic == "orders/dispatched" {
			since = time.Now().Add(-2 * time.Minute)
		}

		sent, err := s.messagesService.repo.HasSentTemplate(order.ID, template.ID, since)
		if err != nil {
			log.Printf("Automation Error: Deduplication check failed for order %d: %v", order.ID, err)
		} else if sent && !allowMultiple {
			log.Printf("Automation Skip: Template %s already sent for order %d (Topic: %s) within dedupe window. Skipping.", template.TemplateName, order.ID, topic)
			return nil
		}
	}

	var mappings map[string]string
	if template.VariableMappings != nil {
		json.Unmarshal(*template.VariableMappings, &mappings)
	}
	if mappings == nil {
		mappings = make(map[string]string)
	}

	var components []interface{}

	// 1. Body Mapping
	requiredCount := s.countRequiredParams(template.Body)
	if requiredCount > 0 {
		var bodyParams []map[string]string
		for i := 1; i <= requiredCount; i++ {
			mapKey := fmt.Sprintf("body_text_0_{{%d}}", i)
			fieldToMap := mappings[mapKey]
			val := resolveVariable(fieldToMap, order)
			
			// Fallback logic for legacy templates that were not mapped yet
			if val == "" {
				if i == 1 {
					val = resolveVariable("customer_name", order)
				} else if i == 2 {
					val = resolveVariable("order_id", order)
				}
			}
			bodyParams = append(bodyParams, map[string]string{"type": "text", "text": val})
		}
		components = append(components, map[string]interface{}{
			"type":       "body",
			"parameters": bodyParams,
		})
	}

	// 2. Buttons Mapping
	if template.Buttons != nil && string(*template.Buttons) != "null" {
		var buttons []map[string]interface{}
		if err := json.Unmarshal(*template.Buttons, &buttons); err == nil {
			for i, btn := range buttons {
				if btn["type"] == "visit_website" {
					url, _ := btn["url"].(string)
					if strings.Contains(url, "{{1}}") {
						mapKey := fmt.Sprintf("button_url_%d_{{1}}", i)
						fieldToMap := mappings[mapKey]
						
						val := resolveVariable(fieldToMap, order)
						if val == "" {
							// Legacy fallback for embedded tracking loop
							val = resolveVariable("internal_order_id", order)
						}

						components = append(components, map[string]interface{}{
							"type":     "button",
							"sub_type": "url",
							"index":    strconv.Itoa(i),
							"parameters": []map[string]interface{}{
								{
									"type": "text",
									"text": val,
								},
							},
						})
					}
				}
				// Advanced: Flow, Copy Code mapping can be added here
			}
		}
	}

	// 3. Header Mapping
	if template.Header != nil && string(*template.Header) != "null" {
		var hData struct {
			Type string `json:"type"`
		}
		json.Unmarshal(*template.Header, &hData)
		hType := strings.ToUpper(hData.Type)

		if hType == "TEXT" {
			var hTextData struct {
				Text string `json:"text"`
			}
			json.Unmarshal(*template.Header, &hTextData)
			
			reqCount := s.countRequiredParams(hTextData.Text)
			if reqCount > 0 {
				var headerParams []map[string]string
				for i := 1; i <= reqCount; i++ {
					mapKey := fmt.Sprintf("header_text_0_{{%d}}", i)
					val := resolveVariable(mappings[mapKey], order)
					headerParams = append(headerParams, map[string]string{"type": "text", "text": val})
				}
				components = append(components, map[string]interface{}{
					"type": "header",
					"parameters": headerParams,
				})
			}
		} else if hType == "DOCUMENT" || hType == "IMAGE" || hType == "VIDEO" {
			headerHandle := mappings["header_handle"]
			
			// Dynamic generation
			if headerHandle == "Dynamic Invoice" && hType == "DOCUMENT" {
				log.Printf("Automation Detail: Generating real invoice PDF for order %s", order.OrderNumber)
				items, err := s.messagesService.repo.GetOrderLineItems(order.ID)
				if err != nil {
					log.Printf("Automation Error: Failed to fetch line items for invoice: %v", err)
				} else {
					var buf bytes.Buffer
					if err := s.invoiceService.GeneratePDF(order, items, &buf); err != nil {
						log.Printf("Automation Error: Failed to generate PDF: %v", err)
					} else {
						filename := "Invoice_" + order.OrderNumber + ".pdf"
						id, err := s.messagesService.metaClient.UploadWhatsAppMedia(buf.Bytes(), filename, "application/pdf")
						if err == nil {
							var idVal interface{} = id
							if numericID, err := strconv.ParseInt(id, 10, 64); err == nil {
								idVal = numericID
							}

							components = append(components, map[string]interface{}{
								"type": "header",
								"parameters": []map[string]interface{}{
									{
										"type": "document",
										"document": map[string]interface{}{
											"id":       idVal,
											"filename": filename,
										},
									},
								},
							})
						} else {
							log.Printf("Automation Error: Failed to upload invoice to Meta: %v", err)
						}
					}
				}
			} else if headerHandle != "" {
				// Static mapped media
				var idVal interface{} = headerHandle
				if numericID, err := strconv.ParseInt(headerHandle, 10, 64); err == nil {
					idVal = numericID
				}
				paramType := strings.ToLower(hType)
				
				paramObj := map[string]interface{}{"id": idVal}
				if paramType == "document" {
					paramObj["filename"] = "Document" // generic fallback name
				}

				components = append(components, map[string]interface{}{
					"type": "header",
					"parameters": []map[string]interface{}{
						{
							"type": paramType,
							paramType: paramObj,
						},
					},
				})
			}
		}
	}

	// 5. Send message
	if order.CustomerPhone == nil || *order.CustomerPhone == "" {
		log.Printf("Skip automation: no phone number for order %s", order.OrderNumber)
		return nil
	}

	cleanPhone := sanitizePhoneNumber(*order.CustomerPhone)
	if len(cleanPhone) < 8 {
		log.Printf("Skip automation: invalid phone number '%s' (sanitized: '%s') for order %s", *order.CustomerPhone, cleanPhone, order.OrderNumber)
		return nil
	}

	compJSON, _ := json.Marshal(components)
	log.Printf("Automation Meta Call: Sending %s to %s (Order: %d). Payload: %s", template.TemplateName, cleanPhone, order.ID, string(compJSON))

	return s.messagesService.SendTemplateMessage(
		storeID,
		template.ID,
		order.ID,
		cleanPhone,
		template.TemplateName,
		template.Language,
		components,
	)
}

// countRequiredParams finds the maximum {{n}} placeholder in the template body.
func (s *WebhookMappingService) countRequiredParams(body string) int {
	re := regexp.MustCompile(`\{\{(\d+)\}\}`)
	matches := re.FindAllStringSubmatch(body, -1)
	max := 0
	for _, m := range matches {
		if len(m) > 1 {
			n, _ := strconv.Atoi(m[1])
			if n > max {
				max = n
			}
		}
	}
	return max
}

func resolveCustomerVariable(field string, customer *entity.Customer) string {
	switch field {
	case "customer_name":
		name := entity.DerefStr(customer.FirstName)
		if name == "" {
			name = entity.DerefStr(customer.LastName)
		}
		if name == "" {
			return "Customer"
		}
		return name
	case "customer_city":
		return entity.DerefStr(customer.City)
	case "customer_state":
		return entity.DerefStr(customer.State)
	case "customer_country":
		return entity.DerefStr(customer.Country)
	case "customer_zip":
		return entity.DerefStr(customer.ZipCode)
	case "customer_address1":
		return entity.DerefStr(customer.Address1)
	case "customer_address2":
		return entity.DerefStr(customer.Address2)
	case "customer_email":
		return entity.DerefStr(customer.Email)
	case "customer_phone":
		return customer.PhoneNumber
	case "customer_total_orders":
		return fmt.Sprintf("%d", customer.TotalOrders)
	case "customer_total_spent":
		return fmt.Sprintf("%.2f", customer.TotalSpent)
	default:
		return ""
	}
}

func (s *WebhookMappingService) ExecuteMarketingSend(storeID string, template *AutomationTemplate, customer *entity.Customer) error {
	var mappings map[string]string
	if template.VariableMappings != nil {
		json.Unmarshal(*template.VariableMappings, &mappings)
	}
	if mappings == nil {
		mappings = make(map[string]string)
	}

	var components []interface{}

	requiredCount := s.countRequiredParams(template.Body)
	if requiredCount > 0 {
		var bodyParams []map[string]string
		for i := 1; i <= requiredCount; i++ {
			mapKey := fmt.Sprintf("body_text_0_{{%d}}", i)
			fieldToMap := mappings[mapKey]
			val := resolveCustomerVariable(fieldToMap, customer)
			
			if val == "" && i == 1 {
				val = resolveCustomerVariable("customer_name", customer)
			}
			bodyParams = append(bodyParams, map[string]string{"type": "text", "text": val})
		}
		components = append(components, map[string]interface{}{
			"type":       "body",
			"parameters": bodyParams,
		})
	}

	if template.Buttons != nil && string(*template.Buttons) != "null" {
		var buttons []map[string]interface{}
		if err := json.Unmarshal(*template.Buttons, &buttons); err == nil {
			for i, btn := range buttons {
				if btn["type"] == "visit_website" {
					url, _ := btn["url"].(string)
					if strings.Contains(url, "{{1}}") {
						mapKey := fmt.Sprintf("button_url_%d_{{1}}", i)
						fieldToMap := mappings[mapKey]
						val := resolveCustomerVariable(fieldToMap, customer)
						components = append(components, map[string]interface{}{
							"type":     "button",
							"sub_type": "url",
							"index":    strconv.Itoa(i),
							"parameters": []map[string]interface{}{{"type": "text", "text": val}},
						})
					}
				}
			}
		}
	}

	if template.Header != nil && string(*template.Header) != "null" {
		var hData struct {
			Type string `json:"type"`
		}
		json.Unmarshal(*template.Header, &hData)
		hType := strings.ToUpper(hData.Type)

		if hType == "DOCUMENT" || hType == "IMAGE" || hType == "VIDEO" {
			headerHandle := mappings["header_handle"]
			if headerHandle != "" {
				var idVal interface{} = headerHandle
				if numericID, err := strconv.ParseInt(headerHandle, 10, 64); err == nil {
					idVal = numericID
				}
				paramType := strings.ToLower(hType)
				paramObj := map[string]interface{}{"id": idVal}
				if paramType == "document" {
					paramObj["filename"] = "Document"
				}
				components = append(components, map[string]interface{}{
					"type": "header",
					"parameters": []map[string]interface{}{
						{
							"type": paramType,
							paramType: paramObj,
						},
					},
				})
			}
		}
	}

	cleanPhone := sanitizePhoneNumber(customer.PhoneNumber)
	if len(cleanPhone) < 8 {
		return nil
	}

	return s.messagesService.SendTemplateMessage(storeID, template.ID, 0, cleanPhone, template.TemplateName, template.Language, components)
}

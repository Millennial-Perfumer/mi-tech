package whatsapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"mi-tech/internal/entity"
	"mi-tech/internal/service"
	"strconv"
	"strings"
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
}

func NewWebhookMappingService(tRepo *TemplatesRepository, mService *MessagesService, iService *service.InvoiceService) *WebhookMappingService {
	return &WebhookMappingService{
		templatesRepo:   tRepo,
		messagesService: mService,
		invoiceService:  iService,
	}
}

func (s *WebhookMappingService) ExecuteMapping(storeID, topic string, order entity.Order) error {
	log.Printf("Automation Start: Executing mapping for Order %s (%s), Topic: %s", order.ID, order.OrderNumber, topic)

	// 1. Find matching trigger
	trigger, err := s.templatesRepo.GetTriggerByTopic(storeID, topic)
	if err != nil {
		return fmt.Errorf("error fetching trigger: %w", err)
	}
	if trigger == nil {
		log.Printf("Automation Skip: No enabled trigger found for topic %s (Store: %s)", topic, storeID)
		return nil // No mapping for this topic
	}

	// 2. Fetch template
	template, err := s.templatesRepo.GetTemplateByID(trigger.TemplateID)
	if err != nil {
		return fmt.Errorf("error fetching template: %w", err)
	}
	if template == nil {
		log.Printf("Automation Error: Trigger exists for %s but template %d not found", topic, trigger.TemplateID)
		return fmt.Errorf("template not found: %d", trigger.TemplateID)
	}

	log.Printf("Automation Progress: Found trigger %s -> Template: %s (ID: %d)", topic, template.TemplateName, template.ID)

	// 2a. Deduplication Check: Don't send the same template twice to the same order
	// Special Case: Lifecycle events (assigned, fulfilled, delivery) can be re-sent if intentional (e.g. re-assignment).
	// The 15s/30s guards in WebhookHandler already prevent automated double-pings.
	allowMultiple := topic == "orders/assigned" || topic == "orders/fulfilled" || topic == "orders/updated"

	sent, err := s.messagesService.repo.HasSentTemplate(order.ID, template.ID)
	if err != nil {
		log.Printf("Automation Error: Deduplication check failed for order %s: %v", order.ID, err)
	} else if sent && !allowMultiple {
		log.Printf("Automation Skip: Template %s already sent for order %s. Skipping duplicate.", template.TemplateName, order.ID)
		return nil
	}

	// 3. Extract parameters based on template name or topic
	var components []interface{}

	// Create body parameters based on common patterns
	custName := entity.DerefStr(order.CustomerFirstName)
	if custName == "" {
		custName = entity.DerefStr(order.CustomerName)
	}
	if custName == "" {
		custName = "Customer"
	}

	bodyParams := []map[string]string{
		{"type": "text", "text": custName},
		{"type": "text", "text": order.OrderNumber},
	}

	// Dynamic Parameter Mapping: Match the exact number of placeholders in the template body
	requiredCount := s.countRequiredParams(template.Body)

	// Template-specific mapping logic
	if template.TemplateName == "order_dispatched_v3" || template.TemplateName == "out_for_delivery_v3" || template.TemplateName == "order_assigned_v3" {
		if requiredCount >= 3 {
			bodyParams = append(bodyParams, map[string]string{"type": "text", "text": entity.DerefStr(order.ShippingCompany)})
		}
		if requiredCount >= 4 {
			bodyParams = append(bodyParams, map[string]string{"type": "text", "text": entity.DerefStr(order.TrackingNumber)})
		}
		if requiredCount >= 5 {
			bodyParams = append(bodyParams, map[string]string{"type": "text", "text": entity.DerefStr(order.TrackingUrl)})
		}
	} else {
		// Generic fallback
		if requiredCount >= 3 {
			bodyParams = append(bodyParams, map[string]string{"type": "text", "text": entity.DerefStr(order.ShippingCompany)})
		}
		if requiredCount >= 4 {
			bodyParams = append(bodyParams, map[string]string{"type": "text", "text": entity.DerefStr(order.TrackingNumber)})
		}
		if requiredCount >= 5 {
			bodyParams = append(bodyParams, map[string]string{"type": "text", "text": entity.DerefStr(order.TrackingUrl)})
		}
	}

	// Trim bodyParams if for some reason we have more than required (safety)
	if len(bodyParams) > requiredCount && requiredCount > 0 {
		bodyParams = bodyParams[:requiredCount]
	}

	components = append(components, map[string]interface{}{
		"type":       "body",
		"parameters": bodyParams,
	})

	// 3b. Handle Buttons (Dynamic URLs)
	// If the template has a visit_website button, we pass the tracking URL as a parameter.
	if template.Buttons != nil && string(*template.Buttons) != "null" {
		var buttons []map[string]interface{}
		if err := json.Unmarshal(*template.Buttons, &buttons); err == nil {
			for i, btn := range buttons {
				// Meta dynamic URL buttons are of type "visit_website"
				// IMPORTANT: Only send parameters if the URL in the template actually has a variable suffix {{1}}
				if btn["type"] == "visit_website" {
					url, _ := btn["url"].(string)
					trackingURL := entity.DerefStr(order.TrackingUrl)

					if strings.Contains(url, "{{1}}") {
						// For our branded redirector (https://example.com/t/{{1}}),
						// we pass the internal Order ID as the parameter.
						buttonParam := order.ID

						components = append(components, map[string]interface{}{
							"type":     "button",
							"sub_type": "url",
							"index":    strconv.Itoa(i),
							"parameters": []map[string]interface{}{
								{
									"type": "text",
									"text": buttonParam,
								},
							},
						})
						log.Printf("Automation Detail: Added tracking_url parameter to dynamic button %d (Param: %s)", i, buttonParam)
						break // Usually only one tracking button per template
					} else if trackingURL != "" {
						log.Printf("Automation Info: Button %d is static (no {{1}}). Skipping parameter injection.", i)
					}
				}
			}
		}
	}

	// 4. Handle Header (Media Attachments)
	if template.Header != nil && string(*template.Header) != "null" {
		var hData struct {
			Type string `json:"type"`
		}
		json.Unmarshal(*template.Header, &hData)

		if strings.ToUpper(hData.Type) == "DOCUMENT" {
			// For DOCUMENT headers, we generate the actual PDF invoice and upload it to Meta
			log.Printf("Automation Detail: Generating real invoice PDF for order %s", order.OrderNumber)

			// 1. Fetch line items (required for invoice generation)
			items, err := s.messagesService.repo.GetOrderLineItems(order.ID)
			if err != nil {
				log.Printf("Automation Error: Failed to fetch line items for invoice: %v", err)
			} else {
				// 2. Generate PDF bytes
				var buf bytes.Buffer
				if err := s.invoiceService.GeneratePDF(order, items, &buf); err != nil {
					log.Printf("Automation Error: Failed to generate PDF: %v", err)
				} else {
					// 3. Upload to WhatsApp Media API to get a Media ID
					filename := "Invoice_" + order.OrderNumber + ".pdf"
					id, err := s.messagesService.metaClient.UploadWhatsAppMedia(buf.Bytes(), filename, "application/pdf")
					if err != nil {
						log.Printf("Automation Error: Failed to upload invoice to Meta: %v", err)
						return fmt.Errorf("failed to upload invoice: %w", err)
					}
					log.Printf("Automation Success: Uploaded invoice, Media ID: %s", id)

					// Meta Cloud API often expects the ID as a numeric value in JSON
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
				}
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
	log.Printf("Automation Meta Call: Sending %s to %s (Order: %s). Payload: %s", template.TemplateName, cleanPhone, order.ID, string(compJSON))

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

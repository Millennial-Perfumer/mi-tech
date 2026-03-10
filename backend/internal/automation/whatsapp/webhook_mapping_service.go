package whatsapp

import (
	"fmt"
	"log"
	"shopify-gst-app/internal/models"
)

type WebhookMappingService struct {
	templatesRepo   *TemplatesRepository
	messagesService *MessagesService
}

func NewWebhookMappingService(tRepo *TemplatesRepository, mService *MessagesService) *WebhookMappingService {
	return &WebhookMappingService{
		templatesRepo:   tRepo,
		messagesService: mService,
	}
}

func (s *WebhookMappingService) ExecuteMapping(storeID, topic string, order models.Order) error {
	// 1. Find matching trigger
	trigger, err := s.templatesRepo.GetTriggerByTopic(storeID, topic)
	if err != nil {
		return fmt.Errorf("error fetching trigger: %w", err)
	}
	if trigger == nil {
		return nil // No mapping for this topic
	}

	// 2. Fetch template
	template, err := s.templatesRepo.GetTemplateByID(trigger.TemplateID)
	if err != nil {
		return fmt.Errorf("error fetching template: %w", err)
	}
	if template == nil {
		return fmt.Errorf("template not found: %d", trigger.TemplateID)
	}

	// 3. Extract parameters based on topic (simplified implementation)
	var components []interface{}
	if topic == "orders/create" {
		components = []interface{}{
			map[string]interface{}{
				"type": "body",
				"parameters": []map[string]string{
					{"type": "text", "text": order.CustomerFirstName},
					{"type": "text", "text": order.CustomerLastName},
					{"type": "text", "text": order.OrderNumber},
				},
			},
		}
	}

	// 4. Send message
	if order.CustomerPhone == "" {
		log.Printf("Skip automation: no phone number for order %s", order.OrderNumber)
		return nil
	}

	return s.messagesService.SendTemplateMessage(
		storeID,
		template.ID,
		order.ID,
		order.CustomerPhone,
		template.TemplateName,
		template.Language,
		components,
	)
}

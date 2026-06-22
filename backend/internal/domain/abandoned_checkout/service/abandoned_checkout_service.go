package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	acDto "mi-tech/internal/domain/abandoned_checkout/dto"
	acEntity "mi-tech/internal/domain/abandoned_checkout/entity"
	acRepo "mi-tech/internal/domain/abandoned_checkout/repository"
	communicationEntity "mi-tech/internal/domain/communication/entity"
	communicationRepo "mi-tech/internal/domain/communication/repository"
	communicationService "mi-tech/internal/domain/communication/service"
	"mi-tech/internal/shared/config"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var templateParamRegex = regexp.MustCompile(`\{\{(\d+)\}\}`)

type AbandonedCheckoutService interface {
	ProcessCheckoutWebhook(ctx context.Context, storeID string, checkout acEntity.AbandonedCheckout) error
	MarkCheckoutCompleted(ctx context.Context, checkoutToken, checkoutID, orderID string) error
	ProcessRecoveryQueue(ctx context.Context) error
	ListCheckouts(ctx context.Context, storeID string, page, limit int, search, status, startDate, endDate string) ([]acEntity.AbandonedCheckout, int64, error)
	TriggerManualRecovery(ctx context.Context, storeID string, id int) error
	GetAnalytics(ctx context.Context, storeID string, startDate, endDate string) (*acDto.AbandonedCheckoutAnalyticsResponse, error)
	DeleteCheckout(ctx context.Context, storeID string, id int) error
}

type abandonedCheckoutService struct {
	repo          acRepo.AbandonedCheckoutRepository
	templatesRepo communicationRepo.TemplatesRepository
	messagesServ  *communicationService.MessagesService
	settings      *config.SettingsProvider
}

func NewAbandonedCheckoutService(
	repo acRepo.AbandonedCheckoutRepository,
	templatesRepo communicationRepo.TemplatesRepository,
	messagesServ *communicationService.MessagesService,
	settings *config.SettingsProvider,
) AbandonedCheckoutService {
	return &abandonedCheckoutService{
		repo:          repo,
		templatesRepo: templatesRepo,
		messagesServ:  messagesServ,
		settings:      settings,
	}
}

func (s *abandonedCheckoutService) ListCheckouts(ctx context.Context, storeID string, page, limit int, search, status, startDate, endDate string) ([]acEntity.AbandonedCheckout, int64, error) {
	return s.repo.List(ctx, storeID, page, limit, search, status, startDate, endDate)
}

func (s *abandonedCheckoutService) TriggerManualRecovery(ctx context.Context, storeID string, id int) error {
	ac, err := s.repo.GetByID(ctx, storeID, id)
	if err != nil {
		return fmt.Errorf("checkout not found: %w", err)
	}

	if ac.Completed {
		return fmt.Errorf("checkout is already completed")
	}

	s.processSingleCheckout(ctx, *ac)
	return nil
}

func (s *abandonedCheckoutService) ProcessCheckoutWebhook(ctx context.Context, storeID string, ac acEntity.AbandonedCheckout) error {
	ac.StoreID = storeID
	ac.Completed = false
	ac.RecoveryStatus = "PENDING"
	ac.RecoveryAttempts = 0
	return s.repo.Upsert(ctx, &ac)
}

func (s *abandonedCheckoutService) MarkCheckoutCompleted(ctx context.Context, checkoutToken, checkoutID, orderID string) error {
	return s.repo.MarkCompleted(ctx, checkoutToken, checkoutID, orderID)
}

func (s *abandonedCheckoutService) ProcessRecoveryQueue(ctx context.Context) error {
	// threshold based on settings config
	delayMins := s.settings.GetAbandonedCheckoutDelayMinutes()
	threshold := time.Now().Add(-time.Duration(delayMins) * time.Minute)
	checkouts, err := s.repo.GetPendingForRecovery(ctx, threshold, 50)
	if err != nil {
		return fmt.Errorf("failed to fetch pending checkouts: %w", err)
	}

	if len(checkouts) == 0 {
		return nil
	}

	log.Printf("Abandoned Checkout Recovery: Found %d checkouts to process", len(checkouts))

	for _, ac := range checkouts {
		s.processSingleCheckout(ctx, ac)
	}

	return nil
}

func (s *abandonedCheckoutService) processSingleCheckout(ctx context.Context, ac acEntity.AbandonedCheckout) {
	// First mark as PROCESSING to lock
	err := s.repo.UpdateRecoveryStatus(ctx, ac.ID, "PROCESSING", ac.RecoveryAttempts+1, "", nil)
	if err != nil {
		log.Printf("Abandoned Checkout Recovery Error: Failed to mark checkout %d as PROCESSING: %v", ac.ID, err)
		return
	}

	// Fetch trigger for topic "checkouts/abandoned"
	trigger, err := s.templatesRepo.GetTriggerByTopic(ac.StoreID, "checkouts/abandoned")
	if err != nil || trigger == nil || !trigger.Enabled {
		log.Printf("Abandoned Checkout Recovery Info: No active trigger for checkouts/abandoned in store %s. Cancelling checkout %d recovery.", ac.StoreID, ac.ID)
		_ = s.repo.UpdateRecoveryStatus(ctx, ac.ID, "CANCELLED", ac.RecoveryAttempts, "No active trigger configured", nil)
		return
	}

	template, err := s.templatesRepo.GetTemplateByID(trigger.TemplateID)
	if err != nil || template == nil {
		log.Printf("Abandoned Checkout Recovery Error: Template %d not found for checkout %d: %v", trigger.TemplateID, ac.ID, err)
		_ = s.repo.UpdateRecoveryStatus(ctx, ac.ID, "FAILED", ac.RecoveryAttempts, "Template not found", nil)
		return
	}

	// Build Meta Cloud API parameters
	components := s.buildComponents(template, ac)

	phone := communicationService.SanitizePhoneNumber(ac.Phone)
	if len(phone) < 8 {
		log.Printf("Abandoned Checkout Recovery Info: Invalid phone number '%s' for checkout %d. Cancelling.", ac.Phone, ac.ID)
		_ = s.repo.UpdateRecoveryStatus(ctx, ac.ID, "CANCELLED", ac.RecoveryAttempts, "Invalid phone number", nil)
		return
	}

	// Check if customer completed any orders after this checkout was abandoned to prevent duplicate recovery messages
	hasRecentOrder, err := s.repo.CheckRecentOrders(ctx, ac.Phone, ac.Email, ac.AbandonedAt)
	if err == nil && hasRecentOrder {
		log.Printf("Abandoned Checkout Recovery Info: Customer already placed an order after abandoning checkout %d. Cancelling recovery.", ac.ID)
		_ = s.repo.UpdateRecoveryStatus(ctx, ac.ID, "CANCELLED", ac.RecoveryAttempts, "Customer completed purchase recently", nil)
		return
	}


	// Send message
	log.Printf("Abandoned Checkout Recovery: Dispatching template %s to %s for checkout %d", template.TemplateName, phone, ac.ID)
	err = s.messagesServ.SendTemplateMessage(
		ac.StoreID,
		template.ID,
		0, // 0 indicates checkout recovery, not tied to a completed order yet
		phone,
		template.TemplateName,
		template.Language,
		components,
	)

	now := time.Now()
	if err != nil {
		log.Printf("Abandoned Checkout Recovery Error: Failed to send template message for checkout %d: %v", ac.ID, err)
		_ = s.repo.UpdateRecoveryStatus(ctx, ac.ID, "FAILED", ac.RecoveryAttempts, err.Error(), nil)
	} else {
		log.Printf("Abandoned Checkout Recovery Success: Message sent for checkout %d", ac.ID)
		_ = s.repo.UpdateRecoveryStatus(ctx, ac.ID, "SENT", ac.RecoveryAttempts, "", &now)
	}
}

func (s *abandonedCheckoutService) buildComponents(template *communicationEntity.AutomationTemplate, ac acEntity.AbandonedCheckout) []interface{} {
	var mappings map[string]string
	if template.VariableMappings != nil {
		json.Unmarshal(*template.VariableMappings, &mappings)
	}
	if mappings == nil {
		mappings = make(map[string]string)
	}

	var components []interface{}

	// Body mapping
	requiredCount := s.countRequiredParams(template.Body)
	if requiredCount > 0 {
		var bodyParams []map[string]string
		for i := 1; i <= requiredCount; i++ {
			mapKey := fmt.Sprintf("body_text_0_{{%d}}", i)
			fieldToMap := mappings[mapKey]
			val := s.resolveCheckoutVariable(fieldToMap, ac)

			// Legacy fallback matching template parameters
			if val == "" {
				if i == 1 {
					val = ac.CustomerName
				} else if i == 2 {
					val = ac.CheckoutURL
				}
			}
			bodyParams = append(bodyParams, map[string]string{"type": "text", "text": val})
		}
		components = append(components, map[string]interface{}{
			"type":       "body",
			"parameters": bodyParams,
		})
	}

	// Buttons mapping
	if template.Buttons != nil && string(*template.Buttons) != "null" {
		var buttons []map[string]interface{}
		if err := json.Unmarshal(*template.Buttons, &buttons); err == nil {
			for i, btn := range buttons {
				if btn["type"] == "visit_website" {
					urlVal, _ := btn["url"].(string)
					if strings.Contains(urlVal, "{{1}}") || strings.Contains(strings.ToLower(urlVal), "%7b%7b1%7d%7d") {
						mapKey := fmt.Sprintf("button_url_%d_{{1}}", i)
						fieldToMap := mappings[mapKey]
						val := s.resolveCheckoutVariable(fieldToMap, ac)

						if val == "" {
							val = ac.CheckoutURL
						}

						// If the value is a full URL, strip the scheme and host to match Meta's dynamic URL logic
						if strings.HasPrefix(val, "http://") || strings.HasPrefix(val, "https://") {
							if parsedURL, err := url.Parse(val); err == nil {
								val = parsedURL.Path
								if parsedURL.RawQuery != "" {
									val += "?" + parsedURL.RawQuery
								}
								if parsedURL.Fragment != "" {
									val += "#" + parsedURL.Fragment
								}
								val = strings.TrimPrefix(val, "/")
							}
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
			}
		}
	}

	return components
}

func (s *abandonedCheckoutService) resolveCheckoutVariable(field string, ac acEntity.AbandonedCheckout) string {
	switch field {
	case "customer_name":
		if ac.CustomerName != "" {
			return ac.CustomerName
		}
		return "Customer"
	case "checkout_url":
		return ac.CheckoutURL
	case "total_price":
		return fmt.Sprintf("%.2f", ac.TotalPrice)
	case "currency":
		return ac.Currency
	default:
		return ""
	}
}

func (s *abandonedCheckoutService) countRequiredParams(body string) int {
	matches := templateParamRegex.FindAllStringSubmatch(body, -1)
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

func (s *abandonedCheckoutService) GetAnalytics(ctx context.Context, storeID string, startDate, endDate string) (*acDto.AbandonedCheckoutAnalyticsResponse, error) {
	return s.repo.GetAnalytics(ctx, storeID, startDate, endDate)
}

func (s *abandonedCheckoutService) DeleteCheckout(ctx context.Context, storeID string, id int) error {
	return s.repo.Delete(ctx, storeID, id)
}

package whatsapp

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"mi-tech/internal/config"
	"mi-tech/internal/entity"
	"mi-tech/internal/service"
)

// FeedbackScheduler coordinates scanning and triggering automatic feedback requests.
type FeedbackScheduler struct {
	settingsProvider  *config.SettingsProvider
	orderService      *service.OrderService
	mappingService    *WebhookMappingService
	templatesRepo     *TemplatesRepository
	lastTriggeredDate string
	mu                sync.Mutex
}

// NewFeedbackScheduler creates a new instance of FeedbackScheduler.
func NewFeedbackScheduler(
	settingsProvider *config.SettingsProvider,
	orderService *service.OrderService,
	mappingService *WebhookMappingService,
	templatesRepo *TemplatesRepository,
) *FeedbackScheduler {
	return &FeedbackScheduler{
		settingsProvider: settingsProvider,
		orderService:     orderService,
		mappingService:   mappingService,
		templatesRepo:    templatesRepo,
	}
}

// Start runs the periodic check loop. It should be run in a goroutine.
func (s *FeedbackScheduler) Start(ctx context.Context) {
	log.Println("FeedbackScheduler: Starting background worker")
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	// Perform initial check on startup
	s.checkAndTrigger(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Println("FeedbackScheduler: Background worker stopping due to context cancellation")
			return
		case <-ticker.C:
			s.checkAndTrigger(ctx)
		}
	}
}

// checkAndTrigger checks settings and time to determine if feedback sending should run.
func (s *FeedbackScheduler) checkAndTrigger(ctx context.Context) {
	if !s.settingsProvider.IsFeedbackAutoTriggerEnabled() {
		return
	}

	configTime := s.settingsProvider.GetFeedbackAutoTriggerTime()
	var hour, min int
	n, err := fmt.Sscanf(configTime, "%d:%d", &hour, &min)
	if err != nil || n != 2 || hour < 0 || hour > 23 || min < 0 || min > 59 {
		log.Printf("FeedbackScheduler: Invalid configured trigger time '%s', defaulting to 10:00. Error: %v", configTime, err)
		hour = 10
		min = 0
	}

	now := time.Now()
	currentDate := now.Format("2006-01-02")

	s.mu.Lock()
	alreadyTriggered := s.lastTriggeredDate == currentDate
	s.mu.Unlock()

	if alreadyTriggered {
		return
	}

	if now.Hour() == hour && now.Minute() == min {
		log.Printf("FeedbackScheduler: Time matched (%02d:%02d), executing daily auto feedback trigger", hour, min)

		s.mu.Lock()
		s.lastTriggeredDate = currentDate
		s.mu.Unlock()

		go s.executeFeedbackScanAndSend(ctx)
	}
}

// executeFeedbackScanAndSend fetches candidates, loads templates, and bulk-sends WhatsApp requests.
func (s *FeedbackScheduler) executeFeedbackScanAndSend(ctx context.Context) {
	log.Println("FeedbackScheduler: Starting automated scan for feedback candidates...")

	delayMins := s.settingsProvider.GetFeedbackAutomationDelayMinutes()
	orders, err := s.orderService.GetOrdersForFeedback(delayMins)
	if err != nil {
		log.Printf("FeedbackScheduler: Error scanning for candidates: %v", err)
		return
	}

	templateName := s.settingsProvider.Get("feedback_whatsapp_template_name")
	if templateName == "" {
		log.Println("FeedbackScheduler: WhatsApp feedback template name is not configured, skipping auto send")
		return
	}

	template, err := s.templatesRepo.GetTemplateByName(config.StoreIDShopify, templateName)
	if err != nil {
		log.Printf("FeedbackScheduler: Error fetching WhatsApp template '%s': %v", templateName, err)
		return
	}
	if template == nil {
		log.Printf("FeedbackScheduler: WhatsApp template '%s' not found, skipping auto send", templateName)
		return
	}

	var eligibleOrders []entity.Order
	for _, order := range orders {
		if order.DeliveredAt == nil {
			continue
		}
		customerPhone := entity.DerefStr(order.CustomerPhone)
		if customerPhone == "" {
			continue
		}
		eligibleOrders = append(eligibleOrders, order)
	}

	if len(eligibleOrders) == 0 {
		log.Println("FeedbackScheduler: No eligible orders found for feedback sending today")
		return
	}

	log.Printf("FeedbackScheduler: Found %d eligible orders. Initiating auto feedback send...", len(eligibleOrders))

	var wg sync.WaitGroup
	sem := make(chan struct{}, 5) // Limit concurrency to 5 parallel sends

	for _, o := range eligibleOrders {
		select {
		case <-ctx.Done():
			log.Println("FeedbackScheduler: Execution cancelled during bulk sending")
			return
		default:
		}

		wg.Add(1)
		go func(order entity.Order) {
			defer wg.Done()

			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				return
			}

			// Fetch full order entity to ensure line items/mappings are resolved correctly
			fullOrder, err := s.orderService.GetOrderEntity(order.ID)
			if err != nil {
				log.Printf("FeedbackScheduler: Error fetching full order %d: %v", order.ID, err)
				return
			}

			storeID := config.StoreIDShopify
			err = s.mappingService.ExecuteManualSend(storeID, template.ID, fullOrder)
			if err != nil {
				log.Printf("FeedbackScheduler: Failed to send feedback for order %d: %v", order.ID, err)
				return
			}

			// Update feedback status to Sent (2)
			_ = s.orderService.UpdateFeedbackStatus(order.ID, 2)
			log.Printf("FeedbackScheduler: Successfully sent feedback request for order %d", order.ID)
		}(o)
	}

	wg.Wait()
	log.Println("FeedbackScheduler: Finished daily auto feedback sending")
}

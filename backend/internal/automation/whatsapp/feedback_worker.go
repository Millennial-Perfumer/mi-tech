package whatsapp

import (
	"log"
	"mi-tech/internal/service"
	"time"
)

type FeedbackWorker struct {
	orderService   *service.OrderService
	mappingService *WebhookMappingService
	interval       time.Duration
}

func NewFeedbackWorker(orderService *service.OrderService, mappingService *WebhookMappingService) *FeedbackWorker {
	return &FeedbackWorker{
		orderService:   orderService,
		mappingService: mappingService,
		interval:       1 * time.Hour, // Hardcoded per user request
	}
}

func (w *FeedbackWorker) Start() {
	log.Printf("Feedback Automation Worker: Starting with interval %v", w.interval)
	ticker := time.NewTicker(w.interval)
	
	// Run once immediately on start
	go w.process()

	for range ticker.C {
		w.process()
	}
}

func (w *FeedbackWorker) process() {
	log.Println("Feedback Automation Worker: Checking for orders ready for feedback...")
	
	// Fetch delay from settings
	delayMins := w.mappingService.settingsProvider.GetFeedbackAutomationDelayMinutes()

	// 1. Get orders delivered delayMins minutes ago with 'pending' feedback status
	orders, err := w.orderService.GetOrdersForFeedback(delayMins)
	if err != nil {
		log.Printf("Feedback Automation Worker: Error fetching orders: %v", err)
		return
	}

	if len(orders) == 0 {
		log.Println("Feedback Automation Worker: No orders require feedback request at this time.")
		return
	}

	log.Printf("Feedback Automation Worker: Found %d orders ready for feedback request.", len(orders))

	for _, order := range orders {
		// 2. Trigger feedback flow template
		// Topic: orders/feedback_request - user must ensure a trigger/template exists for this topic
		err := w.mappingService.ExecuteMapping(order.SourceID, "orders/feedback_request", order)
		if err != nil {
			log.Printf("Feedback Automation Worker: Failed to send feedback for order %d: %v", order.ID, err)
			continue
		}

		// 3. Update status to 'sent' (Status ID: 2)
		if err := w.orderService.UpdateFeedbackStatus(order.ID, 2); err != nil {
			log.Printf("Feedback Automation Worker: Failed to update status for order %d: %v", order.ID, err)
		}
	}
}

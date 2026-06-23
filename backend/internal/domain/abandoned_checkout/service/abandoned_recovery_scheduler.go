package service

import (
	"context"
	"log"
	"time"
)

type AbandonedRecoveryScheduler struct {
	acService AbandonedCheckoutService
}

func NewAbandonedRecoveryScheduler(acService AbandonedCheckoutService) *AbandonedRecoveryScheduler {
	return &AbandonedRecoveryScheduler{
		acService: acService,
	}
}

func (s *AbandonedRecoveryScheduler) Start(ctx context.Context) {
	log.Println("AbandonedRecoveryScheduler: Starting background worker")
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	// Initial trigger after 1 minute of startup to let systems settle
	go func() {
		select {
		case <-ctx.Done():
			return
		case <-time.After(1 * time.Minute):
			if err := s.acService.ProcessRecoveryQueue(ctx); err != nil {
				log.Printf("AbandonedRecoveryScheduler Error: Initial trigger failed: %v", err)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			log.Println("AbandonedRecoveryScheduler: Background worker stopping due to context cancellation")
			return
		case <-ticker.C:
			if err := s.acService.ProcessRecoveryQueue(ctx); err != nil {
				log.Printf("AbandonedRecoveryScheduler Error: Failed to run recovery queue: %v", err)
			}
		}
	}
}

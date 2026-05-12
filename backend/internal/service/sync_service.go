package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"mi-tech/internal/client/shopify"
	"mi-tech/internal/entity"
	"mi-tech/internal/mapper"
	"mi-tech/internal/repository"
)

// SyncService orchestrates fetching orders from Shopify and persisting them.
type SyncService struct {
	shopifyClient   *shopify.Client
	orderRepo       repository.OrderRepository
	customerService *CustomerService
	orchestrator    *SyncOrchestrator
}

// NewSyncService creates a new SyncService.
func NewSyncService(shopifyClient *shopify.Client, orderRepo repository.OrderRepository, customerService *CustomerService, orchestrator *SyncOrchestrator) *SyncService {
	return &SyncService{
		shopifyClient:   shopifyClient,
		orderRepo:       orderRepo,
		customerService: customerService,
		orchestrator:    orchestrator,
	}
}

// Sync fetches new/updated orders from Shopify and upserts them into the database.
func (s *SyncService) Sync(startTime *time.Time, endTime *time.Time) (int, error) {
	if s.shopifyClient == nil {
		return 0, fmt.Errorf("shopify client not initialized")
	}

	var start, end time.Time

	if startTime != nil {
		start = *startTime
	} else {
		start = s.getLastSyncTime()
	}

	if endTime != nil {
		end = *endTime
	} else {
		end = time.Now()
	}

	log.Printf("Starting Shopify order sync fetching orders updated between %s and %s...", 
		start.Format(time.RFC3339), end.Format(time.RFC3339))

	shopifyOrders, err := s.shopifyClient.FetchOrders(start, end)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch from Shopify: %w", err)
	}

	if len(shopifyOrders) == 0 {
		log.Printf("No new or updated orders found from Shopify between %s and %s", 
			start.Format(time.RFC3339), end.Format(time.RFC3339))
		return 0, nil
	}

	log.Printf("Fetched %d orders. Proceeding to sink to database.", len(shopifyOrders))

	// Map GraphQL DTOs → Entities
	var orderEntities []entity.Order
	for _, so := range shopifyOrders {
		order := mapper.GraphQLOrderToEntity(so)
		order.LineItems = mapper.GraphQLLineItemsToEntities(order.ID, so.LineItems)
		orderEntities = append(orderEntities, order)
	}

	affectedIDs, err := s.orderRepo.UpsertBatch(orderEntities)
	if err != nil {
		return 0, fmt.Errorf("failed to upsert batch: %w", err)
	}

	// Trigger global sync for all affected inventory items in batch
	if s.orchestrator != nil && len(affectedIDs) > 0 {
		_ = s.orchestrator.GlobalSyncBatch(context.Background(), affectedIDs, "shopify")
	}

	// Update customer metadata in batch to avoid N+1 queries
	if s.customerService != nil {
		if err := s.customerService.UpdateFromOrdersBatch(context.Background(), orderEntities); err != nil {
			log.Printf("Sync: failed to update customer metadata in batch: %v", err)
		}
	}

	log.Printf("Successfully synchronized %d orders and their items into PostgreSQL.", len(orderEntities))
	return len(orderEntities), nil
}

// ResetOnly wipes all orders locally but does not perform a resync.
func (s *SyncService) ResetOnly() error {
	return s.orderRepo.TruncateAll()
}

// ResetAndSync wipes all orders locally and performs a full sync.
func (s *SyncService) ResetAndSync() (int, error) {
	if s.shopifyClient == nil {
		return 0, fmt.Errorf("shopify client not initialized")
	}
	if err := s.ResetOnly(); err != nil {
		return 0, err
	}
	return s.Sync(nil, nil)
}

func (s *SyncService) getLastSyncTime() time.Time {
	// Fall back to January 1st 2026 baseline
	baseline, _ := time.Parse(time.RFC3339, "2026-01-01T00:00:00Z")
	return baseline
}

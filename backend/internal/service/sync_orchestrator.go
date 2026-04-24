package service

import (
	"context"
	"log"
	"mi-tech/internal/client/amazon"
	"mi-tech/internal/client/shopify"
	"mi-tech/internal/entity"
	"mi-tech/internal/repository"
)

// SyncOrchestrator takes the lead on cross-platform stock consistency.
// It ensures that when stock changes in the MI App, it is pushed to all connected platforms.
type SyncOrchestrator struct {
	inventoryRepo repository.InventoryRepository
	shopifyClient *shopify.Client
	amazonClient  *amazon.Client
}

func NewSyncOrchestrator(
	inventoryRepo repository.InventoryRepository,
	shopifyClient *shopify.Client,
	amazonClient *amazon.Client,
) *SyncOrchestrator {
	return &SyncOrchestrator{
		inventoryRepo: inventoryRepo,
		shopifyClient: shopifyClient,
		amazonClient:  amazonClient,
	}
}

// GlobalSync updates all platforms for a specific internal inventory item.
// sourcePlatform allows us to avoid "echo" updates back to the source.
func (s *SyncOrchestrator) GlobalSync(ctx context.Context, itemID int, sourcePlatform string) error {
	item, err := s.inventoryRepo.GetItemByID(itemID)
	if err != nil {
		return err
	}

	log.Printf("SyncOrchestrator: Triggering global sync for SKU %s (Qty: %d)", item.MISKU, item.CurrentStock)

	for _, m := range item.Mappings {
		if m.Platform == sourcePlatform {
			continue // Don't push back to the platform that triggered the change
		}

		switch m.Platform {
		case "shopify":
			if m.ExternalVariantID != nil {
				log.Printf("SyncOrchestrator: Pushing stock update to Shopify for inventory item %s", *m.ExternalVariantID)
				locationID := s.shopifyClient.GetLocationID()
				if locationID == "" {
					discoveredID, err := s.shopifyClient.DiscoverPrimaryLocationID(ctx)
					if err != nil {
						log.Printf("SyncOrchestrator Warning: Shopify location ID not configured and discovery failed: %v", err)
						continue
					}
					locationID = discoveredID
					log.Printf("SyncOrchestrator: Auto-discovered Shopify location ID: %s", locationID)
				}

				if locationID != "" {
					err := s.shopifyClient.AdjustInventoryLevel(*m.ExternalVariantID, locationID, item.CurrentStock)
					if err != nil {
						log.Printf("SyncOrchestrator Warning: Shopify sync failed for %s: %v", *m.ExternalVariantID, err)
					}
				}
			}
		case "amazon":
			log.Printf("SyncOrchestrator: Pushing stock update to Amazon for SKU %s", m.ExternalSKU)
			err := s.amazonClient.UpdateInventory(m.ExternalSKU, item.CurrentStock)
			if err != nil {
				log.Printf("SyncOrchestrator Warning: Amazon sync failed for %s: %v", m.ExternalSKU, err)
			}
		}
	}

	return nil
}

// AdjustStock adjusts stock internally and then triggers a global sync.
func (s *SyncOrchestrator) AdjustStock(ctx context.Context, itemID int, delta int, sourcePlatform string, reason string, externalOrderID *string) error {
	err := s.inventoryRepo.AdjustStock(itemID, delta)
	if err != nil {
		return err
	}

	// Log the movement
	s.inventoryRepo.LogAdjustment(&entity.InventoryLog{
		InventoryItemID: itemID,
		Delta:           delta,
		Reason:          reason,
		Platform:        sourcePlatform,
		ExternalOrderID: externalOrderID,
	})

	return s.GlobalSync(ctx, itemID, sourcePlatform)
}

// UpdateStock sets the absolute stock internally and then triggers a global sync.
func (s *SyncOrchestrator) UpdateStock(ctx context.Context, itemID int, val int, sourcePlatform string, reason string) error {
	// For absolute updates, we need to know the delta for logging
	item, _ := s.inventoryRepo.GetItemByID(itemID)
	delta := val - item.CurrentStock

	err := s.inventoryRepo.UpdateStockCount(itemID, val)
	if err != nil {
		return err
	}

	// Log the movement
	s.inventoryRepo.LogAdjustment(&entity.InventoryLog{
		InventoryItemID: itemID,
		Delta:           delta,
		Reason:          reason,
		Platform:        sourcePlatform,
	})

	return s.GlobalSync(ctx, itemID, sourcePlatform)
}

// AdjustStockByPlatformSKU resolves an internal item by its platform-specific SKU and then adjusts its stock.
func (s *SyncOrchestrator) AdjustStockByPlatformSKU(ctx context.Context, platform string, sku string, delta int, reason string, externalOrderID *string) error {
	invItem, err := s.inventoryRepo.GetItemByPlatformSKU(platform, sku)
	if err != nil {
		return err
	}
	return s.AdjustStock(ctx, invItem.ID, delta, platform, reason, externalOrderID)
}

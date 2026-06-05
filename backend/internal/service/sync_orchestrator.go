package service

import (
	"context"
	"log"
	"mi-tech/internal/client/amazon"
	"mi-tech/internal/client/shopify"
	"mi-tech/internal/config"
	"mi-tech/internal/entity"
	"mi-tech/internal/repository"
	"os"

	"gorm.io/gorm"
)

// SyncOrchestrator takes the lead on cross-platform stock consistency.
// It ensures that when stock changes in the MI App, it is pushed to all connected platforms.
type SyncOrchestrator struct {
	inventoryRepo repository.InventoryRepository
	shopifyClient *shopify.Client
	amazonClient  *amazon.Client
	settings      *config.SettingsProvider
}

func NewSyncOrchestrator(
	inventoryRepo repository.InventoryRepository,
	shopifyClient *shopify.Client,
	amazonClient *amazon.Client,
	settings *config.SettingsProvider,
) *SyncOrchestrator {
	return &SyncOrchestrator{
		inventoryRepo: inventoryRepo,
		shopifyClient: shopifyClient,
		amazonClient:  amazonClient,
		settings:      settings,
	}
}

func (s *SyncOrchestrator) WithTx(tx *gorm.DB) *SyncOrchestrator {
	if tx == nil {
		return s
	}
	return &SyncOrchestrator{
		inventoryRepo: s.inventoryRepo.WithTx(tx),
		shopifyClient: s.shopifyClient,
		amazonClient:  s.amazonClient,
		settings:      s.settings,
	}
}

func (s *SyncOrchestrator) IsSyncAllowed() bool {
	// 1. Check environment variable override (highest priority for local dev)
	if os.Getenv("DISABLE_INVENTORY_SYNC") == "true" {
		return false
	}

	// 2. Check database configuration
	if s.settings != nil && !s.settings.IsInventorySyncEnabled() {
		return false
	}

	return true
}

// GlobalSync updates all platforms for a specific internal inventory item.
// sourcePlatform allows us to avoid "echo" updates back to the source.
func (s *SyncOrchestrator) GlobalSync(ctx context.Context, itemID int, sourcePlatform string) error {
	item, err := s.inventoryRepo.GetItemByID(itemID)
	if err != nil {
		return err
	}

	log.Printf("SyncOrchestrator: Triggering global sync for SKU %s (Qty: %d)", item.MISKU, item.CurrentStock)

	if !s.IsSyncAllowed() {
		log.Printf("SyncOrchestrator: Global sync is DISABLED via configuration or environment")
		return nil
	}

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

// AdjustStockInternal performs only the local DB updates and logging.
// It does NOT trigger external synchronization.
func (s *SyncOrchestrator) AdjustStockInternal(ctx context.Context, itemID int, delta int, sourcePlatform string, reason string, externalOrderID *string) error {
	// ALWAYS update local stock — this must happen regardless of sync setting
	err := s.inventoryRepo.AdjustStock(itemID, delta)
	if err != nil {
		return err
	}

	// Log the movement
	return s.inventoryRepo.LogAdjustment(&entity.InventoryLog{
		InventoryItemID: itemID,
		Delta:           delta,
		Reason:          reason,
		Platform:        sourcePlatform,
		ExternalOrderID: externalOrderID,
	})
}

// AdjustStock adjusts stock internally and then triggers a global sync.
func (s *SyncOrchestrator) AdjustStock(ctx context.Context, itemID int, delta int, sourcePlatform string, reason string, externalOrderID *string) error {
	if err := s.AdjustStockInternal(ctx, itemID, delta, sourcePlatform, reason, externalOrderID); err != nil {
		return err
	}

	// Only push to external platforms if sync is enabled
	if !s.IsSyncAllowed() {
		log.Printf("SyncOrchestrator: Local stock adjusted for item %d (delta: %d), but external sync is DISABLED", itemID, delta)
		return nil
	}

	return s.GlobalSync(ctx, itemID, sourcePlatform)
}

// UpdateStock sets the absolute stock internally and then triggers a global sync.
func (s *SyncOrchestrator) UpdateStock(ctx context.Context, itemID int, val int, sourcePlatform string, reason string) error {
	if !s.IsSyncAllowed() {
		log.Printf("SyncOrchestrator: Skipping internal stock update (Sync DISABLED)")
		return nil
	}

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

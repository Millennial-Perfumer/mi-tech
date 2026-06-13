package service

import (
	"context"
	"fmt"
	"log/slog"
	"mi-tech/internal/domain/inventory/entity"
	"mi-tech/internal/domain/inventory/repository"
	"mi-tech/internal/shared/config"
	"mi-tech/internal/shared/extclient/shopify"
	"strconv"
	"strings"
	"time"
)

type StockOrchestrator interface {
	AdjustStock(ctx context.Context, itemID int, delta int, sourcePlatform string, reason string, externalOrderID *string) error
	UpdateStock(ctx context.Context, itemID int, val int, sourcePlatform string, reason string) error
}

type AmazonPoller interface {
	SyncOrders(ctx context.Context, start, end *time.Time)
}

type InventoryService struct {
	repo          repository.InventoryRepository
	shopifyClient *shopify.Client
	orchestrator  StockOrchestrator
	settings      *config.SettingsProvider
	amzPoller     AmazonPoller
}

func NewInventoryService(repo repository.InventoryRepository, shopifyClient *shopify.Client, orchestrator StockOrchestrator, settings *config.SettingsProvider, amzPoller AmazonPoller) *InventoryService {
	return &InventoryService{
		repo:          repo,
		shopifyClient: shopifyClient,
		orchestrator:  orchestrator,
		settings:      settings,
		amzPoller:     amzPoller,
	}
}

// SuggestNextSKU finds the highest mi-XX and returns the next one.
func (s *InventoryService) SuggestNextSKU() (string, error) {
	maxSKU, err := s.repo.GetMaxMISKU()
	if err != nil {
		return "", err
	}

	if maxSKU == "" {
		return "mi-01", nil
	}

	// Extract the number from mi-XX
	parts := strings.Split(maxSKU, "-")
	if len(parts) != 2 {
		return "mi-01", nil
	}

	num, err := strconv.Atoi(parts[1])
	if err != nil {
		return "mi-01", nil
	}

	// Generate mi-01, mi-02, ..., mi-10, mi-100...
	nextNum := num + 1
	return fmt.Sprintf("mi-%02d", nextNum), nil
}

// CreateItem handles the manual creation of an internal product.
func (s *InventoryService) CreateItem(ctx context.Context, item *entity.InventoryItem) error {
	if item.MISKU == "" {
		next, err := s.SuggestNextSKU()
		if err != nil {
			return err
		}
		item.MISKU = next
	}
	return s.repo.CreateItem(item)
}

// UpdateItem handles partial updates to an internal product.
func (s *InventoryService) UpdateItem(ctx context.Context, item *entity.InventoryItem) error {
	existing, err := s.repo.GetItemByID(item.ID)
	if err != nil {
		return err
	}

	if item.MISKU != "" {
		existing.MISKU = item.MISKU
	}
	if item.Title != "" {
		existing.Title = item.Title
	}

	return s.repo.UpdateItem(&existing)
}

// MapProduct links an external SKU to an internal item.
func (s *InventoryService) MapProduct(ctx context.Context, internalItemID int, platform, externalSKU, variantID string) error {
	mapping := &entity.InventoryMapping{
		InventoryItemID:   internalItemID,
		Platform:          platform,
		ExternalSKU:       externalSKU,
		ExternalVariantID: &variantID,
	}
	return s.repo.CreateMapping(mapping)
}

// DeleteMapping removes a specific SKU mapping by ID.
func (s *InventoryService) DeleteMapping(ctx context.Context, id int) error {
	return s.repo.DeleteMapping(id)
}

// SyncShopifyProducts fetches all Shopify variants and stages them for sync.
func (s *InventoryService) SyncShopifyProducts(ctx context.Context) ([]entity.InventoryItem, error) {
	shopifyProducts, err := s.shopifyClient.FetchProducts()
	if err != nil {
		return nil, err
	}

	locationID := s.settings.GetShopifyLocationID()
	results := []entity.InventoryItem{}
	for i := range shopifyProducts {
		sp := &shopifyProducts[i]

		// Extract metafields
		var productDesc, productSpec string
		if sp.DescriptionMetafield != nil {
			productDesc = sp.DescriptionMetafield.Value
		}
		if sp.SpecificationMetafield != nil {
			productSpec = sp.SpecificationMetafield.Value
		}

		// Fallback to descriptionHtml if metafield is empty
		finalDesc := productDesc
		if finalDesc == "" {
			finalDesc = sp.DescriptionHtml
		}

		for _, v := range sp.Variants.Edges {
			invItemID := v.Node.InventoryItem.ID

			// Find stock for configured location, with global fallback
			availableStock := 0
			foundLocationMatch := false

			for _, lvl := range v.Node.InventoryItem.InventoryLevels.Edges {
				slog.Info("[DEBUG] Inspecting level", "loc_id", lvl.Node.Location.ID, "loc_name", lvl.Node.Location.Name)

				// Ultra-flexible match:
				// 1. Full GID match
				// 2. Numeric ID suffix match
				// 3. Name-based match (user provided the name in settings or explicitly)
				isMatch := (lvl.Node.Location.ID == locationID) ||
					strings.HasSuffix(lvl.Node.Location.ID, "/"+locationID) ||
					(strings.HasPrefix(locationID, "gid://") && strings.HasSuffix(locationID, "/"+lvl.Node.Location.ID)) ||
					(strings.TrimSpace(strings.ToLower(lvl.Node.Location.Name)) == strings.TrimSpace(strings.ToLower(locationID))) ||
					(lvl.Node.Location.Name == "Millennial Perfumer - WH") // Hardcoded fallback for the user's specific case

				locQty := 0
				for _, q := range lvl.Node.Quantities {
					if q.Name == "available" {
						locQty = q.Quantity
						break
					}
					if locQty == 0 && q.Quantity > 0 {
						locQty = q.Quantity
					}
				}

				if isMatch {
					availableStock = locQty
					foundLocationMatch = true
					slog.Info("[DEBUG] Location matched!", "id_or_name", locationID, "matched_name", lvl.Node.Location.Name, "qty", availableStock)
					break
				} else {
					// Summative fallback logic in case no primary location is found later
					availableStock += locQty
				}
			}

			// If no specific location matched, we show the global sum
			if !foundLocationMatch && availableStock > 0 {
				slog.Info("[DEBUG] No location match, showing global stock", "total", availableStock)
			}

			// Parse Shopify price
			var priceVal float64
			if v.Node.Price != "" {
				if parsedPrice, err := strconv.ParseFloat(v.Node.Price, 64); err == nil {
					priceVal = parsedPrice
				}
			}

			// Use a helper to ensure a fresh pointer for each variant
			item := entity.InventoryItem{
				Title:         fmt.Sprintf("%s - %s", sp.Title, v.Node.Title),
				Description:   stringPtr(finalDesc),
				Specification: stringPtr(productSpec),
				CurrentStock:  availableStock,
				Price:         priceVal,
				Mappings: []entity.InventoryMapping{
					{
						Platform:          "shopify",
						ExternalSKU:       v.Node.SKU,
						ExternalVariantID: &invItemID,
					},
				},
			}
			results = append(results, item)
		}
	}

	return results, nil
}

// PriceSyncStats represents results of variant price updates.
type PriceSyncStats struct {
	TotalProcessed int `json:"total_processed"`
	UpdatedCount   int `json:"updated_count"`
	SkippedCount   int `json:"skipped_count"`
	NotFoundCount  int `json:"not_found_count"`
}

// SyncShopifyPrices fetches live Shopify variants and backfills database inventory prices by SKU match.
func (s *InventoryService) SyncShopifyPrices(ctx context.Context) (PriceSyncStats, error) {
	stats := PriceSyncStats{}

	shopifyProducts, err := s.shopifyClient.FetchProducts()
	if err != nil {
		return stats, fmt.Errorf("failed to fetch products from shopify: %w", err)
	}

	for _, sp := range shopifyProducts {
		for _, v := range sp.Variants.Edges {
			stats.TotalProcessed++
			sku := v.Node.SKU
			priceStr := v.Node.Price

			if sku == "" {
				slog.Warn("Skipping Shopify variant with empty SKU in price sync", "variant_id", v.Node.ID, "title", v.Node.Title)
				stats.SkippedCount++
				continue
			}

			priceVal, err := strconv.ParseFloat(priceStr, 64)
			if err != nil {
				slog.Error("Failed to parse price in price sync", "sku", sku, "price", priceStr, "error", err)
				stats.SkippedCount++
				continue
			}

			// 1. Match via inventory_mappings
			item, err := s.repo.GetItemByPlatformSKU("shopify", sku)
			if err == nil {
				item.Price = priceVal
				if err := s.repo.UpdateItem(&item); err != nil {
					slog.Error("Failed to update item price via mapping in price sync", "sku", sku, "error", err)
				} else {
					slog.Info("Successfully updated price via mapping in price sync", "sku", sku, "price", priceVal)
					stats.UpdatedCount++
				}
				continue
			}

			// 2. Fallback: match via direct mi_sku
			items, err := s.repo.ListItems(sku)
			if err == nil && len(items) > 0 {
				var directMatched bool
				for _, item := range items {
					if strings.EqualFold(item.MISKU, sku) {
						item.Price = priceVal
						if err := s.repo.UpdateItem(&item); err == nil {
							slog.Info("Successfully updated price via direct SKU match in price sync", "sku", sku, "price", priceVal)
							stats.UpdatedCount++
							directMatched = true
							break
						}
					}
				}
				if directMatched {
					continue
				}
			}

			slog.Warn("No mapping or direct SKU match found for Shopify variant in price sync", "sku", sku)
			stats.NotFoundCount++
		}
	}

	return stats, nil
}

func stringPtr(s string) *string {
	return &s
}

// ClearAll wipes the warehouse and mappings.
func (s *InventoryService) ClearAll(ctx context.Context) error {
	return s.repo.DeleteAll()
}

// BulkImport handles massive imports with duplicate detection and sequential SKU generation.
func (s *InventoryService) BulkImport(ctx context.Context, items []entity.InventoryItem) error {
	if len(items) == 0 {
		return nil
	}

	// 1. Fetch existing mappings to detect collisions early
	existingMappings, err := s.repo.ListMappings()
	if err != nil {
		return fmt.Errorf("failed to load existing mappings: %w", err)
	}

	mappedSKUs := make(map[string]bool)
	for _, m := range existingMappings {
		mappedSKUs[strings.ToLower(m.ExternalSKU)] = true
	}

	// 2. Identify latest internal SKU for sequential numbering
	maxSKU, err := s.repo.GetMaxMISKU()
	if err != nil {
		return err
	}

	startNum := 0
	if maxSKU != "" {
		parts := strings.Split(maxSKU, "-")
		if len(parts) == 2 {
			if n, err := strconv.Atoi(parts[1]); err == nil {
				startNum = n
			}
		}
	}

	// 3. Process items and prepare for import
	var toCreate []entity.InventoryItem
	for i := range items {
		// Verify if the platform SKU is already mapped locally
		hasCollision := false
		for _, m := range items[i].Mappings {
			if mappedSKUs[strings.ToLower(m.ExternalSKU)] {
				hasCollision = true
				break
			}
		}

		if hasCollision {
			slog.Warn("Skipping bulk import for item because SKU is already mapped", "title", items[i].Title)
			continue
		}

		// Prevent duplicates within the same batch by updating the map
		for _, m := range items[i].Mappings {
			mappedSKUs[strings.ToLower(m.ExternalSKU)] = true
		}

		// Assign next internal SKU based on Shopify numerals if available
		if items[i].MISKU == "" {
			shopifySKU := ""
			for _, m := range items[i].Mappings {
				if strings.ToLower(m.Platform) == "shopify" {
					shopifySKU = m.ExternalSKU
					break
				}
			}

			if shopifySKU != "" {
				digits := ""
				for _, r := range shopifySKU {
					if r >= '0' && r <= '9' {
						digits += string(r)
					}
				}
				if digits != "" {
					items[i].MISKU = "mi-" + digits
				}
			}

			// Fallback to sequential if no shopify digits found
			if items[i].MISKU == "" {
				startNum++
				items[i].MISKU = fmt.Sprintf("mi-%02d", startNum)
			}
		}

		// Also create a 'pos' mapping for each item using its MISKU
		items[i].Mappings = append(items[i].Mappings, entity.InventoryMapping{
			Platform:    "pos",
			ExternalSKU: items[i].MISKU,
		})

		toCreate = append(toCreate, items[i])
	}

	if len(toCreate) == 0 {
		return nil
	}

	// 4. Batch Import using a single database roundtrip.
	if err := s.repo.BulkCreateItem(toCreate); err != nil {
		slog.Error("Bulk import failure", "err", err)
		return fmt.Errorf("failed to bulk import items: %w", err)
	}

	slog.Info("Bulk import completed", "total", len(items), "imported", len(toCreate))
	return nil
}

// GetInventoryDashboard returns all items for the UI.
func (s *InventoryService) GetInventoryDashboard(search string) ([]entity.InventoryItem, error) {
	return s.repo.ListItems(search)
}

// AdjustStock handles manual stock updates and triggers sync.
func (s *InventoryService) AdjustStock(id int, delta int) error {
	if s.orchestrator != nil {
		return s.orchestrator.AdjustStock(context.Background(), id, delta, "internal", "manual_adjustment", nil)
	}
	return s.repo.AdjustStock(id, delta)
}

// UpdateStockCount sets absolute stock and triggers sync.
func (s *InventoryService) UpdateStockCount(id int, val int) error {
	if s.orchestrator != nil {
		return s.orchestrator.UpdateStock(context.Background(), id, val, "internal", "manual_correction")
	}
	return s.repo.UpdateStockCount(id, val)
}

// GetLogs returns the stock movement history for an item.
func (s *InventoryService) GetLogs(itemID int) ([]entity.InventoryLog, error) {
	return s.repo.GetLogsByItemID(itemID)
}

// SyncAmazonOrders triggers an immediate poll of Amazon orders.
func (s *InventoryService) SyncAmazonOrders(ctx context.Context, start, end *time.Time) {
	if s.amzPoller != nil {
		go s.amzPoller.SyncOrders(ctx, start, end)
	}
}

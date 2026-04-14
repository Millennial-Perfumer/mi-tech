package service

import (
	"context"
	"fmt"
	"mi-tech/internal/client/shopify"
	"mi-tech/internal/entity"
	"mi-tech/internal/repository"
	"strconv"
	"strings"
)

type InventoryService struct {
	repo          repository.InventoryRepository
	shopifyClient *shopify.Client
}

func NewInventoryService(repo repository.InventoryRepository, shopifyClient *shopify.Client) *InventoryService {
	return &InventoryService{
		repo:          repo,
		shopifyClient: shopifyClient,
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

// SyncShopifyProducts fetches all Shopify variants and stages them for sync.
func (s *InventoryService) SyncShopifyProducts(ctx context.Context) ([]entity.InventoryItem, error) {
	shopifyProducts, err := s.shopifyClient.FetchProducts()
	if err != nil {
		return nil, err
	}

	var results []entity.InventoryItem
	for _, sp := range shopifyProducts {
		for _, v := range sp.Variants.Edges {
			item := entity.InventoryItem{
				Title:       fmt.Sprintf("%s - %s", sp.Title, v.Node.Title),
				Description: &sp.Description,
				Mappings: []entity.InventoryMapping{
					{
						Platform:          "shopify",
						ExternalSKU:       v.Node.SKU,
						ExternalVariantID: &v.Node.ID,
					},
				},
			}
			results = append(results, item)
		}
	}

	return results, nil
}

// GetInventoryDashboard returns all items for the UI.
func (s *InventoryService) GetInventoryDashboard(search string) ([]entity.InventoryItem, error) {
	return s.repo.ListItems(search)
}

// AdjustStock handles manual stock updates.
func (s *InventoryService) AdjustStock(id int, delta int) error {
	return s.repo.AdjustStock(id, delta)
}

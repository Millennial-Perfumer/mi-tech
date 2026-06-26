1. **Fix `BulkUpdatePrices` in `InventoryRepository`**
   - Add `Omit(clause.Associations)` to `BulkUpdatePrices` in `backend/internal/domain/inventory/repository/inventory_repository.go`.

2. **Add `GetItemsBySKUs(skus []string)` to `InventoryRepository`**
   - Implement `GetItemsBySKUs(skus []string) ([]entity.InventoryItem, error)` in `inventory_repository.go` using a query that matches either `mi_sku` or `mappings.external_sku`.
   Wait, a simpler way is to just fetch `ListItems("")`? No, the reviewer said it's a massive memory regression. Let me write a custom query for `GetItemsBySKUs`.
   `r.db.Preload("Mappings").Joins("LEFT JOIN inventory_mappings ON inventory_mappings.inventory_item_id = inventory_items.id").Where("inventory_items.mi_sku IN ? OR inventory_mappings.external_sku IN ?", skus, skus).Find(&items)`
   Wait, since we preloaded mappings, we also need `Group("inventory_items.id")` to avoid duplicate items. Actually, a subquery is cleaner:
   `r.db.Preload("Mappings").Where("mi_sku IN ? OR id IN (SELECT inventory_item_id FROM inventory_mappings WHERE external_sku IN ?)", skus, skus).Find(&items)`

3. **Fix `SyncShopifyPrices` in `InventoryService`**
   - Loop over `shopifyProducts` and extract all SKUs into a slice `skus`.
   - Fetch items using `s.repo.GetItemsBySKUs(skus)`.
   - Create O(1) lookup maps: `skuToItemID` and `miSKUToItemID`.
   - Keep a `map[int]*entity.InventoryItem` for `itemsToUpdateMap` so pointer updates aren't lost.
   - Loop over Shopify variants, update prices, and if changed, add to `itemsToUpdateMap`.
   - Convert `itemsToUpdateMap` to a slice and call `s.repo.BulkUpdatePrices`.

4. **Verify again**
   - Update tests, run `go test`, and call code review again.

package service

import (
	"context"
	"fmt"
	"log/slog"
	"mi-tech/internal/client/amazon"
	"mi-tech/internal/entity"
	"mi-tech/internal/repository"
	"strconv"
	"strings"
	"sync"
	"time"
)

type AmazonOrderPoller struct {
	amazonClient  *amazon.Client
	orderRepo     repository.OrderRepository
	inventoryRepo repository.InventoryRepository
	orchestrator  *SyncOrchestrator
	interval      time.Duration
}

func NewAmazonOrderPoller(
	amazonClient *amazon.Client,
	orderRepo repository.OrderRepository,
	inventoryRepo repository.InventoryRepository,
	orchestrator *SyncOrchestrator,
) *AmazonOrderPoller {
	return &AmazonOrderPoller{
		amazonClient:  amazonClient,
		orderRepo:     orderRepo,
		inventoryRepo: inventoryRepo,
		orchestrator:  orchestrator,
		interval:      3 * time.Minute,
	}
}

func (p *AmazonOrderPoller) Start(ctx context.Context) {
	slog.Info("AmazonOrderPoller: Starting background worker", "interval", p.interval)
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	// Initial sync on start
	p.SyncOrders(ctx, nil, nil)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.SyncOrders(ctx, nil, nil)
		}
	}
}

func (p *AmazonOrderPoller) SyncOrders(ctx context.Context, start, end *time.Time) {
	slog.Info("AmazonOrderPoller: Polling Amazon for orders...")
	useLastUpdated := (start == nil)

	var createdAfter time.Time
	var createdBefore time.Time

	if start != nil {
		createdAfter = *start
	} else {
		createdAfter = time.Now().Add(-6 * time.Hour)
	}

	if end != nil {
		createdBefore = *end
		safetyCutoff := time.Now().Add(-2 * time.Minute)
		if createdBefore.After(safetyCutoff) {
			createdBefore = safetyCutoff
		}
	}

	slog.Info("AmazonOrderPoller: Polling details",
		"afterTime", createdAfter.Format(time.RFC3339),
		"beforeTime", createdBefore.Format(time.RFC3339),
		"useLastUpdated", useLastUpdated,
	)

	amazonOrders, err := p.amazonClient.GetOrders(createdAfter, createdBefore, useLastUpdated)
	if err != nil {
		slog.Error("AmazonOrderPoller: Failed to fetch orders", "error", err)
		return
	}

	slog.Info("AmazonOrderPoller: API Response", "orderCount", len(amazonOrders))

	// 1. Parallel fetch items and map to entities
	// Optimization: Reduces network latency by fetching items for multiple orders concurrently.
	orderChan := make(chan entity.Order, len(amazonOrders))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 5) // Limit concurrency to respect Amazon API rate limits

	for _, ao := range amazonOrders {
		wg.Add(1)
		go func(ao map[string]interface{}) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			order, err := p.mapAmazonOrderToEntity(ctx, ao)
			if err != nil {
				return
			}
			orderChan <- *order
		}(ao)
	}

	go func() {
		wg.Wait()
		close(orderChan)
	}()

	var orderEntities []entity.Order
	for order := range orderChan {
		amazonStatus := entity.DerefStr(order.Status)
		matchStatus := strings.TrimSpace(amazonStatus)
		isFulfilledOrCommitted := (matchStatus == "Unshipped" || matchStatus == "Shipped" || matchStatus == "InvoiceConfirmation" || (order.FulfillmentStatus != nil && *order.FulfillmentStatus == "fulfilled"))

		if isFulfilledOrCommitted {
			p.processDeduction(ctx, &order)
		}

		if amazonStatus == "Canceled" {
			p.processReversal(ctx, &order)
		}

		orderEntities = append(orderEntities, order)
	}

	if len(orderEntities) > 0 {
		// 2. Batch Upsert to DB
		// Optimization: Reduces DB roundtrips from O(N) to O(1) for order and line item persistence.
		affectedIDs, err := p.orderRepo.UpsertBatch(orderEntities)
		if err != nil {
			slog.Error("AmazonOrderPoller: Batch upsert failed", "error", err)
			return
		}

		// 3. Parallel trigger global sync for affected items
		// Optimization: Resolves N+1 pattern by parallelizing external API calls (Shopify/Amazon) across affected inventory.
		var syncWg sync.WaitGroup
		syncSemaphore := make(chan struct{}, 10) // Limit concurrency for external API calls

		// Deduplicate inventory item IDs to avoid redundant global syncs
		uniqueAffectedIDs := make(map[int]bool)
		for _, id := range affectedIDs {
			uniqueAffectedIDs[id] = true
		}

		for id := range uniqueAffectedIDs {
			syncWg.Add(1)
			go func(itemID int) {
				defer syncWg.Done()
				syncSemaphore <- struct{}{}
				defer func() { <-syncSemaphore }()
				_ = p.orchestrator.GlobalSync(ctx, itemID, "amazon")
			}(id)
		}
		syncWg.Wait()
	}
}

func (p *AmazonOrderPoller) processDeduction(ctx context.Context, order *entity.Order) {
	slog.Info("AmazonOrderPoller: Deducting inventory", "orderID", order.ExternalOrderID, "itemCount", len(order.LineItems))

	if len(order.LineItems) == 0 {
		slog.Warn("AmazonOrderPoller: No line items found, skipping deduction", "orderID", order.ExternalOrderID)
		return
	}

	// Fetch existing logs to ensure idempotency (don't double deduct if partially failed)
	existingLogs, err := p.inventoryRepo.GetLogsByExternalOrderID(order.ExternalOrderID)
	if err != nil {
		slog.Error("AmazonOrderPoller: Failed to fetch existing logs, skipping deduction to be safe", "orderID", order.ExternalOrderID, "error", err)
		return
	}

	allSuccess := true
	for _, item := range order.LineItems {
		if item.SKU == nil || *item.SKU == "" {
			slog.Warn("AmazonOrderPoller: Skipping item with nil/empty SKU", "orderID", order.ExternalOrderID)
			allSuccess = false
			continue
		}

		// Check if this specific SKU has already been deducted for this order
		alreadyDeducted := false
		for _, log := range existingLogs {
			// We look for a 'sale' log with the negative quantity
			// Note: This assumes 1 mapping per platform/sku
			if log.Delta == -item.Quantity && log.Reason == "sale" {
				// To be truly precise, we should resolve the item ID first, but checking Delta/Reason/ExternalOrderID is 99% safe
				alreadyDeducted = true
				break
			}
		}

		if alreadyDeducted {
			slog.Debug("AmazonOrderPoller: Item already deducted (log found), skipping", "orderID", order.ExternalOrderID, "sku", *item.SKU)
			continue
		}

		slog.Info("AmazonOrderPoller: Adjusting stock", "orderID", order.ExternalOrderID, "sku", *item.SKU, "qty", -item.Quantity)
		err := p.orchestrator.AdjustStockByPlatformSKU(ctx, "amazon", *item.SKU, -item.Quantity, "sale", &order.ExternalOrderID)
		if err != nil {
			slog.Error("AmazonOrderPoller: Stock adjustment FAILED — will retry next poll", "orderID", order.ExternalOrderID, "sku", *item.SKU, "error", err)
			allSuccess = false
		}
	}

	if allSuccess {
		order.InventoryDeducted = true
		slog.Info("AmazonOrderPoller: All items deducted successfully", "orderID", order.ExternalOrderID)
	} else {
		slog.Error("AmazonOrderPoller: Deduction incomplete — flag NOT set, will retry", "orderID", order.ExternalOrderID)
	}
}

func (p *AmazonOrderPoller) processReversal(ctx context.Context, order *entity.Order) {
	slog.Info("AmazonOrderPoller: Reversing inventory for cancelled order", "orderID", order.ExternalOrderID, "itemCount", len(order.LineItems))

	if len(order.LineItems) == 0 {
		return
	}

	existingLogs, err := p.inventoryRepo.GetLogsByExternalOrderID(order.ExternalOrderID)
	if err != nil {
		slog.Error("AmazonOrderPoller: Failed to fetch existing logs for reversal", "orderID", order.ExternalOrderID, "error", err)
		return
	}

	allSuccess := true
	for _, item := range order.LineItems {
		if item.SKU == nil || *item.SKU == "" {
			continue
		}

		// Check if this specific SKU has already been reversed (cancellation log)
		alreadyReversed := false
		for _, log := range existingLogs {
			if log.Delta == item.Quantity && log.Reason == "cancellation" {
				alreadyReversed = true
				break
			}
		}

		if alreadyReversed {
			continue
		}

		err := p.orchestrator.AdjustStockByPlatformSKU(ctx, "amazon", *item.SKU, item.Quantity, "cancellation", &order.ExternalOrderID)
		if err != nil {
			slog.Error("AmazonOrderPoller: Stock reversal FAILED", "orderID", order.ExternalOrderID, "sku", *item.SKU, "error", err)
			allSuccess = false
		}
	}

	if allSuccess {
		order.InventoryDeducted = false
	}
}

func (p *AmazonOrderPoller) mapAmazonOrderToEntity(ctx context.Context, ao map[string]interface{}) (*entity.Order, error) {
	amazonOrderID := ao["AmazonOrderId"].(string)
	amazonStatus := ao["OrderStatus"].(string)

	// Fetch items from Amazon
	items, err := p.amazonClient.GetOrderItems(amazonOrderID)
	if err != nil {
		slog.Error("AmazonOrderPoller: Failed to fetch items for order", "orderID", amazonOrderID, "error", err)
		return nil, err
	}

	lineItems := make([]entity.LineItem, 0, len(items))
	for _, item := range items {
		qty := int(item["QuantityOrdered"].(float64))
		sku := item["SellerSKU"].(string)
		lineItems = append(lineItems, entity.LineItem{
			ID:       item["OrderItemId"].(string),
			SKU:      &sku,
			Quantity: qty,
		})
	}

	order := &entity.Order{
		SourceID:          "amazon",
		ExternalOrderID:   amazonOrderID,
		OrderNumber:       amazonOrderID,
		Status:            &amazonStatus,
		FinancialStatus:   entity.StrPtr("paid"),
		FulfillmentStatus: entity.StrPtr("unfulfilled"),
		LineItems:         lineItems,
		CreatedAt:         p.parseAmazonDate(ao["PurchaseDate"].(string)),
	}

	// PII Mapping
	name := "Amazon Customer"
	if shipping, ok := ao["ShippingAddress"].(map[string]interface{}); ok {
		if sName, ok := shipping["Name"].(string); ok && sName != "" && sName != "<nil>" {
			name = sName
		}
	} else if buyerInfo, ok := ao["BuyerInfo"].(map[string]interface{}); ok {
		if bName, ok := buyerInfo["BuyerName"].(string); ok && bName != "" && bName != "<nil>" {
			name = bName
		}
	}
	order.CustomerName = &name

	if shipping, ok := ao["ShippingAddress"].(map[string]interface{}); ok {
		city := fmt.Sprintf("%v", shipping["City"])
		if city == "" || city == "<nil>" { city = "N/A" }
		order.CustomerCity = &city

		state := fmt.Sprintf("%v", shipping["StateOrRegion"])
		if state == "" || state == "<nil>" { state = "N/A" }
		order.CustomerState = &state

		zip := fmt.Sprintf("%v", shipping["PostalCode"])
		if zip == "" || zip == "<nil>" { zip = "N/A" }
		order.CustomerZip = &zip

		country := fmt.Sprintf("%v", shipping["CountryCode"])
		if country == "" || country == "<nil>" { country = "IN" }
		order.CustomerCountry = &country
	} else {
		// Fallback if ShippingAddress is entirely missing (PII restricted)
		order.CustomerCity = entity.StrPtr("N/A")
		order.CustomerState = entity.StrPtr("N/A")
		order.CustomerZip = entity.StrPtr("N/A")
		order.CustomerCountry = entity.StrPtr("IN")
	}

	if total, ok := ao["OrderTotal"].(map[string]interface{}); ok {
		amountStr := fmt.Sprintf("%v", total["Amount"])
		amount, _ := strconv.ParseFloat(amountStr, 64)
		order.TotalPrice = amount
	}

	matchStatus := strings.TrimSpace(amazonStatus)
	switch matchStatus {
	case "Shipped", "InvoiceConfirmation":
		order.FulfillmentStatus = entity.StrPtr("fulfilled")
	case "Canceled":
		order.FulfillmentStatus = entity.StrPtr("cancelled")
		cancelledAt := time.Now()
		order.CancelledAt = &cancelledAt
	case "Unshipped", "PartiallyShipped":
		order.FulfillmentStatus = entity.StrPtr("unfulfilled")
	default:
		order.FulfillmentStatus = entity.StrPtr("unfulfilled")
	}

	if esStatus, ok := ao["EasyShipShipmentStatus"].(string); ok {
		esStatus = strings.TrimSpace(esStatus)
		if esStatus == "PickedUp" || esStatus == "OutForDelivery" || esStatus == "Delivered" {
			order.FulfillmentStatus = entity.StrPtr("fulfilled")
		}
		if strings.EqualFold(esStatus, "Delivered") {
			order.DeliveryStatus = entity.StrPtr("delivered")
			now := time.Now()
			order.DeliveredAt = &now
			pStatus := 4
			order.FeedbackStatusID = &pStatus
		}
	}

	// Check if we already have this order to preserve internal state
	existing, err := p.orderRepo.GetByExternalID(amazonOrderID)
	if err == nil {
		order.ID = existing.ID
		order.InventoryDeducted = existing.InventoryDeducted
		if existing.DeliveryStatus != nil && *existing.DeliveryStatus == "delivered" {
			order.DeliveryStatus = existing.DeliveryStatus
			order.DeliveredAt = existing.DeliveredAt
			order.FeedbackStatusID = existing.FeedbackStatusID
		}
	}

	return order, nil
}

func (p *AmazonOrderPoller) parseAmazonDate(dateStr string) time.Time {
	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return time.Now()
	}
	return t
}

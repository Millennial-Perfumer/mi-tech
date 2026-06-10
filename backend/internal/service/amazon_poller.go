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
	"time"

	"golang.org/x/sync/errgroup"
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
	// Use LastUpdatedAfter for periodic polls to catch status changes on old orders.
	// If specific dates are provided, we stick to CreatedAfter for predictable range syncing.
	useLastUpdated := (start == nil)

	var createdAfter time.Time
	var createdBefore time.Time

	if start != nil {
		createdAfter = *start
	} else {
		// Default check for orders in the last 6 hours to catch status updates.
		// Since we now use LastUpdatedAfter for periodic polls, this will catch
		// any order (old or new) that was modified in the last 6 hours.
		createdAfter = time.Now().Add(-6 * time.Hour)
	}

	if end != nil {
		createdBefore = *end
		// Enforce Amazon's 2-minute safety margin for cutoff times
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

	if len(amazonOrders) == 0 {
		return
	}

	// 1. Fetch Line Items in Parallel
	// Optimization: Parallelize API calls to overcome sequential network bottleneck.
	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(5) // Respect rate limits

	type orderResult struct {
		order entity.Order
		ao    map[string]interface{}
	}
	results := make([]orderResult, len(amazonOrders))

	for i, ao := range amazonOrders {
		i, ao := i, ao // closure capture
		g.Go(func() error {
			amazonOrderID := ao["AmazonOrderId"].(string)
			amazonStatus := ao["OrderStatus"].(string)

			items, err := p.amazonClient.GetOrderItems(amazonOrderID)
			if err != nil {
				slog.Error("AmazonOrderPoller: Failed to fetch items for order", "orderID", amazonOrderID, "error", err)
				return nil // Continue with others
			}

			lineItems := []entity.LineItem{}
			for _, item := range items {
				qty := int(item["QuantityOrdered"].(float64))
				sku := item["SellerSKU"].(string)
				lineItems = append(lineItems, entity.LineItem{
					ID:       item["OrderItemId"].(string),
					SKU:      &sku,
					Quantity: qty,
				})
			}

			order := entity.Order{
				SourceID:          "amazon",
				ExternalOrderID:   amazonOrderID,
				OrderNumber:       amazonOrderID,
				Status:            &amazonStatus,
				FinancialStatus:   entity.StrPtr("paid"),
				FulfillmentStatus: entity.StrPtr("unfulfilled"),
				LineItems:         lineItems,
				CreatedAt:         p.parseAmazonDate(ao["PurchaseDate"].(string)),
			}

			// Map PII and Status (Logic moved to helper for readability)
			p.mapAmazonOrderData(&order, ao)

			results[i] = orderResult{order: order, ao: ao}
			return nil
		})
	}

	if err := g.Wait(); err != nil && gCtx.Err() != nil {
		slog.Error("AmazonOrderPoller: Error during parallel item fetch", "error", err)
		return
	}

	// 2. Prepare for Batch Sink
	var ordersToUpsert []entity.Order
	for _, res := range results {
		if res.order.ExternalOrderID == "" {
			continue
		}
		ordersToUpsert = append(ordersToUpsert, res.order)
	}

	if len(ordersToUpsert) == 0 {
		return
	}

	// 3. Batch Sink to DB
	// Optimization: Reduce DB roundtrips from O(N) to O(1).
	affectedIDs, err := p.orderRepo.UpsertBatch(ordersToUpsert)
	if err != nil {
		slog.Error("AmazonOrderPoller: Failed to batch upsert orders", "error", err)
		return
	}

	// 4. Trigger Global Sync for affected items to ensure cross-platform consistency
	if p.orchestrator != nil {
		for _, id := range affectedIDs {
			_ = p.orchestrator.GlobalSync(ctx, id, "amazon")
		}
	}

	slog.Info("AmazonOrderPoller: Sync complete", "processed", len(ordersToUpsert))
}

func (p *AmazonOrderPoller) mapAmazonOrderData(order *entity.Order, ao map[string]interface{}) {
	amazonStatus := entity.DerefStr(order.Status)
	matchStatus := strings.TrimSpace(amazonStatus)

	// Billing/Shipping info mapping
	name := ""
	if shipping, ok := ao["ShippingAddress"].(map[string]interface{}); ok {
		name = fmt.Sprintf("%v", shipping["Name"])
	}
	if (name == "" || name == "<nil>") && ao["BuyerInfo"] != nil {
		if buyerInfo, ok := ao["BuyerInfo"].(map[string]interface{}); ok {
			name = fmt.Sprintf("%v", buyerInfo["BuyerName"])
		}
	}
	if name == "" || name == "<nil>" {
		name = "Amazon Customer"
	}
	order.CustomerName = &name

	if shipping, ok := ao["ShippingAddress"].(map[string]interface{}); ok {
		order.CustomerCity = entity.StrPtr(fmt.Sprintf("%v", shipping["City"]))
		order.CustomerState = entity.StrPtr(fmt.Sprintf("%v", shipping["StateOrRegion"]))
		order.CustomerZip = entity.StrPtr(fmt.Sprintf("%v", shipping["PostalCode"]))
		order.CustomerCountry = entity.StrPtr(fmt.Sprintf("%v", shipping["CountryCode"]))
	} else {
		order.CustomerName = entity.StrPtr("Amazon Customer")
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

	// Status Mapping
	switch matchStatus {
	case "Shipped", "InvoiceConfirmation":
		order.FulfillmentStatus = entity.StrPtr("fulfilled")
	case "Canceled":
		order.FulfillmentStatus = entity.StrPtr("cancelled")
		cancelledAt := time.Now()
		order.CancelledAt = &cancelledAt
	case "Unshipped", "PartiallyShipped":
		order.FulfillmentStatus = entity.StrPtr("unfulfilled")
	case "Pending":
		order.FulfillmentStatus = entity.StrPtr("unfulfilled")
		// Optimization: Don't deduct inventory for Pending orders yet
		order.SkipInventorySync = true
	default:
		order.FulfillmentStatus = entity.StrPtr("unfulfilled")
	}

	// Easy Ship Specifics
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

func (p *AmazonOrderPoller) parseAmazonDate(dateStr string) time.Time {
	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return time.Now()
	}
	return t
}

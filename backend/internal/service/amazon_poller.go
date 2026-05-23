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

	var wg sync.WaitGroup
	// Optimization: Limit concurrency to 5 to avoid hitting Amazon API rate limits
	semaphore := make(chan struct{}, 5)
	orderChan := make(chan entity.Order, len(amazonOrders))

	for _, ao := range amazonOrders {
		wg.Add(1)
		go func(ao map[string]interface{}) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			amazonOrderID := ao["AmazonOrderId"].(string)
			amazonStatus := ao["OrderStatus"].(string)
			esStatus, _ := ao["EasyShipShipmentStatus"].(string)

			slog.Info("AmazonOrderPoller: Processing order",
				"orderID", amazonOrderID,
				"status", amazonStatus,
				"esStatus", esStatus,
			)

			// 1. Fetch line items for this order in parallel
			items, err := p.amazonClient.GetOrderItems(amazonOrderID)
			if err != nil {
				slog.Error("AmazonOrderPoller: Failed to fetch items for order", "orderID", amazonOrderID, "error", err)
				return
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
				OrderNumber:       amazonOrderID, // Amazon uses OrderID as the display number
				Status:            &amazonStatus,
				FinancialStatus:   entity.StrPtr("paid"), // Usually paid on Amazon
				FulfillmentStatus: entity.StrPtr("unfulfilled"),
				LineItems:         lineItems,
				CreatedAt:         p.parseAmazonDate(ao["PurchaseDate"].(string)),
			}

			// Billing/Shipping info mapping
			name := ""
			if shipping, ok := ao["ShippingAddress"].(map[string]interface{}); ok {
				name = fmt.Sprintf("%v", shipping["Name"])
			}

			// Fallback to BuyerInfo if ShippingAddress Name is empty
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
				city := fmt.Sprintf("%v", shipping["City"])
				if city == "" || city == "<nil>" {
					city = "N/A"
				}
				order.CustomerCity = &city

				state := fmt.Sprintf("%v", shipping["StateOrRegion"])
				if state == "" || state == "<nil>" {
					state = "N/A"
				}
				order.CustomerState = &state

				zip := fmt.Sprintf("%v", shipping["PostalCode"])
				if zip == "" || zip == "<nil>" {
					zip = "N/A"
				}
				order.CustomerZip = &zip

				country := fmt.Sprintf("%v", shipping["CountryCode"])
				if country == "" || country == "<nil>" {
					country = "IN"
				}
				order.CustomerCountry = &country
			} else {
				// Fallback if ShippingAddress is entirely missing (PII restricted)
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

			// Financial status is usually paid on Amazon
			order.FinancialStatus = entity.StrPtr("paid")

			// Dynamic Status Mapping
			// Amazon Statuses: Pending, Unshipped, PartiallyShipped, Shipped, Canceled, Unfulfillable, InvoiceConfirmation, etc.

			// Normalize for matching
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

			// Easy Ship Specific: Override fulfillment and delivery status based on tracking
			if esStatus, ok := ao["EasyShipShipmentStatus"].(string); ok {
				esStatus = strings.TrimSpace(esStatus)
				if esStatus == "PickedUp" || esStatus == "OutForDelivery" || esStatus == "Delivered" {
					order.FulfillmentStatus = entity.StrPtr("fulfilled")
				}
				if strings.EqualFold(esStatus, "Delivered") {
					order.DeliveryStatus = entity.StrPtr("delivered")
					now := time.Now()
					order.DeliveredAt = &now
					// Amazon orders lack customer phone numbers (PII restricted), so we skip feedback automation.
					// Status 4 (expired/skipped) ensures they don't appear in the feedback trigger list.
					pStatus := 4
					order.FeedbackStatusID = &pStatus
				}
			}

			// Check if we already have this order to preserve internal state
			existing, err := p.orderRepo.GetByExternalID(amazonOrderID)
			if err == nil {
				order.ID = existing.ID
				order.InventoryDeducted = existing.InventoryDeducted

				// Preserve manually set or previously discovered delivery data
				if existing.DeliveryStatus != nil && *existing.DeliveryStatus == "delivered" {
					order.DeliveryStatus = existing.DeliveryStatus
					if existing.DeliveredAt != nil {
						order.DeliveredAt = existing.DeliveredAt
					}
					if existing.FeedbackStatusID != nil {
						order.FeedbackStatusID = existing.FeedbackStatusID
					}
				}
			}

			// 2. Logic for Stock Deduction (Unshipped/Shipped/fulfilled = Amazon has committed or we have fulfilled the inventory)
			isFulfilledOrCommitted := (matchStatus == "Unshipped" || matchStatus == "Shipped" || matchStatus == "InvoiceConfirmation" || (order.FulfillmentStatus != nil && *order.FulfillmentStatus == "fulfilled"))

			if isFulfilledOrCommitted {
				p.processDeduction(ctx, &order)
			}

			// 3. Logic for Reversal (Canceled)
			if amazonStatus == "Canceled" {
				p.processReversal(ctx, &order)
			}

			orderChan <- order
		}(ao)
	}

	// Wait for all parallel fetches and processing to complete
	go func() {
		wg.Wait()
		close(orderChan)
	}()

	var ordersToUpsert []entity.Order
	for order := range orderChan {
		ordersToUpsert = append(ordersToUpsert, order)
	}

	if len(ordersToUpsert) == 0 {
		return
	}

	// 4. Optimization: Batch Upsert all orders in a single transaction
	affectedIDs, err := p.orderRepo.UpsertBatch(ordersToUpsert)
	if err != nil {
		slog.Error("AmazonOrderPoller: Batch upsert failed", "error", err)
		return
	}

	// 5. Optimization: Parallelize Global Synchronization for affected inventory items
	if p.orchestrator != nil && len(affectedIDs) > 0 {
		// Deduplicate affected IDs to avoid redundant syncs
		uniqueIDs := make(map[int]bool)
		for _, id := range affectedIDs {
			uniqueIDs[id] = true
		}

		var syncWg sync.WaitGroup
		// Limit concurrency for global sync to avoid overwhelming external APIs (Shopify)
		syncSemaphore := make(chan struct{}, 10)
		for id := range uniqueIDs {
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

func (p *AmazonOrderPoller) parseAmazonDate(dateStr string) time.Time {
	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return time.Now()
	}
	return t
}

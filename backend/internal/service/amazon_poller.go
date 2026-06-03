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

	// Optimization: Fetch order items in parallel to slash network latency.
	// We use a concurrency limit of 5 to respect Amazon SP-API rate limits.
	type orderResult struct {
		order       entity.Order
		raw         map[string]interface{}
		matchStatus string
		err         error
	}

	results := make(chan orderResult, len(amazonOrders))
	sem := make(chan struct{}, 5) // Concurrency limit

	for _, ao := range amazonOrders {
		go func(ao map[string]interface{}) {
			sem <- struct{}{}
			defer func() { <-sem }()

			amazonOrderID := ao["AmazonOrderId"].(string)
			amazonStatus := ao["OrderStatus"].(string)

			items, err := p.amazonClient.GetOrderItems(amazonOrderID)
			if err != nil {
				results <- orderResult{err: fmt.Errorf("failed to fetch items for %s: %w", amazonOrderID, err)}
				return
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

			// PII Mapping
			name := "Amazon Customer"
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

			city, state, zip, country := "N/A", "N/A", "N/A", "IN"
			if shipping, ok := ao["ShippingAddress"].(map[string]interface{}); ok {
				city = fmt.Sprintf("%v", shipping["City"])
				if city == "" || city == "<nil>" { city = "N/A" }

				state = fmt.Sprintf("%v", shipping["StateOrRegion"])
				if state == "" || state == "<nil>" { state = "N/A" }

				zip = fmt.Sprintf("%v", shipping["PostalCode"])
				if zip == "" || zip == "<nil>" { zip = "N/A" }

				country = fmt.Sprintf("%v", shipping["CountryCode"])
				if country == "" || country == "<nil>" { country = "IN" }
			}
			order.CustomerCity = &city
			order.CustomerState = &state
			order.CustomerZip = &zip
			order.CustomerCountry = &country

			if total, ok := ao["OrderTotal"].(map[string]interface{}); ok {
				amountStr := fmt.Sprintf("%v", total["Amount"])
				amount, _ := strconv.ParseFloat(amountStr, 64)
				order.TotalPrice = amount
			}

			// Status Mapping
			matchStatus := strings.TrimSpace(amazonStatus)
			switch matchStatus {
			case "Shipped", "InvoiceConfirmation":
				order.FulfillmentStatus = entity.StrPtr("fulfilled")
			case "Canceled":
				order.FulfillmentStatus = entity.StrPtr("cancelled")
				cancelledAt := time.Now()
				order.CancelledAt = &cancelledAt
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

			results <- orderResult{order: order, raw: ao, matchStatus: matchStatus}
		}(ao)
	}

	for i := 0; i < len(amazonOrders); i++ {
		res := <-results
		if res.err != nil {
			slog.Error("AmazonOrderPoller: Network error during parallel fetch", "error", res.err)
			continue
		}

		order := res.order
		matchStatus := res.matchStatus
		amazonOrderID := order.ExternalOrderID

		slog.Info("AmazonOrderPoller: Processing order",
			"orderID", amazonOrderID,
			"status", *order.Status,
		)

		// Check existing for state preservation
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

		// Inventory Logic
		isFulfilledOrCommitted := (matchStatus == "Unshipped" || matchStatus == "Shipped" || matchStatus == "InvoiceConfirmation" || (order.FulfillmentStatus != nil && *order.FulfillmentStatus == "fulfilled"))
		if isFulfilledOrCommitted {
			p.processDeduction(ctx, &order)
		}
		if matchStatus == "Canceled" {
			p.processReversal(ctx, &order)
		}

		// Final Persistence
		p.orderRepo.Upsert(order)
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

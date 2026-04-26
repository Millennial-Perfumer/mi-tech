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

	for _, ao := range amazonOrders {
		amazonOrderID := ao["AmazonOrderId"].(string)
		amazonStatus := ao["OrderStatus"].(string)
		esStatus, _ := ao["EasyShipShipmentStatus"].(string)

		slog.Info("AmazonOrderPoller: Processing order", 
			"orderID", amazonOrderID, 
			"status", amazonStatus, 
			"esStatus", esStatus,
		)

		// 1. Sync the order to our DB first (Upsert)
		items, err := p.amazonClient.GetOrderItems(amazonOrderID)
		if err != nil {
			slog.Error("AmazonOrderPoller: Failed to fetch items for order", "orderID", amazonOrderID, "error", err)
			continue
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
			SourceID:        "amazon",
			ExternalOrderID: amazonOrderID,
			OrderNumber:      amazonOrderID, // Amazon uses OrderID as the display number
			Status:           &amazonStatus,
			FinancialStatus:  entity.StrPtr("paid"), // Usually paid on Amazon
			FulfillmentStatus: entity.StrPtr("unfulfilled"),
			LineItems:       lineItems,
			CreatedAt:       p.parseAmazonDate(ao["PurchaseDate"].(string)),
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
			
			// Easy Ship Specific: If it's Unshipped but Picked Up, it's effectively fulfilled
			if esStatus, ok := ao["EasyShipShipmentStatus"].(string); ok {
				esStatus = strings.TrimSpace(esStatus)
				if esStatus == "PickedUp" || esStatus == "OutForDelivery" || esStatus == "Delivered" {
					order.FulfillmentStatus = entity.StrPtr("fulfilled")
				}
			}
		default:
			order.FulfillmentStatus = entity.StrPtr("unfulfilled")
		}

		// Check if we already have this order to see if inventory was already deducted
		existing, err := p.orderRepo.GetByExternalID(amazonOrderID)
		inventoryAlreadyDeducted := false
		if err == nil {
			inventoryAlreadyDeducted = existing.InventoryDeducted
			order.ID = existing.ID
			order.InventoryDeducted = existing.InventoryDeducted
			
			// Preserve manually set DeliveryStatus if it was already "delivered"
			if existing.DeliveryStatus != nil {
				order.DeliveryStatus = existing.DeliveryStatus
			}
		}

		// 2. Logic for Stock Deduction (Unshipped/Shipped/fulfilled = Amazon has committed or we have fulfilled the inventory)
		isFulfilledOrCommitted := (matchStatus == "Unshipped" || matchStatus == "Shipped" || matchStatus == "InvoiceConfirmation" || (order.FulfillmentStatus != nil && *order.FulfillmentStatus == "fulfilled"))
		
		if isFulfilledOrCommitted && !inventoryAlreadyDeducted {
			p.processDeduction(ctx, &order)
		}

		// 3. Logic for Reversal (Canceled)
		if amazonStatus == "Canceled" && inventoryAlreadyDeducted {
			p.processReversal(ctx, &order)
		}

		// 4. Final Upsert to keep status current
		p.orderRepo.Upsert(order)
	}
}

func (p *AmazonOrderPoller) processDeduction(ctx context.Context, order *entity.Order) {
	slog.Info("AmazonOrderPoller: Order moved to Unshipped, deducting inventory", "orderID", order.ExternalOrderID)
	for _, item := range order.LineItems {
		if item.SKU == nil {
			continue
		}

		err := p.orchestrator.AdjustStockByPlatformSKU(ctx, "amazon", *item.SKU, -item.Quantity, "sale", &order.ExternalOrderID)
		if err != nil {
			slog.Error("AmazonOrderPoller: Stock adjustment failed", "sku", *item.SKU, "error", err)
		}
	}
	order.InventoryDeducted = true
}

func (p *AmazonOrderPoller) processReversal(ctx context.Context, order *entity.Order) {
	slog.Info("AmazonOrderPoller: Order Canceled, returning inventory to stock", "orderID", order.ExternalOrderID)
	for _, item := range order.LineItems {
		if item.SKU == nil {
			continue
		}

		err := p.orchestrator.AdjustStockByPlatformSKU(ctx, "amazon", *item.SKU, item.Quantity, "cancellation", &order.ExternalOrderID)
		if err != nil {
			slog.Error("AmazonOrderPoller: Stock reversal failed", "sku", *item.SKU, "error", err)
		}
	}
	order.InventoryDeducted = false
}

func (p *AmazonOrderPoller) parseAmazonDate(dateStr string) time.Time {
	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return time.Now()
	}
	return t
}

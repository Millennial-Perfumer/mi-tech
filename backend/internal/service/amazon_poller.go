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

	var createdAfter, createdBefore time.Time
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

	amazonOrders, err := p.amazonClient.GetOrders(createdAfter, createdBefore, useLastUpdated)
	if err != nil {
		slog.Error("AmazonOrderPoller: Failed to fetch orders", "error", err)
		return
	}

	if len(amazonOrders) == 0 {
		slog.Info("AmazonOrderPoller: No orders found in range")
		return
	}

	// 1. Parallelize Item Fetching (N+1 Optimization)
	// Using bounded semaphore to avoid hitting SP-API rate limits
	type taskResult struct {
		order *entity.Order
		err   error
	}

	resultsChan := make(chan taskResult, len(amazonOrders))
	semaphore := make(chan struct{}, 5)
	var wg sync.WaitGroup

	for _, ao := range amazonOrders {
		wg.Add(1)
		go func(ao map[string]interface{}) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			amazonOrderID := ao["AmazonOrderId"].(string)
			items, err := p.amazonClient.GetOrderItems(amazonOrderID)
			if err != nil {
				resultsChan <- taskResult{err: err}
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

			order := p.mapAmazonOrderToEntity(ao, lineItems)
			resultsChan <- taskResult{order: &order}
		}(ao)
	}

	wg.Wait()
	close(resultsChan)

	var orderEntities []entity.Order
	for res := range resultsChan {
		if res.err != nil {
			slog.Error("AmazonOrderPoller: Item fetch failed", "error", res.err)
			continue
		}
		if res.order != nil {
			orderEntities = append(orderEntities, *res.order)
		}
	}

	if len(orderEntities) == 0 {
		return
	}

	// 2. Batch Upsert (O(1) DB Optimization)
	// Expected Impact: Reduces DB roundtrips from O(N) to O(1).
	affectedIDs, err := p.orderRepo.UpsertBatch(orderEntities)
	if err != nil {
		slog.Error("AmazonOrderPoller: Batch upsert failed", "error", err)
		return
	}

	// 3. Parallelize Global Sync for affected items
	if p.orchestrator != nil && len(affectedIDs) > 0 {
		uniqueIDs := make(map[int]bool)
		for _, id := range affectedIDs {
			uniqueIDs[id] = true
		}

		syncSem := make(chan struct{}, 10)
		var syncWg sync.WaitGroup
		for id := range uniqueIDs {
			syncWg.Add(1)
			go func(itemID int) {
				defer syncWg.Done()
				syncSem <- struct{}{}
				defer func() { <-syncSem }()
				_ = p.orchestrator.GlobalSync(ctx, itemID, "amazon")
			}(id)
		}
		syncWg.Wait()
	}

	slog.Info("AmazonOrderPoller: Sync completed", "processed", len(orderEntities), "affected_items", len(affectedIDs))
}

func (p *AmazonOrderPoller) mapAmazonOrderToEntity(ao map[string]interface{}, lineItems []entity.LineItem) entity.Order {
	amazonOrderID := ao["AmazonOrderId"].(string)
	amazonStatus := ao["OrderStatus"].(string)

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

	// Billing/Shipping info mapping with PII fallbacks
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
		order.CustomerCity = entity.StrPtr(p.valOrNA(shipping["City"]))
		order.CustomerState = entity.StrPtr(p.valOrNA(shipping["StateOrRegion"]))
		order.CustomerZip = entity.StrPtr(p.valOrNA(shipping["PostalCode"]))
		order.CustomerCountry = entity.StrPtr(p.valOrNA(shipping["CountryCode"]))
	} else {
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
	matchStatus := strings.TrimSpace(amazonStatus)
	switch matchStatus {
	case "Shipped", "InvoiceConfirmation":
		order.FulfillmentStatus = entity.StrPtr("fulfilled")
	case "Canceled":
		order.FulfillmentStatus = entity.StrPtr("cancelled")
		cancelledAt := time.Now()
		order.CancelledAt = &cancelledAt
	}

	// Inventory Deduction Intent
	// Unshipped/Shipped/fulfilled = Amazon has committed or we have fulfilled the inventory
	if matchStatus == "Unshipped" || matchStatus == "Shipped" || matchStatus == "InvoiceConfirmation" || (order.FulfillmentStatus != nil && *order.FulfillmentStatus == "fulfilled") {
		order.InventoryDeducted = true
	}
	// Reversal Intent
	if matchStatus == "Canceled" {
		order.InventoryDeducted = false
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
			order.FeedbackStatusID = entity.IntPtr(4)
		}
	}

	return order
}

func (p *AmazonOrderPoller) valOrNA(v interface{}) string {
	s := fmt.Sprintf("%v", v)
	if s == "" || s == "<nil>" {
		return "N/A"
	}
	return s
}


func (p *AmazonOrderPoller) parseAmazonDate(dateStr string) time.Time {
	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return time.Now()
	}
	return t
}

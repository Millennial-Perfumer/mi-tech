package service

import (
	"fmt"
	"log"
	"mi-tech/internal/client/amazon"
	"mi-tech/internal/entity"
	"time"
)

// AmazonSyncService handles periodic fetching and processing of Amazon orders.
type AmazonSyncService struct {
	amazonClient *amazon.Client
	orderService *OrderService
}

func NewAmazonSyncService(amazonClient *amazon.Client, orderService *OrderService) *AmazonSyncService {
	return &AmazonSyncService{
		amazonClient: amazonClient,
		orderService: orderService,
	}
}

// SyncOrders pulls recent orders from Amazon and syncs them into our system.
func (s *AmazonSyncService) SyncOrders(since time.Time) (int, error) {
	log.Printf("AmazonSyncService: Fetching orders created after %s", since.Format(time.RFC3339))
	
	rawOrders, err := s.amazonClient.GetOrders(since, time.Time{}, false)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch amazon orders: %w", err)
	}

	count := 0
	for _, ro := range rawOrders {
		order := s.mapAmazonToEntity(ro)
		
		// Fetch Line Items
		rawItems, err := s.amazonClient.GetOrderItems(order.ExternalOrderID)
		if err != nil {
			log.Printf("AmazonSyncService Warning: Failed to fetch items for order %s: %v", order.ExternalOrderID, err)
			continue
		}

		order.LineItems = s.mapAmazonItemsToEntities(rawItems)
		
		err = s.orderService.UpsertOrder(order)
		if err != nil {
			log.Printf("AmazonSyncService Warning: Failed to upsert order %s: %v", order.ExternalOrderID, err)
			continue
		}
		count++
	}

	return count, nil
}

func (s *AmazonSyncService) mapAmazonItemsToEntities(rawItems []map[string]interface{}) []entity.LineItem {
	var items []entity.LineItem
	for _, ri := range rawItems {
		qty := 0
		if q, ok := ri["QuantityOrdered"].(float64); ok {
			qty = int(q)
		}

		sku := fmt.Sprintf("%v", ri["SellerSKU"])
		item := entity.LineItem{
			Title:    entity.StrPtr(fmt.Sprintf("%v", ri["Title"])),
			SKU:      &sku,
			Quantity: qty,
		}

		if price, ok := ri["ItemPrice"].(map[string]interface{}); ok {
			amount := fmt.Sprintf("%v", price["Amount"])
			f := s.parseFloat(amount)
			item.Price = f
		}

		items = append(items, item)
	}
	return items
}

func (s *AmazonSyncService) mapAmazonToEntity(ro map[string]interface{}) entity.Order {
	extID := fmt.Sprintf("%v", ro["AmazonOrderId"])
	
	order := entity.Order{
		SourceID:         "amazon",
		ExternalOrderID:  extID,
		OrderNumber:      extID, // Amazon uses OrderID as the display number
		Status:           entity.StrPtr(fmt.Sprintf("%v", ro["OrderStatus"])),
		FinancialStatus:  entity.StrPtr("paid"), // Usually paid on Amazon
		FulfillmentStatus: entity.StrPtr("unfulfilled"),
		CreatedAt:        s.parseTime(ro["PurchaseDate"]),
		UpdatedAt:        s.parseTime(ro["LastUpdateDate"]),
	}

	// Billing/Shipping info mapping
	if shipping, ok := ro["ShippingAddress"].(map[string]interface{}); ok {
		order.CustomerName = entity.StrPtr(fmt.Sprintf("%v", shipping["Name"]))
		order.CustomerCity = entity.StrPtr(fmt.Sprintf("%v", shipping["City"]))
		order.CustomerState = entity.StrPtr(fmt.Sprintf("%v", shipping["StateOrRegion"]))
		order.CustomerZip = entity.StrPtr(fmt.Sprintf("%v", shipping["PostalCode"]))
		order.CustomerCountry = entity.StrPtr(fmt.Sprintf("%v", shipping["CountryCode"]))
	}

	if total, ok := ro["OrderTotal"].(map[string]interface{}); ok {
		amountStr := fmt.Sprintf("%v", total["Amount"])
		order.TotalPrice = s.parseFloat(amountStr)
	}

	return order
}

func (s *AmazonSyncService) parseTime(v interface{}) time.Time {
	if s, ok := v.(string); ok {
		t, _ := time.Parse(time.RFC3339, s)
		return t
	}
	return time.Now()
}

func (s *AmazonSyncService) parseFloat(v string) float64 {
	var f float64
	fmt.Sscanf(v, "%f", &f)
	return f
}

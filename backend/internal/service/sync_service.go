package service

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"shopify-gst-app/internal/entity"
	"shopify-gst-app/internal/models"
	"shopify-gst-app/internal/repository"
	"shopify-gst-app/internal/shopify"
)

// SyncService orchestrates fetching orders from Shopify and persisting them.
type SyncService struct {
	shopifyClient *shopify.Client
	orderRepo     repository.OrderRepository
}

// NewSyncService creates a new SyncService.
func NewSyncService(shopifyClient *shopify.Client, orderRepo repository.OrderRepository) *SyncService {
	return &SyncService{
		shopifyClient: shopifyClient,
		orderRepo:     orderRepo,
	}
}

// Sync fetches new/updated orders from Shopify and upserts them into the database.
func (s *SyncService) Sync() (int, error) {
	lastSync := s.getLastSyncTime()
	log.Printf("Starting Shopify order sync fetching orders updated after %s...", lastSync.Format(time.RFC3339))

	shopifyOrders, err := s.shopifyClient.FetchOrders(lastSync)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch from Shopify: %w", err)
	}

	if len(shopifyOrders) == 0 {
		log.Printf("No new or updated orders found from Shopify since %s", lastSync.Format(time.RFC3339))
		return 0, nil
	}

	log.Printf("Fetched %d orders. Proceeding to sink to database.", len(shopifyOrders))

	// Map old models.GraphQLOrderNode → entity.Order
	var orderEntities []entity.Order
	for _, so := range shopifyOrders {
		orderEntities = append(orderEntities, graphQLNodeToEntity(so))
	}

	if err := s.orderRepo.UpsertBatch(orderEntities); err != nil {
		return 0, fmt.Errorf("failed to upsert batch: %w", err)
	}

	log.Printf("Successfully synchronized %d orders and their items into PostgreSQL.", len(orderEntities))
	return len(orderEntities), nil
}

// ResetAndSync wipes all orders locally and performs a full sync.
func (s *SyncService) ResetAndSync() (int, error) {
	if err := s.orderRepo.TruncateAll(); err != nil {
		return 0, err
	}
	return s.Sync()
}

func (s *SyncService) getLastSyncTime() time.Time {
	baseline, _ := time.Parse(time.RFC3339, "2026-03-01T00:00:00Z")
	return baseline
}

// graphQLNodeToEntity bridges models.GraphQLOrderNode → entity.Order.
// This will be replaced by mapper.GraphQLOrderToEntity in Phase 7 when
// shopify.Client is updated to return dto types.
func graphQLNodeToEntity(so models.GraphQLOrderNode) entity.Order {
	var custName, custEmail, custPhone, custCity, custState, custCountry string

	custEmail = so.Email
	if so.Customer != nil {
		custName = strings.TrimSpace(so.Customer.DisplayName)
		if custName == "" {
			custName = strings.TrimSpace(so.Customer.FirstName + " " + so.Customer.LastName)
		}
	}

	if so.ShippingAddress != nil {
		if custName == "" {
			custName = strings.TrimSpace(so.ShippingAddress.Name)
		}
		custPhone = so.ShippingAddress.Phone
		custCity = so.ShippingAddress.City
		custState = so.ShippingAddress.Province
		custCountry = so.ShippingAddress.Country
	} else if so.BillingAddress != nil {
		if custName == "" {
			custName = strings.TrimSpace(so.BillingAddress.Name)
		}
		custPhone = so.BillingAddress.Phone
		custCity = so.BillingAddress.City
		custState = so.BillingAddress.Province
		custCountry = so.BillingAddress.Country
	}

	if custName == "" {
		custName = "Valued Customer"
	}

	status := "unfulfilled"
	if so.DisplayFulfillmentStatus == "FULFILLED" {
		status = "fulfilled"
	} else if so.DisplayFinancialStatus == "PAID" {
		status = "paid"
	}

	financialStatus := strings.ToLower(so.DisplayFinancialStatus)
	fulfillmentStatus := strings.ToLower(so.DisplayFulfillmentStatus)
	deliveryStatus := "pending"
	trackingNumber := ""
	shippingCompany := ""
	trackingUrl := ""

	if len(so.Fulfillments) > 0 {
		f := so.Fulfillments[0]
		if f.DisplayStatus != "" {
			deliveryStatus = strings.ToLower(strings.ReplaceAll(f.DisplayStatus, "_", " "))
		} else {
			deliveryStatus = strings.ToLower(strings.ReplaceAll(f.Status, "_", " "))
		}
		if len(f.TrackingInfo) > 0 {
			trackingNumber = f.TrackingInfo[0].Number
			shippingCompany = f.TrackingInfo[0].Company
			trackingUrl = f.TrackingInfo[0].Url
		}
	}

	sourceID := "shopify"
	switch strings.ToLower(so.SourceName) {
	case "amazon":
		sourceID = "amazon"
	case "pos":
		sourceID = "pos"
	}

	idStr := strings.TrimPrefix(so.ID, "gid://shopify/Order/")

	createdAt, _ := time.Parse(time.RFC3339, so.ProcessedAt)
	updatedAt, _ := time.Parse(time.RFC3339, so.UpdatedAt)

	order := entity.Order{
		ID:                idStr,
		SourceID:          sourceID,
		ExternalOrderID:   idStr,
		OrderNumber:       so.Name,
		TotalPrice:        parseFloat(so.CurrentTotalPriceSet.ShopMoney.Amount),
		SubtotalPrice:     toNullFloat(so.CurrentSubtotalPriceSet.ShopMoney.Amount),
		TotalTax:          toNullFloat(so.CurrentTotalTaxSet.ShopMoney.Amount),
		CreatedAt:         createdAt,
		UpdatedAt:         updatedAt,
		FinancialStatus:   toNS(financialStatus),
		FulfillmentStatus: toNS(fulfillmentStatus),
		DeliveryStatus:    toNS(deliveryStatus),
		TrackingNumber:    toNS(trackingNumber),
		ShippingCompany:   toNS(shippingCompany),
		TrackingUrl:       toNS(trackingUrl),
		Status:            toNS(status),
		CustomerName:      toNS(custName),
		CustomerEmail:     toNS(custEmail),
		CustomerPhone:     toNS(custPhone),
		CustomerCity:      toNS(custCity),
		CustomerState:     toNS(custState),
		CustomerCountry:   toNS(custCountry),
	}

	// Map line items
	for _, edge := range so.LineItems.Edges {
		li := edge.Node
		hsCode := ""
		if li.Variant != nil {
			hsCode = li.Variant.InventoryItem.HarmonizedSystemCode
		}
		itemID := strings.TrimPrefix(li.ID, "gid://shopify/LineItem/")
		order.LineItems = append(order.LineItems, entity.LineItem{
			ID:       itemID,
			OrderID:  idStr,
			Title:    toNS(li.Title),
			SKU:      toNS(li.SKU),
			HSCode:   toNS(hsCode),
			Quantity: li.Quantity,
			Price:    parseFloat(li.OriginalTotalSet.ShopMoney.Amount),
			Discount: parseFloat(li.TotalDiscountSet.ShopMoney.Amount),
		})
	}

	return order
}

// --- helpers ---

func toNS(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

func toNullFloat(s string) sql.NullFloat64 {
	if s == "" {
		return sql.NullFloat64{Valid: false}
	}
	v := parseFloat(s)
	return sql.NullFloat64{Float64: v, Valid: true}
}

func parseFloat(s string) float64 {
	var v float64
	fmt.Sscanf(s, "%f", &v)
	return v
}

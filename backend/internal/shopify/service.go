package shopify

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"shopify-gst-app/internal/db"
)

// SyncService orchestrates fetching orders and saving them to the DB
type SyncService struct {
	client *Client
	db     *sql.DB
}

func NewSyncService(client *Client, db *sql.DB) *SyncService {
	return &SyncService{
		client: client,
		db:     db,
	}
}

// Sync fetches all relevant orders and upserts them into PostgreSQL
func (s *SyncService) Sync() (int, error) {
	// Query to find the latest order updated timestamp
	var lastSync time.Time
	err := s.db.QueryRow("SELECT MAX(updated_at) FROM orders").Scan(&lastSync)

	if err != nil || lastSync.IsZero() {
		// If table is empty or we get an error, fallback to the March 1st 2026 default baseline
		log.Println("No previous sync found, falling back to initial March 2026 baseline")
		lastSync, _ = time.Parse(time.RFC3339, "2026-03-01T00:00:00Z")
	} else {
		// Subtract 5 minutes to create an overlap window to prevent missing orders due to clock drift
		lastSync = lastSync.Add(-5 * time.Minute)
	}

	log.Printf("Starting Shopify order sync fetching orders updated after %s...", lastSync.Format(time.RFC3339))

	// 1. Fetch from Shopify
	shopifyOrders, err := s.client.FetchOrders(lastSync)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch from Shopify: %w", err)
	}

	if len(shopifyOrders) == 0 {
		log.Printf("No new or updated orders found from Shopify since %s", lastSync.Format(time.RFC3339))
		return 0, nil
	}

	log.Printf("Fetched %d orders. Proceeding to sink to database.", len(shopifyOrders))

	// 2. UPSERT to PostgreSQL
	// We use ON CONFLICT (source_id, external_order_id) DO UPDATE to seamlessly handle existing vs new orders
	orderQuery := `
	INSERT INTO orders (
		id, source_id, external_order_id, order_number, total_price, subtotal_price, total_tax, created_at, updated_at,
		customer_name, customer_email, customer_phone, customer_city, customer_state, customer_country, status,
		financial_status, fulfillment_status, delivery_status,
		tracking_number, shipping_company, tracking_url
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22)
	ON CONFLICT (source_id, external_order_id) DO UPDATE SET
		order_number = EXCLUDED.order_number,
		total_price = EXCLUDED.total_price,
		subtotal_price = EXCLUDED.subtotal_price,
		total_tax = EXCLUDED.total_tax,
		updated_at = EXCLUDED.updated_at,
		customer_name = EXCLUDED.customer_name,
		customer_email = EXCLUDED.customer_email,
		customer_phone = EXCLUDED.customer_phone,
		customer_city = EXCLUDED.customer_city,
		customer_state = EXCLUDED.customer_state,
		customer_country = EXCLUDED.customer_country,
		status = EXCLUDED.status,
		financial_status = EXCLUDED.financial_status,
		fulfillment_status = EXCLUDED.fulfillment_status,
		delivery_status = EXCLUDED.delivery_status,
		tracking_number = EXCLUDED.tracking_number,
		shipping_company = EXCLUDED.shipping_company,
		tracking_url = EXCLUDED.tracking_url;
	`

	itemQuery := `
	INSERT INTO order_line_items (id, order_id, title, sku, hs_code, quantity, price, discount)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	ON CONFLICT (id) DO UPDATE SET
		title = EXCLUDED.title,
		sku = EXCLUDED.sku,
		hs_code = EXCLUDED.hs_code,
		quantity = EXCLUDED.quantity,
		price = EXCLUDED.price,
		discount = EXCLUDED.discount;
	`

	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("could not begin db transaction: %w", err)
	}

	stmt, err := tx.Prepare(orderQuery)
	if err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("could not prepare statement: %w", err)
	}
	defer stmt.Close()

	itemStmt, err := tx.Prepare(itemQuery)
	if err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("could not prepare item statement: %w", err)
	}
	defer itemStmt.Close()

	updatedCount := 0
	for _, so := range shopifyOrders {
		var currentCustName, custEmail, custPhone, custCity, custState, custCountry string

		// Fallbacks for restricted PII
		custEmail = so.Email
		if so.Customer != nil {
			currentCustName = strings.TrimSpace(so.Customer.DisplayName)
			if currentCustName == "" {
				currentCustName = strings.TrimSpace(so.Customer.FirstName + " " + so.Customer.LastName)
			}
		}

		if so.ShippingAddress != nil {
			if currentCustName == "" {
				currentCustName = strings.TrimSpace(so.ShippingAddress.Name)
			}
			custPhone = so.ShippingAddress.Phone
			custCity = so.ShippingAddress.City
			custState = so.ShippingAddress.Province
			custCountry = so.ShippingAddress.Country
		} else if so.BillingAddress != nil {
			if currentCustName == "" {
				currentCustName = strings.TrimSpace(so.BillingAddress.Name)
			}
			if custPhone == "" {
				custPhone = so.BillingAddress.Phone
			}
			custCity = so.BillingAddress.City
			custState = so.BillingAddress.Province
			custCountry = so.BillingAddress.Country
		}

		if currentCustName == "" {
			currentCustName = "Valued Customer"
		}

		// determine status for local display
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
			// Use displayStatus for human-readable delivery state
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

		idStr := strings.TrimPrefix(so.ID, "gid://shopify/Order/")

		// Map Shopify sourceName to internal source_id
		sourceID := "shopify"
		switch strings.ToLower(so.SourceName) {
		case "amazon":
			sourceID = "amazon"
		case "pos":
			sourceID = "pos"
		}

		_, err = stmt.Exec(
			idStr,
			sourceID,  // mapped source_id
			idStr,     // external_order_id
			so.Name,
			so.CurrentTotalPriceSet.ShopMoney.Amount,
			so.CurrentSubtotalPriceSet.ShopMoney.Amount,
			so.CurrentTotalTaxSet.ShopMoney.Amount,
			so.ProcessedAt, // Use ProcessedAt for the database record
			so.UpdatedAt,
			currentCustName,
			custEmail,
			custPhone,
			custCity,
			custState,
			custCountry,
			status,
			financialStatus,
			fulfillmentStatus,
			deliveryStatus,
			trackingNumber,
			shippingCompany,
			trackingUrl,
		)
		if err != nil {
			log.Printf("Failed to upsert order %s: %v", idStr, err)
			continue
		}

		// Insert Line Items
		for _, edge := range so.LineItems.Edges {
			li := edge.Node
			hsCode := ""
			if li.Variant != nil {
				hsCode = li.Variant.InventoryItem.HarmonizedSystemCode
			}

			itemIdStr := strings.TrimPrefix(li.ID, "gid://shopify/LineItem/")
			_, err = itemStmt.Exec(
				itemIdStr,
				idStr,
				li.Title,
				li.SKU,
				hsCode,
				li.Quantity,
				li.OriginalTotalSet.ShopMoney.Amount, // Use originalTotalSet as total price for the item
				li.TotalDiscountSet.ShopMoney.Amount,
			)
			if err != nil {
				log.Printf("Failed to upsert line item %s for order %s: %v", itemIdStr, idStr, err)
			}
		}

		updatedCount++
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("could not commit transaction: %w", err)
	}

	log.Printf("Successfully synchronized %d orders and their items into PostgreSQL.", updatedCount)
	return updatedCount, nil
}

// ResetAndSync wipes all orders locally and performs a full synchronization
func (s *SyncService) ResetAndSync() (int, error) {
	err := db.TruncateOrders(s.db)
	if err != nil {
		return 0, err
	}
	return s.Sync()
}

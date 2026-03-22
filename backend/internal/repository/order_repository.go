package repository

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"mi-tech/internal/entity"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// gormOrderRepository is the GORM implementation of OrderRepository.
type gormOrderRepository struct {
	db *gorm.DB
}

// NewOrderRepository creates a new GORM-backed OrderRepository.
func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &gormOrderRepository{db: db}
}

func (r *gormOrderRepository) List(filter OrderFilter) ([]entity.Order, int, error) {
	query := r.db.Model(&entity.Order{})

	if filter.StartDate != "" {
		query = query.Where("created_at >= ?", filter.StartDate)
	}
	if filter.EndDate != "" {
		query = query.Where("created_at <= ?", filter.EndDate)
	}
	if filter.Search != "" {
		searchTerm := "%" + filter.Search + "%"
		query = query.Where("order_number ILIKE ? OR customer_name ILIKE ?", searchTerm, searchTerm)
	}
	if filter.Source != "" {
		query = query.Where("source_id = ?", filter.Source)
	}
	if filter.FinancialStatus != "" {
		if strings.ToLower(filter.FinancialStatus) == "unpaid" {
			query = query.Where("financial_status != ?", "paid")
		} else {
			query = query.Where("financial_status = ?", filter.FinancialStatus)
		}
	}
	if filter.FulfillmentStatus != "" {
		query = query.Where("fulfillment_status = ?", filter.FulfillmentStatus)
	}

	// Count total matching
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count orders: %w", err)
	}

	// Sorting
	orderClause := "created_at DESC"
	if filter.SortBy != "" {
		dir := "DESC"
		if strings.ToUpper(filter.SortOrder) == "ASC" {
			dir = "ASC"
		}
		allowed := map[string]bool{
			"created_at": true, "order_number": true, "total_price": true,
			"customer_name": true, "source_id": true, "financial_status": true,
			"fulfillment_status": true,
		}
		if allowed[filter.SortBy] {
			orderClause = filter.SortBy + " " + dir
		}
	}

	// Pagination
	page := filter.Page
	limit := filter.Limit
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 25
	}
	offset := (page - 1) * limit

	var orders []entity.Order
	if err := query.Order(orderClause).Offset(offset).Limit(limit).Find(&orders).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list orders: %w", err)
	}

	return orders, int(totalCount), nil
}

func (r *gormOrderRepository) GetByID(id int64) (entity.Order, error) {
	var order entity.Order
	err := r.db.First(&order, id).Error
	return order, err
}

func (r *gormOrderRepository) GetByFlexibleID(id string) (entity.Order, error) {
	var order entity.Order
	// 1. Try numeric primary key
	if idInt, err := strconv.ParseInt(id, 10, 64); err == nil {
		if err := r.db.First(&order, idInt).Error; err == nil {
			return order, nil
		}
	}
	// 2. Fallback to external_order_id
	err := r.db.Where("external_order_id = ?", id).First(&order).Error
	return order, err
}

func (r *gormOrderRepository) GetByExternalID(externalID string) (entity.Order, error) {
	var order entity.Order
	err := r.db.Where("external_order_id = ?", externalID).First(&order).Error
	return order, err
}

func (r *gormOrderRepository) isWeak(s *string) bool {
	if s == nil {
		return true
	}
	val := strings.TrimSpace(*s)
	return val == "" || val == "Valued Customer" || val == "pending"
}

func (r *gormOrderRepository) mergePII(existing *entity.Order, incoming *entity.Order) {
	updatedAny := false

	check := func(incomingField **string, existingField **string) {
		if r.isWeak(*incomingField) && !r.isWeak(*existingField) {
			*incomingField = *existingField
		} else if !r.isWeak(*incomingField) && r.isWeak(*existingField) {
			updatedAny = true
		}
	}

	check(&incoming.CustomerName, &existing.CustomerName)
	check(&incoming.CustomerFirstName, &existing.CustomerFirstName)
	check(&incoming.CustomerLastName, &existing.CustomerLastName)
	check(&incoming.CustomerEmail, &existing.CustomerEmail)
	check(&incoming.CustomerPhone, &existing.CustomerPhone)
	check(&incoming.CustomerCity, &existing.CustomerCity)
	check(&incoming.CustomerState, &existing.CustomerState)
	check(&incoming.CustomerCountry, &existing.CustomerCountry)
	check(&incoming.CustomerAddress1, &existing.CustomerAddress1)
	check(&incoming.CustomerAddress2, &existing.CustomerAddress2)
	check(&incoming.CustomerZip, &existing.CustomerZip)

	if updatedAny {
		log.Printf("Repository: Order %s PII upgraded from weak/empty to strong data.", incoming.ExternalOrderID)
	}
}

func (r *gormOrderRepository) Upsert(order entity.Order) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Check if the order already exists to preserve PII
		var existing entity.Order
		err := tx.Where("source_id = ? AND external_order_id = ?", order.SourceID, order.ExternalOrderID).
			Select("id", "customer_name", "customer_first_name", "customer_last_name", "customer_email", "customer_phone",
				"customer_city", "customer_state", "customer_country", "customer_address1", "customer_address2", "customer_zip").
			First(&existing).Error

		if err == nil {
			order.ID = existing.ID // Crucial to link line items correctly and resolve primary key conflict
			r.mergePII(&existing, &order)
		}

		// 2. Upsert the order
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "external_order_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"order_number", "financial_status", "fulfillment_status", "delivery_status",
				"tracking_number", "shipping_company", "tracking_url", "status", "updated_at",
				"cancelled_at", "cancel_reason", "total_price", "subtotal_price",
				"total_tax", "customer_name", "customer_email", "customer_phone",
				"customer_city", "customer_state", "customer_country",
				"customer_address1", "customer_address2", "customer_zip",
				"customer_first_name", "customer_last_name", "raw_payload",
			}),
		}).Omit("LineItems").Create(&order).Error; err != nil {
			return fmt.Errorf("failed to upsert order: %w", err)
		}

		// 3. Explicitly synchronize line items
		// We delete all and re-create to ensure quantities and titles are fresh.
		if err := tx.Where("order_id = ?", order.ID).Delete(&entity.LineItem{}).Error; err != nil {
			return fmt.Errorf("failed to clean old line items: %w", err)
		}

		for _, item := range order.LineItems {
			item.OrderID = order.ID
			if err := tx.Clauses(clause.OnConflict{
				UpdateAll: true,
			}).Create(&item).Error; err != nil {
				return fmt.Errorf("failed to insert line item %s: %w", item.ID, err)
			}
		}

		return nil
	})
}

func (r *gormOrderRepository) UpsertBatch(orders []entity.Order) error {
	if len(orders) == 0 {
		return nil
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Fetch existing orders in one batch to preserve PII
		externalIDs := make([]string, len(orders))
		sourceIDs := make(map[string]bool)
		for i, o := range orders {
			externalIDs[i] = o.ExternalOrderID
			sourceIDs[o.SourceID] = true
		}

		// Collect unique source IDs (usually just one, but let's be safe)
		var uniqueSources []string
		for s := range sourceIDs {
			uniqueSources = append(uniqueSources, s)
		}

		var existingOrders []entity.Order
		err := tx.Where("source_id IN ? AND external_order_id IN ?", uniqueSources, externalIDs).
			Select("id", "source_id", "external_order_id", "customer_name", "customer_first_name", "customer_last_name", "customer_email", "customer_phone",
				"customer_city", "customer_state", "customer_country", "customer_address1", "customer_address2", "customer_zip").
			Find(&existingOrders).Error

		if err != nil {
			return fmt.Errorf("failed to fetch existing orders for merge: %w", err)
		}

		// Create a map for O(1) lookup: key = source_id:external_order_id
		// This protects against map collisions if multiple sources use same external IDs
		existingMap := make(map[string]entity.Order)
		for _, e := range existingOrders {
			key := fmt.Sprintf("%s:%s", e.SourceID, e.ExternalOrderID)
			existingMap[key] = e
		}

		for i := range orders {
			key := fmt.Sprintf("%s:%s", orders[i].SourceID, orders[i].ExternalOrderID)
			// Merge PII if existing order found
			if existing, found := existingMap[key]; found {
				orders[i].ID = existing.ID // Crucial to link line items correctly
				r.mergePII(&existing, &orders[i])
			}
		}

		// 2. Batch Upsert Orders (Omit LineItems to handle them separately)
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "external_order_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"order_number", "total_price", "subtotal_price", "total_tax",
				"updated_at", "customer_name", "customer_email", "customer_phone",
				"customer_city", "customer_state", "customer_country", "status",
				"financial_status", "fulfillment_status", "delivery_status",
				"tracking_number", "shipping_company", "tracking_url",
				"customer_first_name", "customer_last_name", "customer_address1", "customer_address2", "customer_zip",
				"raw_payload",
			}),
		}).Omit("LineItems").Create(&orders).Error; err != nil {
			return fmt.Errorf("failed to batch upsert orders: %w", err)
		}

		// 3. Flatten and Batch Upsert Line Items
		var allLineItems []entity.LineItem
		for i := range orders {
			for j := range orders[i].LineItems {
				orders[i].LineItems[j].OrderID = orders[i].ID
				allLineItems = append(allLineItems, orders[i].LineItems[j])
			}
		}

		if len(allLineItems) > 0 {
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"title", "sku", "hs_code", "quantity", "price", "discount"}),
			}).Create(&allLineItems).Error; err != nil {
				return fmt.Errorf("failed to batch upsert line items: %w", err)
			}
		}
		return nil
	})
}

func (r *gormOrderRepository) UpdateStatus(externalOrderID string, financialStatus, fulfillmentStatus string) error {
	return r.db.Model(&entity.Order{}).
		Where("external_order_id = ?", externalOrderID).
		Updates(map[string]interface{}{
			"financial_status":   financialStatus,
			"fulfillment_status": fulfillmentStatus,
			"status":             fulfillmentStatus,
			"updated_at":         time.Now(),
		}).Error
}

func (r *gormOrderRepository) UpdateOrderStatus(id int64, status string) (int64, error) {
	status = strings.ToUpper(status)
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	if status == "CANCELLED" {
		updates["fulfillment_status"] = "restocked"
		// Only set cancelled_at if not already set
		r.db.Model(&entity.Order{}).Where("id = ? AND cancelled_at IS NULL", id).
			Update("cancelled_at", time.Now())
	} else if status == "FULFILLED" || status == "UNFULFILLED" {
		updates["fulfillment_status"] = strings.ToLower(status)
	}

	result := r.db.Model(&entity.Order{}).Where("id = ?", id).Updates(updates)
	return result.RowsAffected, result.Error
}

func (r *gormOrderRepository) CancelOrder(externalOrderID string, cancelledAt *string, reason string) error {
	updates := map[string]interface{}{
		"status":        "CANCELLED",
		"cancel_reason": reason,
		"updated_at":    time.Now(),
	}
	if cancelledAt != nil {
		updates["cancelled_at"] = *cancelledAt
	}

	return r.db.Model(&entity.Order{}).
		Where("external_order_id = ?", externalOrderID).
		Updates(updates).Error
}

func (r *gormOrderRepository) UpdateTrackingInfo(externalOrderID string, trackingNumber, shippingCompany, trackingUrl, deliveryStatus string) error {
	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}
	if trackingNumber != "" {
		updates["tracking_number"] = trackingNumber
	}
	if shippingCompany != "" {
		updates["shipping_company"] = shippingCompany
	}
	if trackingUrl != "" {
		updates["tracking_url"] = trackingUrl
	}
	if deliveryStatus != "" {
		updates["delivery_status"] = deliveryStatus
	}

	return r.db.Model(&entity.Order{}).
		Where("external_order_id = ?", externalOrderID).
		Updates(updates).Error
}

func (r *gormOrderRepository) ListSources() ([]entity.Source, error) {
	var sources []entity.Source
	err := r.db.Where("enabled = ?", true).Order("name ASC").Find(&sources).Error
	return sources, err
}

func (r *gormOrderRepository) TruncateAll() error {
	if err := r.db.Exec("TRUNCATE TABLE order_line_items CASCADE").Error; err != nil {
		return fmt.Errorf("failed to truncate order_line_items: %w", err)
	}
	if err := r.db.Exec("TRUNCATE TABLE orders CASCADE").Error; err != nil {
		return fmt.Errorf("failed to truncate orders: %w", err)
	}
	if err := r.db.Exec("TRUNCATE TABLE webhook_events CASCADE").Error; err != nil {
		return fmt.Errorf("failed to truncate webhook_events: %w", err)
	}
	log.Println("Successfully truncated orders, order_line_items, and webhook_events tables")
	return nil
}

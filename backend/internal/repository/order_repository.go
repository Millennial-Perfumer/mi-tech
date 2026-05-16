package repository

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"mi-tech/internal/dto"
	"mi-tech/internal/entity"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// gormOrderRepository is the GORM implementation of OrderRepository.
type gormOrderRepository struct {
	db *gorm.DB
}

// orderListAllowedSortColumns defines the permitted columns for sorting to prevent SQL injection.
// Hoisted to package level to avoid redundant map allocations per request.
var orderListAllowedSortColumns = map[string]bool{
	"created_at": true, "order_number": true, "total_price": true,
	"customer_name": true, "source_id": true, "financial_status": true,
	"fulfillment_status": true,
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
		query = query.Where("order_number ILIKE ? OR customer_name ILIKE ? OR customer_email ILIKE ? OR customer_phone ILIKE ? OR tracking_number ILIKE ?",
			searchTerm, searchTerm, searchTerm, searchTerm, searchTerm)
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
		// Security: Use allowlist for sortBy to prevent SQL injection.
		// Performance: Uses pre-allocated package-level map.
		if orderListAllowedSortColumns[filter.SortBy] {
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

func (r *gormOrderRepository) syncInventoryDeltas(tx *gorm.DB, order *entity.Order, oldItems []entity.LineItem) ([]int, error) {
	var affectedIDs []int
	
	// Index by SKU: delta = New - Old
	skuDeltas := make(map[string]int)

	// New items (to be deducted)
	for _, li := range order.LineItems {
		if li.SKU != nil && *li.SKU != "" {
			skuDeltas[*li.SKU] += li.Quantity
		}
	}

	// Old items (to be restocked)
	for _, li := range oldItems {
		if li.SKU != nil && *li.SKU != "" {
			skuDeltas[*li.SKU] -= li.Quantity
		}
	}

	for sku, delta := range skuDeltas {
		if delta == 0 {
			continue
		}

		var mapping entity.InventoryMapping
		err := tx.Where("platform = ? AND external_sku = ?", order.SourceID, sku).First(&mapping).Error
		if err != nil {
			// No mapping for this SKU, skip
			continue
		}

		// Adjust stock: subtract the delta
		// If delta is positive (more sold), stock decreases.
		// If delta is negative (restocked), stock increases.
		if err := tx.Model(&entity.InventoryItem{}).
			Where("id = ?", mapping.InventoryItemID).
			Update("current_stock", gorm.Expr("current_stock - ?", delta)).Error; err != nil {
			return nil, fmt.Errorf("failed to adjust stock for %s (delta %d): %w", sku, delta, err)
		}

		// Log the adjustment for audit trail
		reason := "sale"
		if delta < 0 {
			reason = "adjustment"
		}
		if err := tx.Create(&entity.InventoryLog{
			InventoryItemID: mapping.InventoryItemID,
			Delta:           -delta,
			Reason:          reason,
			Platform:        order.SourceID,
			ExternalOrderID: &order.ExternalOrderID,
		}).Error; err != nil {
			return nil, fmt.Errorf("failed to log inventory adjustment: %w", err)
		}

		affectedIDs = append(affectedIDs, mapping.InventoryItemID)
	}

	order.InventoryDeducted = true
	return affectedIDs, nil
}

func (r *gormOrderRepository) Upsert(order entity.Order) ([]int, error) {
	var affectedIDs []int
	if order.CustomerPhone != nil {
		normalized := entity.NormalizePhone(*order.CustomerPhone)
		order.CustomerPhone = &normalized
	}
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Check if the order already exists to preserve PII
		var existing entity.Order
		err := tx.Where("source_id = ? AND external_order_id = ?", order.SourceID, order.ExternalOrderID).
			Select("id", "customer_name", "customer_first_name", "customer_last_name", "customer_email", "customer_phone",
				"customer_city", "customer_state", "customer_country", "customer_address1", "customer_address2", "customer_zip", "delivered_at", "inventory_deducted").
			First(&existing).Error

		if err == nil {
			order.ID = existing.ID // Crucial to link line items correctly and resolve primary key conflict
			r.mergePII(&existing, &order)
			
			// Preserve delivered_at if already set
			if existing.DeliveredAt != nil {
				order.DeliveredAt = existing.DeliveredAt
			}
			order.InventoryDeducted = existing.InventoryDeducted
		}

		// Stamp delivered_at if status transition detected
		if strings.ToLower(strings.TrimSpace(entity.DerefStr(order.DeliveryStatus))) == "delivered" && order.DeliveredAt == nil {
			now := time.Now()
			order.DeliveredAt = &now
			// Initialize feedback status to 'Pending' (1) if it's not already set
			if order.FeedbackStatusID == nil || *order.FeedbackStatusID == 0 {
				pending := 1
				order.FeedbackStatusID = &pending
			}
		}

		// 1.5 Auto-Link Customer PII if phone is missing but external_id is present
		if r.isWeak(order.CustomerPhone) && !r.isWeak(order.CustomerExternalID) {
			var cust entity.Customer
			if err := tx.Where("external_id = ?", *order.CustomerExternalID).First(&cust).Error; err == nil {
				if cust.PhoneNumber != "" {
					order.CustomerPhone = &cust.PhoneNumber
					log.Printf("Repository: Order %s auto-linked to customer %s to restore phone.", order.OrderNumber, cust.PhoneNumber)

					// Also restore other fields if possible
					if r.isWeak(order.CustomerFirstName) { order.CustomerFirstName = cust.FirstName }
					if r.isWeak(order.CustomerLastName) { order.CustomerLastName = cust.LastName }
					if r.isWeak(order.CustomerEmail) { order.CustomerEmail = cust.Email }
					if r.isWeak(order.CustomerCity) { order.CustomerCity = cust.City }
					if r.isWeak(order.CustomerState) { order.CustomerState = cust.State }
				}
			}
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
				"customer_first_name", "customer_last_name", "raw_payload", "customer_external_id",
				"total_discount", "delivered_at", "inventory_deducted",
			}),
		}).Omit("LineItems").Create(&order).Error; err != nil {
			return fmt.Errorf("failed to upsert order: %w", err)
		}

		// 3. Explicitly synchronize line items in batch
		// First, fetch old line items to calculate inventory deltas
		var oldLineItems []entity.LineItem
		if order.ID != 0 {
			if err := tx.Where("order_id = ?", order.ID).Find(&oldLineItems).Error; err != nil {
				return fmt.Errorf("failed to fetch old line items: %w", err)
			}
		}

		// We delete all and re-create to ensure quantities and titles are fresh.
		if err := tx.Where("order_id = ?", order.ID).Delete(&entity.LineItem{}).Error; err != nil {
			return fmt.Errorf("failed to clean old line items: %w", err)
		}

		if len(order.LineItems) > 0 {
			// Optimization: Set OrderID by index to update original slice elements.
			for i := range order.LineItems {
				order.LineItems[i].OrderID = order.ID
			}
			// Optimization: Batch insert line items in a single O(1) roundtrip.
			if err := tx.Clauses(clause.OnConflict{
				UpdateAll: true,
			}).Create(&order.LineItems).Error; err != nil {
				return fmt.Errorf("failed to batch insert line items: %w", err)
			}
		}

		// 4. Sync Inventory Deltas
		ids, err := r.syncInventoryDeltas(tx, &order, oldLineItems)
		if err != nil {
			return err
		}
		affectedIDs = ids

		// Update order again with inventory_deducted flag
		return tx.Model(&order).Update("inventory_deducted", order.InventoryDeducted).Error
	})
	return affectedIDs, err
}

func (r *gormOrderRepository) UpsertBatch(orders []entity.Order) ([]int, error) {
	var affectedIDs []int
	if len(orders) == 0 {
		return affectedIDs, nil
	}
	// ... (phone normalization)
	err := r.db.Transaction(func(tx *gorm.DB) error {
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
				"customer_city", "customer_state", "customer_country", "customer_address1", "customer_address2", "customer_zip", "delivered_at", "inventory_deducted").
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
				
				// Preserve delivered_at if already set
				if existing.DeliveredAt != nil {
					orders[i].DeliveredAt = existing.DeliveredAt
				}
				orders[i].InventoryDeducted = existing.InventoryDeducted
			}

			// Stamp delivered_at if status transition detected
			if strings.ToLower(strings.TrimSpace(entity.DerefStr(orders[i].DeliveryStatus))) == "delivered" && orders[i].DeliveredAt == nil {
				now := time.Now()
				orders[i].DeliveredAt = &now
				// Initialize feedback status to 'Pending' (1) if it's not already set
				if orders[i].FeedbackStatusID == nil || *orders[i].FeedbackStatusID == 0 {
					pending := 1
					orders[i].FeedbackStatusID = &pending
				}
			}
		}

		// 1.5 Fetch all existing line items for these orders to calculate deltas
		orderIDs := make([]int64, 0)
		for _, o := range orders {
			if o.ID != 0 {
				orderIDs = append(orderIDs, o.ID)
			}
		}
		
		oldLinesByOrder := make(map[int64][]entity.LineItem)
		if len(orderIDs) > 0 {
			var allOldLineItems []entity.LineItem
			if err := tx.Where("order_id IN ?", orderIDs).Find(&allOldLineItems).Error; err != nil {
				return fmt.Errorf("failed to fetch old line items: %w", err)
			}
			for _, li := range allOldLineItems {
				oldLinesByOrder[li.OrderID] = append(oldLinesByOrder[li.OrderID], li)
			}

			// Clean old line items to handle removals correctly (Batch Sync)
			if err := tx.Where("order_id IN ?", orderIDs).Delete(&entity.LineItem{}).Error; err != nil {
				return fmt.Errorf("failed to clean old line items: %w", err)
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
				"raw_payload", "customer_external_id", "total_discount", "delivered_at", "inventory_deducted",
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

		// 4. Optimized Batch Sync Inventory Deltas
		type skuKey struct {
			Platform string
			SKU      string
		}
		totalSKUDeltas := make(map[skuKey]int)
		pairs := make([][]interface{}, 0)

		for i := range orders {
			// Calculate deltas for this order: New - Old
			// Note: We use the same logic as syncInventoryDeltas to preserve behavior
			orderSKUDeltas := make(map[string]int)
			for _, li := range orders[i].LineItems {
				if li.SKU != nil && *li.SKU != "" {
					orderSKUDeltas[*li.SKU] += li.Quantity
				}
			}
			for _, li := range oldLinesByOrder[orders[i].ID] {
				if li.SKU != nil && *li.SKU != "" {
					orderSKUDeltas[*li.SKU] -= li.Quantity
				}
			}

			for sku, delta := range orderSKUDeltas {
				if delta == 0 {
					continue
				}
				key := skuKey{Platform: orders[i].SourceID, SKU: sku}
				if _, exists := totalSKUDeltas[key]; !exists {
					pairs = append(pairs, []interface{}{key.Platform, key.SKU})
				}
				totalSKUDeltas[key] += delta
			}
			orders[i].InventoryDeducted = true
		}

		if len(pairs) > 0 {
			var mappings []entity.InventoryMapping
			// Optimization: Batch fetch all relevant mappings in a single O(1) round-trip using tuple IN
			if err := tx.Where("(platform, external_sku) IN ?", pairs).Find(&mappings).Error; err != nil {
				return fmt.Errorf("failed to fetch mappings in batch: %w", err)
			}

			itemDeltas := make(map[int]int)
			for _, m := range mappings {
				key := skuKey{Platform: m.Platform, SKU: m.ExternalSKU}
				itemDeltas[m.InventoryItemID] += totalSKUDeltas[key]
			}

			for itemID, delta := range itemDeltas {
				if delta == 0 {
					continue
				}
				// Optimization: Aggregate deltas by InventoryItemID to minimize update queries
				if err := tx.Model(&entity.InventoryItem{}).
					Where("id = ?", itemID).
					Update("current_stock", gorm.Expr("current_stock - ?", delta)).Error; err != nil {
					return fmt.Errorf("failed to adjust stock for item %d: %w", itemID, err)
				}

				// Note: Batch upsert logs are aggregated by itemID for performance.
				// For detailed order-level logs, use individual upserts.
				if err := tx.Create(&entity.InventoryLog{
					InventoryItemID: itemID,
					Delta:           -delta,
					Reason:          "batch_sync",
					Platform:        uniqueSources[0], // Assuming batch is usually from one source
				}).Error; err != nil {
					return fmt.Errorf("failed to log batch inventory adjustment: %w", err)
				}

				affectedIDs = append(affectedIDs, itemID)
			}
		}

		// Optimization: Batch update inventory_deducted flag for all processed orders
		orderIDsForFlag := make([]int64, len(orders))
		for i := range orders {
			orderIDsForFlag[i] = orders[i].ID
		}
		if err := tx.Model(&entity.Order{}).Where("id IN ?", orderIDsForFlag).Update("inventory_deducted", true).Error; err != nil {
			return fmt.Errorf("failed to batch update inventory_deducted flag: %w", err)
		}

		return nil
	})
	return affectedIDs, err
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

func (r *gormOrderRepository) UpdateFinancialStatus(id int64, status string) error {
	return r.db.Model(&entity.Order{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"financial_status": status,
			"updated_at":       time.Now(),
		}).Error
}

func (r *gormOrderRepository) UpdateOrderStatus(id int64, status string) (int64, error) {
	status = strings.ToLower(status)
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

	if status == "DELIVERED" {
		updates["delivery_status"] = "delivered"
		updates["delivered_at"] = time.Now()
		// Initialize feedback status to 'Pending' (1) if it's not already set
		updates["feedback_status_id"] = gorm.Expr("COALESCE(feedback_status_id, 1)")
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
	commonUpdates := map[string]interface{}{
		"updated_at": time.Now(),
	}
	if trackingNumber != "" {
		commonUpdates["tracking_number"] = trackingNumber
	}
	if shippingCompany != "" {
		commonUpdates["shipping_company"] = shippingCompany
	}
	if trackingUrl != "" {
		commonUpdates["tracking_url"] = trackingUrl
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&entity.Order{}).Where("external_order_id = ?", externalOrderID).Updates(commonUpdates).Error; err != nil {
			return err
		}
		if deliveryStatus != "" {
			updates := map[string]interface{}{
				"delivery_status": deliveryStatus,
				"updated_at":      time.Now(),
			}
			if deliveryStatus == "delivered" {
				updates["delivered_at"] = time.Now()
				// Initialize feedback status to 'Pending' (1) if it's not already set
				updates["feedback_status_id"] = gorm.Expr("COALESCE(feedback_status_id, 1)")
			}
			// Protect 'delivered' status from being overwritten by 'in_transit' or other earlier states
			return tx.Model(&entity.Order{}).
				Where("external_order_id = ? AND (delivery_status != 'delivered' OR delivery_status IS NULL)", externalOrderID).
				Updates(updates).Error
		}
		return nil
	})
}

func (r *gormOrderRepository) UpdateOrderDetails(id int64, order entity.Order) error {
	return r.db.Model(&entity.Order{}).Where("id = ?", id).Updates(map[string]interface{}{
		"customer_first_name": order.CustomerFirstName,
		"customer_last_name":  order.CustomerLastName,
		"customer_name":       order.CustomerName,
		"customer_email":      order.CustomerEmail,
		"customer_phone":      order.CustomerPhone,
		"customer_address1":   order.CustomerAddress1,
		"customer_address2":   order.CustomerAddress2,
		"customer_city":       order.CustomerCity,
		"customer_state":      order.CustomerState,
		"customer_zip":        order.CustomerZip,
		"customer_country":    order.CustomerCountry,
		"updated_at":          time.Now(),
	}).Error
}

func (r *gormOrderRepository) GetCustomerStats(phone string) (totalOrders int, totalSpent float64, err error) {
	row := r.db.Raw("SELECT COUNT(*), COALESCE(SUM(total_price), 0) FROM orders WHERE customer_phone = ? AND COALESCE(LOWER(status), '') != ?", phone, "cancelled").Row()
	err = row.Scan(&totalOrders, &totalSpent)
	return
}

func (r *gormOrderRepository) GetCustomersStats(phones []string) (map[string]struct{ Count int; Sum float64 }, error) {
	if len(phones) == 0 {
		return make(map[string]struct{ Count int; Sum float64 }), nil
	}

	type result struct {
		Phone string
		Count int
		Sum   float64
	}
	var results []result
	err := r.db.Model(&entity.Order{}).
		Where("customer_phone IN ? AND COALESCE(LOWER(status), '') != ?", phones, "cancelled").
		Select("customer_phone as phone, COUNT(*) as count, COALESCE(SUM(total_price), 0) as sum").
		Group("customer_phone").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	statsMap := make(map[string]struct{ Count int; Sum float64 })
	for _, res := range results {
		statsMap[res.Phone] = struct {
			Count int
			Sum   float64
		}{Count: res.Count, Sum: res.Sum}
	}
	return statsMap, nil
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

func (r *gormOrderRepository) MarkAsDelivered(id int64) error {
	now := time.Now()
	return r.db.Model(&entity.Order{}).Where("id = ?", id).Updates(map[string]interface{}{
		"delivery_status":    "delivered",
		"delivered_at":       now,
		"updated_at":         now,
		"feedback_status_id": gorm.Expr("COALESCE(feedback_status_id, 1)"),
	}).Error
}

func (r *gormOrderRepository) GetOrdersForFeedback(delayMinutes int) ([]entity.Order, error) {
	var orders []entity.Order
	// Logic: Delivered but feedback is still 'pending' (status_id = 1 or NULL) and delivered_at <= delayMinutes ago
	threshold := time.Now().Add(time.Duration(-delayMinutes) * time.Minute)
	
	// Use TRIM(LOWER()) to be resilient against trailing spaces found in DB ('delivered ')
	err := r.db.Where("TRIM(LOWER(delivery_status)) = ? AND (feedback_status_id = ? OR feedback_status_id IS NULL) AND delivered_at <= ?", "delivered", 1, threshold).
		Order("delivered_at DESC").
		Find(&orders).Error
	return orders, err
}

func (r *gormOrderRepository) UpdateFeedbackStatus(id int64, statusID int) error {
	updates := map[string]interface{}{
		"feedback_status_id": statusID,
	}
	if statusID == 2 { // Sent
		updates["feedback_sent_at"] = time.Now()
	}
	return r.db.Model(&entity.Order{}).Where("id = ?", id).Updates(updates).Error
}

func (r *gormOrderRepository) GetByIDAndPhone(id int64, phone string) (entity.Order, error) {
	var order entity.Order
	// Use LIKE to handle variations in phone format (e.g. with country code)
	// We check if the stored phone ends with the provided phone or vice versa
	searchPhone := phone
	if len(phone) > 10 {
		searchPhone = phone[len(phone)-10:]
	}
	err := r.db.Where("id = ? AND (customer_phone LIKE ? OR ? LIKE '%' || customer_phone)", id, "%"+searchPhone, phone).First(&order).Error
	return order, err
}

func (r *gormOrderRepository) SaveCustomerFeedback(feedback entity.CustomerFeedback) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "order_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"rating", "message", "updated_at"}),
	}).Create(&feedback).Error
}

func (r *gormOrderRepository) GetCustomerFeedback() ([]dto.FeedbackResponse, error) {
	var results []dto.FeedbackResponse
	err := r.db.Table("customer_feedback").
		Select("customer_feedback.id, customer_feedback.order_id, orders.order_number, orders.customer_name, customer_feedback.rating, customer_feedback.message as comment, customer_feedback.created_at").
		Joins("JOIN orders ON orders.id = customer_feedback.order_id").
		Order("customer_feedback.created_at DESC").
		Scan(&results).Error
	return results, err
}

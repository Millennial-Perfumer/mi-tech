package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"mi-tech/internal/domain/order/entity"

	"gorm.io/gorm"
)

// EventFilter bounds history reads so an order/customer timeline cannot
// accidentally return an unbounded result set.
type EventFilter struct {
	OrderID         int64
	CustomerID      int64
	ExternalOrderID string
	CustomerPhone   string
	Search          string
	EventType       string
	StartDate       string
	EndDate         string
	Page            int
	Limit           int
}

// EventRepository is the read surface for order/customer audit history.
// Writes are kept on the same transaction as the current-state mutation by
// the concrete repositories in this package.
type EventRepository interface {
	ListOrderEvents(ctx context.Context, filter EventFilter) ([]entity.OrderEvent, int64, error)
	ListCustomerEvents(ctx context.Context, filter EventFilter) ([]entity.CustomerEvent, int64, error)
}

type gormEventRepository struct {
	db *gorm.DB
}

// NewEventRepository creates the history repository used by REST and MCP
// read handlers.
func NewEventRepository(db *gorm.DB) EventRepository {
	return &gormEventRepository{db: db}
}

func newEventRepository(db *gorm.DB) *gormEventRepository {
	return &gormEventRepository{db: db}
}

// RecordOrderChanges lets another domain that mutates the shared orders table
// (for example B2B invoice synchronization) use the same audit semantics.
func RecordOrderChanges(tx *gorm.DB, before, after *entity.Order, source, actorType string) error {
	return newEventRepository(tx).recordOrderChanges(tx, before, after, source, actorType)
}

func (r *gormEventRepository) ListOrderEvents(ctx context.Context, filter EventFilter) ([]entity.OrderEvent, int64, error) {
	query := r.db.WithContext(ctx).Model(&entity.OrderEvent{})
	if filter.OrderID > 0 {
		query = query.Where("order_id = ?", filter.OrderID)
	}
	if filter.ExternalOrderID != "" {
		query = query.Where("external_order_id = ?", filter.ExternalOrderID)
	}
	if filter.EventType != "" {
		query = query.Where("event_type = ?", filter.EventType)
	}
	var err error
	if query, err = addEventDateFilters(query, filter.StartDate, filter.EndDate); err != nil {
		return nil, 0, err
	}
	if filter.Search != "" {
		term := "%" + filter.Search + "%"
		query = query.Where("external_order_id ILIKE ? OR before_data::text ILIKE ? OR after_data::text ILIKE ? OR diff_data::text ILIKE ?", term, term, term, term)
	}

	return listOrderEvents(query, filter)
}

func (r *gormEventRepository) ListCustomerEvents(ctx context.Context, filter EventFilter) ([]entity.CustomerEvent, int64, error) {
	query := r.db.WithContext(ctx).Model(&entity.CustomerEvent{})
	if filter.CustomerID > 0 {
		query = query.Where("customer_id = ?", filter.CustomerID)
	}
	if filter.OrderID > 0 {
		query = query.Where("order_id = ?", filter.OrderID)
	}
	if filter.CustomerPhone != "" {
		query = query.Where("customer_phone = ?", filter.CustomerPhone)
	}
	if filter.EventType != "" {
		query = query.Where("event_type = ?", filter.EventType)
	}
	var err error
	if query, err = addEventDateFilters(query, filter.StartDate, filter.EndDate); err != nil {
		return nil, 0, err
	}
	if filter.Search != "" {
		term := "%" + filter.Search + "%"
		query = query.Where("customer_phone ILIKE ? OR before_data::text ILIKE ? OR after_data::text ILIKE ? OR diff_data::text ILIKE ?", term, term, term, term)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count customer events: %w", err)
	}

	page, limit := eventPagination(filter.Page, filter.Limit)
	var events []entity.CustomerEvent
	if err := query.Order("occurred_at DESC").Order("id DESC").Offset((page - 1) * limit).Limit(limit).Find(&events).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list customer events: %w", err)
	}
	return events, total, nil
}

func listOrderEvents(query *gorm.DB, filter EventFilter) ([]entity.OrderEvent, int64, error) {
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count order events: %w", err)
	}

	page, limit := eventPagination(filter.Page, filter.Limit)
	var events []entity.OrderEvent
	if err := query.Order("occurred_at DESC").Order("id DESC").Offset((page - 1) * limit).Limit(limit).Find(&events).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list order events: %w", err)
	}
	return events, total, nil
}

func eventPagination(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	return page, limit
}

// addEventDateFilters treats date-only end dates as inclusive. For example,
// end_date=2026-08-30 includes events throughout August 30 by using the
// beginning of August 31 as an exclusive upper bound.
func addEventDateFilters(query *gorm.DB, startDate, endDate string) (*gorm.DB, error) {
	const dateFormat = "2006-01-02"

	var start time.Time
	if startDate != "" {
		parsed, err := time.Parse(dateFormat, startDate)
		if err != nil {
			return nil, fmt.Errorf("invalid start date %q: %w", startDate, err)
		}
		start = parsed
		query = query.Where("occurred_at >= ?", start)
	}

	if endDate != "" {
		end, err := time.Parse(dateFormat, endDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end date %q: %w", endDate, err)
		}
		if !start.IsZero() && start.After(end) {
			return nil, fmt.Errorf("start date %q is after end date %q", startDate, endDate)
		}
		query = query.Where("occurred_at < ?", end.AddDate(0, 0, 1))
	}

	return query, nil
}

// recordOrderChanges records one or more semantic events for a transition.
// It intentionally records the before/after values for only the affected
// field group, which keeps timelines readable and avoids duplicating raw PII.
func (r *gormEventRepository) recordOrderChanges(tx *gorm.DB, before, after *entity.Order, source, actorType string) error {
	if after == nil {
		return nil
	}
	if source == "" {
		source = "system"
	}
	if actorType == "" {
		actorType = "system"
	}

	afterSnapshot := orderSnapshot(after)
	if before == nil {
		return r.createOrderEvent(tx, after, "order_created", source, actorType, nil, afterSnapshot)
	}

	beforeSnapshot := orderSnapshot(before)
	groups := map[string][]string{
		"tracking_changed": {
			"tracking_number", "shipping_company", "tracking_url",
		},
		"delivery_status_changed":    {"delivery_status", "delivered_at"},
		"fulfillment_status_changed": {"fulfillment_status"},
		"payment_status_changed":     {"financial_status"},
		"customer_details_changed": {
			"customer_name", "customer_first_name", "customer_last_name", "customer_email",
			"customer_phone", "customer_city", "customer_state", "customer_country",
			"customer_address1", "customer_address2", "customer_zip", "customer_external_id",
		},
		"order_status_changed":    {"status", "cancelled_at", "cancel_reason"},
		"order_totals_changed":    {"total_price", "subtotal_price", "total_tax", "total_discount"},
		"inventory_state_changed": {"inventory_deducted"},
	}

	recorded := false
	for eventType, fields := range groups {
		beforeGroup := selectFields(beforeSnapshot, fields)
		afterGroup := selectFields(afterSnapshot, fields)
		if reflect.DeepEqual(beforeGroup, afterGroup) {
			continue
		}
		if err := r.createOrderEvent(tx, after, eventType, source, actorType, beforeGroup, afterGroup); err != nil {
			return err
		}
		recorded = true
	}

	if !reflect.DeepEqual(beforeSnapshot, afterSnapshot) && !recorded {
		if err := r.createOrderEvent(tx, after, "order_updated", source, actorType, beforeSnapshot, afterSnapshot); err != nil {
			return err
		}
	}
	return nil
}

func (r *gormEventRepository) createOrderEvent(tx *gorm.DB, order *entity.Order, eventType, source, actorType string, before, after map[string]any) error {
	beforeData, afterData, diffData, err := eventJSON(before, after)
	if err != nil {
		return fmt.Errorf("failed to serialize order event: %w", err)
	}

	var orderID *int64
	if order.ID > 0 {
		id := order.ID
		orderID = &id
	}
	var externalID *string
	if order.ExternalOrderID != "" {
		value := order.ExternalOrderID
		externalID = &value
	}
	now := time.Now().UTC()
	event := &entity.OrderEvent{
		OrderID:         orderID,
		ExternalOrderID: externalID,
		EventType:       eventType,
		Source:          source,
		ActorType:       actorType,
		BeforeData:      beforeData,
		AfterData:       afterData,
		DiffData:        diffData,
		OccurredAt:      now,
		CreatedAt:       now,
	}
	return tx.Create(event).Error
}

func (r *gormEventRepository) recordCustomerChanges(tx *gorm.DB, before, after *entity.Customer, source, actorType string, orderID *int64) error {
	if after == nil {
		return nil
	}
	if source == "" {
		source = "system"
	}
	if actorType == "" {
		actorType = "system"
	}

	afterSnapshot := customerSnapshot(after)
	if before == nil {
		return r.createCustomerEvent(tx, after, orderID, "customer_created", source, actorType, nil, afterSnapshot)
	}

	beforeSnapshot := customerSnapshot(before)
	groups := map[string][]string{
		"customer_identity_changed": {"phone_number", "external_id"},
		"customer_details_changed": {
			"first_name", "last_name", "email", "address1", "address2", "city", "state", "country", "zip_code",
		},
		"customer_stats_changed": {"total_orders", "total_spent"},
	}
	recorded := false
	for eventType, fields := range groups {
		beforeGroup := selectFields(beforeSnapshot, fields)
		afterGroup := selectFields(afterSnapshot, fields)
		if reflect.DeepEqual(beforeGroup, afterGroup) {
			continue
		}
		if err := r.createCustomerEvent(tx, after, orderID, eventType, source, actorType, beforeGroup, afterGroup); err != nil {
			return err
		}
		recorded = true
	}
	if !reflect.DeepEqual(beforeSnapshot, afterSnapshot) && !recorded {
		return r.createCustomerEvent(tx, after, orderID, "customer_updated", source, actorType, beforeSnapshot, afterSnapshot)
	}
	return nil
}

func (r *gormEventRepository) createCustomerEvent(tx *gorm.DB, customer *entity.Customer, orderID *int64, eventType, source, actorType string, before, after map[string]any) error {
	beforeData, afterData, diffData, err := eventJSON(before, after)
	if err != nil {
		return fmt.Errorf("failed to serialize customer event: %w", err)
	}

	var customerID *int64
	if customer.ID > 0 {
		id := customer.ID
		customerID = &id
	}
	var phone *string
	if customer.PhoneNumber != "" {
		value := customer.PhoneNumber
		phone = &value
	}
	now := time.Now().UTC()
	event := &entity.CustomerEvent{
		CustomerID:    customerID,
		OrderID:       orderID,
		CustomerPhone: phone,
		EventType:     eventType,
		Source:        source,
		ActorType:     actorType,
		BeforeData:    beforeData,
		AfterData:     afterData,
		DiffData:      diffData,
		OccurredAt:    now,
		CreatedAt:     now,
	}
	return tx.Create(event).Error
}

func eventJSON(before, after map[string]any) (*json.RawMessage, *json.RawMessage, *json.RawMessage, error) {
	var beforeData, afterData, diffData *json.RawMessage
	if before != nil {
		value, err := json.Marshal(before)
		if err != nil {
			return nil, nil, nil, err
		}
		raw := json.RawMessage(value)
		beforeData = &raw
	}
	if after != nil {
		value, err := json.Marshal(after)
		if err != nil {
			return nil, nil, nil, err
		}
		raw := json.RawMessage(value)
		afterData = &raw
	}
	if before != nil && after != nil {
		diff := make(map[string]any)
		keys := make(map[string]struct{}, len(before)+len(after))
		for key := range before {
			keys[key] = struct{}{}
		}
		for key := range after {
			keys[key] = struct{}{}
		}
		for key := range keys {
			if !reflect.DeepEqual(before[key], after[key]) {
				diff[key] = map[string]any{"before": before[key], "after": after[key]}
			}
		}
		value, err := json.Marshal(diff)
		if err != nil {
			return nil, nil, nil, err
		}
		raw := json.RawMessage(value)
		diffData = &raw
	}
	return beforeData, afterData, diffData, nil
}

func selectFields(snapshot map[string]any, fields []string) map[string]any {
	selected := make(map[string]any, len(fields))
	for _, field := range fields {
		selected[field] = snapshot[field]
	}
	return selected
}

func orderSnapshot(order *entity.Order) map[string]any {
	return map[string]any{
		"id":                   order.ID,
		"source_id":            order.SourceID,
		"external_order_id":    order.ExternalOrderID,
		"order_number":         order.OrderNumber,
		"total_price":          order.TotalPrice,
		"subtotal_price":       optionalFloat(order.SubtotalPrice),
		"total_tax":            optionalFloat(order.TotalTax),
		"total_discount":       order.TotalDiscount,
		"financial_status":     optionalString(order.FinancialStatus),
		"fulfillment_status":   optionalString(order.FulfillmentStatus),
		"delivery_status":      optionalString(order.DeliveryStatus),
		"tracking_number":      optionalString(order.TrackingNumber),
		"shipping_company":     optionalString(order.ShippingCompany),
		"tracking_url":         optionalString(order.TrackingUrl),
		"status":               optionalString(order.Status),
		"cancelled_at":         optionalTime(order.CancelledAt),
		"cancel_reason":        optionalString(order.CancelReason),
		"customer_name":        optionalString(order.CustomerName),
		"customer_first_name":  optionalString(order.CustomerFirstName),
		"customer_last_name":   optionalString(order.CustomerLastName),
		"customer_email":       optionalString(order.CustomerEmail),
		"customer_phone":       optionalString(order.CustomerPhone),
		"customer_city":        optionalString(order.CustomerCity),
		"customer_state":       optionalString(order.CustomerState),
		"customer_country":     optionalString(order.CustomerCountry),
		"customer_address1":    optionalString(order.CustomerAddress1),
		"customer_address2":    optionalString(order.CustomerAddress2),
		"customer_zip":         optionalString(order.CustomerZip),
		"customer_external_id": optionalString(order.CustomerExternalID),
		"delivered_at":         optionalTime(order.DeliveredAt),
		"feedback_status_id":   optionalInt(order.FeedbackStatusID),
		"inventory_deducted":   order.InventoryDeducted,
	}
}

func customerSnapshot(customer *entity.Customer) map[string]any {
	return map[string]any{
		"id":           customer.ID,
		"phone_number": customer.PhoneNumber,
		"first_name":   optionalString(customer.FirstName),
		"last_name":    optionalString(customer.LastName),
		"email":        optionalString(customer.Email),
		"address1":     optionalString(customer.Address1),
		"address2":     optionalString(customer.Address2),
		"city":         optionalString(customer.City),
		"state":        optionalString(customer.State),
		"country":      optionalString(customer.Country),
		"zip_code":     optionalString(customer.ZipCode),
		"total_orders": customer.TotalOrders,
		"total_spent":  customer.TotalSpent,
		"source_id":    customer.SourceID,
		"external_id":  optionalString(customer.ExternalID),
	}
}

func optionalString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func optionalFloat(value *float64) float64 {
	if value == nil {
		return 0
	}
	return *value
}

func optionalInt(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

func optionalTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.UTC().Format(time.RFC3339Nano)
}

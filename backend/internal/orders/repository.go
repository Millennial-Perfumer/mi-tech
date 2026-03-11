package orders

import (
	"database/sql"
	"fmt"
	"shopify-gst-app/internal/models"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) UpsertOrder(order models.Order) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Upsert Order
	query := `
		INSERT INTO orders (
			id, source_id, external_order_id, order_number, total_price, subtotal_price, total_tax, 
			currency, financial_status, fulfillment_status, status, created_at, updated_at, 
			cancelled_at, cancel_reason, customer_name, customer_email, customer_phone, 
			customer_city, customer_state, customer_country, raw_payload,
			customer_address1, customer_address2, customer_zip,
			customer_first_name, customer_last_name, delivery_status,
			tracking_number, shipping_company
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30)
		ON CONFLICT (source_id, external_order_id) DO UPDATE SET
			financial_status = EXCLUDED.financial_status,
			fulfillment_status = EXCLUDED.fulfillment_status,
			delivery_status = EXCLUDED.delivery_status,
			tracking_number = EXCLUDED.tracking_number,
			shipping_company = EXCLUDED.shipping_company,
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at,
			cancelled_at = EXCLUDED.cancelled_at,
			cancel_reason = EXCLUDED.cancel_reason,
			total_price = EXCLUDED.total_price,
			subtotal_price = EXCLUDED.subtotal_price,
			total_tax = EXCLUDED.total_tax,
			customer_address1 = EXCLUDED.customer_address1,
			customer_address2 = EXCLUDED.customer_address2,
			customer_zip = EXCLUDED.customer_zip,
			customer_first_name = EXCLUDED.customer_first_name,
			customer_last_name = EXCLUDED.customer_last_name
		RETURNING id
	`
	err = tx.QueryRow(query,
		order.ID, order.SourceID, order.ExternalOrderID, order.OrderNumber, order.TotalPrice, order.SubtotalPrice, order.TotalTax,
		order.Currency, order.FinancialStatus, order.FulfillmentStatus, order.Status, order.CreatedAt, order.UpdatedAt,
		order.CancelledAt, order.CancelReason, order.CustomerName, order.CustomerEmail, order.CustomerPhone,
		order.CustomerCity, order.CustomerState, order.CustomerCountry, order.RawPayload,
		order.CustomerAddress1, order.CustomerAddress2, order.CustomerZip,
		order.CustomerFirstName, order.CustomerLastName, order.DeliveryStatus,
		order.TrackingNumber, order.ShippingCompany,
	).Scan(&order.ID)
	if err != nil {
		return fmt.Errorf("failed to upsert order: %w", err)
	}

	// 2. Clear old items and insert new ones (for simplicity and idempotency)
	_, err = tx.Exec("DELETE FROM order_line_items WHERE order_id = $1", order.ID)
	if err != nil {
		return fmt.Errorf("failed to clean old line items: %w", err)
	}

	for _, item := range order.LineItems {
		itemQuery := `
			INSERT INTO order_line_items (
				id, order_id, product_id, variant_id, title, sku, hs_code, quantity, price, discount
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`
		_, err = tx.Exec(itemQuery,
			item.ID, order.ID, item.ProductID, item.VariantID, item.Title, item.SKU, item.HSCode, item.Quantity, item.Price, item.Discount,
		)
		if err != nil {
			return fmt.Errorf("failed to insert line item %s: %w", item.ID, err)
		}
	}

	return tx.Commit()
}

func (r *Repository) UpdateStatus(externalOrderID string, financialStatus, fulfillmentStatus string) error {
	query := `
		UPDATE orders 
		SET financial_status = $1, fulfillment_status = $2, status = $2, updated_at = CURRENT_TIMESTAMP
		WHERE external_order_id = $3
	`
	_, err := r.db.Exec(query, financialStatus, fulfillmentStatus, externalOrderID)
	return err
}

func (r *Repository) CancelOrder(externalOrderID string, cancelledAt *string, reason string) error {
	query := `
		UPDATE orders 
		SET status = 'CANCELLED', cancelled_at = $1, cancel_reason = $2, updated_at = CURRENT_TIMESTAMP
		WHERE external_order_id = $3
	`
	_, err := r.db.Exec(query, cancelledAt, reason, externalOrderID)
	return err
}

func (r *Repository) GetOrder(id string) (models.Order, error) {
	var o models.Order
	err := r.db.QueryRow(`
		SELECT id, order_number, total_price, subtotal_price, total_tax, created_at, 
		       customer_name, customer_email, customer_phone, 
		       customer_city, customer_state, customer_country, status, raw_payload,
		       customer_address1, customer_address2, customer_zip,
		       customer_first_name, customer_last_name, financial_status, fulfillment_status, delivery_status,
		       tracking_number, shipping_company
		FROM orders WHERE id = $1 OR external_order_id = $1
	`, id).Scan(&o.ID, &o.OrderNumber, &o.TotalPrice, &o.SubtotalPrice, &o.TotalTax, &o.CreatedAt,
		&o.CustomerName, &o.CustomerEmail, &o.CustomerPhone,
		&o.CustomerCity, &o.CustomerState, &o.CustomerCountry, &o.Status, &o.RawPayload,
		&o.CustomerAddress1, &o.CustomerAddress2, &o.CustomerZip,
		&o.CustomerFirstName, &o.CustomerLastName, &o.FinancialStatus, &o.FulfillmentStatus, &o.DeliveryStatus,
		&o.TrackingNumber, &o.ShippingCompany)
	return o, err
}

// SaveWebhookEvent inserts a new webhook event log
func (r *Repository) SaveWebhookEvent(event *models.WebhookEvent) error {
	query := `
		INSERT INTO webhook_events (source_id, topic, external_id, webhook_delivery_id, payload)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`
	return r.db.QueryRow(query, event.SourceID, event.Topic, event.ExternalID, event.WebhookDeliveryID, event.Payload).Scan(&event.ID, &event.CreatedAt)
}

// IsWebhookProcessed checks if a webhook delivery ID already exists
func (r *Repository) IsWebhookProcessed(deliveryID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM webhook_events WHERE webhook_delivery_id = $1)`
	err := r.db.QueryRow(query, deliveryID).Scan(&exists)
	return exists, err
}

// LinkWebhookToOrder links a webhook event to an internal order and marks it processed
func (r *Repository) LinkWebhookToOrder(deliveryID string, orderID string) error {
	query := `UPDATE webhook_events SET order_id = $1, processed = true WHERE webhook_delivery_id = $2`
	_, err := r.db.Exec(query, orderID, deliveryID)
	return err
}

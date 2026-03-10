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
		INSERT INTO shopify_orders (
			id, shopify_order_id, order_number, total_price, subtotal_price, total_tax, 
			currency, financial_status, fulfillment_status, status, created_at, updated_at, 
			cancelled_at, cancel_reason, customer_name, customer_email, customer_phone, 
			customer_city, customer_state, customer_country, raw_payload,
			customer_address1, customer_address2, customer_zip,
			customer_first_name, customer_last_name
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26)
		ON CONFLICT (id) DO UPDATE SET
			financial_status = EXCLUDED.financial_status,
			fulfillment_status = EXCLUDED.fulfillment_status,
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
	`
	_, err = tx.Exec(query,
		order.ID, order.ShopifyOrderID, order.OrderNumber, order.TotalPrice, order.SubtotalPrice, order.TotalTax,
		order.Currency, order.FinancialStatus, order.FulfillmentStatus, order.Status, order.CreatedAt, order.UpdatedAt,
		order.CancelledAt, order.CancelReason, order.CustomerName, order.CustomerEmail, order.CustomerPhone,
		order.CustomerCity, order.CustomerState, order.CustomerCountry, order.RawPayload,
		order.CustomerAddress1, order.CustomerAddress2, order.CustomerZip,
		order.CustomerFirstName, order.CustomerLastName,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert order: %w", err)
	}

	// 2. Clear old items and insert new ones (for simplicity and idempotency)
	_, err = tx.Exec("DELETE FROM shopify_order_line_items WHERE order_id = $1", order.ID)
	if err != nil {
		return fmt.Errorf("failed to clean old line items: %w", err)
	}

	for _, item := range order.LineItems {
		itemQuery := `
			INSERT INTO shopify_order_line_items (
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

func (r *Repository) UpdateStatus(shopifyOrderID string, financialStatus, fulfillmentStatus string) error {
	query := `
		UPDATE shopify_orders 
		SET financial_status = $1, fulfillment_status = $2, status = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3 OR shopify_order_id = $3
	`
	_, err := r.db.Exec(query, financialStatus, fulfillmentStatus, shopifyOrderID)
	return err
}

func (r *Repository) CancelOrder(shopifyOrderID string, cancelledAt *string, reason string) error {
	query := `
		UPDATE shopify_orders 
		SET status = 'CANCELLED', cancelled_at = $1, cancel_reason = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3 OR shopify_order_id = $3
	`
	_, err := r.db.Exec(query, cancelledAt, reason, shopifyOrderID)
	return err
}

func (r *Repository) GetOrder(id string) (models.Order, error) {
	var o models.Order
	err := r.db.QueryRow(`
		SELECT id, order_number, total_price, subtotal_price, total_tax, created_at, 
		       customer_name, customer_email, customer_phone, 
		       customer_city, customer_state, customer_country, status, raw_payload,
		       customer_address1, customer_address2, customer_zip,
		       customer_first_name, customer_last_name
		FROM shopify_orders WHERE id = $1 OR shopify_order_id = $1
	`, id).Scan(&o.ID, &o.OrderNumber, &o.TotalPrice, &o.SubtotalPrice, &o.TotalTax, &o.CreatedAt,
		&o.CustomerName, &o.CustomerEmail, &o.CustomerPhone,
		&o.CustomerCity, &o.CustomerState, &o.CustomerCountry, &o.Status, &o.RawPayload,
		&o.CustomerAddress1, &o.CustomerAddress2, &o.CustomerZip,
		&o.CustomerFirstName, &o.CustomerLastName)
	return o, err
}

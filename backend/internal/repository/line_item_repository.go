package repository

import (
	"database/sql"
	"fmt"

	"shopify-gst-app/internal/entity"
)

// pgLineItemRepository is the PostgreSQL implementation of LineItemRepository.
type pgLineItemRepository struct {
	db *sql.DB
}

// NewLineItemRepository creates a new PostgreSQL-backed LineItemRepository.
func NewLineItemRepository(db *sql.DB) LineItemRepository {
	return &pgLineItemRepository{db: db}
}

func (r *pgLineItemRepository) GetByOrderID(orderID string) ([]entity.LineItem, error) {
	rows, err := r.db.Query(`
		SELECT id, order_id, COALESCE(title, ''), COALESCE(sku, ''), COALESCE(hs_code, ''), 
		       quantity, price, COALESCE(discount, 0)
		FROM order_line_items WHERE order_id = $1
	`, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to query line items: %w", err)
	}
	defer rows.Close()

	var items []entity.LineItem
	for rows.Next() {
		var li entity.LineItem
		var title, sku, hsCode string
		if err := rows.Scan(&li.ID, &li.OrderID, &title, &sku, &hsCode, &li.Quantity, &li.Price, &li.Discount); err != nil {
			return nil, fmt.Errorf("failed to scan line item: %w", err)
		}
		li.Title = toNullStr(title)
		li.SKU = toNullStr(sku)
		li.HSCode = toNullStr(hsCode)
		items = append(items, li)
	}

	return items, nil
}

func (r *pgLineItemRepository) UpsertBatch(txIface interface{}, orderID string, items []entity.LineItem) error {
	tx, ok := txIface.(*sql.Tx)
	if !ok {
		return fmt.Errorf("invalid transaction type")
	}

	for _, item := range items {
		_, err := tx.Exec(`
			INSERT INTO order_line_items (id, order_id, product_id, variant_id, title, sku, hs_code, quantity, price, discount)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			ON CONFLICT (id) DO UPDATE SET
				title = EXCLUDED.title,
				sku = EXCLUDED.sku,
				hs_code = EXCLUDED.hs_code,
				quantity = EXCLUDED.quantity,
				price = EXCLUDED.price,
				discount = EXCLUDED.discount
		`, item.ID, orderID, item.ProductID, item.VariantID, item.Title, item.SKU, item.HSCode, item.Quantity, item.Price, item.Discount)
		if err != nil {
			return fmt.Errorf("failed to upsert line item %s: %w", item.ID, err)
		}
	}

	return nil
}

func (r *pgLineItemRepository) DeleteByOrderID(txIface interface{}, orderID string) error {
	tx, ok := txIface.(*sql.Tx)
	if !ok {
		return fmt.Errorf("invalid transaction type")
	}
	_, err := tx.Exec("DELETE FROM order_line_items WHERE order_id = $1", orderID)
	return err
}

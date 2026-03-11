package repository

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"shopify-gst-app/internal/entity"
)

// pgOrderRepository is the PostgreSQL implementation of OrderRepository.
type pgOrderRepository struct {
	db *sql.DB
}

// NewOrderRepository creates a new PostgreSQL-backed OrderRepository.
func NewOrderRepository(db *sql.DB) OrderRepository {
	return &pgOrderRepository{db: db}
}

func (r *pgOrderRepository) List(filter OrderFilter) ([]entity.Order, int, error) {
	// 1. Build where clause
	whereClause := " WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if filter.StartDate != "" {
		whereClause += fmt.Sprintf(" AND created_at >= $%d", argIdx)
		args = append(args, filter.StartDate)
		argIdx++
	}
	if filter.EndDate != "" {
		whereClause += fmt.Sprintf(" AND created_at <= $%d", argIdx)
		args = append(args, filter.EndDate)
		argIdx++
	}
	if filter.Search != "" {
		whereClause += fmt.Sprintf(" AND (order_number ILIKE $%d OR customer_name ILIKE $%d)", argIdx, argIdx)
		args = append(args, "%"+filter.Search+"%")
		argIdx++
	}
	if filter.Source != "" {
		whereClause += fmt.Sprintf(" AND source_id = $%d", argIdx)
		args = append(args, filter.Source)
		argIdx++
	}
	if filter.FinancialStatus != "" {
		if strings.ToLower(filter.FinancialStatus) == "unpaid" {
			whereClause += " AND financial_status != 'paid'"
		} else {
			whereClause += fmt.Sprintf(" AND financial_status = $%d", argIdx)
			args = append(args, filter.FinancialStatus)
			argIdx++
		}
	}
	if filter.FulfillmentStatus != "" {
		whereClause += fmt.Sprintf(" AND fulfillment_status = $%d", argIdx)
		args = append(args, filter.FulfillmentStatus)
		argIdx++
	}

	// 2. Count total matching orders
	countQuery := "SELECT COUNT(*) FROM orders" + whereClause
	var totalCount int
	if err := r.db.QueryRow(countQuery, args...).Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("failed to count orders: %w", err)
	}

	// 3. Fetch paginated orders
	page := filter.Page
	limit := filter.Limit
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 25
	}
	offset := (page - 1) * limit

	// Determine sorting
	sortBy := "created_at"
	sortOrder := "DESC"
	
	allowedSortCols := map[string]string{
		"order_number":       "order_number",
		"customer_name":     "customer_name",
		"created_at":        "created_at",
		"total_price":       "total_price",
		"financial_status":   "financial_status",
		"fulfillment_status": "fulfillment_status",
		"source_id":          "source_id",
	}
	
	if col, ok := allowedSortCols[filter.SortBy]; ok {
		sortBy = col
	}
	if strings.ToUpper(filter.SortOrder) == "ASC" {
		sortOrder = "ASC"
	}

	selectQuery := fmt.Sprintf(`
		SELECT id, order_number, total_price, created_at, 
		       COALESCE(customer_name, ''), COALESCE(customer_city, ''), COALESCE(customer_state, ''), COALESCE(customer_country, ''), 
		       COALESCE(status, ''),
		       COALESCE(financial_status, ''), COALESCE(fulfillment_status, ''), COALESCE(delivery_status, ''), 
		       COALESCE(tracking_number, ''), COALESCE(shipping_company, ''), COALESCE(tracking_url, ''),
		       source_id
		FROM orders 
		%s
		ORDER BY %s %s 
		LIMIT %d OFFSET %d
	`, whereClause, sortBy, sortOrder, limit, offset)

	rows, err := r.db.Query(selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query orders: %w", err)
	}
	defer rows.Close()

	var orders []entity.Order
	for rows.Next() {
		var o entity.Order
		var custName, custCity, custState, custCountry, status string
		var finStatus, fulStatus, delStatus, trackNum, shipCo, trackUrl string

		if err := rows.Scan(
			&o.ID, &o.OrderNumber, &o.TotalPrice, &o.CreatedAt,
			&custName, &custCity, &custState, &custCountry, &status,
			&finStatus, &fulStatus, &delStatus,
			&trackNum, &shipCo, &trackUrl,
			&o.SourceID,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan order row: %w", err)
		}

		o.CustomerName = toNullStr(custName)
		o.CustomerCity = toNullStr(custCity)
		o.CustomerState = toNullStr(custState)
		o.CustomerCountry = toNullStr(custCountry)
		o.Status = toNullStr(status)
		o.FinancialStatus = toNullStr(finStatus)
		o.FulfillmentStatus = toNullStr(fulStatus)
		o.DeliveryStatus = toNullStr(delStatus)
		o.TrackingNumber = toNullStr(trackNum)
		o.ShippingCompany = toNullStr(shipCo)
		o.TrackingUrl = toNullStr(trackUrl)

		orders = append(orders, o)
	}

	return orders, totalCount, nil
}

func (r *pgOrderRepository) GetByID(id string) (entity.Order, error) {
	return r.getOrder("id", id)
}

func (r *pgOrderRepository) GetByExternalID(externalID string) (entity.Order, error) {
	return r.getOrder("external_order_id", externalID)
}

func (r *pgOrderRepository) getOrder(column, value string) (entity.Order, error) {
	var o entity.Order
	query := fmt.Sprintf(`
		SELECT id, order_number, total_price, COALESCE(subtotal_price, 0), COALESCE(total_tax, 0), 
		       created_at, source_id,
		       COALESCE(customer_name, ''), COALESCE(customer_email, ''), COALESCE(customer_phone, ''), 
		       COALESCE(customer_city, ''), COALESCE(customer_state, ''), COALESCE(customer_country, ''),
		       COALESCE(status, ''), raw_payload,
		       COALESCE(customer_address1, ''), COALESCE(customer_address2, ''), COALESCE(customer_zip, ''),
		       COALESCE(customer_first_name, ''), COALESCE(customer_last_name, ''),
		       COALESCE(financial_status, ''), COALESCE(fulfillment_status, ''), COALESCE(delivery_status, ''),
		       COALESCE(tracking_number, ''), COALESCE(shipping_company, ''), COALESCE(tracking_url, '')
		FROM orders WHERE %s = $1 OR external_order_id = $1
	`, column)

	var custName, custEmail, custPhone, custCity, custState, custCountry string
	var status, addr1, addr2, zip, firstName, lastName string
	var finStatus, fulStatus, delStatus, trackNum, shipCo, trackUrl string

	err := r.db.QueryRow(query, value).Scan(
		&o.ID, &o.OrderNumber, &o.TotalPrice, &o.SubtotalPrice, &o.TotalTax,
		&o.CreatedAt, &o.SourceID,
		&custName, &custEmail, &custPhone,
		&custCity, &custState, &custCountry,
		&status, &o.RawPayload,
		&addr1, &addr2, &zip,
		&firstName, &lastName,
		&finStatus, &fulStatus, &delStatus,
		&trackNum, &shipCo, &trackUrl,
	)
	if err != nil {
		return o, err
	}

	o.CustomerName = toNullStr(custName)
	o.CustomerEmail = toNullStr(custEmail)
	o.CustomerPhone = toNullStr(custPhone)
	o.CustomerCity = toNullStr(custCity)
	o.CustomerState = toNullStr(custState)
	o.CustomerCountry = toNullStr(custCountry)
	o.Status = toNullStr(status)
	o.CustomerAddress1 = toNullStr(addr1)
	o.CustomerAddress2 = toNullStr(addr2)
	o.CustomerZip = toNullStr(zip)
	o.CustomerFirstName = toNullStr(firstName)
	o.CustomerLastName = toNullStr(lastName)
	o.FinancialStatus = toNullStr(finStatus)
	o.FulfillmentStatus = toNullStr(fulStatus)
	o.DeliveryStatus = toNullStr(delStatus)
	o.TrackingNumber = toNullStr(trackNum)
	o.ShippingCompany = toNullStr(shipCo)
	o.TrackingUrl = toNullStr(trackUrl)

	return o, nil
}

func (r *pgOrderRepository) Upsert(order entity.Order) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	if err := r.upsertInTx(tx, order); err != nil {
		return err
	}

	// Clear old line items and insert new ones
	if _, err := tx.Exec("DELETE FROM order_line_items WHERE order_id = $1", order.ID); err != nil {
		return fmt.Errorf("failed to clean old line items: %w", err)
	}

	for _, item := range order.LineItems {
		_, err := tx.Exec(`
			INSERT INTO order_line_items (id, order_id, product_id, variant_id, title, sku, hs_code, quantity, price, discount)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`, item.ID, order.ID, item.ProductID, item.VariantID, item.Title, item.SKU, item.HSCode, item.Quantity, item.Price, item.Discount)
		if err != nil {
			return fmt.Errorf("failed to insert line item %s: %w", item.ID, err)
		}
	}

	return tx.Commit()
}

func (r *pgOrderRepository) UpsertBatch(orders []entity.Order) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	orderStmt, err := tx.Prepare(upsertOrderSQL)
	if err != nil {
		return fmt.Errorf("failed to prepare order stmt: %w", err)
	}
	defer orderStmt.Close()

	itemStmt, err := tx.Prepare(upsertLineItemSQL)
	if err != nil {
		return fmt.Errorf("failed to prepare item stmt: %w", err)
	}
	defer itemStmt.Close()

	for _, o := range orders {
		_, err := orderStmt.Exec(
			o.ID, o.SourceID, o.ExternalOrderID, o.OrderNumber,
			o.TotalPrice, o.SubtotalPrice, o.TotalTax,
			o.CreatedAt, o.UpdatedAt,
			o.CustomerName, o.CustomerEmail, o.CustomerPhone,
			o.CustomerCity, o.CustomerState, o.CustomerCountry,
			o.Status, o.FinancialStatus, o.FulfillmentStatus, o.DeliveryStatus,
			o.TrackingNumber, o.ShippingCompany, o.TrackingUrl,
		)
		if err != nil {
			log.Printf("Failed to upsert order %s: %v", o.ID, err)
			continue
		}

		for _, li := range o.LineItems {
			_, err := itemStmt.Exec(
				li.ID, o.ID, li.Title, li.SKU, li.HSCode,
				li.Quantity, li.Price, li.Discount,
			)
			if err != nil {
				log.Printf("Failed to upsert line item %s for order %s: %v", li.ID, o.ID, err)
			}
		}
	}

	return tx.Commit()
}

func (r *pgOrderRepository) UpdateStatus(externalOrderID string, financialStatus, fulfillmentStatus string) error {
	query := `
		UPDATE orders 
		SET financial_status = $1, fulfillment_status = $2, status = $2, updated_at = CURRENT_TIMESTAMP
		WHERE external_order_id = $3
	`
	_, err := r.db.Exec(query, financialStatus, fulfillmentStatus, externalOrderID)
	return err
}

func (r *pgOrderRepository) UpdateOrderStatus(id string, status string) (int64, error) {
	status = strings.ToUpper(status)
	var query string
	var args []interface{}

	if status == "CANCELLED" {
		query = `UPDATE orders SET status = $1, fulfillment_status = 'restocked', cancelled_at = COALESCE(cancelled_at, CURRENT_TIMESTAMP), updated_at = CURRENT_TIMESTAMP WHERE id = $2`
		args = []interface{}{status, id}
	} else if status == "FULFILLED" || status == "UNFULFILLED" {
		query = `UPDATE orders SET status = $1, fulfillment_status = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3`
		args = []interface{}{status, strings.ToLower(status), id}
	} else {
		query = `UPDATE orders SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
		args = []interface{}{status, id}
	}

	res, err := r.db.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *pgOrderRepository) CancelOrder(externalOrderID string, cancelledAt *string, reason string) error {
	query := `
		UPDATE orders 
		SET status = 'CANCELLED', cancelled_at = $1, cancel_reason = $2, updated_at = CURRENT_TIMESTAMP
		WHERE external_order_id = $3
	`
	_, err := r.db.Exec(query, cancelledAt, reason, externalOrderID)
	return err
}

func (r *pgOrderRepository) TruncateAll() error {
	if _, err := r.db.Exec("TRUNCATE TABLE order_line_items CASCADE"); err != nil {
		return fmt.Errorf("failed to truncate order_line_items: %w", err)
	}
	if _, err := r.db.Exec("TRUNCATE TABLE orders CASCADE"); err != nil {
		return fmt.Errorf("failed to truncate orders: %w", err)
	}
	if _, err := r.db.Exec("TRUNCATE TABLE webhook_events CASCADE"); err != nil {
		return fmt.Errorf("failed to truncate webhook_events: %w", err)
	}
	log.Println("Successfully truncated orders, order_line_items, and webhook_events tables")
	return nil
}

// upsertInTx performs the order upsert within an existing transaction (used by Upsert).
func (r *pgOrderRepository) upsertInTx(tx *sql.Tx, order entity.Order) error {
	err := tx.QueryRow(upsertOrderWithReturningSQL,
		order.ID, order.SourceID, order.ExternalOrderID, order.OrderNumber,
		order.TotalPrice, order.SubtotalPrice, order.TotalTax,
		order.Currency, order.FinancialStatus, order.FulfillmentStatus,
		order.Status, order.CreatedAt, order.UpdatedAt,
		order.CancelledAt, order.CancelReason,
		order.CustomerName, order.CustomerEmail, order.CustomerPhone,
		order.CustomerCity, order.CustomerState, order.CustomerCountry,
		order.RawPayload,
		order.CustomerAddress1, order.CustomerAddress2, order.CustomerZip,
		order.CustomerFirstName, order.CustomerLastName,
		order.DeliveryStatus, order.TrackingNumber, order.ShippingCompany,
	).Scan(&order.ID)
	if err != nil {
		return fmt.Errorf("failed to upsert order: %w", err)
	}
	return nil
}

// --- SQL constants ---

const upsertOrderSQL = `
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

const upsertOrderWithReturningSQL = `
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

const upsertLineItemSQL = `
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

// toNullStr converts a possibly empty string to sql.NullString
func toNullStr(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

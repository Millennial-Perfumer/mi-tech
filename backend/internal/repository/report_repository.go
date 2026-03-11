package repository

import (
	"database/sql"
	"fmt"
	"time"
)

// pgMetricsRepository is the PostgreSQL implementation of MetricsRepository.
type pgMetricsRepository struct {
	db *sql.DB
}

// NewMetricsRepository creates a new PostgreSQL-backed MetricsRepository.
func NewMetricsRepository(db *sql.DB) MetricsRepository {
	return &pgMetricsRepository{db: db}
}

func (r *pgMetricsRepository) GetDashboardMetrics(startDate, endDate string) (totalRevenue, cgst, sgst, igst float64, totalOrders, cancelledOrders, fulfilledOrders, unfulfilledOrders int, err error) {
	start, end := parseDateRange(startDate, endDate)

	query := `
		SELECT 
			COALESCE(SUM(total_price), 0) as total_revenue,
			COALESCE(SUM(CASE WHEN LOWER(customer_state) = 'tamil nadu' THEN (total_price - ROUND(total_price / 1.18, 2)) / 2 ELSE 0 END), 0) as cgst,
			COALESCE(SUM(CASE WHEN LOWER(customer_state) = 'tamil nadu' THEN (total_price - ROUND(total_price / 1.18, 2)) / 2 ELSE 0 END), 0) as sgst,
			COALESCE(SUM(CASE WHEN LOWER(customer_state) != 'tamil nadu' THEN (total_price - ROUND(total_price / 1.18, 2)) ELSE 0 END), 0) as igst,
			COUNT(id) as total_orders,
			COUNT(id) FILTER (WHERE LOWER(status) = 'cancelled') as cancelled_orders,
			COUNT(id) FILTER (WHERE LOWER(status) = 'fulfilled') as fulfilled_orders,
			COUNT(id) FILTER (WHERE LOWER(status) = 'unfulfilled') as unfulfilled_orders
		FROM orders 
		WHERE created_at >= $1 AND created_at <= $2
	`
	err = r.db.QueryRow(query, start, end).Scan(
		&totalRevenue, &cgst, &sgst, &igst, &totalOrders, &cancelledOrders, &fulfilledOrders, &unfulfilledOrders,
	)
	return
}

// pgReportRepository is the PostgreSQL implementation of ReportRepository.
type pgReportRepository struct {
	db *sql.DB
}

// NewReportRepository creates a new PostgreSQL-backed ReportRepository.
func NewReportRepository(db *sql.DB) ReportRepository {
	return &pgReportRepository{db: db}
}

func (r *pgReportRepository) GetGSTSummary(startDate, endDate string) (totalOrders, cancelledOrders, fulfilledOrders, unfulfilledOrders, paidOrders int, totalRevenue, totalTaxable, totalTax float64, err error) {
	start, end := parseDateRange(startDate, endDate)

	query := `
		SELECT 
			COUNT(id),
			COUNT(id) FILTER (WHERE LOWER(status) = 'cancelled'),
			COUNT(id) FILTER (WHERE LOWER(fulfillment_status) = 'fulfilled'),
			COUNT(id) FILTER (WHERE LOWER(fulfillment_status) != 'fulfilled' AND LOWER(status) != 'cancelled'),
			COUNT(id) FILTER (WHERE LOWER(financial_status) = 'paid'),
			COALESCE(SUM(total_price) FILTER (WHERE LOWER(status) != 'cancelled'), 0) as revenue,
			COALESCE(SUM(ROUND(total_price / 1.18, 2)) FILTER (WHERE LOWER(status) != 'cancelled'), 0) as taxable,
			COALESCE(SUM(total_price - ROUND(total_price / 1.18, 2)) FILTER (WHERE LOWER(status) != 'cancelled'), 0) as tax
		FROM orders 
		WHERE created_at >= $1 AND created_at <= $2
	`
	err = r.db.QueryRow(query, start, end).Scan(
		&totalOrders, &cancelledOrders, &fulfilledOrders, &unfulfilledOrders, &paidOrders,
		&totalRevenue, &totalTaxable, &totalTax,
	)
	return
}

func (r *pgReportRepository) GetStateSummary(startDate, endDate string) ([]StateSummaryResult, error) {
	start, end := parseDateRange(startDate, endDate)

	query := `
		SELECT 
			COALESCE(customer_state, 'N/A'),
			COUNT(id) as orders,
			COALESCE(SUM(ROUND(total_price / 1.18, 2)), 0) as taxable,
			COALESCE(SUM(total_price - ROUND(total_price / 1.18, 2)), 0) as gst,
			COALESCE(SUM(total_price), 0) as revenue
		FROM orders
		WHERE created_at >= $1 AND created_at <= $2 AND LOWER(status) != 'cancelled'
		GROUP BY customer_state
		ORDER BY revenue DESC
	`
	rows, err := r.db.Query(query, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query state summary: %w", err)
	}
	defer rows.Close()

	var results []StateSummaryResult
	for rows.Next() {
		var s StateSummaryResult
		if err := rows.Scan(&s.State, &s.Orders, &s.TaxableValue, &s.TotalGST, &s.Revenue); err != nil {
			return nil, err
		}
		results = append(results, s)
	}
	return results, nil
}

func (r *pgReportRepository) GetHSNSummary(startDate, endDate string) ([]HSNSummaryResult, error) {
	start, end := parseDateRange(startDate, endDate)

	query := `
		WITH OrderSubtotals AS (
			SELECT order_id, SUM(price * quantity - discount) as line_sum
			FROM order_line_items
			GROUP BY order_id
		),
		LineItemShares AS (
			SELECT 
				li.order_id,
				COALESCE(li.hs_code, '33029019') as hs_code,
				li.quantity,
				(li.price * li.quantity - li.discount) as line_val,
				os.line_sum,
				o.total_price,
				COALESCE(o.customer_state, 'N/A') as state
			FROM order_line_items li
			JOIN orders o ON li.order_id = o.id
			JOIN OrderSubtotals os ON li.order_id = os.order_id
			WHERE o.created_at >= $1 AND o.created_at <= $2 AND LOWER(o.status) != 'cancelled' AND os.line_sum > 0
		)
		SELECT 
			hs_code,
			COUNT(DISTINCT order_id) as prod_count, -- Approximate count
			SUM(quantity) as qty,
			SUM(ROUND((line_val / line_sum) * (total_price / 1.18), 2)) as taxable,
			SUM(ROUND((line_val / line_sum) * total_price, 2) - ROUND((line_val / line_sum) * (total_price / 1.18), 2)) as gst,
			SUM(ROUND((line_val / line_sum) * total_price, 2)) as revenue,
			state
		FROM LineItemShares
		GROUP BY hs_code, state
		ORDER BY revenue DESC
	`
	rows, err := r.db.Query(query, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query HSN summary: %w", err)
	}
	defer rows.Close()

	var results []HSNSummaryResult
	for rows.Next() {
		var h HSNSummaryResult
		if err := rows.Scan(&h.HSNCode, &h.ProductCount, &h.QtySold, &h.TaxableValue, &h.TotalGST, &h.Revenue, &h.State); err != nil {
			return nil, err
		}
		results = append(results, h)
	}
	return results, nil
}

func (r *pgReportRepository) GetDocumentsIssued(startDate, endDate string) (minOrder, maxOrder *int64, total, cancelled int, err error) {
	start, end := parseDateRange(startDate, endDate)

	query := `
		SELECT 
			MIN(NULLIF(regexp_replace(order_number, '[^0-9]', '', 'g'), '')::bigint) as min_val,
			MAX(NULLIF(regexp_replace(order_number, '[^0-9]', '', 'g'), '')::bigint) as max_val,
			COUNT(id) as total,
			COUNT(id) FILTER (WHERE LOWER(status) = 'cancelled') as cancelled
		FROM orders
		WHERE created_at >= $1 AND created_at <= $2
	`
	var minV, maxV sql.NullInt64
	err = r.db.QueryRow(query, start, end).Scan(&minV, &maxV, &total, &cancelled)
	if err != nil {
		return
	}
	if minV.Valid {
		minOrder = &minV.Int64
	}
	if maxV.Valid {
		maxOrder = &maxV.Int64
	}
	return
}

func (r *pgReportRepository) GetTaxByState(startDate, endDate string) ([]StateTaxResult, error) {
	start, end := parseDateRange(startDate, endDate)

	query := `
		SELECT 
			COALESCE(customer_state, 'N/A'),
			SUM(total_price - (total_price / 1.18)) as sum_tax
		FROM orders
		WHERE created_at >= $1 AND created_at <= $2 AND LOWER(status) != 'cancelled'
		GROUP BY customer_state
	`
	rows, err := r.db.Query(query, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query tax by state: %w", err)
	}
	defer rows.Close()

	var results []StateTaxResult
	for rows.Next() {
		var s StateTaxResult
		if err := rows.Scan(&s.State, &s.Tax); err != nil {
			return nil, err
		}
		results = append(results, s)
	}
	return results, nil
}

// --- Shared helper ---

func parseDateRange(startDate, endDate string) (time.Time, time.Time) {
	start := parseISO(startDate)
	end := parseISO(endDate)

	if start.IsZero() {
		now := time.Now()
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	}
	if end.IsZero() {
		end = time.Now()
	}
	return start, end
}

func parseISO(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	// Try RFC3339 first
	t, err := time.Parse(time.RFC3339, s)
	if err == nil {
		return t
	}
	// Try with milliseconds (common in ISO strings from JS)
	t, err = time.Parse("2006-01-02T15:04:05.000Z", s)
	if err == nil {
		return t
	}
	return time.Time{}
}

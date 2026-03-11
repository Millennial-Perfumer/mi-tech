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

func (r *pgMetricsRepository) GetDashboardMetrics(startDate, endDate string) (totalRevenue float64, totalOrders, cancelledOrders, fulfilledOrders, unfulfilledOrders int, err error) {
	start, end := parseDateRange(startDate, endDate)

	query := `
		SELECT 
			COALESCE(SUM(total_price), 0) as total_revenue,
			COUNT(id) as total_orders,
			COUNT(id) FILTER (WHERE LOWER(status) = 'cancelled') as cancelled_orders,
			COUNT(id) FILTER (WHERE LOWER(status) = 'fulfilled') as fulfilled_orders,
			COUNT(id) FILTER (WHERE LOWER(status) = 'unfulfilled') as unfulfilled_orders
		FROM orders 
		WHERE created_at >= $1 AND created_at <= $2
	`
	err = r.db.QueryRow(query, start, end).Scan(
		&totalRevenue, &totalOrders, &cancelledOrders, &fulfilledOrders, &unfulfilledOrders,
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
			COUNT(id) as total_orders,
			COUNT(id) FILTER (WHERE LOWER(status) = 'cancelled') as cancelled_orders,
			COUNT(id) FILTER (WHERE LOWER(status) = 'fulfilled') as fulfilled_orders,
			COUNT(id) FILTER (WHERE LOWER(status) = 'unfulfilled') as unfulfilled_orders,
			COUNT(id) FILTER (WHERE LOWER(status) = 'paid') as paid_orders,
			COALESCE(SUM(total_price) FILTER (WHERE LOWER(status) != 'cancelled'), 0) as total_revenue,
			COALESCE(SUM(subtotal_price) FILTER (WHERE LOWER(status) != 'cancelled'), 0) as total_taxable,
			COALESCE(SUM(total_tax) FILTER (WHERE LOWER(status) != 'cancelled'), 0) as total_tax
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
			customer_state,
			COUNT(id) as orders,
			COALESCE(SUM(subtotal_price), 0) as taxable,
			COALESCE(SUM(total_tax), 0) as gst,
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
		SELECT 
			li.hs_code,
			COUNT(DISTINCT li.id) as product_count,
			SUM(li.quantity) as qty,
			SUM((li.price * li.quantity - li.discount) / 1.18) as taxable,
			SUM((li.price * li.quantity - li.discount) - (li.price * li.quantity - li.discount) / 1.18) as gst,
			SUM(li.price * li.quantity - li.discount) as revenue,
			o.customer_state
		FROM order_line_items li
		JOIN orders o ON li.order_id = o.id
		WHERE o.created_at >= $1 AND o.created_at <= $2 AND LOWER(o.status) != 'cancelled'
		GROUP BY li.hs_code, o.customer_state
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
			customer_state,
			SUM(total_tax) as sum_tax
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
	start, _ := time.Parse(time.RFC3339, startDate)
	end, _ := time.Parse(time.RFC3339, endDate)
	if start.IsZero() {
		now := time.Now()
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	}
	if end.IsZero() {
		end = time.Now()
	}
	return start, end
}

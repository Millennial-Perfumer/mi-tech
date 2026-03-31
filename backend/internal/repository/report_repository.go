package repository

import (
	"database/sql"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// gormMetricsRepository is the GORM implementation of MetricsRepository.
type gormMetricsRepository struct {
	db *gorm.DB
}

// NewMetricsRepository creates a new GORM-backed MetricsRepository.
func NewMetricsRepository(db *gorm.DB) MetricsRepository {
	return &gormMetricsRepository{db: db}
}

func (r *gormMetricsRepository) GetDashboardMetrics(startDate, endDate string) (totalRevenue, cgst, sgst, igst float64, totalOrders, cancelledOrders, fulfilledOrders, unfulfilledOrders int, err error) {
	start, end := parseDateRange(startDate, endDate)

	query := `
		SELECT 
			COALESCE(SUM(total_price) FILTER (WHERE status IS NULL OR LOWER(status) != 'cancelled'), 0) as total_revenue,
			COALESCE(SUM(CASE WHEN LOWER(customer_state) IN ('tamil nadu', 'tn') THEN (total_price - ROUND(total_price / 1.18, 2)) / 2 ELSE 0 END) FILTER (WHERE status IS NULL OR LOWER(status) != 'cancelled'), 0) as cgst,
			COALESCE(SUM(CASE WHEN LOWER(customer_state) IN ('tamil nadu', 'tn') THEN (total_price - ROUND(total_price / 1.18, 2)) / 2 ELSE 0 END) FILTER (WHERE status IS NULL OR LOWER(status) != 'cancelled'), 0) as sgst,
			COALESCE(SUM(CASE WHEN LOWER(customer_state) NOT IN ('tamil nadu', 'tn') THEN (total_price - ROUND(total_price / 1.18, 2)) ELSE 0 END) FILTER (WHERE status IS NULL OR LOWER(status) != 'cancelled'), 0) as igst,
			COUNT(id) as total_orders,
			COUNT(id) FILTER (WHERE LOWER(status) = 'cancelled') as cancelled_orders,
			COUNT(id) FILTER (WHERE LOWER(fulfillment_status) = 'fulfilled') as fulfilled_orders,
			COUNT(id) FILTER (WHERE LOWER(fulfillment_status) != 'fulfilled' AND (status IS NULL OR LOWER(status) != 'cancelled')) as unfulfilled_orders
		FROM orders 
		WHERE created_at >= ? AND created_at <= ?
	`

	type metricsResult struct {
		TotalRevenue      float64
		CGST              float64
		SGST              float64
		IGST              float64
		TotalOrders       int
		CancelledOrders   int
		FulfilledOrders   int
		UnfulfilledOrders int
	}

	var result metricsResult
	err = r.db.Raw(query, start, end).Scan(&result).Error
	if err != nil {
		return
	}

	return result.TotalRevenue, result.CGST, result.SGST, result.IGST,
		result.TotalOrders, result.CancelledOrders, result.FulfilledOrders, result.UnfulfilledOrders, nil
}

// gormReportRepository is the GORM implementation of ReportRepository.
type gormReportRepository struct {
	db *gorm.DB
}

// NewReportRepository creates a new GORM-backed ReportRepository.
func NewReportRepository(db *gorm.DB) ReportRepository {
	return &gormReportRepository{db: db}
}

func (r *gormReportRepository) GetGSTSummary(startDate, endDate string) (totalOrders, cancelledOrders, fulfilledOrders, unfulfilledOrders, paidOrders int, totalRevenue, totalTaxable, totalTax float64, err error) {
	start, end := parseDateRange(startDate, endDate)

	query := `
		SELECT 
			COUNT(id),
			COUNT(id) FILTER (WHERE LOWER(status) = 'cancelled'),
			COUNT(id) FILTER (WHERE LOWER(fulfillment_status) = 'fulfilled'),
			COUNT(id) FILTER (WHERE LOWER(fulfillment_status) != 'fulfilled' AND (status IS NULL OR LOWER(status) != 'cancelled')),
			COUNT(id) FILTER (WHERE LOWER(financial_status) = 'paid'),
			COALESCE(SUM(total_price) FILTER (WHERE status IS NULL OR LOWER(status) != 'cancelled'), 0) as revenue,
			COALESCE(SUM(ROUND(total_price / 1.18, 2)) FILTER (WHERE status IS NULL OR LOWER(status) != 'cancelled'), 0) as taxable,
			COALESCE(SUM(total_price - ROUND(total_price / 1.18, 2)) FILTER (WHERE status IS NULL OR LOWER(status) != 'cancelled'), 0) as tax
		FROM orders 
		WHERE created_at >= ? AND created_at <= ?
	`

	type summaryResult struct {
		TotalOrders       int
		CancelledOrders   int
		FulfilledOrders   int
		UnfulfilledOrders int
		PaidOrders        int
		Revenue           float64
		Taxable           float64
		Tax               float64
	}

	var result summaryResult
	row := r.db.Raw(query, start, end).Row()
	err = row.Scan(
		&result.TotalOrders, &result.CancelledOrders, &result.FulfilledOrders,
		&result.UnfulfilledOrders, &result.PaidOrders,
		&result.Revenue, &result.Taxable, &result.Tax,
	)
	if err != nil {
		return
	}

	return result.TotalOrders, result.CancelledOrders, result.FulfilledOrders,
		result.UnfulfilledOrders, result.PaidOrders,
		result.Revenue, result.Taxable, result.Tax, nil
}

func (r *gormReportRepository) GetStateSummary(startDate, endDate string) ([]StateSummaryResult, error) {
	start, end := parseDateRange(startDate, endDate)

	query := `
		SELECT 
			COALESCE(customer_state, 'N/A') as state,
			COUNT(id) as orders,
			COALESCE(SUM(ROUND(total_price / 1.18, 2)), 0) as taxable_value,
			COALESCE(SUM(total_price - ROUND(total_price / 1.18, 2)), 0) as total_gst,
			COALESCE(SUM(total_price), 0) as revenue
		FROM orders
		WHERE created_at >= ? AND created_at <= ? AND (status IS NULL OR LOWER(status) != 'cancelled')
		GROUP BY customer_state
		ORDER BY revenue DESC
	`

	var results []StateSummaryResult
	if err := r.db.Raw(query, start, end).Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to query state summary: %w", err)
	}
	return results, nil
}

func (r *gormReportRepository) GetHSNSummary(startDate, endDate string) ([]HSNSummaryResult, error) {
	start, end := parseDateRange(startDate, endDate)

	// Optimization: Replaces a global CTE that aggregated the entire order_line_items table
	// with a window function (SUM(...) OVER (PARTITION BY li.order_id)) applied
	// within a date-filtered subquery.
	// Expected Impact: Prevents full table scan of order_line_items, reducing query time from O(TotalItems) to O(FilteredItems).
	query := `
		WITH LineItemShares AS (
			SELECT 
				li.order_id,
				COALESCE(li.hs_code, '33029019') as hs_code,
				li.quantity,
				(li.price * li.quantity - li.discount) as line_val,
				SUM(li.price * li.quantity - li.discount) OVER (PARTITION BY li.order_id) as line_sum,
				o.total_price,
				COALESCE(o.customer_state, 'N/A') as state
			FROM order_line_items li
			JOIN orders o ON li.order_id = o.id
			WHERE o.created_at >= ? AND o.created_at <= ?
			  AND (o.status IS NULL OR LOWER(o.status) != 'cancelled')
		)
		SELECT 
			hs_code as hsn_code,
			COUNT(DISTINCT order_id) as product_count,
			SUM(quantity) as qty_sold,
			SUM(ROUND((line_val / line_sum) * (total_price / 1.18), 2)) as taxable_value,
			SUM(ROUND((line_val / line_sum) * total_price, 2) - ROUND((line_val / line_sum) * (total_price / 1.18), 2)) as total_gst,
			SUM(ROUND((line_val / line_sum) * total_price, 2)) as revenue,
			state
		FROM LineItemShares
		WHERE line_sum > 0
		GROUP BY hs_code, state
		ORDER BY revenue DESC
	`

	var results []HSNSummaryResult
	if err := r.db.Raw(query, start, end).Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to query HSN summary: %w", err)
	}
	return results, nil
}

func (r *gormReportRepository) GetDocumentsIssued(startDate, endDate string) (minOrder, maxOrder *int64, total, cancelled int, err error) {
	start, end := parseDateRange(startDate, endDate)

	query := `
		SELECT 
			MIN(NULLIF(regexp_replace(order_number, '[^0-9]', '', 'g'), '')::bigint) as min_val,
			MAX(NULLIF(regexp_replace(order_number, '[^0-9]', '', 'g'), '')::bigint) as max_val,
			COUNT(id) as total,
			COUNT(id) FILTER (WHERE LOWER(status) = 'cancelled') as cancelled
		FROM orders
		WHERE created_at >= ? AND created_at <= ?
	`

	var minV, maxV sql.NullInt64
	row := r.db.Raw(query, start, end).Row()
	err = row.Scan(&minV, &maxV, &total, &cancelled)
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

func (r *gormReportRepository) GetTaxByState(startDate, endDate string) ([]StateTaxResult, error) {
	start, end := parseDateRange(startDate, endDate)

	query := `
		SELECT 
			COALESCE(customer_state, 'N/A') as state,
			SUM(total_price - ROUND(total_price / 1.18, 2)) as tax
		FROM orders
		WHERE created_at >= ? AND created_at <= ? AND (status IS NULL OR LOWER(status) != 'cancelled')
		GROUP BY customer_state
	`

	var results []StateTaxResult
	if err := r.db.Raw(query, start, end).Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to query tax by state: %w", err)
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
	t, err := time.Parse(time.RFC3339, s)
	if err == nil {
		return t
	}
	t, err = time.Parse("2006-01-02T15:04:05.000Z", s)
	if err == nil {
		return t
	}
	return time.Time{}
}

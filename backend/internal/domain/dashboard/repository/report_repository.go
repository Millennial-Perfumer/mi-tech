package repository

import (
	"database/sql"
	"fmt"
	"time"

	"mi-tech/internal/domain/dashboard/dto"

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

func (r *gormMetricsRepository) GetDashboardMetrics(startDate, endDate string, sourceIDs []string) (dto.DashboardMetrics, error) {
	start, end := parseDateRange(startDate, endDate)

	sourceFilter := ""
	args := []interface{}{start, end}
	if len(sourceIDs) > 0 {
		sourceFilter = " AND source_id IN ?"
		args = append(args, sourceIDs)
	}

	query := `
		SELECT 
			COALESCE(SUM(total_price) FILTER (WHERE NOT (LOWER(COALESCE(status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(fulfillment_status, '')) IN ('cancelled', 'canceled'))), 0) as total_revenue,
			COALESCE(SUM(CASE WHEN LOWER(customer_state) IN ('tamil nadu', 'tn', 'tamilnadu') THEN (total_price - ROUND(total_price / 1.18, 2)) / 2 ELSE 0 END) FILTER (WHERE NOT (LOWER(COALESCE(status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(fulfillment_status, '')) IN ('cancelled', 'canceled'))), 0) as cgst,
			COALESCE(SUM(CASE WHEN LOWER(customer_state) IN ('tamil nadu', 'tn', 'tamilnadu') THEN (total_price - ROUND(total_price / 1.18, 2)) / 2 ELSE 0 END) FILTER (WHERE NOT (LOWER(COALESCE(status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(fulfillment_status, '')) IN ('cancelled', 'canceled'))), 0) as sgst,
			COALESCE(SUM(CASE WHEN LOWER(customer_state) NOT IN ('tamil nadu', 'tn', 'tamilnadu') THEN (total_price - ROUND(total_price / 1.18, 2)) ELSE 0 END) FILTER (WHERE NOT (LOWER(COALESCE(status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(fulfillment_status, '')) IN ('cancelled', 'canceled'))), 0) as igst,
			COUNT(id) as total_orders,
			COUNT(id) FILTER (WHERE LOWER(status) IN ('cancelled', 'canceled') OR LOWER(fulfillment_status) IN ('cancelled', 'canceled')) as cancelled_orders,
			COUNT(id) FILTER (WHERE LOWER(fulfillment_status) = 'fulfilled') as fulfilled_orders,
			COUNT(id) FILTER (WHERE LOWER(COALESCE(fulfillment_status, '')) != 'fulfilled' AND NOT (LOWER(COALESCE(status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(fulfillment_status, '')) IN ('cancelled', 'canceled'))) as unfulfilled_orders,
			COALESCE(SUM(total_discount), 0) as total_discount,
			COUNT(id) FILTER (WHERE LOWER(financial_status) = 'paid') as paid_orders,
			COUNT(id) FILTER (WHERE LOWER(financial_status) = 'pending') as pending_orders,
			COUNT(id) FILTER (WHERE LOWER(financial_status) = 'partially_paid') as partial_orders
		FROM orders 
		WHERE created_at >= ? AND created_at <= ?` + sourceFilter

	type metricsResult struct {
		TotalRevenue      float64
		CGST              float64
		SGST              float64
		IGST              float64
		TotalOrders       int
		CancelledOrders   int
		FulfilledOrders   int
		UnfulfilledOrders int
		TotalDiscount     float64
		PaidOrders        int
		PendingOrders     int
		PartialOrders     int
	}

	var res metricsResult
	err := r.db.Raw(query, args...).Scan(&res).Error
	if err != nil {
		return dto.DashboardMetrics{}, err
	}

	// Channel Breakdown Query
	var channelMetrics []dto.ChannelMetrics
	channelQuery := `
		SELECT 
			source_id,
			COALESCE(SUM(total_price), 0) as revenue,
			COUNT(id) as orders,
			COALESCE(AVG(total_price), 0) as aov
		FROM orders
		WHERE created_at >= ? AND created_at <= ? AND NOT (LOWER(COALESCE(status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(fulfillment_status, '')) IN ('cancelled', 'canceled'))` + sourceFilter + `
		GROUP BY source_id
		ORDER BY revenue DESC`

	err = r.db.Raw(channelQuery, args...).Scan(&channelMetrics).Error
	if err != nil {
		return dto.DashboardMetrics{}, err
	}

	discountPercent := 0.0
	if res.TotalRevenue > 0 {
		discountPercent = (res.TotalDiscount / (res.TotalRevenue + res.TotalDiscount)) * 100
	}

	return dto.DashboardMetrics{
		TotalRevenue:      res.TotalRevenue,
		TotalInvoices:     res.TotalOrders,
		TotalGSTCollected: res.CGST + res.SGST + res.IGST,
		CGSTCollected:     res.CGST,
		SGSTCollected:     res.SGST,
		IGSTCollected:     res.IGST,
		TotalOrders:       res.TotalOrders,
		CancelledOrders:   res.CancelledOrders,
		FulfilledOrders:   res.FulfilledOrders,
		UnfulfilledOrders: res.UnfulfilledOrders,
		TotalDiscount:     res.TotalDiscount,
		DiscountPercent:   discountPercent,
		ChannelBreakdown:  channelMetrics,
		PaymentBreakdown: dto.PaymentHealth{
			Paid:      res.PaidOrders,
			Pending:   res.PendingOrders,
			Partial:   res.PartialOrders,
			Cancelled: res.CancelledOrders,
		},
	}, nil
}

func (r *gormMetricsRepository) GetTopProducts(startDate, endDate string, sourceIDs []string, limit int) ([]dto.TopProductRow, error) {
	start, end := parseDateRange(startDate, endDate)
	sourceFilter := ""
	args := []interface{}{start, end}
	if len(sourceIDs) > 0 {
		sourceFilter = " AND o.source_id IN ?"
		args = append(args, sourceIDs)
	}

	query := `
		SELECT 
			li.sku,
			li.title,
			SUM(li.quantity) as quantity,
			SUM(li.price * li.quantity - li.discount) as revenue
		FROM order_line_items li
		JOIN orders o ON li.order_id = o.id
		WHERE o.created_at >= ? AND o.created_at <= ? AND NOT (LOWER(COALESCE(o.status, '')) IN ('cancelled', 'canceled'))` + sourceFilter + `
		GROUP BY li.sku, li.title
		ORDER BY quantity DESC
		LIMIT ?`
	args = append(args, limit)

	var results []dto.TopProductRow
	err := r.db.Raw(query, args...).Scan(&results).Error
	return results, err
}

func (r *gormMetricsRepository) GetRevenueTrend(startDate, endDate string, sourceIDs []string) ([]dto.RevenueTrendRow, error) {
	start, end := parseDateRange(startDate, endDate)
	sourceFilter := ""
	args := []interface{}{start, end}
	if len(sourceIDs) > 0 {
		sourceFilter = " AND source_id IN ?"
		args = append(args, sourceIDs)
	}

	query := `
		SELECT 
			TO_CHAR(created_at, 'YYYY-MM-DD') as date,
			SUM(total_price) as revenue,
			COUNT(id) as orders
		FROM orders
		WHERE created_at >= ? AND created_at <= ? AND NOT (LOWER(COALESCE(status, '')) IN ('cancelled', 'canceled'))` + sourceFilter + `
		GROUP BY date
		ORDER BY date ASC`

	var results []dto.RevenueTrendRow
	err := r.db.Raw(query, args...).Scan(&results).Error
	return results, err
}

func (r *gormMetricsRepository) GetGeoDistribution(startDate, endDate string, sourceIDs []string, limit int) ([]dto.GeoDistributionRow, error) {
	start, end := parseDateRange(startDate, endDate)
	sourceFilter := ""
	args := []interface{}{start, end}
	if len(sourceIDs) > 0 {
		sourceFilter = " AND source_id IN ?"
		args = append(args, sourceIDs)
	}

	query := `
		SELECT 
			INITCAP(COALESCE(customer_state, 'Unknown')) as state,
			COUNT(id) as orders,
			SUM(total_price) as revenue
		FROM orders
		WHERE created_at >= ? AND created_at <= ? AND NOT (LOWER(COALESCE(status, '')) IN ('cancelled', 'canceled'))` + sourceFilter + `
		GROUP BY state
		ORDER BY orders DESC
		LIMIT ?`
	args = append(args, limit)

	var results []dto.GeoDistributionRow
	err := r.db.Raw(query, args...).Scan(&results).Error
	return results, err
}

// gormReportRepository is the GORM implementation of ReportRepository.
type gormReportRepository struct {
	db *gorm.DB
}

// NewReportRepository creates a new GORM-backed ReportRepository.
func NewReportRepository(db *gorm.DB) ReportRepository {
	return &gormReportRepository{db: db}
}

func (r *gormReportRepository) GetGSTSummary(startDate, endDate string) (GSTSummaryResult, error) {
	start, end := parseDateRange(startDate, endDate)

	query := `
		SELECT 
			COUNT(id),
			COUNT(id) FILTER (WHERE LOWER(status) IN ('cancelled', 'canceled') OR LOWER(fulfillment_status) IN ('cancelled', 'canceled')),
			COUNT(id) FILTER (WHERE LOWER(fulfillment_status) = 'fulfilled'),
			COUNT(id) FILTER (WHERE LOWER(COALESCE(fulfillment_status, '')) != 'fulfilled' AND NOT (LOWER(COALESCE(status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(fulfillment_status, '')) IN ('cancelled', 'canceled'))),
			COUNT(id) FILTER (WHERE LOWER(financial_status) = 'paid'),
			COALESCE(SUM(total_price) FILTER (WHERE NOT (LOWER(COALESCE(status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(fulfillment_status, '')) IN ('cancelled', 'canceled'))), 0) as revenue,
			COALESCE(SUM(ROUND(total_price / 1.18, 2)) FILTER (WHERE NOT (LOWER(COALESCE(status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(fulfillment_status, '')) IN ('cancelled', 'canceled'))), 0) as taxable,
			COALESCE(SUM(total_price - ROUND(total_price / 1.18, 2)) FILTER (WHERE NOT (LOWER(COALESCE(status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(fulfillment_status, '')) IN ('cancelled', 'canceled'))), 0) as tax,
			COALESCE(SUM(CASE WHEN LOWER(customer_state) IN ('tamil nadu', 'tn', 'tamilnadu') THEN (total_price - ROUND(total_price / 1.18, 2)) / 2 ELSE 0 END) FILTER (WHERE NOT (LOWER(COALESCE(status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(fulfillment_status, '')) IN ('cancelled', 'canceled'))), 0) as cgst,
			COALESCE(SUM(CASE WHEN LOWER(customer_state) IN ('tamil nadu', 'tn', 'tamilnadu') THEN (total_price - ROUND(total_price / 1.18, 2)) / 2 ELSE 0 END) FILTER (WHERE NOT (LOWER(COALESCE(status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(fulfillment_status, '')) IN ('cancelled', 'canceled'))), 0) as sgst,
			COALESCE(SUM(CASE WHEN LOWER(customer_state) NOT IN ('tamil nadu', 'tn', 'tamilnadu') THEN (total_price - ROUND(total_price / 1.18, 2)) ELSE 0 END) FILTER (WHERE NOT (LOWER(COALESCE(status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(fulfillment_status, '')) IN ('cancelled', 'canceled'))), 0) as igst
		FROM orders 
		WHERE created_at >= ? AND created_at <= ?
	`

	var result GSTSummaryResult
	row := r.db.Raw(query, start, end).Row()
	err := row.Scan(
		&result.TotalOrders, &result.CancelledOrders, &result.FulfilledOrders,
		&result.UnfulfilledOrders, &result.PaidOrders,
		&result.TotalRevenue, &result.TotalTaxable, &result.TotalTax,
		&result.CGST, &result.SGST, &result.IGST,
	)
	if err != nil {
		return GSTSummaryResult{}, err
	}

	return result, nil
}

func (r *gormReportRepository) GetStateSummary(startDate, endDate string) ([]StateSummaryResult, error) {
	start, end := parseDateRange(startDate, endDate)

	query := `
		SELECT 
			INITCAP(COALESCE(customer_state, 'N/A')) as state,
			COUNT(id) as orders,
			COALESCE(SUM(ROUND(total_price / 1.18, 2)), 0) as taxable_value,
			COALESCE(SUM(total_price - ROUND(total_price / 1.18, 2)), 0) as total_gst,
			COALESCE(SUM(total_price), 0) as revenue
		FROM orders
		WHERE created_at >= ? AND created_at <= ? AND NOT (LOWER(COALESCE(status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(fulfillment_status, '')) IN ('cancelled', 'canceled'))
		GROUP BY INITCAP(COALESCE(customer_state, 'N/A'))
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

	query := `
		WITH LineItemShares AS (
			SELECT 
				li.order_id,
				COALESCE(li.hs_code, '33029019') as hs_code,
				li.quantity,
				(li.price * li.quantity - li.discount) as line_val,
				SUM(li.price * li.quantity - li.discount) OVER (PARTITION BY li.order_id) as line_sum,
				o.total_price,
				ROUND(o.total_price / 1.18, 2) as order_taxable,
				(o.total_price - ROUND(o.total_price / 1.18, 2)) as order_tax,
				INITCAP(COALESCE(o.customer_state, 'N/A')) as state
			FROM order_line_items li
			JOIN orders o ON li.order_id = o.id
			WHERE o.created_at >= ? AND o.created_at <= ? AND NOT (LOWER(COALESCE(o.status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(o.fulfillment_status, '')) IN ('cancelled', 'canceled'))
		)
		SELECT 
			hs_code as hsn_code,
			COUNT(DISTINCT order_id) as product_count,
			SUM(quantity) as qty_sold,
			ROUND(SUM((line_val / line_sum) * order_taxable), 2) as taxable_value,
			ROUND(SUM((line_val / line_sum) * order_tax), 2) as total_gst,
			ROUND(SUM((line_val / line_sum) * total_price), 2) as revenue,
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
			COUNT(id) FILTER (WHERE LOWER(status) IN ('cancelled', 'canceled') OR LOWER(fulfillment_status) IN ('cancelled', 'canceled')) as cancelled
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

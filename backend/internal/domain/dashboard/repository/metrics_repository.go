package repository

import (
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
		sourceFilter = " AND o.source_id IN ?"
		args = append(args, sourceIDs)
	}

	query := `
		SELECT 
			COALESCE(SUM(o.total_price) FILTER (WHERE NOT (LOWER(COALESCE(o.status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(o.fulfillment_status, '')) IN ('cancelled', 'canceled'))), 0) as total_revenue,
			COALESCE(SUM(CASE WHEN COALESCE(s.code, '33') = '33' THEN (o.total_price - ROUND(o.total_price / 1.18, 2)) / 2 ELSE 0 END) FILTER (WHERE NOT (LOWER(COALESCE(o.status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(o.fulfillment_status, '')) IN ('cancelled', 'canceled'))), 0) as cgst,
			COALESCE(SUM(CASE WHEN COALESCE(s.code, '33') = '33' THEN (o.total_price - ROUND(o.total_price / 1.18, 2)) / 2 ELSE 0 END) FILTER (WHERE NOT (LOWER(COALESCE(o.status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(o.fulfillment_status, '')) IN ('cancelled', 'canceled'))), 0) as sgst,
			COALESCE(SUM(CASE WHEN COALESCE(s.code, '33') != '33' THEN (o.total_price - ROUND(o.total_price / 1.18, 2)) ELSE 0 END) FILTER (WHERE NOT (LOWER(COALESCE(o.status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(o.fulfillment_status, '')) IN ('cancelled', 'canceled'))), 0) as igst,
			COUNT(o.id) as total_orders,
			COUNT(o.id) FILTER (WHERE LOWER(o.status) IN ('cancelled', 'canceled') OR LOWER(COALESCE(o.fulfillment_status, '')) IN ('cancelled', 'canceled')) as cancelled_orders,
			COUNT(o.id) FILTER (WHERE LOWER(o.fulfillment_status) = 'fulfilled') as fulfilled_orders,
			COUNT(o.id) FILTER (WHERE LOWER(COALESCE(o.fulfillment_status, '')) != 'fulfilled' AND NOT (LOWER(COALESCE(o.status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(o.fulfillment_status, '')) IN ('cancelled', 'canceled'))) as unfulfilled_orders,
			COALESCE(SUM(o.total_discount), 0) as total_discount,
			COUNT(o.id) FILTER (WHERE LOWER(o.financial_status) = 'paid') as paid_orders,
			COUNT(o.id) FILTER (WHERE LOWER(o.financial_status) = 'pending') as pending_orders,
			COUNT(o.id) FILTER (WHERE LOWER(o.financial_status) = 'partially_paid') as partial_orders
		FROM orders o
		LEFT JOIN gst_state_codes s ON LOWER(TRIM(o.customer_state)) = ANY(s.aliases)
		WHERE o.created_at >= ? AND o.created_at <= ?` + sourceFilter

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
			SUM(li.price * li.quantity - li.discount - COALESCE(li.order_discount, 0)) as revenue
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

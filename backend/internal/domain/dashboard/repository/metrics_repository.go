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
		sourceFilter = " AND t.source_id IN ?"
		args = append(args, sourceIDs)
	}

	query := `
		SELECT 
			COALESCE(SUM(t.total_price), 0) as total_revenue,
			COALESCE(SUM(CASE WHEN COALESCE(s.code, '33') = '33' THEN (t.total_price - ROUND(t.total_price / 1.18, 2)) / 2 ELSE 0 END), 0) as cgst,
			COALESCE(SUM(CASE WHEN COALESCE(s.code, '33') = '33' THEN (t.total_price - ROUND(t.total_price / 1.18, 2)) / 2 ELSE 0 END), 0) as sgst,
			COALESCE(SUM(CASE WHEN COALESCE(s.code, '33') != '33' THEN (t.total_price - ROUND(t.total_price / 1.18, 2)) ELSE 0 END), 0) as igst,
			COUNT(t.transaction_id) as total_orders,
			COUNT(t.transaction_id) FILTER (WHERE LOWER(t.order_status) IN ('cancelled', 'canceled')) as cancelled_orders,
			COUNT(t.transaction_id) FILTER (WHERE LOWER(t.order_status) = 'fulfilled') as fulfilled_orders,
			COUNT(t.transaction_id) FILTER (WHERE LOWER(COALESCE(t.order_status, '')) != 'fulfilled' AND NOT (LOWER(COALESCE(t.order_status, '')) IN ('cancelled', 'canceled'))) as unfulfilled_orders,
			COALESCE(SUM(t.total_discount), 0) as total_discount,
			COUNT(t.transaction_id) FILTER (WHERE LOWER(t.payment_status) = 'paid') as paid_orders,
			COUNT(t.transaction_id) FILTER (WHERE LOWER(t.payment_status) IN ('pending', 'unpaid')) as pending_orders,
			COUNT(t.transaction_id) FILTER (WHERE LOWER(t.payment_status) IN ('partially_paid', 'partial')) as partial_orders
		FROM unified_revenue_transactions t
		LEFT JOIN gst_state_codes s ON LOWER(TRIM(t.state)) = ANY(s.aliases)
		WHERE t.tx_date >= ? AND t.tx_date <= ?` + sourceFilter

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
			COUNT(transaction_id) as orders,
			COALESCE(AVG(total_price), 0) as aov
		FROM unified_revenue_transactions t
		WHERE tx_date >= ? AND tx_date <= ?` + sourceFilter + `
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
		sourceFilter = " AND t.source_id IN ?"
		args = append(args, sourceIDs)
	}

	query := `
		SELECT 
			t.sku,
			t.title,
			SUM(t.quantity) as quantity,
			SUM(t.price * t.quantity - t.discount - t.order_discount) as revenue
		FROM (
			SELECT 
				li.sku,
				li.title,
				li.quantity::numeric as quantity,
				li.price,
				li.discount,
				COALESCE(li.order_discount, 0) as order_discount,
				o.created_at as tx_date,
				o.source_id
			FROM order_line_items li
			JOIN orders o ON li.order_id = o.id
			WHERE NOT (LOWER(COALESCE(o.status, '')) IN ('cancelled', 'canceled'))

			UNION ALL

			SELECT 
				bi.sku,
				bi.item_details as title,
				bi.quantity as quantity,
				bi.rate as price,
				0.00 as discount,
				0.00 as order_discount,
				inv.created_at as tx_date,
				'B2B' as source_id
			FROM b2b_invoice_items bi
			JOIN b2b_invoices inv ON bi.invoice_id = inv.id
			WHERE inv.status = 'ISSUED'
		) t
		WHERE t.tx_date >= ? AND t.tx_date <= ?` + sourceFilter + `
		GROUP BY t.sku, t.title
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
		sourceFilter = " AND t.source_id IN ?"
		args = append(args, sourceIDs)
	}

	query := `
		SELECT 
			TO_CHAR(t.tx_date, 'YYYY-MM-DD') as date,
			SUM(t.total_price) as revenue,
			COUNT(t.transaction_id) as orders
		FROM unified_revenue_transactions t
		WHERE t.tx_date >= ? AND t.tx_date <= ?` + sourceFilter + `
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
		sourceFilter = " AND t.source_id IN ?"
		args = append(args, sourceIDs)
	}

	query := `
		SELECT 
			INITCAP(COALESCE(t.state, 'Unknown')) as state,
			COUNT(t.transaction_id) as orders,
			SUM(t.total_price) as revenue
		FROM unified_revenue_transactions t
		WHERE t.tx_date >= ? AND t.tx_date <= ?` + sourceFilter + `
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

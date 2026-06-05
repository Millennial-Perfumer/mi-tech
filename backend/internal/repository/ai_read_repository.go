package repository

import (
	"mi-tech/internal/entity"

	"gorm.io/gorm"
)

// AIReadRepository defines the contract for aggregate data retrieval for AI analysis.
type AIReadRepository interface {
	GetRevenueSummary(startDate, endDate string) (entity.AIRevenueSummary, error)
	GetRevenueByChannel(startDate, endDate string) ([]entity.AIChannelRevenue, error)
	GetRevenueByState(startDate, endDate string) ([]entity.AIStateRevenue, error)
	GetDailyRevenueTrend(startDate, endDate string) ([]entity.AIDailyRevenue, error)
	GetTopProducts(startDate, endDate string, limit int) ([]entity.AIProductRank, error)
	GetProductPerformance(sku string) (entity.AIProductStats, error)
	GetCustomerSegmentation() (entity.AICustomerSegments, error)
	GetTopCustomers(limit int) ([]entity.AITopCustomer, error)
	GetInventorySnapshot() ([]entity.AIInventoryStatus, error)
	GetBusinessSnapshot() (entity.AIBusinessSnapshot, error)
	ExecuteRawQuery(query string) ([]map[string]interface{}, error)
	ListTables() ([]string, error)
	DescribeTable(tableName string) ([]map[string]interface{}, error)
}

type gormAIReadRepository struct {
	db    *gorm.DB
	guard *QueryGuard
}

func NewAIReadRepository(db *gorm.DB) AIReadRepository {
	return &gormAIReadRepository{
		db:    db,
		guard: NewQueryGuard(),
	}
}

func (r *gormAIReadRepository) GetRevenueSummary(startDate, endDate string) (entity.AIRevenueSummary, error) {
	startDate, endDate = normalizeDates(startDate, endDate)
	var result entity.AIRevenueSummary
	query := `
		SELECT 
			COALESCE(SUM(total_price), 0) as total_revenue,
			COUNT(*) as total_orders
		FROM orders 
		WHERE created_at >= ? AND created_at <= ? 
		AND NOT (LOWER(COALESCE(status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(fulfillment_status, '')) IN ('cancelled', 'canceled'))
	`
	if err := r.guard.IsSafe(query); err != nil {
		return result, err
	}

	err := r.db.Raw(query, startDate, endDate).Scan(&result).Error
	result.StartDate = startDate
	result.EndDate = endDate
	return result, err
}

func (r *gormAIReadRepository) GetRevenueByChannel(startDate, endDate string) ([]entity.AIChannelRevenue, error) {
	startDate, endDate = normalizeDates(startDate, endDate)
	var results []entity.AIChannelRevenue
	query := `
		SELECT 
			source_id as channel,
			COALESCE(SUM(total_price), 0) as revenue,
			COUNT(*) as orders
		FROM orders 
		WHERE created_at >= ? AND created_at <= ? 
		AND NOT (LOWER(COALESCE(status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(fulfillment_status, '')) IN ('cancelled', 'canceled'))
		GROUP BY source_id
		ORDER BY revenue DESC
	`
	if err := r.guard.IsSafe(query); err != nil {
		return nil, err
	}

	err := r.db.Raw(query, startDate, endDate).Scan(&results).Error
	return results, err
}

func (r *gormAIReadRepository) GetRevenueByState(startDate, endDate string) ([]entity.AIStateRevenue, error) {
	startDate, endDate = normalizeDates(startDate, endDate)
	var results []entity.AIStateRevenue
	query := `
		SELECT 
			customer_state as state,
			COALESCE(SUM(total_price), 0) as revenue,
			COUNT(*) as orders
		FROM orders 
		WHERE created_at >= ? AND created_at <= ? 
		AND NOT (LOWER(COALESCE(status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(fulfillment_status, '')) IN ('cancelled', 'canceled'))
		GROUP BY customer_state
		ORDER BY revenue DESC
	`
	if err := r.guard.IsSafe(query); err != nil {
		return nil, err
	}

	err := r.db.Raw(query, startDate, endDate).Scan(&results).Error
	return results, err
}

func (r *gormAIReadRepository) GetDailyRevenueTrend(startDate, endDate string) ([]entity.AIDailyRevenue, error) {
	startDate, endDate = normalizeDates(startDate, endDate)
	var results []entity.AIDailyRevenue
	query := `
		SELECT 
			DATE(created_at) as date,
			COALESCE(SUM(total_price), 0) as revenue
		FROM orders 
		WHERE created_at >= ? AND created_at <= ? 
		AND NOT (LOWER(COALESCE(status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(fulfillment_status, '')) IN ('cancelled', 'canceled'))
		GROUP BY DATE(created_at)
		ORDER BY date ASC
	`
	if err := r.guard.IsSafe(query); err != nil {
		return nil, err
	}

	err := r.db.Raw(query, startDate, endDate).Scan(&results).Error
	return results, err
}

func (r *gormAIReadRepository) GetTopProducts(startDate, endDate string, limit int) ([]entity.AIProductRank, error) {
	startDate, endDate = normalizeDates(startDate, endDate)
	var results []entity.AIProductRank
	query := `
		SELECT 
			COALESCE(li.title, '') as name,
			COALESCE(li.sku, '') as sku,
			COALESCE(SUM(li.quantity), 0) as qty_sold,
			COALESCE(SUM(li.price * li.quantity), 0) as revenue
		FROM order_line_items li
		JOIN orders ON li.order_id = orders.id
		WHERE orders.created_at >= ? AND orders.created_at <= ?
		AND NOT (LOWER(COALESCE(orders.status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(orders.fulfillment_status, '')) IN ('cancelled', 'canceled'))
		GROUP BY li.title, li.sku
		ORDER BY qty_sold DESC
		LIMIT ?
	`
	if err := r.guard.IsSafe(query); err != nil {
		return nil, err
	}

	err := r.db.Raw(query, startDate, endDate, limit).Scan(&results).Error
	return results, err
}

func (r *gormAIReadRepository) GetProductPerformance(sku string) (entity.AIProductStats, error) {
	var stats entity.AIProductStats
	// Stock and general info from inventory_items
	queryStock := `
		SELECT 
			mi_sku as sku, title as name, current_stock
		FROM inventory_items
		WHERE mi_sku = ?
		LIMIT 1
	`
	if err := r.guard.IsSafe(queryStock); err != nil {
		return stats, err
	}
	r.db.Raw(queryStock, sku).Scan(&stats)

	// Sales volume from order_line_items
	querySales := `
		SELECT 
			COALESCE(SUM(li.quantity), 0) as total_sold,
			COALESCE(SUM(li.price * li.quantity), 0) as inventory_value
		FROM order_line_items li
		JOIN orders ON li.order_id = orders.id
		WHERE li.sku = ? AND NOT (LOWER(COALESCE(orders.status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(orders.fulfillment_status, '')) IN ('cancelled', 'canceled'))
	`
	if err := r.guard.IsSafe(querySales); err != nil {
		return stats, err
	}
	var salesInfo struct {
		TotalSold      int
		InventoryValue float64
	}
	r.db.Raw(querySales, sku).Scan(&salesInfo)
	stats.TotalSold = salesInfo.TotalSold
	stats.InventoryValue = salesInfo.InventoryValue

	// Avg daily sales (last 90 days)
	queryAvg := `
		SELECT COALESCE(SUM(li.quantity), 0) / 90.0 as avg_sales
		FROM order_line_items li
		JOIN orders ON li.order_id = orders.id
		WHERE li.sku = ? AND NOT (LOWER(COALESCE(orders.status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(orders.fulfillment_status, '')) IN ('cancelled', 'canceled'))
		AND orders.created_at >= NOW() - INTERVAL '90 days'
	`
	if err := r.guard.IsSafe(queryAvg); err != nil {
		return stats, err
	}
	r.db.Raw(queryAvg, sku).Scan(&stats.AverageDaily)

	return stats, nil
}

func (r *gormAIReadRepository) GetCustomerSegmentation() (entity.AICustomerSegments, error) {
	var result entity.AICustomerSegments

	var total int64
	r.db.Model(&entity.Customer{}).Count(&total)
	result.TotalCustomers = int(total)

	r.db.Raw("SELECT COUNT(*) FROM customers WHERE created_at >= NOW() - INTERVAL '30 days'").Scan(&result.NewCustomers)

	// Repeat Rate
	var totalCustomersWithOrders int64
	r.db.Raw("SELECT COUNT(DISTINCT customer_phone) FROM orders").Scan(&totalCustomersWithOrders)
	var repeatCustomers int64
	r.db.Raw("SELECT COUNT(*) FROM (SELECT customer_phone FROM orders GROUP BY customer_phone HAVING COUNT(*) > 1) as t").Scan(&repeatCustomers)

	if totalCustomersWithOrders > 0 {
		result.RepeatRate = float64(repeatCustomers) / float64(totalCustomersWithOrders)
	}

	// Channel Split
	rows, _ := r.db.Raw("SELECT source_id as channel, COUNT(*) as count FROM customers GROUP BY source_id").Rows()
	defer rows.Close()
	result.ChannelSplit = make(map[string]int)
	for rows.Next() {
		var channel string
		var count int
		rows.Scan(&channel, &count)
		result.ChannelSplit[channel] = count
	}

	return result, nil
}

func (r *gormAIReadRepository) GetTopCustomers(limit int) ([]entity.AITopCustomer, error) {
	var results []entity.AITopCustomer
	query := `
		SELECT 
			COALESCE(customer_name, '') as name,
			COALESCE(customer_phone, '') as phone,
			COALESCE(SUM(total_price), 0) as total_spend,
			COUNT(*) as order_count
		FROM orders
		WHERE NOT (LOWER(COALESCE(status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(fulfillment_status, '')) IN ('cancelled', 'canceled'))
		GROUP BY customer_name, customer_phone
		ORDER BY total_spend DESC
		LIMIT ?
	`
	if err := r.guard.IsSafe(query); err != nil {
		return nil, err
	}
	err := r.db.Raw(query, limit).Scan(&results).Error
	return results, err
}

func (r *gormAIReadRepository) GetInventorySnapshot() ([]entity.AIInventoryStatus, error) {
	var results []entity.AIInventoryStatus
	query := `
		SELECT 
			mi_sku as sku,
			title as name,
			current_stock as stock,
			COALESCE(specification, '') as specification,
			CASE 
				WHEN current_stock <= 0 THEN 'Out of Stock'
				WHEN current_stock <= 10 THEN 'Low Stock'
				ELSE 'In Stock'
			END as status
		FROM inventory_items
		ORDER BY current_stock ASC
	`
	if err := r.guard.IsSafe(query); err != nil {
		return nil, err
	}
	err := r.db.Raw(query).Scan(&results).Error
	return results, err
}

func (r *gormAIReadRepository) GetBusinessSnapshot() (entity.AIBusinessSnapshot, error) {
	var snap entity.AIBusinessSnapshot

	// MTD
	r.db.Raw(`
		SELECT COALESCE(SUM(total_price), 0), COUNT(*) 
		FROM orders 
		WHERE NOT (LOWER(COALESCE(status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(fulfillment_status, '')) IN ('cancelled', 'canceled')) 
		AND created_at >= date_trunc('month', current_date)
	`).Row().Scan(&snap.MTDRevenue, &snap.MTDOrders)

	// Today
	r.db.Raw(`
		SELECT COALESCE(SUM(total_price), 0), COUNT(*) 
		FROM orders 
		WHERE NOT (LOWER(COALESCE(status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(fulfillment_status, '')) IN ('cancelled', 'canceled')) 
		AND created_at >= current_date
	`).Row().Scan(&snap.TodayRevenue, &snap.TodayOrders)

	// Low Stock
	r.db.Raw(`SELECT COUNT(*) FROM inventory_items WHERE current_stock <= 10`).Scan(&snap.LowStockCount)

	// Pending (Unfulfilled)
	r.db.Raw(`SELECT COUNT(*) FROM orders WHERE NOT (LOWER(COALESCE(status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(fulfillment_status, '')) IN ('cancelled', 'canceled')) AND fulfillment_status != 'fulfilled'`).Scan(&snap.PendingOrders)

	return snap, nil
}

func (r *gormAIReadRepository) ExecuteRawQuery(sql string) ([]map[string]interface{}, error) {
	if err := r.guard.IsSafe(sql); err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	err := r.db.Raw(sql).Scan(&results).Error
	return results, err
}

func (r *gormAIReadRepository) ListTables() ([]string, error) {
	var tables []string
	query := "SELECT table_name FROM information_schema.tables WHERE table_schema='public'"
	err := r.db.Raw(query).Scan(&tables).Error
	return tables, err
}

func (r *gormAIReadRepository) DescribeTable(tableName string) ([]map[string]interface{}, error) {
	var columns []map[string]interface{}
	query := "SELECT column_name, data_type, is_nullable FROM information_schema.columns WHERE table_name = ?"
	err := r.db.Raw(query, tableName).Scan(&columns).Error
	return columns, err
}

func normalizeDates(start, end string) (string, string) {
	if len(start) == 10 {
		start = start + " 00:00:00"
	}
	if len(end) == 10 {
		end = end + " 23:59:59"
	}
	return start, end
}

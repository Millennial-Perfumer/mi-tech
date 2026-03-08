package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

type MetricsHandler struct {
	db *sql.DB
}

func NewMetricsHandler(db *sql.DB) *MetricsHandler {
	return &MetricsHandler{
		db: db,
	}
}

func (h *MetricsHandler) GetDashboardMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	var startDate, endDate time.Time
	var err error

	if startDateStr != "" {
		startDate, err = time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			http.Error(w, "Invalid start_date format", http.StatusBadRequest)
			return
		}
	} else {
		// Default to MTD (Month to Date) if not provided
		now := time.Now()
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	}

	if endDateStr != "" {
		endDate, err = time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			http.Error(w, "Invalid end_date format", http.StatusBadRequest)
			return
		}
	} else {
		endDate = time.Now()
	}

	query := `
		SELECT 
			COALESCE(SUM(total_price), 0) as total_revenue,
			COUNT(id) as total_orders,
			COUNT(id) FILTER (WHERE LOWER(status) = 'cancelled') as cancelled_orders,
			COUNT(id) FILTER (WHERE LOWER(status) = 'fulfilled') as fulfilled_orders,
			COUNT(id) FILTER (WHERE LOWER(status) = 'unfulfilled') as unfulfilled_orders
		FROM shopify_orders 
		WHERE created_at >= $1 AND created_at <= $2
	`

	var totalRevenue float64
	var totalOrders, cancelledOrders, fulfilledOrders, unfulfilledOrders int

	err = h.db.QueryRow(query, startDate, endDate).Scan(
		&totalRevenue,
		&totalOrders,
		&cancelledOrders,
		&fulfilledOrders,
		&unfulfilledOrders,
	)
	if err != nil {
		http.Error(w, "Failed to calculate metrics", http.StatusInternalServerError)
		return
	}

	// GST Calculation (Assuming uniform 18% for now)
	// Price = Base + GST
	// GST = Total - (Total / 1.18)
	totalGST := totalRevenue - (totalRevenue / 1.18)

	// Assuming 50/50 split between CGST and SGST for local state sales,
	// and IGST for inter-state. For this simplified dashboard, we'll
	// distribute it 40% IGST, 30% CGST, 30% SGST as an example placeholder
	// since we don't have address states in the DB yet.
	igst := totalGST * 0.40
	cgst := totalGST * 0.30
	sgst := totalGST * 0.30

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"metrics": map[string]interface{}{
			"total_revenue":       totalRevenue,
			"total_invoices":      totalOrders, // 1 invoice per order for now
			"total_gst_collected": totalGST,
			"cgst_collected":      cgst,
			"sgst_collected":      sgst,
			"igst_collected":      igst,
			"total_orders":        totalOrders,
			"cancelled_orders":    cancelledOrders,
			"fulfilled_orders":    fulfilledOrders,
			"unfulfilled_orders":  unfulfilledOrders,
		},
	})
}

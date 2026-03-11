package handler

import (
	"encoding/json"
	"net/http"

	"shopify-gst-app/internal/service"
)

// MetricsHandler is a thin HTTP adapter for dashboard metrics.
type MetricsHandler struct {
	metricsService *service.MetricsService
}

// NewMetricsHandler creates a new MetricsHandler.
func NewMetricsHandler(metricsService *service.MetricsService) *MetricsHandler {
	return &MetricsHandler{metricsService: metricsService}
}

// GetDashboardMetrics handles GET /api/dashboard/metrics.
func (h *MetricsHandler) GetDashboardMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	metrics, err := h.metricsService.GetDashboardMetrics(startDate, endDate)
	if err != nil {
		http.Error(w, "Failed to calculate metrics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"metrics": metrics,
	})
}

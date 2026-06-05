package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"mi-tech/internal/service"
)

// MetricsHandler is a thin HTTP adapter for dashboard metrics.
type MetricsHandler struct {
	metricsService *service.MetricsService
}

// NewMetricsHandler creates a new MetricsHandler.
func NewMetricsHandler(metricsService *service.MetricsService) *MetricsHandler {
	return &MetricsHandler{metricsService: metricsService}
}

func parseSourceIDs(r *http.Request) []string {
	sourcesStr := r.URL.Query().Get("source_ids")
	if sourcesStr == "" {
		return nil
	}
	return strings.Split(sourcesStr, ",")
}

// GetDashboardMetrics handles GET /api/metrics/dashboard.
func (h *MetricsHandler) GetDashboardMetrics(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	sourceIDs := parseSourceIDs(r)

	metrics, err := h.metricsService.GetDashboardMetrics(startDate, endDate, sourceIDs)
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

// GetTopProducts handles GET /api/metrics/top-products.
func (h *MetricsHandler) GetTopProducts(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	sourceIDs := parseSourceIDs(r)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 5
	}

	products, err := h.metricsService.GetTopProducts(startDate, endDate, sourceIDs, limit)
	if err != nil {
		http.Error(w, "Failed to fetch top products", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"products": products,
	})
}

// GetRevenueTrend handles GET /api/metrics/revenue-trend.
func (h *MetricsHandler) GetRevenueTrend(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	sourceIDs := parseSourceIDs(r)

	trend, err := h.metricsService.GetRevenueTrend(startDate, endDate, sourceIDs)
	if err != nil {
		http.Error(w, "Failed to fetch revenue trend", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"trend":   trend,
	})
}

// GetGeoDistribution handles GET /api/metrics/geo-distribution.
func (h *MetricsHandler) GetGeoDistribution(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	sourceIDs := parseSourceIDs(r)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 5
	}

	geo, err := h.metricsService.GetGeoDistribution(startDate, endDate, sourceIDs, limit)
	if err != nil {
		http.Error(w, "Failed to fetch geo distribution", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"distribution": geo,
	})
}

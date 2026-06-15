package repository

import (
	"mi-tech/internal/domain/dashboard/dto"
)

// MetricsRepository defines data access for dashboard metric queries.
type MetricsRepository interface {
	GetDashboardMetrics(startDate, endDate string, sourceIDs []string) (dto.DashboardMetrics, error)
	GetTopProducts(startDate, endDate string, sourceIDs []string, limit int) ([]dto.TopProductRow, error)
	GetRevenueTrend(startDate, endDate string, sourceIDs []string) ([]dto.RevenueTrendRow, error)
	GetGeoDistribution(startDate, endDate string, sourceIDs []string, limit int) ([]dto.GeoDistributionRow, error)
}

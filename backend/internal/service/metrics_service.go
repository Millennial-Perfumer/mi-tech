package service

import (
	"mi-tech/internal/dto"
	"mi-tech/internal/repository"
)

// MetricsService handles dashboard metrics business logic.
type MetricsService struct {
	metricsRepo repository.MetricsRepository
}

// NewMetricsService creates a new MetricsService.
func NewMetricsService(metricsRepo repository.MetricsRepository) *MetricsService {
	return &MetricsService{metricsRepo: metricsRepo}
}

// GetDashboardMetrics calculates and returns all dashboard metrics including GST splits.
func (s *MetricsService) GetDashboardMetrics(startDate, endDate string, sourceIDs []string) (dto.DashboardMetrics, error) {
	return s.metricsRepo.GetDashboardMetrics(startDate, endDate, sourceIDs)
}

// GetTopProducts returns the top selling products.
func (s *MetricsService) GetTopProducts(startDate, endDate string, sourceIDs []string, limit int) ([]dto.TopProductRow, error) {
	return s.metricsRepo.GetTopProducts(startDate, endDate, sourceIDs, limit)
}

// GetRevenueTrend returns the revenue trend over time.
func (s *MetricsService) GetRevenueTrend(startDate, endDate string, sourceIDs []string) ([]dto.RevenueTrendRow, error) {
	return s.metricsRepo.GetRevenueTrend(startDate, endDate, sourceIDs)
}

// GetGeoDistribution returns the geographic distribution of orders.
func (s *MetricsService) GetGeoDistribution(startDate, endDate string, sourceIDs []string, limit int) ([]dto.GeoDistributionRow, error) {
	return s.metricsRepo.GetGeoDistribution(startDate, endDate, sourceIDs, limit)
}

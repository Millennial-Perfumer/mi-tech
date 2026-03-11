package service

import (
	"shopify-gst-app/internal/dto"
	"shopify-gst-app/internal/repository"
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
func (s *MetricsService) GetDashboardMetrics(startDate, endDate string) (dto.DashboardMetrics, error) {
	totalRevenue, cgst, sgst, igst, totalOrders, cancelledOrders, fulfilledOrders, unfulfilledOrders, err :=
		s.metricsRepo.GetDashboardMetrics(startDate, endDate)
	if err != nil {
		return dto.DashboardMetrics{}, err
	}

	return dto.DashboardMetrics{
		TotalRevenue:      totalRevenue,
		TotalInvoices:     totalOrders,
		TotalGSTCollected: cgst + sgst + igst,
		CGSTCollected:     cgst,
		SGSTCollected:     sgst,
		IGSTCollected:     igst,
		TotalOrders:       totalOrders,
		CancelledOrders:   cancelledOrders,
		FulfilledOrders:   fulfilledOrders,
		UnfulfilledOrders: unfulfilledOrders,
	}, nil
}

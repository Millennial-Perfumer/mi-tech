package test

import (
	"testing"

	"mi-tech/internal/domain/dashboard/dto"
	"mi-tech/internal/domain/dashboard/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockMetricsRepository struct {
	mock.Mock
}

func (m *mockMetricsRepository) GetDashboardMetrics(startDate, endDate string, sourceIDs []string) (dto.DashboardMetrics, error) {
	args := m.Called(startDate, endDate, sourceIDs)
	return args.Get(0).(dto.DashboardMetrics), args.Error(1)
}

func (m *mockMetricsRepository) GetTopProducts(startDate, endDate string, sourceIDs []string, limit int) ([]dto.TopProductRow, error) {
	args := m.Called(startDate, endDate, sourceIDs, limit)
	return args.Get(0).([]dto.TopProductRow), args.Error(1)
}

func (m *mockMetricsRepository) GetRevenueTrend(startDate, endDate string, sourceIDs []string) ([]dto.RevenueTrendRow, error) {
	args := m.Called(startDate, endDate, sourceIDs)
	return args.Get(0).([]dto.RevenueTrendRow), args.Error(1)
}

func (m *mockMetricsRepository) GetGeoDistribution(startDate, endDate string, sourceIDs []string, limit int) ([]dto.GeoDistributionRow, error) {
	args := m.Called(startDate, endDate, sourceIDs, limit)
	return args.Get(0).([]dto.GeoDistributionRow), args.Error(1)
}

func TestMetricsService_GetDashboardMetrics(t *testing.T) {
	mockRepo := new(mockMetricsRepository)
	srv := service.NewMetricsService(mockRepo)

	expectedMetrics := dto.DashboardMetrics{
		TotalRevenue:      1000.0,
		TotalInvoices:     10,
		TotalGSTCollected: 180.0,
	}

	mockRepo.On("GetDashboardMetrics", "2023-01-01", "2023-01-31", []string{}).
		Return(expectedMetrics, nil)

	metrics, err := srv.GetDashboardMetrics("2023-01-01", "2023-01-31", []string{})

	assert.NoError(t, err)
	assert.Equal(t, 1000.0, metrics.TotalRevenue)
	assert.Equal(t, 10, metrics.TotalInvoices)
	mockRepo.AssertExpectations(t)
}

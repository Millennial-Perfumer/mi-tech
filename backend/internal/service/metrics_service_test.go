package service

import (
	"testing"

	"mi-tech/internal/dto"
	"mi-tech/internal/service/mocks"

	"github.com/stretchr/testify/assert"
)

func TestMetricsService_GetDashboardMetrics(t *testing.T) {
	mockRepo := new(mocks.MockMetricsRepository)
	service := NewMetricsService(mockRepo)

	expectedMetrics := dto.DashboardMetrics{
		TotalRevenue:      1000.0,
		TotalInvoices:     10,
		TotalGSTCollected: 180.0,
	}

	mockRepo.On("GetDashboardMetrics", "2023-01-01", "2023-01-31", []string{}).
		Return(expectedMetrics, nil)

	metrics, err := service.GetDashboardMetrics("2023-01-01", "2023-01-31", []string{})

	assert.NoError(t, err)
	assert.Equal(t, 1000.0, metrics.TotalRevenue)
	assert.Equal(t, 10, metrics.TotalInvoices)
	mockRepo.AssertExpectations(t)
}

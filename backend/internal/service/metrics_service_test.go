package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockMetricsRepository struct {
	mock.Mock
}

func (m *MockMetricsRepository) GetDashboardMetrics(startDate, endDate string) (float64, float64, float64, float64, int, int, int, int, error) {
	args := m.Called(startDate, endDate)
	return args.Get(0).(float64), args.Get(1).(float64), args.Get(2).(float64), args.Get(3).(float64), args.Int(4), args.Int(5), args.Int(6), args.Int(7), args.Error(8)
}

func TestMetricsService_GetDashboardMetrics(t *testing.T) {
	mockRepo := new(MockMetricsRepository)
	service := NewMetricsService(mockRepo)

	mockRepo.On("GetDashboardMetrics", "2023-01-01", "2023-01-31").
		Return(1000.0, 90.0, 90.0, 0.0, 10, 1, 5, 4, nil)

	metrics, err := service.GetDashboardMetrics("2023-01-01", "2023-01-31")

	assert.NoError(t, err)
	assert.Equal(t, 1000.0, metrics.TotalRevenue)
	assert.Equal(t, 10, metrics.TotalOrders)
	mockRepo.AssertExpectations(t)
}

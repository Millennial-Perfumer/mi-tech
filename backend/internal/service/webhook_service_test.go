package service

import (
	"testing"

	"mi-tech/internal/entity"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockWebhookEventRepository struct {
	mock.Mock
}

func (m *MockWebhookEventRepository) Save(event *entity.WebhookEvent) error {
	args := m.Called(event)
	return args.Error(0)
}

func (m *MockWebhookEventRepository) IsProcessed(deliveryID string) (bool, error) {
	args := m.Called(deliveryID)
	return args.Bool(0), args.Error(1)
}

func (m *MockWebhookEventRepository) LinkToOrder(deliveryID string, orderID int64) error {
	args := m.Called(deliveryID, orderID)
	return args.Error(0)
}

func TestWebhookService_IsProcessed(t *testing.T) {
	mockWebhookRepo := new(MockWebhookEventRepository)
	service := &WebhookService{
		webhookEventRepo: mockWebhookRepo,
	}

	mockWebhookRepo.On("IsProcessed", "del_1").Return(true, nil)
	mockWebhookRepo.On("IsProcessed", "del_2").Return(false, nil)

	p1, _ := service.IsProcessed("del_1")
	p2, _ := service.IsProcessed("del_2")

	assert.True(t, p1)
	assert.False(t, p2)
}

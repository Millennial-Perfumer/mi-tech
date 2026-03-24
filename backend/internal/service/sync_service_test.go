package service

import (
	"testing"

	"mi-tech/internal/service/mocks"

	"github.com/stretchr/testify/assert"
)

func TestSyncService_Truncate(t *testing.T) {
	mockOrderRepo := new(mocks.MockOrderRepository)
	service := NewSyncService(nil, mockOrderRepo, nil)

	mockOrderRepo.On("TruncateAll").Return(nil)

	err := service.orderRepo.TruncateAll()
	assert.NoError(t, err)
	mockOrderRepo.AssertExpectations(t)
}

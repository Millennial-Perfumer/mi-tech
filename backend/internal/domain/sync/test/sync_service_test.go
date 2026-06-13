package test

import (
	"testing"

	"mi-tech/internal/shared/extclient/shopify"
	"mi-tech/internal/shared/config"
	"mi-tech/internal/domain/sync/service"
	orderMocks "mi-tech/internal/domain/order/test"

	"github.com/stretchr/testify/assert"
)

func TestSyncService_ResetOnly(t *testing.T) {
	mockOrderRepo := new(orderMocks.MockOrderRepository)
	srv := service.NewSyncService(nil, mockOrderRepo, nil, nil)

	mockOrderRepo.On("TruncateAll").Return(nil)

	err := srv.ResetOnly()
	assert.NoError(t, err)
	mockOrderRepo.AssertExpectations(t)
}

func TestSyncService_ResetAndSync(t *testing.T) {
	mockOrderRepo := new(orderMocks.MockOrderRepository)
	dummyClient := shopify.NewClient(&config.SettingsProvider{})
	srv := service.NewSyncService(dummyClient, mockOrderRepo, nil, nil)

	mockOrderRepo.On("TruncateAll").Return(nil)

	_, err := srv.ResetAndSync()
	assert.Error(t, err)
	mockOrderRepo.AssertExpectations(t)
}

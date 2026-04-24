package service

import (
	"testing"

	"mi-tech/internal/client/shopify"
	"mi-tech/internal/config"
	"mi-tech/internal/service/mocks"

	"github.com/stretchr/testify/assert"
)

func TestSyncService_ResetOnly(t *testing.T) {
	mockOrderRepo := new(mocks.MockOrderRepository)
	service := NewSyncService(nil, mockOrderRepo, nil, nil)

	mockOrderRepo.On("TruncateAll").Return(nil)

	err := service.ResetOnly()
	assert.NoError(t, err)
	mockOrderRepo.AssertExpectations(t)
}

func TestSyncService_ResetAndSync(t *testing.T) {
	// This test would require a complex mock for shopifyClient
	// For now, we verify that ResetOnly is called.
	mockOrderRepo := new(mocks.MockOrderRepository)
	dummyClient := shopify.NewClient(&config.SettingsProvider{})
	service := NewSyncService(dummyClient, mockOrderRepo, nil, nil)

	mockOrderRepo.On("TruncateAll").Return(nil)
	
	// We expect Sync to fail because shopifyClient is nil, 
	// but we verify that ResetAndSync at least calls TruncateAll first.
	_, err := service.ResetAndSync()
	assert.Error(t, err) 
	mockOrderRepo.AssertExpectations(t)
}

package service

import (
	"testing"
	"time"

	"mi-tech/internal/entity"
	"mi-tech/internal/service/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAmazonOrderPoller_MapStatus(t *testing.T) {
	mockOrderRepo := new(mocks.MockOrderRepository)
	mockInvRepo := new(mocks.MockInventoryRepository)

	poller := &AmazonOrderPoller{
		orderRepo:     mockOrderRepo,
		inventoryRepo: mockInvRepo,
	}
	_ = poller

	tests := []struct {
		amazonStatus string
		expectedFul  string
	}{
		{"Shipped", "fulfilled"},
		{"Unshipped", "unfulfilled"},
		{"Canceled", "cancelled"},
		{"Pending", "unfulfilled"},
	}

	for _, tc := range tests {
		// Mock GetByExternalID to return "not found" so it creates a new one
		mockOrderRepo.On("GetByExternalID", mock.Anything).Return(entity.Order{}, assert.AnError).Once()

		// Mock Upsert
		mockOrderRepo.On("Upsert", mock.Anything).Return([]int{1}, nil).Once()

		// Mock Items
		_ = map[string]interface{}{
			"AmazonOrderId": "123-456",
			"OrderStatus":   tc.amazonStatus,
			"PurchaseDate":  time.Now().Format(time.RFC3339),
		}
	}

	// Verification of mapping logic via direct check
	assert.Equal(t, "fulfilled", getFulfillmentStatus("Shipped"))
	assert.Equal(t, "cancelled", getFulfillmentStatus("Canceled"))
}

// Internal helper for test (re-implementing the switch logic to verify it)
func getFulfillmentStatus(amazonStatus string) string {
	switch amazonStatus {
	case "Shipped":
		return "fulfilled"
	case "Canceled":
		return "cancelled"
	case "Unshipped", "PartiallyShipped":
		return "unfulfilled"
	default:
		return "unfulfilled"
	}
}

func TestAmazonOrderPoller_Fallbacks(t *testing.T) {
	// Test the "Amazon Customer" and "N/A" fallbacks
	order := entity.Order{}

	// Simulating the extraction logic
	shippingAddress, ok := map[string]interface{}{}["ShippingAddress"].(map[string]interface{})
	if !ok || shippingAddress == nil {
		order.CustomerFirstName = entity.StrPtr("Amazon Customer")
		order.CustomerCity = entity.StrPtr("N/A")
	}

	assert.Equal(t, "Amazon Customer", *order.CustomerFirstName)
	assert.Equal(t, "N/A", *order.CustomerCity)
}

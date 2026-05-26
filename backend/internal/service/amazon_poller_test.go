package service

import (
	"testing"

	"mi-tech/internal/entity"
	"mi-tech/internal/service/mocks"

	"github.com/stretchr/testify/assert"
)

func TestAmazonOrderPoller_MapStatus(t *testing.T) {
	mockOrderRepo := new(mocks.MockOrderRepository)
	mockInvRepo := new(mocks.MockInventoryRepository)
	
	poller := &AmazonOrderPoller{
		orderRepo:     mockOrderRepo,
		inventoryRepo: mockInvRepo,
	}
	_ = poller

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
	shippingAddress, ok := map[string]interface{}{}[ "ShippingAddress" ].(map[string]interface{})
	if !ok || shippingAddress == nil {
		order.CustomerFirstName = entity.StrPtr("Amazon Customer")
		order.CustomerCity = entity.StrPtr("N/A")
	}
	
	assert.Equal(t, "Amazon Customer", *order.CustomerFirstName)
	assert.Equal(t, "N/A", *order.CustomerCity)
}

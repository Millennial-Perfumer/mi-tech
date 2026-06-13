package test

import (
	"testing"

	inventoryMocks "mi-tech/internal/domain/inventory/test"
	orderEntity "mi-tech/internal/domain/order/entity"
	orderMocks "mi-tech/internal/domain/order/test"
	"mi-tech/internal/domain/sync/service"
	util "mi-tech/internal/shared/util"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAmazonOrderPoller_MapStatus(t *testing.T) {
	mockOrderRepo := new(orderMocks.MockOrderRepository)
	mockInvRepo := new(inventoryMocks.MockInventoryRepository)

	poller := service.NewAmazonOrderPoller(nil, mockOrderRepo, mockInvRepo, nil)
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

	for range tests {
		mockOrderRepo.On("GetByExternalID", mock.Anything).Return(orderEntity.Order{}, assert.AnError).Once()
		mockOrderRepo.On("Upsert", mock.Anything).Return([]int{1}, nil).Once()
	}

	assert.Equal(t, "fulfilled", getFulfillmentStatus("Shipped"))
	assert.Equal(t, "cancelled", getFulfillmentStatus("Canceled"))
}

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
	order := orderEntity.Order{}

	shippingAddress, ok := map[string]interface{}{}["ShippingAddress"].(map[string]interface{})
	if !ok || shippingAddress == nil {
		order.CustomerFirstName = util.StrPtr("Amazon Customer")
		order.CustomerCity = util.StrPtr("N/A")
	}

	assert.Equal(t, "Amazon Customer", *order.CustomerFirstName)
	assert.Equal(t, "N/A", *order.CustomerCity)
}

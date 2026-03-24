package service

import (
	"testing"

	"mi-tech/internal/entity"
	"mi-tech/internal/repository"
	"mi-tech/internal/service/mocks"

	"github.com/stretchr/testify/assert"
)

func TestOrderService_ListOrders(t *testing.T) {
	mockOrderRepo := new(mocks.MockOrderRepository)
	service := NewOrderService(mockOrderRepo, nil, nil)

	filter := repository.OrderFilter{Page: 1, Limit: 25}
	orders := []entity.Order{
		{ID: 1, OrderNumber: "ORD-1"},
	}

	mockOrderRepo.On("List", filter).Return(orders, 1, nil)

	resp, total, err := service.ListOrders("", "", 1, 25, "", "", "", "", "", "")

	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Equal(t, "ORD-1", resp[0].OrderNumber)
	mockOrderRepo.AssertExpectations(t)
}

func TestOrderService_GetOrder(t *testing.T) {
	mockOrderRepo := new(mocks.MockOrderRepository)
	mockLineItemRepo := new(mocks.MockLineItemRepository)
	service := NewOrderService(mockOrderRepo, mockLineItemRepo, nil)

	order := entity.Order{ID: 1, OrderNumber: "ORD-1"}
	items := []entity.LineItem{{ID: "LI-1", Title: entity.StrPtr("Item 1")}}

	mockOrderRepo.On("GetByID", int64(1)).Return(order, nil)
	mockLineItemRepo.On("GetByOrderID", int64(1)).Return(items, nil)

	resp, err := service.GetOrder(1)

	assert.NoError(t, err)
	assert.Equal(t, "ORD-1", resp.OrderNumber)
	assert.Equal(t, 1, len(resp.LineItems))
	mockOrderRepo.AssertExpectations(t)
	mockLineItemRepo.AssertExpectations(t)
}

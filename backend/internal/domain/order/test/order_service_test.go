package test

import (
	"context"
	"testing"

	"mi-tech/internal/domain/order/dto"
	"mi-tech/internal/domain/order/entity"
	"mi-tech/internal/domain/order/repository"
	"mi-tech/internal/domain/order/service"
	"mi-tech/internal/shared/util"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOrderService_ListOrders(t *testing.T) {
	mockOrderRepo := new(MockOrderRepository)
	svc := service.NewOrderService(mockOrderRepo, nil, nil, nil, nil)

	filter := repository.OrderFilter{Page: 1, Limit: 25}
	orders := []entity.Order{
		{ID: 1, OrderNumber: "ORD-1"},
	}

	mockOrderRepo.On("List", filter).Return(orders, 1, nil)

	resp, total, err := svc.ListOrders("", "", 1, 25, "", "", "", "", "", "", "")

	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Equal(t, "ORD-1", resp[0].OrderNumber)
	mockOrderRepo.AssertExpectations(t)
}

func TestOrderService_GetOrder(t *testing.T) {
	mockOrderRepo := new(MockOrderRepository)
	mockLineItemRepo := new(MockLineItemRepository)
	svc := service.NewOrderService(mockOrderRepo, mockLineItemRepo, nil, nil, nil)

	order := entity.Order{ID: 1, OrderNumber: "ORD-1"}
	items := []entity.LineItem{{ID: "LI-1", Title: util.StrPtr("Item 1")}}

	mockOrderRepo.On("GetByID", int64(1)).Return(order, nil)
	mockLineItemRepo.On("GetByOrderID", int64(1)).Return(items, nil)

	resp, err := svc.GetOrder(1)

	assert.NoError(t, err)
	assert.Equal(t, "ORD-1", resp.OrderNumber)
	assert.Equal(t, 1, len(resp.LineItems))
	mockOrderRepo.AssertExpectations(t)
	mockLineItemRepo.AssertExpectations(t)
}

func TestOrderService_CreateManualOrder(t *testing.T) {
	mockOrderRepo := new(MockOrderRepository)
	mockLineItemRepo := new(MockLineItemRepository)
	svc := service.NewOrderService(mockOrderRepo, mockLineItemRepo, nil, nil, nil)

	req := dto.OrderCreateRequest{
		CustomerName:  "Aamir Siddiqui",
		CustomerPhone: "9876543210",
		TotalPrice:    118.0, // 18% inclusive GST: GST = 18, Subtotal = 100
		LineItems: []dto.LineItemCreateRequest{
			{
				MISKU:    "mi-01",
				Title:    "Oud Perfume",
				Quantity: 2,
				Price:    59.0,
			},
		},
	}

	mockOrderRepo.On("GetNextPOSSequence", "POS1").Return("POS1-001", nil)

	mockOrderRepo.On("Upsert", mock.MatchedBy(func(o entity.Order) bool {
		return o.OrderNumber == "POS1-001" && o.TotalPrice == 118.0 && *o.SubtotalPrice == 100.0 && *o.TotalTax == 18.0
	})).Return([]int{5}, nil)

	mockOrderRepo.On("GetByExternalID", "pos-POS1-001").Return(entity.Order{
		ID:          42,
		OrderNumber: "POS1-001",
		TotalPrice:  118.0,
	}, nil)

	mockLineItemRepo.On("GetByOrderID", int64(42)).Return([]entity.LineItem{
		{
			ID:       "pos-POS1-001-0",
			OrderID:  42,
			SKU:      util.StrPtr("mi-01"),
			Quantity: 2,
			Price:    59.0,
		},
	}, nil)

	resp, err := svc.CreateManualOrder(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, "POS1-001", resp.OrderNumber)
	assert.Equal(t, "118.00", resp.TotalPrice)
	assert.Equal(t, 1, len(resp.LineItems))
	assert.Equal(t, "mi-01", resp.LineItems[0].SKU)
	mockOrderRepo.AssertExpectations(t)
	mockLineItemRepo.AssertExpectations(t)
}

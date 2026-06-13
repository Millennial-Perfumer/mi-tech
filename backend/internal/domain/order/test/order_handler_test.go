package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"mi-tech/internal/domain/order/entity"
	"mi-tech/internal/domain/order/handler"
	"mi-tech/internal/domain/order/repository"
	"mi-tech/internal/domain/order/service"

	"github.com/stretchr/testify/assert"
)

func TestOrderHandler_GetOrders_Success(t *testing.T) {
	mockOrderRepo := new(MockOrderRepository)
	orderService := service.NewOrderService(mockOrderRepo, nil, nil, nil, nil)
	h := handler.NewOrderHandler(orderService, nil, nil)

	filter := repository.OrderFilter{Page: 1, Limit: 25}
	orders := []entity.Order{{ID: 1, OrderNumber: "ORD-1"}}
	mockOrderRepo.On("List", filter).Return(orders, 1, nil)

	req := httptest.NewRequest("GET", "/api/orders?page=1&limit=25", nil)
	w := httptest.NewRecorder()

	h.GetOrders(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, true, resp["success"])
	assert.Equal(t, float64(1), resp["total_count"])
}

func TestOrderHandler_GetSources(t *testing.T) {
	mockOrderRepo := new(MockOrderRepository)
	orderService := service.NewOrderService(mockOrderRepo, nil, nil, nil, nil)
	h := handler.NewOrderHandler(orderService, nil, nil)

	sources := []entity.Source{{ID: "s1", Name: "Source 1"}}
	mockOrderRepo.On("ListSources").Return(sources, nil)

	req := httptest.NewRequest("GET", "/api/sources", nil)
	w := httptest.NewRecorder()

	h.GetSources(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, true, resp["success"])
	assert.Len(t, resp["sources"], 1)
}

func TestOrderHandler_UpdatePaymentStatus(t *testing.T) {
	mockOrderRepo := new(MockOrderRepository)
	orderService := service.NewOrderService(mockOrderRepo, nil, nil, nil, nil)
	h := handler.NewOrderHandler(orderService, nil, nil)

	mockOrderRepo.On("UpdateFinancialStatus", int64(123), "paid").Return(nil)

	body, _ := json.Marshal(map[string]string{"status": "paid"})
	req := httptest.NewRequest("PUT", "/api/orders/payment-status?id=123", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	h.UpdatePaymentStatus(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, true, resp["success"])
	assert.Equal(t, "Payment status updated successfully", resp["message"])
}

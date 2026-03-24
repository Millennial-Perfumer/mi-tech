package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"mi-tech/internal/entity"
	"mi-tech/internal/repository"
	"mi-tech/internal/service"
	"mi-tech/internal/service/mocks"

	"github.com/stretchr/testify/assert"
)

func TestOrderHandler_GetOrders_Success(t *testing.T) {
	mockOrderRepo := new(mocks.MockOrderRepository)
	orderService := service.NewOrderService(mockOrderRepo, nil, nil)
	handler := NewOrderHandler(orderService, nil)

	filter := repository.OrderFilter{Page: 1, Limit: 25}
	orders := []entity.Order{{ID: 1, OrderNumber: "ORD-1"}}
	mockOrderRepo.On("List", filter).Return(orders, 1, nil)

	req := httptest.NewRequest("GET", "/api/orders?page=1&limit=25", nil)
	w := httptest.NewRecorder()

	handler.GetOrders(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, true, resp["success"])
	assert.Equal(t, float64(1), resp["total_count"])
}

func TestOrderHandler_GetSources(t *testing.T) {
	mockOrderRepo := new(mocks.MockOrderRepository)
	orderService := service.NewOrderService(mockOrderRepo, nil, nil)
	handler := NewOrderHandler(orderService, nil)

	sources := []entity.Source{{ID: "s1", Name: "Source 1"}}
	mockOrderRepo.On("ListSources").Return(sources, nil)

	req := httptest.NewRequest("GET", "/api/sources", nil)
	w := httptest.NewRecorder()

	handler.GetSources(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, true, resp["success"])
	assert.Len(t, resp["sources"], 1)
}

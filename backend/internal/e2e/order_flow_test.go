package e2e

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"mi-tech/internal/handler"
	"mi-tech/internal/repository"
	"mi-tech/internal/service"
	"mi-tech/internal/testutil"

	"github.com/stretchr/testify/assert"
)

func TestEndToEnd_LoginAndListOrders(t *testing.T) {
	db, err := testutil.SetupTestDB()
	if err != nil {
		t.Skip("E2E skipping: Database not available")
	}
	defer testutil.CleanupTestDB(db)

	orderRepo := repository.NewOrderRepository(db)
	lineItemRepo := repository.NewLineItemRepository(db)
	customerRepo := repository.NewCustomerRepository(db)
	customerService := service.NewCustomerService(customerRepo, orderRepo, nil)
	orderService := service.NewOrderService(orderRepo, lineItemRepo, customerService)
	orderHandler := handler.NewOrderHandler(orderService, nil)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/orders", orderHandler.GetOrders)

	// 1. Check initially empty
	req := httptest.NewRequest("GET", "/api/orders", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["total_count"])

	// 2. Seed an order directly in DB
	db.Exec("INSERT INTO orders (external_order_id, order_number, source_id, created_at, total_price) VALUES (?, ?, ?, NOW(), ?)",
		"e2e_1", "ORD-E2E", "shopify", 150.0)

	// 3. Check again
	req = httptest.NewRequest("GET", "/api/orders", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(1), resp["total_count"])
	orders := resp["orders"].([]interface{})
	assert.Equal(t, "ORD-E2E", orders[0].(map[string]interface{})["order_number"])
}

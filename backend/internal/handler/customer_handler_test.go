package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"mi-tech/internal/repository"
	"mi-tech/internal/service"
	"mi-tech/internal/testutil"

	"github.com/stretchr/testify/assert"
)

func TestCustomerHandler_ListCustomers(t *testing.T) {
	db, err := testutil.SetupTestDB()
	if err != nil {
		t.Skip("DB not available")
	}
	defer testutil.CleanupTestDB(db)

	customerRepo := repository.NewCustomerRepository(db)
	customerService := service.NewCustomerService(customerRepo, nil, nil)
	handler := NewCustomerHandler(customerService)

	// Seed one (and clear previous to ensure total=1)
	db.Exec("DELETE FROM customers")
	db.Exec("INSERT INTO customers (phone_number, first_name) VALUES (?, ?)", "+91123", "Alice")

	req := httptest.NewRequest("GET", "/api/customers?page=1&pageSize=10", nil)
	w := httptest.NewRecorder()

	handler.ListCustomers(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, true, resp["success"])
	assert.Equal(t, float64(1), resp["total"])
}

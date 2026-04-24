package service

import (
	"context"
	"testing"

	"mi-tech/internal/entity"
	"mi-tech/internal/repository"
	"mi-tech/internal/testutil"

	"github.com/stretchr/testify/assert"
)

func TestCustomerService_NormalizePhone(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"9876543210", "+919876543210"},
		{"919876543210", "+919876543210"},
		{"+919876543210", "+919876543210"},
		{"  9876543210  ", "+919876543210"},
		{"987-654-3210", "+919876543210"},
	}

	for _, tc := range tests {
		assert.Equal(t, tc.expected, entity.NormalizePhone(tc.input))
	}
}

func TestCustomerService_ToTitleCase(t *testing.T) {
	service := NewCustomerService(nil, nil, nil)
	assert.Equal(t, "John Doe", service.toTitleCase("JOHN DOE"))
	assert.Equal(t, "Alice Smith", service.toTitleCase("alice smith"))
	assert.Equal(t, "Bob", service.toTitleCase("bOB"))
}

func TestCustomerService_UpdateFromOrder(t *testing.T) {
	db, err := testutil.SetupTestDB()
	if err != nil {
		t.Skip("DB not available")
	}
	defer testutil.CleanupTestDB(db)

	repo := repository.NewCustomerRepository(db)
	service := NewCustomerService(repo, nil, nil)

	phone := "9876543210"
	order := &entity.Order{
		CustomerPhone:     entity.StrPtr(phone),
		CustomerFirstName: entity.StrPtr("JOHN"),
		CustomerLastName:  entity.StrPtr("DOE"),
		CustomerCity:      entity.StrPtr("Delhi"),
		TotalPrice:        100.50,
	}

	err = service.UpdateFromOrder(context.Background(), order)
	assert.NoError(t, err)

	fetched, _ := repo.GetByPhone(context.Background(), "+919876543210")
	assert.NotNil(t, fetched)
	assert.Equal(t, "John", *fetched.FirstName)
	assert.Equal(t, "Doe", *fetched.LastName)
}
func TestCustomerService_UpdateCustomer_PreserveStats(t *testing.T) {
	db, err := testutil.SetupTestDB()
	if err != nil {
		t.Skip("DB not available")
	}
	defer testutil.CleanupTestDB(db)

	repo := repository.NewCustomerRepository(db)
	service := NewCustomerService(repo, nil, nil)

	// 1. Create a customer with stats
	phone := "+919998887770"
	cust := &entity.Customer{
		PhoneNumber: phone,
		FirstName:   entity.StrPtr("Old"),
		LastName:    entity.StrPtr("Name"),
		TotalOrders: 5,
		TotalSpent:  500.25,
	}
	err = repo.UpsertByPhone(context.Background(), cust)
	assert.NoError(t, err)
	assert.NotZero(t, cust.ID)

	// 2. Call UpdateCustomer with only a name change
	updateReq := &entity.Customer{
		ID:        cust.ID,
		FirstName: entity.StrPtr("New"),
	}
	err = service.UpdateCustomer(context.Background(), updateReq, false)
	assert.NoError(t, err)

	// 3. Verify name updated but stats preserved
	fetched, _ := repo.GetByID(context.Background(), cust.ID)
	assert.Equal(t, "New", *fetched.FirstName)
	assert.Equal(t, "Name", *fetched.LastName) // Preserved by patching
	assert.Equal(t, 5, fetched.TotalOrders)   // Preserved!
	assert.Equal(t, 500.25, fetched.TotalSpent) // Preserved!
}

func TestCustomerService_UpdateFromOrder_RecalculatesStats(t *testing.T) {
	db, err := testutil.SetupTestDB()
	if err != nil {
		t.Skip("DB not available")
	}
	defer testutil.CleanupTestDB(db)

	customerRepo := repository.NewCustomerRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	service := NewCustomerService(customerRepo, orderRepo, nil)

	ctx := context.Background()
	phone := "+919998881111"

	// 1. Create a customer
	cust := &entity.Customer{PhoneNumber: phone, FirstName: entity.StrPtr("Test")}
	err = customerRepo.UpsertByPhone(ctx, cust)
	assert.NoError(t, err)

	// 2. Create 2 orders for this customer in the DB
	status := "pending"
	order1 := entity.Order{
		ExternalOrderID: "ext_001",
		SourceID:        "shopify",
		CustomerPhone:   &phone,
		OrderNumber:     "ORD-001",
		TotalPrice:      150.0,
		Status:          &status,
	}
	order2 := entity.Order{
		ExternalOrderID: "ext_002",
		SourceID:        "shopify",
		CustomerPhone:   &phone,
		OrderNumber:     "ORD-002",
		TotalPrice:      200.0,
		Status:          &status,
	}
	
	// We use orderRepo.Upsert to put them in the DB
	_, err = orderRepo.Upsert(order1)
	assert.NoError(t, err)
	_, err = orderRepo.Upsert(order2)
	assert.NoError(t, err)

	// 3. Call UpdateFromOrder
	err = service.UpdateFromOrder(ctx, &order1)
	assert.NoError(t, err)

	// 4. Verify customer stats are updated correctly (150 + 200 = 350)
	fetched, _ := customerRepo.GetByPhone(ctx, phone)
	assert.NotNil(t, fetched, "Customer should be found")
	assert.Equal(t, 2, fetched.TotalOrders)
	assert.Equal(t, 350.0, fetched.TotalSpent)
}

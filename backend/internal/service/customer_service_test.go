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
	service := NewCustomerService(nil, nil, nil)

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
		assert.Equal(t, tc.expected, service.normalizePhone(tc.input))
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

package repository

import (
	"context"
	"testing"

	"mi-tech/internal/entity"
	"mi-tech/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type CustomerRepositoryTestSuite struct {
	suite.Suite
	db   *gorm.DB
	repo *CustomerRepository
}

func (s *CustomerRepositoryTestSuite) SetupSuite() {
	db, err := testutil.SetupTestDB()
	if err != nil {
		s.T().Skip("Skipping CustomerRepository tests: database not available")
	}
	s.db = db
	s.repo = NewCustomerRepository(db)
}

func (s *CustomerRepositoryTestSuite) TearDownSuite() {
	if s.db != nil {
		testutil.CleanupTestDB(s.db)
	}
}

func (s *CustomerRepositoryTestSuite) SetupTest() {
	s.db.Exec("TRUNCATE TABLE customers CASCADE")
}

func (s *CustomerRepositoryTestSuite) TestUpsertAndGet() {
	ctx := context.Background()
	phone := "+919876543210"
	customer := &entity.Customer{
		PhoneNumber: phone,
		FirstName:   entity.StrPtr("Alice"),
		LastName:    entity.StrPtr("Doe"),
		Email:       entity.StrPtr("alice@example.com"),
		TotalOrders: 5,
		TotalSpent:  500.0,
	}

	// 1. Create
	err := s.repo.UpsertByPhone(ctx, customer)
	assert.NoError(s.T(), err)

	// 2. Get
	fetched, err := s.repo.GetByPhone(ctx, phone)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "Alice", *fetched.FirstName)
	assert.Equal(s.T(), 500.0, fetched.TotalSpent)

	// 3. Update via Upsert (Delta check)
	customer.TotalSpent = 600.0
	err = s.repo.UpsertByPhone(ctx, customer)
	assert.NoError(s.T(), err)

	updated, _ := s.repo.GetByPhone(ctx, phone)
	assert.Equal(s.T(), 600.0, updated.TotalSpent)
}

func (s *CustomerRepositoryTestSuite) TestUpdateStats() {
	ctx := context.Background()
	phone := "+919876543210"
	s.repo.UpsertByPhone(ctx, &entity.Customer{
		PhoneNumber: phone,
		TotalOrders: 1,
		TotalSpent:  100.0,
	})

	err := s.repo.UpdateStats(ctx, phone, 2, 250.5)
	assert.NoError(s.T(), err)

	fetched, _ := s.repo.GetByPhone(ctx, phone)
	assert.Equal(s.T(), 3, fetched.TotalOrders)
	assert.Equal(s.T(), 350.5, fetched.TotalSpent)
}

func (s *CustomerRepositoryTestSuite) TestListWithFilters() {
	ctx := context.Background()
	s.repo.UpsertByPhone(ctx, &entity.Customer{PhoneNumber: "111", FirstName: entity.StrPtr("Alice"), City: entity.StrPtr("Delhi"), TotalSpent: 100})
	s.repo.UpsertByPhone(ctx, &entity.Customer{PhoneNumber: "222", FirstName: entity.StrPtr("Bob"), City: entity.StrPtr("Mumbai"), TotalSpent: 500})

	// Filter by city
	customers, count, err := s.repo.List(ctx, "", "", "", "", 0, 0, 0, "Delhi", "", "", "", "", false, false, false, 0, 10)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int64(1), count)
	assert.Equal(s.T(), "Alice", *customers[0].FirstName)

	// Filter by min spent
	customers, count, _ = s.repo.List(ctx, "", "", "", "", 400, 0, 0, "", "", "", "", "", false, false, false, 0, 10)
	assert.Equal(s.T(), int64(1), count)
	assert.Equal(s.T(), "Bob", *customers[0].FirstName)
}

func (s *CustomerRepositoryTestSuite) TestListSQLInjection() {
	ctx := context.Background()
	s.repo.UpsertByPhone(ctx, &entity.Customer{PhoneNumber: "111", FirstName: entity.StrPtr("Alice")})

	// Attempting SQL injection via sortBy
	// If vulnerable, this might be passed directly to the ORDER BY clause.
	// In some databases, this could be used to extract data via error-based injection or timing attacks.
	injection := "email, (SELECT CASE WHEN (1=1) THEN pg_sleep(0.1) ELSE pg_sleep(0) END)"

	// We just want to see if it executes without error, which it shouldn't if we have an allowlist.
	// Currently it might execute (and sleep briefly) or fail if the syntax is wrong for the specific DB.
	// But the goal is that after the fix, this specific string should be rejected/ignored.
	_, _, err := s.repo.List(ctx, "", injection, "ASC", "", 0, 0, 0, "", "", "", "", "", false, false, false, 0, 10)

	// If there's an error from the DB, it confirms it's trying to execute it.
	// If there's no error, it might still have executed it.
	if err != nil {
		s.T().Logf("List with injection returned error: %v", err)
	}
}

func TestCustomerRepositorySuite(t *testing.T) {
	suite.Run(t, new(CustomerRepositoryTestSuite))
}

package test

import (
	"fmt"
	"testing"

	"mi-tech/internal/domain/order/entity"
	"mi-tech/internal/domain/order/repository"
	"mi-tech/internal/domain/shared/testutil"
	"mi-tech/internal/domain/shared/util"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type OrderRepositoryTestSuite struct {
	suite.Suite
	db   *gorm.DB
	repo repository.OrderRepository
}

func (s *OrderRepositoryTestSuite) SetupSuite() {
	db, err := testutil.SetupTestDB()
	if err != nil {
		s.T().Skip("Skipping OrderRepository tests: database not available")
	}
	s.db = db
	s.repo = repository.NewOrderRepository(db)
}

func (s *OrderRepositoryTestSuite) TearDownSuite() {
	if s.db != nil {
		testutil.CleanupTestDB(s.db)
	}
}

func (s *OrderRepositoryTestSuite) SetupTest() {
	// Clean up tables before each test
	s.db.Exec("TRUNCATE TABLE order_line_items CASCADE")
	s.db.Exec("TRUNCATE TABLE orders CASCADE")
}

func (s *OrderRepositoryTestSuite) TestUpsertAndGet() {
	customerName := "John Doe"
	order := entity.Order{
		ExternalOrderID: "ext_123",
		OrderNumber:     "ORD-123",
		SourceID:        "shopify",
		CustomerName:    &customerName,
		TotalPrice:      100.0,
		LineItems: []entity.LineItem{
			{ID: "li_1", Title: util.StrPtr("Product 1"), Quantity: 1, Price: 100.0},
		},
	}

	// Test Upsert
	_, err := s.repo.Upsert(order)
	assert.NoError(s.T(), err)

	// Test GetByExternalID
	fetched, err := s.repo.GetByExternalID("ext_123")
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "ORD-123", fetched.OrderNumber)
	assert.Equal(s.T(), "John Doe", *fetched.CustomerName)

	// Test GetByFlexibleID (Numeric)
	fetchedFlex, err := s.repo.GetByFlexibleID(fmt.Sprintf("%d", fetched.ID))
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), fetched.ID, fetchedFlex.ID)

	// Test GetByFlexibleID (External)
	fetchedFlexExt, err := s.repo.GetByFlexibleID("ext_123")
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), fetched.ID, fetchedFlexExt.ID)
}

func (s *OrderRepositoryTestSuite) TestListWithFilters() {
	// Seed some orders
	c1 := "User One"
	c2 := "User Two"
	_, _ = s.repo.Upsert(entity.Order{ExternalOrderID: "e1", OrderNumber: "O1", SourceID: "shopify", CustomerName: &c1})
	_, _ = s.repo.Upsert(entity.Order{ExternalOrderID: "e2", OrderNumber: "O2", SourceID: "amazon", CustomerName: &c2})

	// Filter by Source
	orders, count, err := s.repo.List(repository.OrderFilter{Source: "shopify"})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 1, count)
	assert.Equal(s.T(), "O1", orders[0].OrderNumber)

	// Filter by Search
	orders, count, err = s.repo.List(repository.OrderFilter{Search: "User Two"})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 1, count)
	assert.Equal(s.T(), "O2", orders[0].OrderNumber)
}

func (s *OrderRepositoryTestSuite) TestUpdateStatus() {
	_, _ = s.repo.Upsert(entity.Order{ExternalOrderID: "ext_update", OrderNumber: "U1", SourceID: "shopify"})

	err := s.repo.UpdateStatus("ext_update", "paid", "fulfilled")
	assert.NoError(s.T(), err)

	fetched, _ := s.repo.GetByExternalID("ext_update")
	assert.Equal(s.T(), "paid", *fetched.FinancialStatus)
	assert.Equal(s.T(), "fulfilled", *fetched.FulfillmentStatus)
}

func (s *OrderRepositoryTestSuite) TestUpsertPIIMerge() {
	weakName := "Valued Customer"
	strongName := "John Strong"

	// 1. Upsert with strong name
	_, _ = s.repo.Upsert(entity.Order{ExternalOrderID: "pii_1", SourceID: "shopify", CustomerName: &strongName})

	// 2. Upsert same order with weak name
	_, _ = s.repo.Upsert(entity.Order{ExternalOrderID: "pii_1", SourceID: "shopify", CustomerName: &weakName})

	// 3. Verify strong name is preserved
	fetched, _ := s.repo.GetByExternalID("pii_1")
	assert.Equal(s.T(), strongName, *fetched.CustomerName)
}

func (s *OrderRepositoryTestSuite) TestUpsertBatch() {
	orders := []entity.Order{
		{ExternalOrderID: "b1", OrderNumber: "BN1", SourceID: "shopify"},
		{ExternalOrderID: "b2", OrderNumber: "BN2", SourceID: "shopify"},
	}

	_, err := s.repo.UpsertBatch(orders)
	assert.NoError(s.T(), err)

	_, count, _ := s.repo.List(repository.OrderFilter{})
	assert.Equal(s.T(), 2, count)
}

func TestOrderRepositorySuite(t *testing.T) {
	suite.Run(t, new(OrderRepositoryTestSuite))
}

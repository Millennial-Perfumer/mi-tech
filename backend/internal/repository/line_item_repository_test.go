package repository

import (
	"testing"

	"mi-tech/internal/entity"
	"mi-tech/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type LineItemRepositoryTestSuite struct {
	suite.Suite
	db   *gorm.DB
	repo LineItemRepository
}

func (s *LineItemRepositoryTestSuite) SetupSuite() {
	db, err := testutil.SetupTestDB()
	if err != nil {
		s.T().Skip("Skipping LineItemRepository tests: database not available")
	}
	s.db = db
	s.repo = NewLineItemRepository(db)
}

func (s *LineItemRepositoryTestSuite) TearDownSuite() {
	if s.db != nil {
		testutil.CleanupTestDB(s.db)
	}
}

func (s *LineItemRepositoryTestSuite) SetupTest() {
	s.db.Exec("TRUNCATE TABLE order_line_items CASCADE")
}

func (s *LineItemRepositoryTestSuite) TestUpsertAndGet() {
	orderID := int64(1001)
	items := []entity.LineItem{
		{ID: "item_1", Title: entity.StrPtr("Product 1"), Quantity: 2, Price: 50.0},
		{ID: "item_2", Title: entity.StrPtr("Product 2"), Quantity: 1, Price: 30.0},
	}

	// 1. UpsertBatch
	err := s.repo.UpsertBatch(s.db, orderID, items)
	assert.NoError(s.T(), err)

	// 2. GetByOrderID
	fetched, err := s.repo.GetByOrderID(orderID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 2, len(fetched))

	// 3. Update existing
	items[0].Quantity = 5
	s.repo.UpsertBatch(s.db, orderID, items)
	fetched, _ = s.repo.GetByOrderID(orderID)
	for _, item := range fetched {
		if item.ID == "item_1" {
			assert.Equal(s.T(), 5, item.Quantity)
		}
	}
}

func (s *LineItemRepositoryTestSuite) TestDeleteByOrderID() {
	orderID := int64(2002)
	s.repo.UpsertBatch(s.db, orderID, []entity.LineItem{{ID: "del_1", Title: entity.StrPtr("T1")}})

	err := s.repo.DeleteByOrderID(s.db, orderID)
	assert.NoError(s.T(), err)

	fetched, _ := s.repo.GetByOrderID(orderID)
	assert.Equal(s.T(), 0, len(fetched))
}

func TestLineItemRepositorySuite(t *testing.T) {
	suite.Run(t, new(LineItemRepositoryTestSuite))
}

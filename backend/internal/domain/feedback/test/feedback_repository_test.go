package test

import (
	"testing"
	"time"

	"mi-tech/internal/domain/feedback/entity"
	"mi-tech/internal/domain/feedback/repository"
	orderEntity "mi-tech/internal/domain/order/entity"
	"mi-tech/internal/shared/testutil"
	"mi-tech/internal/shared/util"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type FeedbackRepositoryTestSuite struct {
	suite.Suite
	db   *gorm.DB
	repo repository.FeedbackRepository
}

func (s *FeedbackRepositoryTestSuite) SetupSuite() {
	db, err := testutil.SetupTestDB()
	if err != nil {
		s.T().Skip("Skipping FeedbackRepository tests: database not available")
	}
	s.db = db
	s.repo = repository.NewFeedbackRepository(db)
}

func (s *FeedbackRepositoryTestSuite) TearDownSuite() {
	if s.db != nil {
		testutil.CleanupTestDB(s.db)
	}
}

func (s *FeedbackRepositoryTestSuite) SetupTest() {
	s.db.Exec("TRUNCATE TABLE customer_feedback RESTART IDENTITY CASCADE")
	s.db.Exec("TRUNCATE TABLE orders RESTART IDENTITY CASCADE")
}

func (s *FeedbackRepositoryTestSuite) TestFeedbackOperations() {
	// 1. Create a dummy order so we can link feedback to it
	order := orderEntity.Order{
		SourceID:        "shopify",
		ExternalOrderID: "ext-123",
		OrderNumber:     "1001",
		TotalPrice:      100.0,
		CustomerName:    util.StrPtr("John Doe"),
		CustomerPhone:   util.StrPtr("+1234567890"),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	err := s.db.Create(&order).Error
	assert.NoError(s.T(), err)

	// 2. Save customer feedback
	feedback := entity.CustomerFeedback{
		OrderID:       order.ID,
		CustomerPhone: "+1234567890",
		Rating:        5,
		Message:       "Great service!",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	err = s.repo.SaveCustomerFeedback(feedback)
	assert.NoError(s.T(), err)

	// 3. Retrieve customer feedback and verify it enriched order details
	feedbacks, err := s.repo.GetCustomerFeedback()
	assert.NoError(s.T(), err)
	assert.Len(s.T(), feedbacks, 1)
	assert.Equal(s.T(), "1001", feedbacks[0].OrderNumber)
	assert.Equal(s.T(), "John Doe", feedbacks[0].CustomerName)
	assert.Equal(s.T(), "Great service!", feedbacks[0].Comment)
	assert.Equal(s.T(), 5, feedbacks[0].Rating)
	assert.Nil(s.T(), feedbacks[0].AdminComment)

	// 4. Update feedback admin comment
	err = s.repo.UpdateFeedbackAdminComment(feedbacks[0].ID, "Thank you for the review!")
	assert.NoError(s.T(), err)

	// Verify update
	feedbacks2, err := s.repo.GetCustomerFeedback()
	assert.NoError(s.T(), err)
	assert.Len(s.T(), feedbacks2, 1)
	assert.NotNil(s.T(), feedbacks2[0].AdminComment)
	assert.Equal(s.T(), "Thank you for the review!", *feedbacks2[0].AdminComment)
}

func TestFeedbackRepositorySuite(t *testing.T) {
	suite.Run(t, new(FeedbackRepositoryTestSuite))
}

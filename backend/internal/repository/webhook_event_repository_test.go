package repository

import (
	"testing"

	"mi-tech/internal/entity"
	"mi-tech/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type WebhookEventRepositoryTestSuite struct {
	suite.Suite
	db   *gorm.DB
	repo WebhookEventRepository
}

func (s *WebhookEventRepositoryTestSuite) SetupSuite() {
	db, err := testutil.SetupTestDB()
	if err != nil {
		s.T().Skip("Skipping WebhookEventRepository tests: database not available")
	}
	s.db = db
	s.repo = NewWebhookEventRepository(db)
}

func (s *WebhookEventRepositoryTestSuite) TearDownSuite() {
	if s.db != nil {
		testutil.CleanupTestDB(s.db)
	}
}

func (s *WebhookEventRepositoryTestSuite) SetupTest() {
	s.db.Exec("TRUNCATE TABLE webhook_events CASCADE")
}

func (s *WebhookEventRepositoryTestSuite) TestSaveAndCheckProcessed() {
	deliveryID := "del_123"
	event := &entity.WebhookEvent{
		WebhookDeliveryID: deliveryID,
		Topic:             "orders/create",
	}

	// 1. Save
	err := s.repo.Save(event)
	assert.NoError(s.T(), err)

	// 2. IsProcessed
	processed, err := s.repo.IsProcessed(deliveryID)
	assert.NoError(s.T(), err)
	assert.True(s.T(), processed)

	processedNon, _ := s.repo.IsProcessed("non_existent")
	assert.False(s.T(), processedNon)
}

func (s *WebhookEventRepositoryTestSuite) TestLinkToOrder() {
	deliveryID := "del_link"
	s.repo.Save(&entity.WebhookEvent{
		WebhookDeliveryID: deliveryID,
		Topic:             "orders/update",
	})

	err := s.repo.LinkToOrder(deliveryID, 12345)
	assert.NoError(s.T(), err)

	var event entity.WebhookEvent
	s.db.Where("webhook_delivery_id = ?", deliveryID).First(&event)
	assert.Equal(s.T(), int64(12345), event.OrderID)
	assert.True(s.T(), event.Processed)
}

func TestWebhookEventRepositorySuite(t *testing.T) {
	suite.Run(t, new(WebhookEventRepositoryTestSuite))
}

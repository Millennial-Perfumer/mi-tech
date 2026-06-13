package test

import (
	"encoding/json"
	"testing"
	"time"

	orderEntity "mi-tech/internal/domain/order/entity"
	"mi-tech/internal/shared/testutil"
	webhookEntity "mi-tech/internal/domain/webhook/entity"
	"mi-tech/internal/domain/webhook/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type WebhookEventRepositoryTestSuite struct {
	suite.Suite
	db   *gorm.DB
	repo repository.WebhookEventRepository
}

func (s *WebhookEventRepositoryTestSuite) SetupSuite() {
	db, err := testutil.SetupTestDB()
	if err != nil {
		s.T().Skip("Skipping WebhookEventRepository tests: database not available")
	}
	s.db = db
	s.repo = repository.NewWebhookEventRepository(db)
}

func (s *WebhookEventRepositoryTestSuite) TearDownSuite() {
	if s.db != nil {
		testutil.CleanupTestDB(s.db)
	}
}

func (s *WebhookEventRepositoryTestSuite) SetupTest() {
	s.db.Exec("TRUNCATE TABLE webhook_events CASCADE")
	s.db.Exec("TRUNCATE TABLE orders CASCADE")
}

func (s *WebhookEventRepositoryTestSuite) TestSaveAndCheckProcessed() {
	deliveryID := "del_123"
	payload := json.RawMessage(`{}`)
	event := &webhookEntity.WebhookEvent{
		SourceID:          "shopify",
		ExternalID:        "ext_123",
		Payload:           &payload,
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
	// First insert dummy order to satisfy foreign key constraint
	order := &orderEntity.Order{
		ID:          12345,
		SourceID:    "shopify",
		OrderNumber: "1001",
		TotalPrice:  100.0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err := s.db.Create(order).Error
	assert.NoError(s.T(), err)

	deliveryID := "del_link"
	payload := json.RawMessage(`{}`)
	s.repo.Save(&webhookEntity.WebhookEvent{
		SourceID:          "shopify",
		ExternalID:        "ext_link",
		Payload:           &payload,
		WebhookDeliveryID: deliveryID,
		Topic:             "orders/update",
	})

	err = s.repo.LinkToOrder(deliveryID, 12345)
	assert.NoError(s.T(), err)

	var event webhookEntity.WebhookEvent
	s.db.Where("webhook_delivery_id = ?", deliveryID).First(&event)
	assert.NotNil(s.T(), event.OrderID)
	assert.Equal(s.T(), int64(12345), *event.OrderID)
	assert.True(s.T(), event.Processed)
}

func TestWebhookEventRepositorySuite(t *testing.T) {
	suite.Run(t, new(WebhookEventRepositoryTestSuite))
}

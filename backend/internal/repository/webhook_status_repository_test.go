package repository

import (
	"testing"

	"mi-tech/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type WebhookStatusRepositoryTestSuite struct {
	suite.Suite
	db   *gorm.DB
	repo WebhookStatusRepository
}

func (s *WebhookStatusRepositoryTestSuite) SetupSuite() {
	db, err := testutil.SetupTestDB()
	if err != nil {
		s.T().Skip("Skipping WebhookStatusRepository tests: database not available")
	}
	s.db = db
	s.repo = NewWebhookStatusRepository(db)
}

func (s *WebhookStatusRepositoryTestSuite) TearDownSuite() {
	if s.db != nil {
		testutil.CleanupTestDB(s.db)
	}
}

func (s *WebhookStatusRepositoryTestSuite) SetupTest() {
	s.db.Exec("TRUNCATE TABLE webhook_status CASCADE")
	// Seed initial row with ID 1
	s.db.Exec("INSERT INTO webhook_status (id, topic, status, last_received) VALUES (1, 'initial', 'inactive', '1970-01-01 00:00:00')")
}

func (s *WebhookStatusRepositoryTestSuite) TestGetAndUpdate() {
	// 1. Get
	topic, status, _, err := s.repo.Get()
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "initial", topic)
	assert.Equal(s.T(), "inactive", status)

	// 2. UpdateActivity
	err = s.repo.UpdateActivity("orders/create")
	assert.NoError(s.T(), err)

	topic, status, _, _ = s.repo.Get()
	assert.Equal(s.T(), "orders/create", topic)
	assert.Equal(s.T(), "active", status)
}

func TestWebhookStatusRepositorySuite(t *testing.T) {
	suite.Run(t, new(WebhookStatusRepositoryTestSuite))
}

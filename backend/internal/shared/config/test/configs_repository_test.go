package test

import (
	"testing"

	"mi-tech/internal/shared/config/repository"
	"mi-tech/internal/shared/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type ConfigsRepositoryTestSuite struct {
	suite.Suite
	db   *gorm.DB
	repo repository.ConfigsRepository
}

func (s *ConfigsRepositoryTestSuite) SetupSuite() {
	db, err := testutil.SetupTestDB()
	if err != nil {
		s.T().Skip("Skipping ConfigsRepository tests: database not available")
	}
	s.db = db
	s.repo = repository.NewConfigsRepository(db)
}

func (s *ConfigsRepositoryTestSuite) TearDownSuite() {
	if s.db != nil {
		testutil.CleanupTestDB(s.db)
	}
}

func (s *ConfigsRepositoryTestSuite) SetupTest() {
	s.db.Exec("TRUNCATE TABLE app_configs CASCADE")
}

func (s *ConfigsRepositoryTestSuite) TestSetAndGet() {
	key := "test_key"
	value := "test_value"

	// 1. Set
	err := s.repo.Set(key, value)
	assert.NoError(s.T(), err)

	// 2. Get
	fetched, err := s.repo.Get(key)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), value, fetched)

	// 3. Update
	newValue := "new_value"
	s.repo.Set(key, newValue)
	fetched, _ = s.repo.Get(key)
	assert.Equal(s.T(), newValue, fetched)
}

func (s *ConfigsRepositoryTestSuite) TestGetAllMasked() {
	// 1. Seed
	s.db.Exec("INSERT INTO app_configs (key, value, is_secret) VALUES (?, ?, ?)", "secret_key", "very_secret_12345", true)
	s.db.Exec("INSERT INTO app_configs (key, value, is_secret) VALUES (?, ?, ?)", "public_key", "public_value", false)

	// 2. GetAll
	configs, err := s.repo.GetAll()
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 2, len(configs))

	for _, c := range configs {
		if c.Key == "secret_key" {
			assert.Contains(s.T(), c.Value, "•")
			assert.Contains(s.T(), c.Value, "2345") // last 4 should be visible
		} else {
			assert.Equal(s.T(), "public_value", c.Value)
		}
	}
}

func TestConfigsRepositorySuite(t *testing.T) {
	suite.Run(t, new(ConfigsRepositoryTestSuite))
}

package repository

import (
	"testing"

	"mi-tech/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type SettingsRepositoryTestSuite struct {
	suite.Suite
	db   *gorm.DB
	repo *SettingsRepository
}

func (s *SettingsRepositoryTestSuite) SetupSuite() {
	db, err := testutil.SetupTestDB()
	if err != nil {
		s.T().Skip("Skipping SettingsRepository tests: database not available")
	}
	s.db = db
	s.repo = NewSettingsRepository(db)
}

func (s *SettingsRepositoryTestSuite) TearDownSuite() {
	if s.db != nil {
		testutil.CleanupTestDB(s.db)
	}
}

func (s *SettingsRepositoryTestSuite) SetupTest() {
	s.db.Exec("TRUNCATE TABLE app_settings CASCADE")
}

func (s *SettingsRepositoryTestSuite) TestSetAndGet() {
	key := "test_setting"
	value := "setting_value"

	// 1. Set
	err := s.repo.Set(key, value)
	assert.NoError(s.T(), err)

	// 2. Get
	fetched, err := s.repo.Get(key)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), value, fetched)

	// 3. Update
	newValue := "updated_setting"
	s.repo.Set(key, newValue)
	fetched, _ = s.repo.Get(key)
	assert.Equal(s.T(), newValue, fetched)
}

func (s *SettingsRepositoryTestSuite) TestDateRange() {
	start := "2023-01-01"
	end := "2023-01-31"

	// 1. Set
	err := s.repo.SetDateRange(start, end)
	assert.NoError(s.T(), err)

	// 2. Get
	fetchedStart, fetchedEnd, err := s.repo.GetDateRange()
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), start, fetchedStart)
	assert.Equal(s.T(), end, fetchedEnd)
}

func (s *SettingsRepositoryTestSuite) TestGetAll() {
	s.repo.Set("k1", "v1")
	s.repo.Set("k2", "v2")

	var settings []AppSetting
	err := s.repo.GetAll(&settings)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 2, len(settings))
}

func TestSettingsRepositorySuite(t *testing.T) {
	suite.Run(t, new(SettingsRepositoryTestSuite))
}

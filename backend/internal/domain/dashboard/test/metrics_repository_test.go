package test

import (
	"testing"
	"time"

	"mi-tech/internal/domain/dashboard/repository"
	orderEntity "mi-tech/internal/domain/order/entity"
	"mi-tech/internal/shared/testutil"
	util "mi-tech/internal/shared/util"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type MetricsRepositoryTestSuite struct {
	suite.Suite
	db          *gorm.DB
	metricsRepo repository.MetricsRepository
}

func (s *MetricsRepositoryTestSuite) SetupSuite() {
	db, err := testutil.SetupTestDB()
	if err != nil {
		s.T().Skip("Skipping Metrics tests: database not available")
	}
	s.db = db
	s.metricsRepo = repository.NewMetricsRepository(db)
}

func (s *MetricsRepositoryTestSuite) TearDownSuite() {
	if s.db != nil {
		testutil.CleanupTestDB(s.db)
	}
}

func (s *MetricsRepositoryTestSuite) SetupTest() {
	s.db.Exec("TRUNCATE TABLE orders CASCADE")
}

func (s *MetricsRepositoryTestSuite) TestGetDashboardMetrics() {
	// Seed orders
	now := time.Now().Add(5 * time.Minute).Format(time.RFC3339)
	tn := "Tamil Nadu"
	ka := "Karnataka"

	err := s.db.Create(&orderEntity.Order{
		SourceID: "shopify", ExternalOrderID: "m1", TotalPrice: 118.0, CustomerState: &tn, CreatedAt: time.Now(),
	}).Error
	assert.NoError(s.T(), err)

	err = s.db.Create(&orderEntity.Order{
		SourceID: "shopify", ExternalOrderID: "m2", TotalPrice: 118.0, CustomerState: &ka, CreatedAt: time.Now(),
	}).Error
	assert.NoError(s.T(), err)

	err = s.db.Create(&orderEntity.Order{
		SourceID: "shopify", ExternalOrderID: "m3", TotalPrice: 100.0, Status: util.StrPtr("CANCELLED"), CreatedAt: time.Now(),
	}).Error
	assert.NoError(s.T(), err)

	metrics, err := s.metricsRepo.GetDashboardMetrics("", now, []string{})
	assert.NoError(s.T(), err)

	// Total revenue should exclude cancelled: 118 + 118 = 236
	assert.Equal(s.T(), 236.0, metrics.TotalRevenue)
	assert.Equal(s.T(), 3, metrics.TotalInvoices)
	assert.Equal(s.T(), 1, metrics.CancelledOrders)
	assert.Equal(s.T(), 0, metrics.FulfilledOrders)
	assert.Equal(s.T(), 2, metrics.UnfulfilledOrders)

	// TN order (118): Tax is 18. CGST = 9, SGST = 9
	// KA order (118): Tax is 18. IGST = 18
	assert.Equal(s.T(), 9.0, metrics.CGSTCollected)
	assert.Equal(s.T(), 9.0, metrics.SGSTCollected)
	assert.Equal(s.T(), 18.0, metrics.IGSTCollected)
}

func TestMetricsRepositorySuite(t *testing.T) {
	suite.Run(t, new(MetricsRepositoryTestSuite))
}

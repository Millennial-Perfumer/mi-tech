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

type MetricsReportRepositoryTestSuite struct {
	suite.Suite
	db          *gorm.DB
	metricsRepo repository.MetricsRepository
	reportRepo  repository.ReportRepository
}

func (s *MetricsReportRepositoryTestSuite) SetupSuite() {
	db, err := testutil.SetupTestDB()
	if err != nil {
		s.T().Skip("Skipping Metrics/Report tests: database not available")
	}
	s.db = db
	s.metricsRepo = repository.NewMetricsRepository(db)
	s.reportRepo = repository.NewReportRepository(db)
}

func (s *MetricsReportRepositoryTestSuite) TearDownSuite() {
	if s.db != nil {
		testutil.CleanupTestDB(s.db)
	}
}

func (s *MetricsReportRepositoryTestSuite) SetupTest() {
	s.db.Exec("TRUNCATE TABLE orders CASCADE")
}

func (s *MetricsReportRepositoryTestSuite) TestGetDashboardMetrics() {
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

func (s *MetricsReportRepositoryTestSuite) TestGetGSTSummary() {
	now := time.Now().Add(5 * time.Minute).Format(time.RFC3339)
	tn := "Tamil Nadu"
	err := s.db.Create(&orderEntity.Order{
		SourceID: "shopify", ExternalOrderID: "r1", TotalPrice: 118.0, CustomerState: &tn, FinancialStatus: util.StrPtr("paid"), CreatedAt: time.Now(),
	}).Error
	assert.NoError(s.T(), err)

	res, err := s.reportRepo.GetGSTSummary("", now)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 1, res.TotalOrders)
	assert.Equal(s.T(), 1, res.PaidOrders)
	assert.Equal(s.T(), 118.0, res.TotalRevenue)
	assert.Equal(s.T(), 100.0, res.TotalTaxable)
	assert.Equal(s.T(), 18.0, res.TotalTax)
	assert.Equal(s.T(), 9.0, res.CGST)
	assert.Equal(s.T(), 9.0, res.SGST)
	assert.Equal(s.T(), 0.0, res.IGST)
}

func (s *MetricsReportRepositoryTestSuite) TestGetStateSummary() {
	now := time.Now().Add(5 * time.Minute).Format(time.RFC3339)
	tn := "Tamil Nadu"
	ka := "Karnataka"
	err := s.db.Create(&orderEntity.Order{SourceID: "shopify", ExternalOrderID: "s1", TotalPrice: 100, CustomerState: &tn, CreatedAt: time.Now()}).Error
	assert.NoError(s.T(), err)
	err = s.db.Create(&orderEntity.Order{SourceID: "shopify", ExternalOrderID: "s2", TotalPrice: 200, CustomerState: &tn, CreatedAt: time.Now()}).Error
	assert.NoError(s.T(), err)
	err = s.db.Create(&orderEntity.Order{SourceID: "shopify", ExternalOrderID: "s3", TotalPrice: 50, CustomerState: &ka, CreatedAt: time.Now()}).Error
	assert.NoError(s.T(), err)

	results, err := s.reportRepo.GetStateSummary("", now)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 2, len(results))

	for _, r := range results {
		if r.State == "Tamil Nadu" {
			assert.Equal(s.T(), 2, r.Orders)
			assert.Equal(s.T(), 300.0, r.Revenue)
		}
	}
}

func (s *MetricsReportRepositoryTestSuite) TestGetHSNSummary() {
	now := time.Now().Add(5 * time.Minute).Format(time.RFC3339)
	tn := "Tamil Nadu"

	// Create order
	order := orderEntity.Order{
		SourceID:        "shopify",
		ExternalOrderID: "h1",
		TotalPrice:      118.0,
		CustomerState:   &tn,
		CreatedAt:       time.Now(),
	}
	err := s.db.Create(&order).Error
	assert.NoError(s.T(), err)

	// Create line items
	err = s.db.Create(&orderEntity.LineItem{
		OrderID:  order.ID,
		HSCode:   util.StrPtr("123456"),
		Price:    100.0,
		Quantity: 1,
		Discount: 0,
	}).Error
	assert.NoError(s.T(), err)

	results, err := s.reportRepo.GetHSNSummary("", now)
	assert.NoError(s.T(), err)
	assert.NotEmpty(s.T(), results)

	found := false
	for _, r := range results {
		if r.HSNCode == "123456" {
			found = true
			assert.Equal(s.T(), 100.0, r.TaxableValue)
			assert.Equal(s.T(), 18.0, r.TotalGST)
			assert.Equal(s.T(), 118.0, r.Revenue)
		}
	}
	assert.True(s.T(), found)
}

func TestMetricsReportRepositorySuite(t *testing.T) {
	suite.Run(t, new(MetricsReportRepositoryTestSuite))
}

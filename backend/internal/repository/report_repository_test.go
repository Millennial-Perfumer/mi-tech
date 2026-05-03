package repository

import (
	"testing"
	"time"

	"mi-tech/internal/entity"
	"mi-tech/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type MetricsReportRepositoryTestSuite struct {
	suite.Suite
	db          *gorm.DB
	metricsRepo MetricsRepository
	reportRepo  ReportRepository
}

func (s *MetricsReportRepositoryTestSuite) SetupSuite() {
	db, err := testutil.SetupTestDB()
	if err != nil {
		s.T().Skip("Skipping Metrics/Report tests: database not available")
	}
	s.db = db
	s.metricsRepo = NewMetricsRepository(db)
	s.reportRepo = NewReportRepository(db)
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
	now := time.Now().Format(time.RFC3339)
	tn := "Tamil Nadu"
	ka := "Karnataka"

	s.db.Create(&entity.Order{
		ExternalOrderID: "m1", TotalPrice: 118.0, CustomerState: &tn, CreatedAt: time.Now(),
	})
	s.db.Create(&entity.Order{
		ExternalOrderID: "m2", TotalPrice: 118.0, CustomerState: &ka, CreatedAt: time.Now(),
	})
	s.db.Create(&entity.Order{
		ExternalOrderID: "m3", TotalPrice: 100.0, Status: entity.StrPtr("CANCELLED"), CreatedAt: time.Now(),
	})

	rev, cgst, sgst, igst, total, cancelled, fulfilled, unfulfilled, err := s.metricsRepo.GetDashboardMetrics("", now)
	assert.NoError(s.T(), err)

	// Total revenue should exclude cancelled: 118 + 118 = 236
	assert.Equal(s.T(), 236.0, rev)
	assert.Equal(s.T(), 3, total)
	assert.Equal(s.T(), 1, cancelled)
	assert.Equal(s.T(), 0, fulfilled)
	assert.Equal(s.T(), 2, unfulfilled)

	// TN order (118): Tax is 18. CGST = 9, SGST = 9
	// KA order (118): Tax is 18. IGST = 18
	assert.Equal(s.T(), 9.0, cgst)
	assert.Equal(s.T(), 9.0, sgst)
	assert.Equal(s.T(), 18.0, igst)
}

func (s *MetricsReportRepositoryTestSuite) TestGetGSTSummary() {
	now := time.Now().Format(time.RFC3339)
	tn := "Tamil Nadu"
	s.db.Create(&entity.Order{
		ExternalOrderID: "r1", TotalPrice: 118.0, CustomerState: &tn, FinancialStatus: entity.StrPtr("paid"), CreatedAt: time.Now(),
	})

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
	now := time.Now().Format(time.RFC3339)
	tn := "Tamil Nadu"
	ka := "Karnataka"
	s.db.Create(&entity.Order{ExternalOrderID: "s1", TotalPrice: 100, CustomerState: &tn, CreatedAt: time.Now()})
	s.db.Create(&entity.Order{ExternalOrderID: "s2", TotalPrice: 200, CustomerState: &tn, CreatedAt: time.Now()})
	s.db.Create(&entity.Order{ExternalOrderID: "s3", TotalPrice: 50, CustomerState: &ka, CreatedAt: time.Now()})

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
	now := time.Now().Format(time.RFC3339)
	tn := "Tamil Nadu"

	// Create order
	order := entity.Order{
		ExternalOrderID: "h1",
		TotalPrice:      118.0,
		CustomerState:   &tn,
		CreatedAt:       time.Now(),
	}
	s.db.Create(&order)

	// Create line items
	s.db.Create(&entity.LineItem{
		OrderID:  order.ID,
		HSCode:   entity.StrPtr("123456"),
		Price:    100.0,
		Quantity: 1,
		Discount: 0,
	})

	results, err := s.reportRepo.GetHSNSummary("", now)
	assert.NoError(s.T(), err)
	assert.NotEmpty(s.T(), results)

	found := false
	for _, r := range results {
		if r.HSNCode == "123456" {
			found = true
			assert.Equal(s.T(), 100.0, r.TaxableValue)
			assert.Equal(s.T(), 18.0, r.TotalGST)
			assert.Equal(s.T(), 9.0, r.CGST)
			assert.Equal(s.T(), 9.0, r.SGST)
			assert.Equal(s.T(), 0.0, r.IGST)
			assert.Equal(s.T(), 118.0, r.Revenue)
		}
	}
	assert.True(s.T(), found)
}

func TestMetricsReportRepositorySuite(t *testing.T) {
	suite.Run(t, new(MetricsReportRepositoryTestSuite))
}

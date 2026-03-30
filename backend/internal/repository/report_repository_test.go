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

	total, _, _, _, paid, rev, taxable, tax, err := s.reportRepo.GetGSTSummary("", now)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 1, total)
	assert.Equal(s.T(), 1, paid)
	assert.Equal(s.T(), 118.0, rev)
	assert.Equal(s.T(), 100.0, taxable)
	assert.Equal(s.T(), 18.0, tax)
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
	hsn1 := "123456"
	hsn2 := "789012"

	// Order 1: TN, Total 118, 2 line items
	o1 := entity.Order{
		ExternalOrderID: "o1", TotalPrice: 118.0, CustomerState: &tn, CreatedAt: time.Now(),
	}
	s.db.Create(&o1)
	s.db.Create(&entity.LineItem{ID: "li1", OrderID: o1.ID, HSCode: &hsn1, Quantity: 1, Price: 50.0})
	s.db.Create(&entity.LineItem{ID: "li2", OrderID: o1.ID, HSCode: &hsn2, Quantity: 1, Price: 50.0})

	// Order 2: TN, Total 236, 1 line item (same HSN as li1)
	o2 := entity.Order{
		ExternalOrderID: "o2", TotalPrice: 236.0, CustomerState: &tn, CreatedAt: time.Now(),
	}
	s.db.Create(&o2)
	s.db.Create(&entity.LineItem{ID: "li3", OrderID: o2.ID, HSCode: &hsn1, Quantity: 2, Price: 100.0})

	results, err := s.reportRepo.GetHSNSummary("", now)
	assert.NoError(s.T(), err)

	// Expectations:
	// HSN 123456:
	//   Order 1 share: line_val=50, line_sum=100. taxable = (50/100)*(118/1.18) = 50. revenue = (50/100)*118 = 59. qty = 1
	//   Order 2 share: line_val=200, line_sum=200. taxable = (200/200)*(236/1.18) = 200. revenue = (200/200)*236 = 236. qty = 2
	//   Total taxable = 250, Total revenue = 295, Total qty = 3, Product count = 2

	// HSN 789012:
	//   Order 1 share: line_val=50, line_sum=100. taxable = (50/100)*(118/1.18) = 50. revenue = (50/100)*118 = 59. qty = 1
	//   Total taxable = 50, Total revenue = 59, Total qty = 1, Product count = 1

	assert.Equal(s.T(), 2, len(results))
	for _, r := range results {
		if r.HSNCode == hsn1 {
			assert.Equal(s.T(), 3, r.QtySold)
			assert.Equal(s.T(), 2, r.ProductCount)
			assert.InDelta(s.T(), 250.0, r.TaxableValue, 0.01)
			assert.InDelta(s.T(), 295.0, r.Revenue, 0.01)
		} else if r.HSNCode == hsn2 {
			assert.Equal(s.T(), 1, r.QtySold)
			assert.Equal(s.T(), 1, r.ProductCount)
			assert.InDelta(s.T(), 50.0, r.TaxableValue, 0.01)
			assert.InDelta(s.T(), 59.0, r.Revenue, 0.01)
		}
	}
}

func TestMetricsReportRepositorySuite(t *testing.T) {
	suite.Run(t, new(MetricsReportRepositoryTestSuite))
}

package test

import (
	"testing"
	"time"

	"mi-tech/internal/domain/gst/repository"
	orderEntity "mi-tech/internal/domain/order/entity"
	"mi-tech/internal/shared/testutil"
	util "mi-tech/internal/shared/util"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type GSTRepositoryTestSuite struct {
	suite.Suite
	db         *gorm.DB
	reportRepo repository.GSTRepository
}

func (s *GSTRepositoryTestSuite) SetupSuite() {
	db, err := testutil.SetupTestDB()
	if err != nil {
		s.T().Skip("Skipping GST reports tests: database not available")
	}
	s.db = db
	s.reportRepo = repository.NewGSTRepository(db)
}

func (s *GSTRepositoryTestSuite) TearDownSuite() {
	if s.db != nil {
		testutil.CleanupTestDB(s.db)
	}
}

func (s *GSTRepositoryTestSuite) SetupTest() {
	s.db.Exec("TRUNCATE TABLE orders CASCADE")
}

func (s *GSTRepositoryTestSuite) TestGetGSTSummary() {
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

func (s *GSTRepositoryTestSuite) TestGetStateSummary() {
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

func (s *GSTRepositoryTestSuite) TestGetHSNSummary() {
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

func (s *GSTRepositoryTestSuite) TestGetDocumentsIssued() {
	now := time.Now().Add(5 * time.Minute).Format(time.RFC3339)

	// Create a shopify order (should be included)
	inv1 := "SY-2853"
	err := s.db.Create(&orderEntity.Order{
		SourceID:        "shopify",
		ExternalOrderID: "sh1",
		OrderNumber:     "INV-2853",
		InvoiceNumber:   &inv1,
		TotalPrice:      100.0,
		CreatedAt:       time.Now(),
	}).Error
	assert.NoError(s.T(), err)

	// Create another shopify order (should be included)
	inv2 := "SY-2855"
	err = s.db.Create(&orderEntity.Order{
		SourceID:        "shopify",
		ExternalOrderID: "sh2",
		OrderNumber:     "INV-2855",
		InvoiceNumber:   &inv2,
		TotalPrice:      150.0,
		CreatedAt:       time.Now(),
	}).Error
	assert.NoError(s.T(), err)

	// Create an amazon order (should be EXCLUDED)
	inv3 := "AMZ-1"
	err = s.db.Create(&orderEntity.Order{
		SourceID:        "amazon",
		ExternalOrderID: "am1",
		OrderNumber:     "406-2823602-6234752",
		InvoiceNumber:   &inv3,
		TotalPrice:      200.0,
		CreatedAt:       time.Now(),
	}).Error
	assert.NoError(s.T(), err)

	minVal, maxVal, total, cancelled, err := s.reportRepo.GetShopifyDocumentsIssued("", now)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int64(2853), *minVal)
	assert.Equal(s.T(), int64(2855), *maxVal)
	assert.Equal(s.T(), 2, total) // Only Shopify orders counted
	assert.Equal(s.T(), 0, cancelled)

	amzMinVal, amzMaxVal, amzTotal, amzCancelled, err := s.reportRepo.GetAmazonDocumentsIssued("", now)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int64(1), *amzMinVal)
	assert.Equal(s.T(), int64(1), *amzMaxVal)
	assert.Equal(s.T(), 1, amzTotal)
	assert.Equal(s.T(), 0, amzCancelled)
}

func (s *GSTRepositoryTestSuite) TestGetGSTR1B2CS() {
	now := time.Now().Add(5 * time.Minute).Format(time.RFC3339)
	tn := "Tamil Nadu"
	ka := "Karnataka"

	// TN Order (Intrastate)
	err := s.db.Create(&orderEntity.Order{
		SourceID:        "shopify",
		ExternalOrderID: "b2cs_sh1",
		TotalPrice:      118.0, // Taxable: 100, GST: 18
		CustomerState:   &tn,
		CreatedAt:       time.Now(),
	}).Error
	assert.NoError(s.T(), err)

	// KA Order (Interstate)
	err = s.db.Create(&orderEntity.Order{
		SourceID:        "amazon",
		ExternalOrderID: "b2cs_amz1",
		TotalPrice:      236.0, // Taxable: 200, GST: 36
		CustomerState:   &ka,
		CreatedAt:       time.Now(),
	}).Error
	assert.NoError(s.T(), err)

	rows, err := s.reportRepo.GetGSTR1B2CS("", now)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), rows, 2)

	var foundTN, foundKA bool
	for _, row := range rows {
		if row.POS == "33" {
			foundTN = true
			assert.Equal(s.T(), "INTRA", row.SplyTy)
			assert.Equal(s.T(), 18.0, row.Rt)
			assert.Equal(s.T(), 100.0, row.TxVal)
			assert.Equal(s.T(), 9.0, row.Camt)
			assert.Equal(s.T(), 9.0, row.Samt)
			assert.Equal(s.T(), 0.0, row.Iamt)
			assert.Equal(s.T(), "OE", row.Typ) // Shopify = OE
		} else if row.POS == "29" {
			foundKA = true
			assert.Equal(s.T(), "INTER", row.SplyTy)
			assert.Equal(s.T(), 18.0, row.Rt)
			assert.Equal(s.T(), 200.0, row.TxVal)
			assert.Equal(s.T(), 0.0, row.Camt)
			assert.Equal(s.T(), 0.0, row.Samt)
			assert.Equal(s.T(), 36.0, row.Iamt)
			assert.Equal(s.T(), "E", row.Typ) // Amazon = E
		}
	}
	assert.True(s.T(), foundTN)
	assert.True(s.T(), foundKA)
}

func (s *GSTRepositoryTestSuite) TestGetGSTR1HSN() {
	now := time.Now().Add(5 * time.Minute).Format(time.RFC3339)
	tn := "Tamil Nadu"

	order := orderEntity.Order{
		SourceID:        "shopify",
		ExternalOrderID: "hsn_sh1",
		TotalPrice:      118.0,
		CustomerState:   &tn,
		CreatedAt:       time.Now(),
	}
	err := s.db.Create(&order).Error
	assert.NoError(s.T(), err)

	err = s.db.Create(&orderEntity.LineItem{
		OrderID:  order.ID,
		HSCode:   util.StrPtr("330290"),
		Price:    100.0,
		Quantity: 1,
		Discount: 0,
	}).Error
	assert.NoError(s.T(), err)

	rows, err := s.reportRepo.GetGSTR1HSN("", now)
	assert.NoError(s.T(), err)
	assert.NotEmpty(s.T(), rows)

	var found bool
	for _, r := range rows {
		if r.HsnSc == "330290" {
			found = true
			assert.Equal(s.T(), 1.0, r.Qty)
			assert.Equal(s.T(), 100.0, r.TxVal)
			assert.Equal(s.T(), 9.0, r.Camt)
			assert.Equal(s.T(), 9.0, r.Samt)
			assert.Equal(s.T(), 0.0, r.Iamt)
			assert.Equal(s.T(), "PCS", r.Uqc)
		}
	}
	assert.True(s.T(), found)
}

func TestGSTRepositorySuite(t *testing.T) {
	suite.Run(t, new(GSTRepositoryTestSuite))
}

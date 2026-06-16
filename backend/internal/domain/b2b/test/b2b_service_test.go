package test

import (
	"testing"
	"time"

	"mi-tech/internal/domain/b2b/entity"
	"mi-tech/internal/domain/b2b/repository"
	"mi-tech/internal/domain/b2b/service"
	configRepo "mi-tech/internal/shared/config/repository"
	"mi-tech/internal/shared/config"
	"mi-tech/internal/shared/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type B2BServiceTestSuite struct {
	suite.Suite
	db         *gorm.DB
	b2bService *service.B2BService
	settings   *config.SettingsProvider
}

func (s *B2BServiceTestSuite) SetupSuite() {
	db, err := testutil.SetupTestDB()
	if err != nil {
		s.T().Skip("Skipping B2B tests: database not available")
	}
	s.db = db

	// Initialize configs
	configsRepo := configRepo.NewConfigsRepository(db)
	s.settings = config.NewSettingsProvider(configsRepo)

	// Seed business details
	db.Exec("INSERT INTO app_settings (key, value) VALUES ('business_name', 'PARFUM TRADERS') ON CONFLICT (key) DO UPDATE SET value = 'PARFUM TRADERS'")
	db.Exec("INSERT INTO app_settings (key, value) VALUES ('business_gstin', '33AUSPR1909H1ZC') ON CONFLICT (key) DO UPDATE SET value = '33AUSPR1909H1ZC'")
	db.Exec("INSERT INTO app_settings (key, value) VALUES ('business_address_line1', 'No. 9/21, Chennai') ON CONFLICT (key) DO UPDATE SET value = 'No. 9/21, Chennai'")
	db.Exec("INSERT INTO app_settings (key, value) VALUES ('business_address_line2', 'Tamil Nadu') ON CONFLICT (key) DO UPDATE SET value = 'Tamil Nadu'")

	b2bRepo := repository.NewB2BRepository(db)
	s.b2bService = service.NewB2BService(b2bRepo, s.settings, db)
}

func (s *B2BServiceTestSuite) TearDownSuite() {
	if s.db != nil {
		testutil.CleanupTestDB(s.db)
	}
}

func (s *B2BServiceTestSuite) SetupTest() {
	s.db.Exec("TRUNCATE TABLE b2b_invoice_items CASCADE")
	s.db.Exec("TRUNCATE TABLE b2b_invoices CASCADE")
	s.db.Exec("TRUNCATE TABLE b2b_customers CASCADE")
}

func (s *B2BServiceTestSuite) TestCreateCustomer() {
	cust := &entity.B2BCustomer{
		LegalName:      "Alpha Corp",
		GSTIN:          "33ABCDE1234F1Z5", // TN GSTIN
		BillingAddress: "123 Street, Chennai",
	}

	err := s.b2bService.CreateCustomer(cust)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "33", cust.StateCode)
	assert.Equal(s.T(), "Tamil Nadu", cust.State)
	assert.Equal(s.T(), "ABCDE1234F", *cust.PAN)

	// Invalid GSTIN should fail
	badCust := &entity.B2BCustomer{
		LegalName:      "Beta Corp",
		GSTIN:          "INVALIDGSTIN",
		BillingAddress: "123 Street",
	}
	err = s.b2bService.CreateCustomer(badCust)
	assert.Error(s.T(), err)
}

func (s *B2BServiceTestSuite) TestCreateInvoiceAndTaxSplits() {
	cust := &entity.B2BCustomer{
		LegalName:      "Beta Industries",
		GSTIN:          "29ABCDE1234F1Z5", // KA GSTIN (Inter-state)
		BillingAddress: "456 Blvd, Bangalore",
	}
	err := s.b2bService.CreateCustomer(cust)
	assert.NoError(s.T(), err)

	inv := &entity.B2BInvoice{
		CustomerID:  &cust.ID,
		InvoiceDate: time.Date(2026, 6, 15, 10, 0, 0, 0, time.UTC),
		Items: []entity.B2BInvoiceItem{
			{ItemDetails: "Perfume A", Quantity: 10, Rate: 500.00},
			{ItemDetails: "Perfume B", Quantity: 5, Rate: 1000.00},
		},
	}

	err = s.b2bService.CreateInvoice(inv)
	assert.NoError(s.T(), err)

	// Subtotal = (10*500) + (5*1000) = 10000
	assert.Equal(s.T(), 10000.00, inv.SubtotalPrice)

	// Inter-state (KA client, TN seller) -> IGST 18% = 1800. CGST/SGST = 0.
	assert.Equal(s.T(), 18.00, inv.IGSTRate)
	assert.Equal(s.T(), 1800.00, inv.IGSTAmount)
	assert.Equal(s.T(), 0.00, inv.CGSTAmount)
	assert.Equal(s.T(), 0.00, inv.SGSTAmount)
	assert.Equal(s.T(), 11800.00, inv.TotalPrice)

	// Transition to ISSUED and check numbering
	issued, err := s.b2bService.IssueInvoice(inv.ID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "ISSUED", issued.Status)
	assert.Equal(s.T(), "26-27", *issued.FinancialYear)
	assert.Equal(s.T(), "MP/26-27/0001", *issued.InvoiceNumber)
}

func TestB2BServiceSuite(t *testing.T) {
	suite.Run(t, new(B2BServiceTestSuite))
}

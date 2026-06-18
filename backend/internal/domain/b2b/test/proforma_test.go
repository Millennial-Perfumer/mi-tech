package test

import (
	"encoding/json"
	"testing"
	"time"

	"mi-tech/internal/domain/b2b/entity"
	"mi-tech/internal/domain/b2b/repository"
	"mi-tech/internal/domain/b2b/service"
	"mi-tech/internal/shared/config"
	configRepo "mi-tech/internal/shared/config/repository"
	"mi-tech/internal/shared/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type B2BProformaTestSuite struct {
	suite.Suite
	db         *gorm.DB
	b2bService *service.B2BService
	settings   *config.SettingsProvider
}

func (s *B2BProformaTestSuite) SetupSuite() {
	db, err := testutil.SetupTestDB()
	if err != nil {
		s.T().Skip("Skipping Proforma tests: database not available")
	}
	s.db = db

	configsRepo := configRepo.NewConfigsRepository(db)
	s.settings = config.NewSettingsProvider(configsRepo)

	db.Exec("INSERT INTO app_settings (key, value) VALUES ('business_name', 'PARFUM TRADERS') ON CONFLICT (key) DO UPDATE SET value = 'PARFUM TRADERS'")
	db.Exec("INSERT INTO app_settings (key, value) VALUES ('business_gstin', '33AUSPR1909H1ZC') ON CONFLICT (key) DO UPDATE SET value = '33AUSPR1909H1ZC'")
	db.Exec("INSERT INTO app_settings (key, value) VALUES ('business_address_line1', 'No. 9/21, Chennai') ON CONFLICT (key) DO UPDATE SET value = 'No. 9/21, Chennai'")
	db.Exec("INSERT INTO app_settings (key, value) VALUES ('business_address_line2', 'Tamil Nadu') ON CONFLICT (key) DO UPDATE SET value = 'Tamil Nadu'")

	b2bRepo := repository.NewB2BRepository(db)
	s.b2bService = service.NewB2BService(b2bRepo, s.settings, db)
}

func (s *B2BProformaTestSuite) TearDownSuite() {
	if s.db != nil {
		testutil.CleanupTestDB(s.db)
	}
}

func (s *B2BProformaTestSuite) SetupTest() {
	s.db.Exec("TRUNCATE TABLE b2b_invoice_items CASCADE")
	s.db.Exec("TRUNCATE TABLE b2b_invoices CASCADE")
	s.db.Exec("TRUNCATE TABLE b2b_proforma_invoice_items CASCADE")
	s.db.Exec("TRUNCATE TABLE b2b_proforma_invoices CASCADE")
	s.db.Exec("TRUNCATE TABLE b2b_customers CASCADE")
}

func (s *B2BProformaTestSuite) TestProformaLifecycle() {
	// 1. Create B2B Customer
	cust := &entity.B2BCustomer{
		LegalName:      "Beta Industries",
		GSTIN:          "29ABCDE1234F1Z5", // KA GSTIN (Inter-state)
		BillingAddress: "456 Blvd, Bangalore",
	}
	err := s.b2bService.CreateCustomer(cust)
	assert.NoError(s.T(), err)

	// 2. Create Proforma (DRAFT)
	pf := &entity.B2BProformaInvoice{
		CustomerID:  &cust.ID,
		NoteDate:    time.Date(2026, 6, 15, 10, 0, 0, 0, time.UTC),
		AdvancePaid: 5000.00,
		Items: []entity.B2BProformaInvoiceItem{
			{ItemDetails: "Vanilla Extract", Quantity: 10, Rate: 1000.00},
			{ItemDetails: "Oud Oil", Quantity: 2, Rate: 5000.00},
		},
	}

	err = s.b2bService.CreateProforma(pf)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "DRAFT", pf.Status)
	assert.Equal(s.T(), 20000.00, pf.SubtotalPrice)
	assert.Equal(s.T(), 18.00, pf.IGSTRate)
	assert.Equal(s.T(), 3600.00, pf.IGSTAmount)
	assert.Equal(s.T(), 23600.00, pf.TotalPrice)

	// 3. Issue Proforma
	issued, err := s.b2bService.IssueProforma(pf.ID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "SENT", issued.Status)
	assert.Equal(s.T(), "26-27", *issued.FinancialYear)
	assert.Equal(s.T(), "PT/PI/26-27/001", *issued.ProformaNumber)

	// 4. Create Revision
	rev, err := s.b2bService.CreateRevision(issued.ID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "DRAFT", rev.Status)
	assert.Equal(s.T(), 2, rev.RevisionNumber)
	assert.Equal(s.T(), issued.ID, *rev.ParentProformaID)

	// Modify and update revision
	rev.Items[0].Quantity = 12 // Bump quantity
	err = s.b2bService.UpdateProforma(rev)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 22000.00, rev.SubtotalPrice) // (12 * 1000) + (2 * 5000)

	var pfCount int64
	s.db.Model(&entity.B2BProformaInvoice{}).Count(&pfCount)
	assert.Equal(s.T(), int64(2), pfCount) // Must be exactly 2 (the parent/original and the revision)


	// Issue revision
	issuedRev, err := s.b2bService.IssueProforma(rev.ID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "SENT", issuedRev.Status)
	assert.Equal(s.T(), "PT/PI/26-27/001-R2", *issuedRev.ProformaNumber)

	// 5. Accept and Convert
	err = s.b2bService.AcceptProforma(issuedRev.ID)
	assert.NoError(s.T(), err)

	taxInv, err := s.b2bService.ConvertToTaxInvoice(issuedRev.ID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "DRAFT", taxInv.Status)
	assert.Equal(s.T(), issuedRev.ID, *taxInv.ProformaID)
	assert.Equal(s.T(), 5000.00, taxInv.AdvanceAdjusted)

	// Total price = 22000 + 18% GST (3960) = 25960.00
	assert.Equal(s.T(), 25960.00, taxInv.TotalPrice)
	// Balance = Total - AdvanceAdjusted = 25960.00 - 5000.00 = 20960.00
	assert.Equal(s.T(), 20960.00, taxInv.BalanceAmount)

	// Verify proforma status is now CONVERTED_TO_INVOICE
	updatedPf, err := s.b2bService.GetProformaByID(issuedRev.ID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "CONVERTED_TO_INVOICE", updatedPf.Status)
}

func (s *B2BProformaTestSuite) TestProformaExpiration() {
	cust := &entity.B2BCustomer{
		LegalName:      "Beta Industries",
		GSTIN:          "29ABCDE1234F1Z5",
		BillingAddress: "456 Blvd, Bangalore",
	}
	err := s.b2bService.CreateCustomer(cust)
	assert.NoError(s.T(), err)

	validUntil := time.Now().Add(-24 * time.Hour) // Yesterday
	pf := &entity.B2BProformaInvoice{
		CustomerID: &cust.ID,
		NoteDate:   time.Now().Add(-48 * time.Hour),
		ValidUntil: &validUntil,
		Items: []entity.B2BProformaInvoiceItem{
			{ItemDetails: "Vanilla Extract", Quantity: 1, Rate: 100.00},
		},
	}

	err = s.b2bService.CreateProforma(pf)
	assert.NoError(s.T(), err)

	issued, err := s.b2bService.IssueProforma(pf.ID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "SENT", issued.Status)

	expiredCount, err := s.b2bService.MarkExpiredProformas()
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int64(1), expiredCount)

	updatedPf, err := s.b2bService.GetProformaByID(pf.ID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "EXPIRED", updatedPf.Status)
}

func (s *B2BProformaTestSuite) TestProformaUnmarshal() {
	jsonData := `{"id": 42, "note_date": "2026-06-15", "status": "DRAFT", "customer_gstin": "29ABCDE1234F1Z5"}`
	var pf entity.B2BProformaInvoice
	err := json.Unmarshal([]byte(jsonData), &pf)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int64(42), pf.ID)
	assert.Equal(s.T(), "DRAFT", pf.Status)
	assert.Equal(s.T(), "29ABCDE1234F1Z5", pf.CustomerGSTIN)
}

func TestB2BProformaSuite(t *testing.T) {
	suite.Run(t, new(B2BProformaTestSuite))
}


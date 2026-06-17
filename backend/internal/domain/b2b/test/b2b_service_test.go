package test

import (
	"encoding/json"
	"fmt"
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
	assert.Equal(s.T(), "PT/26-27/001", *issued.InvoiceNumber)
}

func (s *B2BServiceTestSuite) TestB2BInvoiceJSONUnmarshal() {
	jsonData := `{
		"invoice_date": "2026-01-19",
		"customer_gstin": "33AAFCK8756M1ZY",
		"customer_name": "KAFA CLOTHING INDIA PRIVATE LIMITED",
		"due_date": "2026-01-19",
		"total_price": 86540.3
	}`

	var inv entity.B2BInvoice
	err := json.Unmarshal([]byte(jsonData), &inv)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 2026, inv.InvoiceDate.Year())
	assert.Equal(s.T(), time.Month(1), inv.InvoiceDate.Month())
	assert.Equal(s.T(), 19, inv.InvoiceDate.Day())
	assert.NotNil(s.T(), inv.DueDate)
	assert.Equal(s.T(), 2026, inv.DueDate.Year())
	assert.Equal(s.T(), time.Month(1), inv.DueDate.Month())
	assert.Equal(s.T(), 19, inv.DueDate.Day())
}

func (s *B2BServiceTestSuite) TestB2BProformaInvoiceJSONUnmarshal() {
	jsonData := `{
		"note_date": "2026-01-19T10:00:00Z",
		"valid_until": "2026-01-25"
	}`

	var pf entity.B2BProformaInvoice
	err := json.Unmarshal([]byte(jsonData), &pf)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 2026, pf.NoteDate.Year())
	assert.Equal(s.T(), time.Month(1), pf.NoteDate.Month())
	assert.Equal(s.T(), 19, pf.NoteDate.Day())
	assert.NotNil(s.T(), pf.ValidUntil)
	assert.Equal(s.T(), 2026, pf.ValidUntil.Year())
	assert.Equal(s.T(), time.Month(1), pf.ValidUntil.Month())
	assert.Equal(s.T(), 25, pf.ValidUntil.Day())
}

func (s *B2BServiceTestSuite) TestDeductAndRevertInventory() {
	cust := &entity.B2BCustomer{
		LegalName:      "Alpha Corp",
		GSTIN:          "33ABCDE1234F1Z5",
		BillingAddress: "123 Street, Chennai",
	}
	err := s.b2bService.CreateCustomer(cust)
	assert.NoError(s.T(), err)

	s.db.Exec("INSERT INTO inventory_items (id, mi_sku, title, current_stock, price, hsn_code, created_at, updated_at) VALUES (999, 'mi-test-sku', 'Test Product', 100, 50.0, '33029019', NOW(), NOW())")

	prodID := int64(999)
	skuStr := "mi-test-sku"
	inv := &entity.B2BInvoice{
		CustomerID:  &cust.ID,
		InvoiceDate: time.Now(),
		Items: []entity.B2BInvoiceItem{
			{
				ProductID:   &prodID,
				SKU:         &skuStr,
				ItemDetails: "Test Product Item",
				Quantity:    10,
				Rate:        50.0,
				Amount:      500.0,
			},
		},
	}
	err = s.b2bService.CreateInvoice(inv)
	assert.NoError(s.T(), err)

	issued, err := s.b2bService.IssueInvoice(inv.ID)
	assert.NoError(s.T(), err)
	assert.False(s.T(), issued.InventoryDeducted)

	err = s.b2bService.DeductInventory(issued.ID)
	assert.NoError(s.T(), err)

	var stock int
	s.db.Table("inventory_items").Where("id = 999").Pluck("current_stock", &stock)
	assert.Equal(s.T(), 90, stock)

	updated, err := s.b2bService.GetInvoiceByID(issued.ID)
	assert.NoError(s.T(), err)
	assert.True(s.T(), updated.InventoryDeducted)

	err = s.b2bService.RevertInventory(issued.ID)
	assert.NoError(s.T(), err)

	s.db.Table("inventory_items").Where("id = 999").Pluck("current_stock", &stock)
	assert.Equal(s.T(), 100, stock)

	updated, err = s.b2bService.GetInvoiceByID(issued.ID)
	assert.NoError(s.T(), err)
	assert.False(s.T(), updated.InventoryDeducted)

	s.db.Exec("DELETE FROM inventory_logs WHERE inventory_item_id = 999")
	s.db.Exec("DELETE FROM inventory_items WHERE id = 999")
}

func (s *B2BServiceTestSuite) TestInvoiceDeleteCascadesAndSync() {
	cust := &entity.B2BCustomer{
		LegalName:      "Delete Sync Corp",
		GSTIN:          "33ABCDE1234F1Z5",
		BillingAddress: "123 Street, Chennai",
	}
	err := s.b2bService.CreateCustomer(cust)
	assert.NoError(s.T(), err)

	inv := &entity.B2BInvoice{
		CustomerID:  &cust.ID,
		InvoiceDate: time.Now(),
		Items: []entity.B2BInvoiceItem{
			{ItemDetails: "Test Item", Quantity: 5, Rate: 100.0},
		},
	}
	err = s.b2bService.CreateInvoice(inv)
	assert.NoError(s.T(), err)

	var orderCount int64
	extID := fmt.Sprintf("B2B-%d", inv.ID)
	err = s.db.Table("orders").Where("source_id = ? AND external_order_id = ?", "b2b", extID).Count(&orderCount).Error
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int64(1), orderCount)

	var orderLineCount int64
	err = s.db.Table("order_line_items").Where("order_id IN (SELECT id FROM orders WHERE source_id = ? AND external_order_id = ?)", "b2b", extID).Count(&orderLineCount).Error
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int64(1), orderLineCount)

	cn := &entity.B2BCreditNote{
		InvoiceID:     &inv.ID,
		CustomerGSTIN: cust.GSTIN,
		CustomerName:  cust.LegalName,
		NoteDate:      time.Now(),
		Items: []entity.B2BCreditNoteItem{
			{ItemDetails: "Returned Item", Quantity: 1, Rate: 100.0},
		},
	}
	err = s.b2bService.CreateCreditNote(cn)
	assert.NoError(s.T(), err)

	dn := &entity.B2BDebitNote{
		InvoiceID:     &inv.ID,
		CustomerGSTIN: cust.GSTIN,
		CustomerName:  cust.LegalName,
		NoteDate:      time.Now(),
		Items: []entity.B2BDebitNoteItem{
			{ItemDetails: "Extra Charge", Quantity: 1, Rate: 20.0},
		},
	}
	err = s.b2bService.CreateDebitNote(dn)
	assert.NoError(s.T(), err)

	var cnCount, dnCount int64
	s.db.Table("b2b_credit_notes").Where("invoice_id = ?", inv.ID).Count(&cnCount)
	s.db.Table("b2b_debit_notes").Where("invoice_id = ?", inv.ID).Count(&dnCount)
	assert.Equal(s.T(), int64(1), cnCount)
	assert.Equal(s.T(), int64(1), dnCount)

	err = s.b2bService.DeleteInvoice(inv.ID)
	assert.NoError(s.T(), err)

	s.db.Table("b2b_credit_notes").Where("invoice_id = ?", inv.ID).Count(&cnCount)
	s.db.Table("b2b_debit_notes").Where("invoice_id = ?", inv.ID).Count(&dnCount)
	assert.Equal(s.T(), int64(0), cnCount)
	assert.Equal(s.T(), int64(0), dnCount)

	var cnItemCount, dnItemCount int64
	s.db.Table("b2b_credit_note_items").Where("credit_note_id = ?", cn.ID).Count(&cnItemCount)
	s.db.Table("b2b_debit_note_items").Where("debit_note_id = ?", dn.ID).Count(&dnItemCount)
	assert.Equal(s.T(), int64(0), cnItemCount)
	assert.Equal(s.T(), int64(0), dnItemCount)

	s.db.Table("orders").Where("source_id = ? AND external_order_id = ?", "b2b", extID).Count(&orderCount)
	assert.Equal(s.T(), int64(0), orderCount)
}

func TestB2BServiceSuite(t *testing.T) {
	suite.Run(t, new(B2BServiceTestSuite))
}

package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"mi-tech/internal/domain/b2b/entity"
	"mi-tech/internal/domain/b2b/helper"
)

// Invoices CRUD
func (s *B2BService) ListInvoices(startDate, endDate string, status string) ([]entity.B2BInvoice, error) {
	return s.repo.ListInvoices(startDate, endDate, status)
}

func (s *B2BService) GetInvoiceByID(id int64) (entity.B2BInvoice, error) {
	return s.repo.GetInvoiceByID(id)
}

func (s *B2BService) CreateInvoice(inv *entity.B2BInvoice) error {
	inv.Status = "DRAFT"
	inv.PaymentStatus = "UNPAID"
	inv.PaidAmount = 0.00
	
	if err := s.calculateInvoiceTotals(inv); err != nil {
		return err
	}
	return s.repo.CreateInvoice(inv)
}

func (s *B2BService) UpdateInvoice(inv *entity.B2BInvoice) error {
	existing, err := s.repo.GetInvoiceByID(inv.ID)
	if err != nil {
		return err
	}
	if existing.Status != "DRAFT" {
		return fmt.Errorf("invoice cannot be updated once it is in %s state", existing.Status)
	}

	inv.Status = "DRAFT"
	if err := s.calculateInvoiceTotals(inv); err != nil {
		return err
	}
	return s.repo.UpdateInvoice(inv)
}

func (s *B2BService) DeleteInvoice(id int64) error {
	return s.repo.DeleteInvoice(id)
}

// Transition Invoice status to ISSUED and assign invoice number
func (s *B2BService) IssueInvoice(id int64) (*entity.B2BInvoice, error) {
	var invoice entity.B2BInvoice
	err := s.db.Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)
		var err error
		invoice, err = txRepo.GetInvoiceByID(id)
		if err != nil {
			return err
		}

		if invoice.Status != "DRAFT" {
			return fmt.Errorf("invoice is already %s and cannot be issued", invoice.Status)
		}

		// Calculate fiscal year based on invoice date
		fy := helper.GetFinancialYear(invoice.InvoiceDate)
		seq, err := txRepo.GetNextSequenceForFY(fy)
		if err != nil {
			return err
		}

		invNumber := fmt.Sprintf("PT/%s/%03d", fy, seq)
		invoice.InvoiceNumber = &invNumber
		invoice.InvoiceSequence = &seq
		invoice.FinancialYear = &fy
		invoice.Status = "ISSUED"

		// Recalculate totals to be absolutely sure
		if err := s.calculateInvoiceTotals(&invoice); err != nil {
			return err
		}

		if err := txRepo.UpdateInvoice(&invoice); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return &invoice, nil
}

// Cancel invoice
func (s *B2BService) CancelInvoice(id int64) error {
	existing, err := s.repo.GetInvoiceByID(id)
	if err != nil {
		return err
	}
	if existing.Status != "ISSUED" {
		return fmt.Errorf("only ISSUED invoices can be CANCELLED; current status is %s", existing.Status)
	}

	existing.Status = "CANCELLED"
	return s.repo.UpdateInvoice(&existing)
}

// Update payment details
func (s *B2BService) UpdatePayment(id int64, paidAmount float64, method string) (*entity.B2BInvoice, error) {
	invoice, err := s.repo.GetInvoiceByID(id)
	if err != nil {
		return nil, err
	}
	if invoice.Status != "ISSUED" {
		return nil, fmt.Errorf("payments can only be registered on ISSUED invoices")
	}

	if paidAmount < 0 {
		return nil, fmt.Errorf("paid amount cannot be negative")
	}
	if paidAmount > invoice.TotalPrice {
		return nil, fmt.Errorf("paid amount %.2f exceeds total price %.2f", paidAmount, invoice.TotalPrice)
	}

	invoice.PaidAmount = paidAmount
	invoice.BalanceAmount = invoice.TotalPrice - paidAmount
	invoice.PaymentMethod = &method
	
	now := time.Now()
	invoice.PaymentDate = &now

	if invoice.BalanceAmount <= 0 {
		invoice.PaymentStatus = "PAID"
	} else if invoice.PaidAmount > 0 {
		invoice.PaymentStatus = "PARTIAL"
	} else {
		invoice.PaymentStatus = "UNPAID"
	}

	if err := s.repo.UpdateInvoice(&invoice); err != nil {
		return nil, err
	}
	return &invoice, nil
}

// Helpers
func (s *B2BService) calculateInvoiceTotals(inv *entity.B2BInvoice) error {
	// Populating Seller properties
	sellerGSTIN := strings.TrimSpace(strings.ToUpper(s.settings.GetBusinessGSTIN()))
	if sellerGSTIN == "" {
		return errors.New("business GSTIN is not configured in app settings")
	}
	inv.SellerGSTIN = sellerGSTIN
	inv.SellerName = s.settings.GetBusinessName()
	inv.SellerAddress = s.settings.GetBusinessAddressLine1() + " " + s.settings.GetBusinessAddressLine2()
	inv.SellerStateCode = sellerGSTIN[0:2]
	sellerState, err := s.getStateNameByCode(inv.SellerStateCode)
	if err != nil {
		return fmt.Errorf("failed to resolve seller state: %w", err)
	}
	inv.SellerState = sellerState

	// Retrieve client information to snap properties
	if inv.CustomerID != nil {
		cust, err := s.repo.GetCustomerByID(*inv.CustomerID)
		if err == nil {
			inv.CustomerGSTIN = cust.GSTIN
			inv.CustomerName = cust.LegalName
			inv.CustomerEmail = cust.Email
			inv.CustomerPhone = cust.Phone
			inv.CustomerState = cust.State
			inv.CustomerStateCode = cust.StateCode
			inv.CustomerAddress = cust.BillingAddress
		}
	}

	inv.CustomerGSTIN = strings.TrimSpace(strings.ToUpper(inv.CustomerGSTIN))
	if !helper.IsValidGSTIN(inv.CustomerGSTIN) {
		return fmt.Errorf("invalid customer GSTIN format: %s", inv.CustomerGSTIN)
	}
	inv.CustomerStateCode = inv.CustomerGSTIN[0:2]

	// Compute subtotal from items
	var subtotal float64
	for i := range inv.Items {
		inv.Items[i].Amount = inv.Items[i].Quantity * inv.Items[i].Rate
		subtotal += inv.Items[i].Amount
	}
	inv.SubtotalPrice = subtotal

	// Apply discount
	if inv.DiscountPercent > 0 {
		inv.DiscountAmount = (inv.SubtotalPrice * inv.DiscountPercent) / 100.00
	}
	taxableAmount := inv.SubtotalPrice - inv.DiscountAmount

	// Determine Tax splits
	var totalTax float64
	var defaultTaxRate float64 = 18.00 // Default B2B invoice tax rate is 18%
	
	// Reset tax splits
	inv.CGSTRate = 0
	inv.CGSTAmount = 0
	inv.SGSTRate = 0
	inv.SGSTAmount = 0
	inv.IGSTRate = 0
	inv.IGSTAmount = 0

	if inv.SellerStateCode == inv.CustomerStateCode {
		// Intra-state
		inv.CGSTRate = defaultTaxRate / 2.00
		inv.CGSTAmount = (taxableAmount * inv.CGSTRate) / 100.00
		inv.SGSTRate = defaultTaxRate / 2.00
		inv.SGSTAmount = (taxableAmount * inv.SGSTRate) / 100.00
		totalTax = inv.CGSTAmount + inv.SGSTAmount
	} else {
		// Inter-state
		inv.IGSTRate = defaultTaxRate
		inv.IGSTAmount = (taxableAmount * inv.IGSTRate) / 100.00
		totalTax = inv.IGSTAmount
	}

	// TDS/TCS calculation
	inv.TDSTCSAmount = 0
	if inv.TDSTCSType == "TDS" || inv.TDSTCSType == "TCS" {
		inv.TDSTCSAmount = (taxableAmount * inv.TDSTCSRate) / 100.00
	}

	// Final sum
	finalTotal := taxableAmount + totalTax + inv.TransportationCharge
	if inv.TDSTCSType == "TCS" {
		finalTotal += inv.TDSTCSAmount
	} else if inv.TDSTCSType == "TDS" {
		finalTotal -= inv.TDSTCSAmount
	}
	inv.TotalPrice = finalTotal
	inv.BalanceAmount = inv.TotalPrice - inv.PaidAmount

	return nil
}

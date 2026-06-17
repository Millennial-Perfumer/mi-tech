package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"mi-tech/internal/domain/b2b/entity"
	"mi-tech/internal/domain/b2b/helper"
	"mi-tech/internal/domain/b2b/repository"
)

// Credit Notes
func (s *B2BService) ListCreditNotes(invoiceID int64) ([]entity.B2BCreditNote, error) {
	return s.repo.ListCreditNotes(invoiceID)
}

func (s *B2BService) GetCreditNoteByID(id int64) (entity.B2BCreditNote, error) {
	return s.repo.GetCreditNoteByID(id)
}

func (s *B2BService) CreateCreditNote(cn *entity.B2BCreditNote) error {
	cn.Status = "DRAFT"
	if err := s.calculateCreditNoteTotals(cn); err != nil {
		return err
	}
	return s.repo.CreateCreditNote(cn)
}

func (s *B2BService) UpdateCreditNote(cn *entity.B2BCreditNote) error {
	existing, err := s.repo.GetCreditNoteByID(cn.ID)
	if err != nil {
		return err
	}
	if existing.Status != "DRAFT" {
		return fmt.Errorf("credit note cannot be updated once in %s status", existing.Status)
	}
	cn.Status = "DRAFT"
	if err := s.calculateCreditNoteTotals(cn); err != nil {
		return err
	}
	return s.repo.UpdateCreditNote(cn)
}

func (s *B2BService) DeleteCreditNote(id int64) error {
	return s.repo.DeleteCreditNote(id)
}

func (s *B2BService) IssueCreditNote(id int64) (*entity.B2BCreditNote, error) {
	var cn entity.B2BCreditNote
	err := s.db.Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)
		var err error
		cn, err = txRepo.GetCreditNoteByID(id)
		if err != nil {
			return err
		}

		if cn.Status != "DRAFT" {
			return fmt.Errorf("credit note is already %s and cannot be issued", cn.Status)
		}

		fy := helper.GetFinancialYear(cn.NoteDate)
		seq, err := txRepo.GetNextCreditNoteSequenceForFY(fy)
		if err != nil {
			return err
		}

		cnNumber := fmt.Sprintf("CN/%s/%05d", fy, seq)
		cn.CreditNoteNumber = &cnNumber
		cn.CreditNoteSequence = &seq
		cn.FinancialYear = &fy
		cn.Status = "ISSUED"

		if err := s.calculateCreditNoteTotals(&cn); err != nil {
			return err
		}

		if err := txRepo.UpdateCreditNote(&cn); err != nil {
			return err
		}

		// Adjust the invoice balance if linked
		if cn.InvoiceID != nil && *cn.InvoiceID > 0 {
			if err := s.UpdateInvoiceOutstanding(txRepo, *cn.InvoiceID); err != nil {
				return fmt.Errorf("failed to adjust invoice balance: %w", err)
			}
		}

		// Log financial audit trail
		audit := &entity.B2BFinancialAuditLog{
			Action:      "ISSUE_CREDIT_NOTE",
			EntityType:  "CREDIT_NOTE",
			EntityID:    cn.ID,
			UserID:      "system",
			Description: fmt.Sprintf("Issued Credit Note %s for Invoice %s, Total: %.2f", cnNumber, *cn.InvoiceNumber, cn.TotalPrice),
			NewValue:    "ISSUED",
		}
		_ = txRepo.SaveAuditLog(audit)

		return nil
	})

	if err != nil {
		return nil, err
	}
	return &cn, nil
}

func (s *B2BService) CancelCreditNote(id int64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)
		existing, err := txRepo.GetCreditNoteByID(id)
		if err != nil {
			return err
		}
		if existing.Status != "ISSUED" {
			return fmt.Errorf("only ISSUED credit notes can be cancelled")
		}

		existing.Status = "CANCELLED"
		if err := txRepo.UpdateCreditNote(&existing); err != nil {
			return err
		}

		// Re-adjust the invoice balance if linked
		if existing.InvoiceID != nil && *existing.InvoiceID > 0 {
			if err := s.UpdateInvoiceOutstanding(txRepo, *existing.InvoiceID); err != nil {
				return fmt.Errorf("failed to adjust invoice balance: %w", err)
			}
		}

		// Log financial audit trail
		audit := &entity.B2BFinancialAuditLog{
			Action:      "CANCEL_CREDIT_NOTE",
			EntityType:  "CREDIT_NOTE",
			EntityID:    existing.ID,
			UserID:      "system",
			Description: fmt.Sprintf("Cancelled Credit Note %s", *existing.CreditNoteNumber),
			OldValue:    "ISSUED",
			NewValue:    "CANCELLED",
		}
		_ = txRepo.SaveAuditLog(audit)

		return nil
	})
}

// Debit Notes
func (s *B2BService) ListDebitNotes(invoiceID int64) ([]entity.B2BDebitNote, error) {
	return s.repo.ListDebitNotes(invoiceID)
}

func (s *B2BService) GetDebitNoteByID(id int64) (entity.B2BDebitNote, error) {
	return s.repo.GetDebitNoteByID(id)
}

func (s *B2BService) CreateDebitNote(dn *entity.B2BDebitNote) error {
	dn.Status = "DRAFT"
	if err := s.calculateDebitNoteTotals(dn); err != nil {
		return err
	}
	return s.repo.CreateDebitNote(dn)
}

func (s *B2BService) UpdateDebitNote(dn *entity.B2BDebitNote) error {
	existing, err := s.repo.GetDebitNoteByID(dn.ID)
	if err != nil {
		return err
	}
	if existing.Status != "DRAFT" {
		return fmt.Errorf("debit note cannot be updated once in %s status", existing.Status)
	}
	dn.Status = "DRAFT"
	if err := s.calculateDebitNoteTotals(dn); err != nil {
		return err
	}
	return s.repo.UpdateDebitNote(dn)
}

func (s *B2BService) DeleteDebitNote(id int64) error {
	return s.repo.DeleteDebitNote(id)
}

func (s *B2BService) IssueDebitNote(id int64) (*entity.B2BDebitNote, error) {
	var dn entity.B2BDebitNote
	err := s.db.Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)
		var err error
		dn, err = txRepo.GetDebitNoteByID(id)
		if err != nil {
			return err
		}

		if dn.Status != "DRAFT" {
			return fmt.Errorf("debit note is already %s and cannot be issued", dn.Status)
		}

		fy := helper.GetFinancialYear(dn.NoteDate)
		seq, err := txRepo.GetNextDebitNoteSequenceForFY(fy)
		if err != nil {
			return err
		}

		dnNumber := fmt.Sprintf("DN/%s/%05d", fy, seq)
		dn.DebitNoteNumber = &dnNumber
		dn.DebitNoteSequence = &seq
		dn.FinancialYear = &fy
		dn.Status = "ISSUED"

		if err := s.calculateDebitNoteTotals(&dn); err != nil {
			return err
		}

		if err := txRepo.UpdateDebitNote(&dn); err != nil {
			return err
		}

		// Adjust the invoice balance if linked
		if dn.InvoiceID != nil && *dn.InvoiceID > 0 {
			if err := s.UpdateInvoiceOutstanding(txRepo, *dn.InvoiceID); err != nil {
				return fmt.Errorf("failed to adjust invoice balance: %w", err)
			}
		}

		// Log financial audit trail
		audit := &entity.B2BFinancialAuditLog{
			Action:      "ISSUE_DEBIT_NOTE",
			EntityType:  "DEBIT_NOTE",
			EntityID:    dn.ID,
			UserID:      "system",
			Description: fmt.Sprintf("Issued Debit Note %s for Invoice %s, Total: %.2f", dnNumber, *dn.InvoiceNumber, dn.TotalPrice),
			NewValue:    "ISSUED",
		}
		_ = txRepo.SaveAuditLog(audit)

		return nil
	})

	if err != nil {
		return nil, err
	}
	return &dn, nil
}

func (s *B2BService) CancelDebitNote(id int64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)
		existing, err := txRepo.GetDebitNoteByID(id)
		if err != nil {
			return err
		}
		if existing.Status != "ISSUED" {
			return fmt.Errorf("only ISSUED debit notes can be cancelled")
		}

		existing.Status = "CANCELLED"
		if err := txRepo.UpdateDebitNote(&existing); err != nil {
			return err
		}

		// Re-adjust the invoice balance if linked
		if existing.InvoiceID != nil && *existing.InvoiceID > 0 {
			if err := s.UpdateInvoiceOutstanding(txRepo, *existing.InvoiceID); err != nil {
				return fmt.Errorf("failed to adjust invoice balance: %w", err)
			}
		}

		// Log financial audit trail
		audit := &entity.B2BFinancialAuditLog{
			Action:      "CANCEL_DEBIT_NOTE",
			EntityType:  "DEBIT_NOTE",
			EntityID:    existing.ID,
			UserID:      "system",
			Description: fmt.Sprintf("Cancelled Debit Note %s", *existing.DebitNoteNumber),
			OldValue:    "ISSUED",
			NewValue:    "CANCELLED",
		}
		_ = txRepo.SaveAuditLog(audit)

		return nil
	})
}

// GST Periods
func (s *B2BService) ListGSTPeriods() ([]entity.GSTPeriod, error) {
	return s.repo.ListGSTPeriods()
}

func (s *B2BService) GetGSTPeriod(month, year int) (entity.GSTPeriod, error) {
	return s.repo.GetGSTPeriod(month, year)
}

func (s *B2BService) SaveGSTPeriod(p *entity.GSTPeriod) error {
	return s.repo.SaveGSTPeriod(p)
}

func (s *B2BService) IsPeriodLocked(t time.Time) (bool, error) {
	month := int(t.Month())
	year := t.Year()
	period, err := s.repo.GetGSTPeriod(month, year)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return period.Status == "LOCKED", nil
}

// Outstanding update logic
func (s *B2BService) UpdateInvoiceOutstanding(txRepo repository.B2BRepository, invoiceID int64) error {
	inv, err := txRepo.GetInvoiceByID(invoiceID)
	if err != nil {
		return err
	}

	cns, err := txRepo.ListCreditNotes(invoiceID)
	if err != nil {
		return err
	}
	var totalCN float64
	for _, cn := range cns {
		if cn.Status == "ISSUED" {
			totalCN += cn.TotalPrice
		}
	}

	dns, err := txRepo.ListDebitNotes(invoiceID)
	if err != nil {
		return err
	}
	var totalDN float64
	for _, dn := range dns {
		if dn.Status == "ISSUED" {
			totalDN += dn.TotalPrice
		}
	}

	inv.BalanceAmount = (inv.TotalPrice + totalDN) - inv.PaidAmount - totalCN
	if inv.BalanceAmount <= 0 {
		inv.PaymentStatus = "PAID"
	} else if inv.PaidAmount > 0 || totalCN > 0 || totalDN > 0 {
		inv.PaymentStatus = "PARTIAL"
	} else {
		inv.PaymentStatus = "UNPAID"
	}
	return txRepo.UpdateInvoice(&inv)
}

// Helpers
func (s *B2BService) calculateCreditNoteTotals(cn *entity.B2BCreditNote) error {
	sellerGSTIN := strings.TrimSpace(strings.ToUpper(s.settings.GetBusinessGSTIN()))
	if sellerGSTIN == "" {
		return errors.New("business GSTIN is not configured")
	}
	cn.SellerGSTIN = sellerGSTIN
	cn.SellerName = s.settings.GetBusinessName()
	cn.SellerAddress = s.settings.GetBusinessAddressLine1() + " " + s.settings.GetBusinessAddressLine2()
	cn.SellerStateCode = sellerGSTIN[0:2]
	sellerState, _ := s.getStateNameByCode(cn.SellerStateCode)
	cn.SellerState = sellerState

	if cn.InvoiceID != nil && *cn.InvoiceID > 0 {
		inv, err := s.repo.GetInvoiceByID(*cn.InvoiceID)
		if err == nil {
			cn.InvoiceNumber = inv.InvoiceNumber
			cn.CustomerID = inv.CustomerID
			cn.CustomerGSTIN = inv.CustomerGSTIN
			cn.CustomerName = inv.CustomerName
			cn.CustomerEmail = inv.CustomerEmail
			cn.CustomerPhone = inv.CustomerPhone
			cn.CustomerState = inv.CustomerState
			cn.CustomerStateCode = inv.CustomerStateCode
			cn.CustomerAddress = inv.CustomerAddress
		}
	}

	cn.CustomerGSTIN = strings.TrimSpace(strings.ToUpper(cn.CustomerGSTIN))
	cn.CustomerStateCode = cn.CustomerGSTIN[0:2]

	var subtotal float64
	for i := range cn.Items {
		cn.Items[i].Amount = cn.Items[i].Quantity * cn.Items[i].Rate
		subtotal += cn.Items[i].Amount
	}
	cn.SubtotalPrice = subtotal

	if cn.DiscountPercent > 0 {
		cn.DiscountAmount = (cn.SubtotalPrice * cn.DiscountPercent) / 100.00
	}
	taxableAmount := cn.SubtotalPrice - cn.DiscountAmount

	var totalTax float64
	var defaultTaxRate float64 = 18.00

	cn.CGSTRate = 0
	cn.CGSTAmount = 0
	cn.SGSTRate = 0
	cn.SGSTAmount = 0
	cn.IGSTRate = 0
	cn.IGSTAmount = 0

	if cn.SellerStateCode == cn.CustomerStateCode {
		cn.CGSTRate = defaultTaxRate / 2.00
		cn.CGSTAmount = (taxableAmount * cn.CGSTRate) / 100.00
		cn.SGSTRate = defaultTaxRate / 2.00
		cn.SGSTAmount = (taxableAmount * cn.SGSTRate) / 100.00
		totalTax = cn.CGSTAmount + cn.SGSTAmount
	} else {
		cn.IGSTRate = defaultTaxRate
		cn.IGSTAmount = (taxableAmount * cn.IGSTRate) / 100.00
		totalTax = cn.IGSTAmount
	}

	cn.TotalPrice = taxableAmount + totalTax
	return nil
}

func (s *B2BService) calculateDebitNoteTotals(dn *entity.B2BDebitNote) error {
	sellerGSTIN := strings.TrimSpace(strings.ToUpper(s.settings.GetBusinessGSTIN()))
	if sellerGSTIN == "" {
		return errors.New("business GSTIN is not configured")
	}
	dn.SellerGSTIN = sellerGSTIN
	dn.SellerName = s.settings.GetBusinessName()
	dn.SellerAddress = s.settings.GetBusinessAddressLine1() + " " + s.settings.GetBusinessAddressLine2()
	dn.SellerStateCode = sellerGSTIN[0:2]
	sellerState, _ := s.getStateNameByCode(dn.SellerStateCode)
	dn.SellerState = sellerState

	if dn.InvoiceID != nil && *dn.InvoiceID > 0 {
		inv, err := s.repo.GetInvoiceByID(*dn.InvoiceID)
		if err == nil {
			dn.InvoiceNumber = inv.InvoiceNumber
			dn.CustomerID = inv.CustomerID
			dn.CustomerGSTIN = inv.CustomerGSTIN
			dn.CustomerName = inv.CustomerName
			dn.CustomerEmail = inv.CustomerEmail
			dn.CustomerPhone = inv.CustomerPhone
			dn.CustomerState = inv.CustomerState
			dn.CustomerStateCode = inv.CustomerStateCode
			dn.CustomerAddress = inv.CustomerAddress
		}
	}

	dn.CustomerGSTIN = strings.TrimSpace(strings.ToUpper(dn.CustomerGSTIN))
	dn.CustomerStateCode = dn.CustomerGSTIN[0:2]

	var subtotal float64
	for i := range dn.Items {
		dn.Items[i].Amount = dn.Items[i].Quantity * dn.Items[i].Rate
		subtotal += dn.Items[i].Amount
	}
	dn.SubtotalPrice = subtotal

	if dn.DiscountPercent > 0 {
		dn.DiscountAmount = (dn.SubtotalPrice * dn.DiscountPercent) / 100.00
	}
	taxableAmount := dn.SubtotalPrice - dn.DiscountAmount

	var totalTax float64
	var defaultTaxRate float64 = 18.00

	dn.CGSTRate = 0
	dn.CGSTAmount = 0
	dn.SGSTRate = 0
	dn.SGSTAmount = 0
	dn.IGSTRate = 0
	dn.IGSTAmount = 0

	if dn.SellerStateCode == dn.CustomerStateCode {
		dn.CGSTRate = defaultTaxRate / 2.00
		dn.CGSTAmount = (taxableAmount * dn.CGSTRate) / 100.00
		dn.SGSTRate = defaultTaxRate / 2.00
		dn.SGSTAmount = (taxableAmount * dn.SGSTRate) / 100.00
		totalTax = dn.CGSTAmount + dn.SGSTAmount
	} else {
		dn.IGSTRate = defaultTaxRate
		dn.IGSTAmount = (taxableAmount * dn.IGSTRate) / 100.00
		totalTax = dn.IGSTAmount
	}

	dn.TotalPrice = taxableAmount + totalTax
	return nil
}

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

// ListProformas lists proformas with filters
func (s *B2BService) ListProformas(startDate, endDate string, status string) ([]entity.B2BProformaInvoice, error) {
	return s.repo.ListProformas(startDate, endDate, status)
}

// GetProformaByID retrieves a single proforma invoice
func (s *B2BService) GetProformaByID(id int64) (entity.B2BProformaInvoice, error) {
	return s.repo.GetProformaByID(id)
}

// CreateProforma creates a new proforma invoice in DRAFT status
func (s *B2BService) CreateProforma(pf *entity.B2BProformaInvoice) error {
	pf.Status = "DRAFT"
	pf.RevisionNumber = 1

	locked, err := s.IsPeriodLocked(pf.NoteDate)
	if err == nil && locked {
		return fmt.Errorf("cannot create proforma invoice in a locked GST filing period")
	}

	if err := s.calculateProformaTotals(pf); err != nil {
		return err
	}
	return s.repo.CreateProforma(pf)
}

// UpdateProforma updates an existing DRAFT proforma invoice
func (s *B2BService) UpdateProforma(pf *entity.B2BProformaInvoice) error {
	existing, err := s.repo.GetProformaByID(pf.ID)
	if err != nil {
		return err
	}
	if existing.Status != "DRAFT" {
		return fmt.Errorf("proforma invoice cannot be updated once it is in %s state", existing.Status)
	}

	locked, err := s.IsPeriodLocked(existing.NoteDate)
	if err == nil && locked {
		return fmt.Errorf("cannot modify proforma invoice in a locked GST filing period")
	}

	pf.Status = "DRAFT"
	if err := s.calculateProformaTotals(pf); err != nil {
		return err
	}
	return s.repo.UpdateProforma(pf)
}

// DeleteProforma deletes a DRAFT proforma invoice
func (s *B2BService) DeleteProforma(id int64) error {
	existing, err := s.repo.GetProformaByID(id)
	if err == nil {
		locked, err := s.IsPeriodLocked(existing.NoteDate)
		if err == nil && locked {
			return fmt.Errorf("cannot delete proforma invoice from a locked GST filing period")
		}
	}
	return s.repo.DeleteProforma(id)
}

// GetNextProformaNumber returns the next proforma number sequence
func (s *B2BService) GetNextProformaNumber(pfDate string) (string, error) {
	var t time.Time
	var err error
	if pfDate != "" {
		t, err = time.Parse("2006-01-02", pfDate)
		if err != nil {
			t = time.Now()
		}
	} else {
		t = time.Now()
	}
	fy := helper.GetFinancialYear(t)
	seq, err := s.repo.GetNextProformaSequenceForFY(fy)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("PT/PI/%s/%03d", fy, seq), nil
}

// IssueProforma transitions status to SENT and assigns sequential number
func (s *B2BService) IssueProforma(id int64) (*entity.B2BProformaInvoice, error) {
	var pf entity.B2BProformaInvoice
	err := s.db.Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)
		var err error
		pf, err = txRepo.GetProformaByID(id)
		if err != nil {
			return err
		}

		if pf.Status != "DRAFT" {
			return fmt.Errorf("proforma is already %s and cannot be issued", pf.Status)
		}

		locked, err := s.IsPeriodLocked(pf.NoteDate)
		if err == nil && locked {
			return fmt.Errorf("cannot issue proforma in a locked GST filing period")
		}

		// Calculate number sequence
		fy := helper.GetFinancialYear(pf.NoteDate)
		
		if pf.ParentProformaID != nil {
			// It's a revision! Use parent's number and sequence with revision suffix
			var parent entity.B2BProformaInvoice
			if err := tx.First(&parent, *pf.ParentProformaID).Error; err != nil {
				return fmt.Errorf("failed to fetch parent proforma: %w", err)
			}
			
			parentNum := ""
			if parent.ProformaNumber != nil {
				parentNum = *parent.ProformaNumber
				// Strip any existing revision suffix if there is one
				if idx := strings.Index(parentNum, "-R"); idx != -1 {
					parentNum = parentNum[:idx]
				}
			} else {
				parentNum = fmt.Sprintf("PT/PI/%s/%03d", fy, 0)
			}

			pfNum := fmt.Sprintf("%s-R%d", parentNum, pf.RevisionNumber)
			pf.ProformaNumber = &pfNum
			pf.ProformaSequence = parent.ProformaSequence
		} else {
			// Standard sequence
			seq, err := txRepo.GetNextProformaSequenceForFY(fy)
			if err != nil {
				return err
			}
			pfNum := fmt.Sprintf("PT/PI/%s/%03d", fy, seq)
			pf.ProformaNumber = &pfNum
			pf.ProformaSequence = &seq
		}

		pf.FinancialYear = &fy
		pf.Status = "SENT"

		if err := s.calculateProformaTotals(&pf); err != nil {
			return err
		}

		if err := txRepo.UpdateProforma(&pf); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return &pf, nil
}

// AcceptProforma transitions status to ACCEPTED
func (s *B2BService) AcceptProforma(id int64) error {
	existing, err := s.repo.GetProformaByID(id)
	if err != nil {
		return err
	}
	if existing.Status != "SENT" {
		return fmt.Errorf("only SENT proformas can be ACCEPTED; current status is %s", existing.Status)
	}

	locked, err := s.IsPeriodLocked(existing.NoteDate)
	if err == nil && locked {
		return fmt.Errorf("cannot accept proforma in a locked GST filing period")
	}

	existing.Status = "ACCEPTED"
	return s.repo.UpdateProforma(&existing)
}

// RejectProforma transitions status to REJECTED
func (s *B2BService) RejectProforma(id int64) error {
	existing, err := s.repo.GetProformaByID(id)
	if err != nil {
		return err
	}
	if existing.Status != "SENT" {
		return fmt.Errorf("only SENT proformas can be REJECTED; current status is %s", existing.Status)
	}

	locked, err := s.IsPeriodLocked(existing.NoteDate)
	if err == nil && locked {
		return fmt.Errorf("cannot reject proforma in a locked GST filing period")
	}

	existing.Status = "REJECTED"
	return s.repo.UpdateProforma(&existing)
}

// CancelProforma transitions status to CANCELLED
func (s *B2BService) CancelProforma(id int64) error {
	existing, err := s.repo.GetProformaByID(id)
	if err != nil {
		return err
	}
	if existing.Status != "SENT" && existing.Status != "ACCEPTED" {
		return fmt.Errorf("only SENT or ACCEPTED proformas can be CANCELLED; current status is %s", existing.Status)
	}

	locked, err := s.IsPeriodLocked(existing.NoteDate)
	if err == nil && locked {
		return fmt.Errorf("cannot cancel proforma in a locked GST filing period")
	}

	existing.Status = "CANCELLED"
	return s.repo.UpdateProforma(&existing)
}

// CreateRevision creates a new revision draft from a SENT or REJECTED proforma
func (s *B2BService) CreateRevision(id int64) (*entity.B2BProformaInvoice, error) {
	var revision entity.B2BProformaInvoice
	err := s.db.Transaction(func(tx *gorm.DB) error {
		existing, err := s.repo.WithTx(tx).GetProformaByID(id)
		if err != nil {
			return err
		}

		if existing.Status == "DRAFT" || existing.Status == "CONVERTED_TO_INVOICE" {
			return fmt.Errorf("cannot revise proforma in %s status", existing.Status)
		}

		locked, err := s.IsPeriodLocked(existing.NoteDate)
		if err == nil && locked {
			return fmt.Errorf("cannot revise proforma in a locked GST filing period")
		}

		// Find parent ID to group revision tree
		parentID := existing.ID
		if existing.ParentProformaID != nil {
			parentID = *existing.ParentProformaID
		}

		// Count existing revisions to set revision number
		var count int64
		err = tx.Model(&entity.B2BProformaInvoice{}).
			Where("id = ? OR parent_proforma_id = ?", parentID, parentID).
			Count(&count).Error
		if err != nil {
			return err
		}

		revision = entity.B2BProformaInvoice{
			Status:            "DRAFT",
			RevisionNumber:    int(count) + 1,
			ParentProformaID:  &parentID,
			NoteDate:          existing.NoteDate,
			ValidUntil:        existing.ValidUntil,
			CustomerID:        existing.CustomerID,
			CustomerGSTIN:     existing.CustomerGSTIN,
			CustomerName:      existing.CustomerName,
			CustomerEmail:     existing.CustomerEmail,
			CustomerPhone:     existing.CustomerPhone,
			CustomerState:     existing.CustomerState,
			CustomerStateCode: existing.CustomerStateCode,
			CustomerAddress:   existing.CustomerAddress,
			CustomerShippingAddress: existing.CustomerShippingAddress,
			SellerGSTIN:       existing.SellerGSTIN,
			SellerName:        existing.SellerName,
			SellerState:       existing.SellerState,
			SellerStateCode:   existing.SellerStateCode,
			SellerAddress:     existing.SellerAddress,
			SubtotalPrice:     existing.SubtotalPrice,
			DiscountPercent:   existing.DiscountPercent,
			DiscountAmount:    existing.DiscountAmount,
			CGSTRate:          existing.CGSTRate,
			CGSTAmount:        existing.CGSTAmount,
			SGSTRate:          existing.SGSTRate,
			SGSTAmount:        existing.SGSTAmount,
			IGSTRate:          existing.IGSTRate,
			IGSTAmount:        existing.IGSTAmount,
			TotalPrice:        existing.TotalPrice,
			AdvancePaid:       existing.AdvancePaid,
		}

		for _, item := range existing.Items {
			revision.Items = append(revision.Items, entity.B2BProformaInvoiceItem{
				ProductID:   item.ProductID,
				ItemDetails: item.ItemDetails,
				SKU:         item.SKU,
				HSNCode:     item.HSNCode,
				Quantity:    item.Quantity,
				Rate:        item.Rate,
				Amount:      item.Amount,
			})
		}

		if err := s.repo.WithTx(tx).CreateProforma(&revision); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return &revision, nil
}

// ConvertToTaxInvoice converts an ACCEPTED/SENT proforma invoice to a draft Tax Invoice
func (s *B2BService) ConvertToTaxInvoice(id int64) (*entity.B2BInvoice, error) {
	var invoice entity.B2BInvoice
	err := s.db.Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)
		pf, err := txRepo.GetProformaByID(id)
		if err != nil {
			return err
		}

		if pf.Status == "CONVERTED_TO_INVOICE" {
			return fmt.Errorf("proforma invoice is already converted to a tax invoice")
		}
		if pf.Status != "ACCEPTED" && pf.Status != "SENT" {
			return fmt.Errorf("only ACCEPTED or SENT proforma invoices can be converted; current status is %s", pf.Status)
		}

		locked, err := s.IsPeriodLocked(pf.NoteDate)
		if err == nil && locked {
			return fmt.Errorf("cannot convert proforma in a locked GST filing period")
		}

		// Map to B2BInvoice
		invoice = entity.B2BInvoice{
			InvoiceDate:       time.Now(),
			Status:            "DRAFT",
			PaymentStatus:     "UNPAID",
			CustomerID:        pf.CustomerID,
			CustomerGSTIN:     pf.CustomerGSTIN,
			CustomerName:      pf.CustomerName,
			CustomerEmail:     pf.CustomerEmail,
			CustomerPhone:     pf.CustomerPhone,
			CustomerState:     pf.CustomerState,
			CustomerStateCode: pf.CustomerStateCode,
			CustomerAddress:   pf.CustomerAddress,
			CustomerShippingAddress: pf.CustomerShippingAddress,
			SellerGSTIN:       pf.SellerGSTIN,
			SellerName:        pf.SellerName,
			SellerState:       pf.SellerState,
			SellerStateCode:   pf.SellerStateCode,
			SellerAddress:     pf.SellerAddress,
			SubtotalPrice:     pf.SubtotalPrice,
			DiscountPercent:   pf.DiscountPercent,
			DiscountAmount:    pf.DiscountAmount,
			CGSTRate:          pf.CGSTRate,
			CGSTAmount:        pf.CGSTAmount,
			SGSTRate:          pf.SGSTRate,
			SGSTAmount:        pf.SGSTAmount,
			IGSTRate:          pf.IGSTRate,
			IGSTAmount:        pf.IGSTAmount,
			TotalPrice:        pf.TotalPrice,
			ProformaID:        &pf.ID,
			AdvanceAdjusted:   pf.AdvancePaid,
		}

		// Adjust BalanceAmount and PaymentStatus based on AdvancePaid
		if invoice.AdvanceAdjusted > 0 {
			invoice.BalanceAmount = invoice.TotalPrice - invoice.AdvanceAdjusted
			if invoice.BalanceAmount <= 0 {
				invoice.PaymentStatus = "PAID"
				invoice.PaidAmount = invoice.TotalPrice // Fully paid by advance
			} else {
				invoice.PaymentStatus = "PARTIAL"
				invoice.PaidAmount = invoice.AdvanceAdjusted
			}
		} else {
			invoice.BalanceAmount = invoice.TotalPrice
		}

		for _, item := range pf.Items {
			invoice.Items = append(invoice.Items, entity.B2BInvoiceItem{
				ProductID:   item.ProductID,
				ItemDetails: item.ItemDetails,
				SKU:         item.SKU,
				HSNCode:     item.HSNCode,
				Quantity:    item.Quantity,
				Rate:        item.Rate,
				Amount:      item.Amount,
			})
		}

		if err := txRepo.CreateInvoice(&invoice); err != nil {
			return err
		}

		// Mark proforma as converted
		pf.Status = "CONVERTED_TO_INVOICE"
		if err := txRepo.UpdateProforma(&pf); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return &invoice, nil
}

// MarkExpiredProformas marks proforma invoices as EXPIRED if their valid_until date is in the past
func (s *B2BService) MarkExpiredProformas() (int64, error) {
	tx := s.db.Model(&entity.B2BProformaInvoice{}).
		Where("status IN ('DRAFT', 'SENT') AND valid_until < CURRENT_DATE").
		Update("status", "EXPIRED")
	return tx.RowsAffected, tx.Error
}

// Helper calculations
func (s *B2BService) calculateProformaTotals(pf *entity.B2BProformaInvoice) error {
	sellerGSTIN := strings.TrimSpace(strings.ToUpper(s.settings.GetBusinessGSTIN()))
	if sellerGSTIN == "" {
		return errors.New("business GSTIN is not configured in app settings")
	}
	pf.SellerGSTIN = sellerGSTIN
	pf.SellerName = s.settings.GetBusinessName()
	pf.SellerAddress = s.settings.GetBusinessAddressLine1() + " " + s.settings.GetBusinessAddressLine2()
	pf.SellerStateCode = sellerGSTIN[0:2]
	sellerState, err := s.getStateNameByCode(pf.SellerStateCode)
	if err != nil {
		return fmt.Errorf("failed to resolve seller state: %w", err)
	}
	pf.SellerState = sellerState

	if pf.CustomerID != nil {
		cust, err := s.repo.GetCustomerByID(*pf.CustomerID)
		if err == nil {
			if pf.CustomerGSTIN == "" {
				pf.CustomerGSTIN = cust.GSTIN
			}
			if pf.CustomerName == "" {
				pf.CustomerName = cust.LegalName
			}
			if pf.CustomerEmail == nil || *pf.CustomerEmail == "" {
				pf.CustomerEmail = cust.Email
			}
			if pf.CustomerPhone == nil || *pf.CustomerPhone == "" {
				pf.CustomerPhone = cust.Phone
			}
			if pf.CustomerState == "" {
				pf.CustomerState = cust.State
			}
			if pf.CustomerStateCode == "" {
				pf.CustomerStateCode = cust.StateCode
			}
			if pf.CustomerAddress == "" {
				pf.CustomerAddress = cust.BillingAddress
			}
			if pf.CustomerShippingAddress == "" && cust.ShippingAddress != nil {
				pf.CustomerShippingAddress = *cust.ShippingAddress
			}
		}
	}

	pf.CustomerGSTIN = strings.TrimSpace(strings.ToUpper(pf.CustomerGSTIN))
	if !helper.IsValidGSTIN(pf.CustomerGSTIN) {
		return fmt.Errorf("invalid customer GSTIN format: %s", pf.CustomerGSTIN)
	}
	pf.CustomerStateCode = pf.CustomerGSTIN[0:2]

	var subtotal float64
	for i := range pf.Items {
		pf.Items[i].Amount = pf.Items[i].Quantity * pf.Items[i].Rate
		subtotal += pf.Items[i].Amount
	}
	pf.SubtotalPrice = subtotal

	if pf.DiscountPercent > 0 {
		pf.DiscountAmount = (pf.SubtotalPrice * pf.DiscountPercent) / 100.00
	}
	taxableAmount := pf.SubtotalPrice - pf.DiscountAmount

	var totalTax float64
	var defaultTaxRate float64 = 18.00

	pf.CGSTRate = 0
	pf.CGSTAmount = 0
	pf.SGSTRate = 0
	pf.SGSTAmount = 0
	pf.IGSTRate = 0
	pf.IGSTAmount = 0

	if pf.SellerStateCode == pf.CustomerStateCode {
		pf.CGSTRate = defaultTaxRate / 2.00
		pf.CGSTAmount = (taxableAmount * pf.CGSTRate) / 100.00
		pf.SGSTRate = defaultTaxRate / 2.00
		pf.SGSTAmount = (taxableAmount * pf.SGSTRate) / 100.00
		totalTax = pf.CGSTAmount + pf.SGSTAmount
	} else {
		pf.IGSTRate = defaultTaxRate
		pf.IGSTAmount = (taxableAmount * pf.IGSTRate) / 100.00
		totalTax = pf.IGSTAmount
	}

	pf.TotalPrice = taxableAmount + totalTax
	return nil
}

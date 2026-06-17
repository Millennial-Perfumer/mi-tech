package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"mi-tech/internal/domain/b2b/entity"
	"mi-tech/internal/domain/b2b/helper"
	inventoryEntity "mi-tech/internal/domain/inventory/entity"
)

// Invoices CRUD
func (s *B2BService) ListInvoices(startDate, endDate string, status string) ([]entity.B2BInvoice, error) {
	return s.repo.ListInvoices(startDate, endDate, status)
}

func (s *B2BService) GetInvoiceByID(id int64) (entity.B2BInvoice, error) {
	return s.repo.GetInvoiceByID(id)
}

// GetNextInvoiceNumber returns what the next invoice number would be for the given date's fiscal year.
// This is a preview — the actual atomic assignment happens inside IssueInvoice.
func (s *B2BService) GetNextInvoiceNumber(invoiceDate string) (string, error) {
	var t time.Time
	var err error
	if invoiceDate != "" {
		t, err = time.Parse("2006-01-02", invoiceDate)
		if err != nil {
			t = time.Now()
		}
	} else {
		t = time.Now()
	}
	fy := helper.GetFinancialYear(t)
	seq, err := s.repo.GetNextSequenceForFY(fy)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("PT/%s/%03d", fy, seq), nil
}

func (s *B2BService) CreateInvoice(inv *entity.B2BInvoice) error {
	inv.Status = "DRAFT"
	inv.PaymentStatus = "UNPAID"
	inv.PaidAmount = 0.00

	locked, err := s.IsPeriodLocked(inv.InvoiceDate)
	if err == nil && locked {
		return fmt.Errorf("cannot create invoice in a locked GST filing period")
	}

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

	locked, err := s.IsPeriodLocked(existing.InvoiceDate)
	if err == nil && locked {
		return fmt.Errorf("cannot modify invoice in a locked GST filing period")
	}

	inv.Status = "DRAFT"
	if err := s.calculateInvoiceTotals(inv); err != nil {
		return err
	}
	return s.repo.UpdateInvoice(inv)
}

func (s *B2BService) DeleteInvoice(id int64) error {
	existing, err := s.repo.GetInvoiceByID(id)
	if err == nil {
		locked, err := s.IsPeriodLocked(existing.InvoiceDate)
		if err == nil && locked {
			return fmt.Errorf("cannot delete invoice from a locked GST filing period")
		}
	}
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

		locked, err := s.IsPeriodLocked(invoice.InvoiceDate)
		if err == nil && locked {
			return fmt.Errorf("cannot issue invoice in a locked GST filing period")
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

	locked, err := s.IsPeriodLocked(existing.InvoiceDate)
	if err == nil && locked {
		return fmt.Errorf("cannot cancel invoice in a locked GST filing period")
	}

	existing.Status = "CANCELLED"
	return s.repo.UpdateInvoice(&existing)
}

// Update payment details - Allowed even if period is locked
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

	// Recalculate based on CN/DN dynamically to keep balance exact
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

	// Retrieve client information to snap properties (only overwrite if empty to allow custom overrides)
	if inv.CustomerID != nil {
		cust, err := s.repo.GetCustomerByID(*inv.CustomerID)
		if err == nil {
			if inv.CustomerGSTIN == "" {
				inv.CustomerGSTIN = cust.GSTIN
			}
			if inv.CustomerName == "" {
				inv.CustomerName = cust.LegalName
			}
			if inv.CustomerEmail == nil || *inv.CustomerEmail == "" {
				inv.CustomerEmail = cust.Email
			}
			if inv.CustomerPhone == nil || *inv.CustomerPhone == "" {
				inv.CustomerPhone = cust.Phone
			}
			if inv.CustomerState == "" {
				inv.CustomerState = cust.State
			}
			if inv.CustomerStateCode == "" {
				inv.CustomerStateCode = cust.StateCode
			}
			if inv.CustomerAddress == "" {
				inv.CustomerAddress = cust.BillingAddress
			}
			if inv.CustomerShippingAddress == "" && cust.ShippingAddress != nil {
				inv.CustomerShippingAddress = *cust.ShippingAddress
			}
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
	inv.BalanceAmount = inv.TotalPrice - inv.PaidAmount - inv.AdvanceAdjusted

	return nil
}

// DeductInventory manually deducts warehouse stock levels for an issued B2B invoice's items.
func (s *B2BService) DeductInventory(id int64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)
		invoice, err := txRepo.GetInvoiceByID(id)
		if err != nil {
			return err
		}

		if invoice.Status != "ISSUED" {
			return fmt.Errorf("inventory can only be deducted for ISSUED invoices")
		}

		if invoice.InventoryDeducted {
			return fmt.Errorf("inventory has already been deducted for this invoice")
		}

		locked, err := s.IsPeriodLocked(invoice.InvoiceDate)
		if err == nil && locked {
			return fmt.Errorf("cannot deduct inventory in a locked GST filing period")
		}

		// Perform deduction for each item
		for _, item := range invoice.Items {
			var itemID int
			var found bool

			// Try to find the inventory item
			if item.ProductID != nil && *item.ProductID > 0 {
				var count int64
				if err := tx.Model(&inventoryEntity.InventoryItem{}).Where("id = ?", *item.ProductID).Count(&count).Error; err == nil && count > 0 {
					itemID = int(*item.ProductID)
					found = true
				}
			}

			if !found && item.SKU != nil && *item.SKU != "" {
				var invItem inventoryEntity.InventoryItem
				if err := tx.Where("mi_sku = ?", *item.SKU).First(&invItem).Error; err == nil {
					itemID = invItem.ID
					found = true
				}
			}

			if !found {
				return fmt.Errorf("could not find warehouse product for item %q (SKU: %v)", item.ItemDetails, item.SKU)
			}

			// Adjust stock: subtract quantity
			qtyInt := int(item.Quantity)
			if qtyInt <= 0 {
				continue // skip zero or negative quantity items
			}

			if err := tx.Model(&inventoryEntity.InventoryItem{}).
				Where("id = ?", itemID).
				Update("current_stock", gorm.Expr("current_stock - ?", qtyInt)).Error; err != nil {
				return fmt.Errorf("failed to deduct stock for product ID %d: %w", itemID, err)
			}

			// Write to inventory_logs
			logEntry := inventoryEntity.InventoryLog{
				InventoryItemID: itemID,
				Delta:           -qtyInt,
				Reason:          "sale",
				Platform:        "B2B",
				ExternalOrderID: invoice.InvoiceNumber,
				CreatedAt:       time.Now(),
			}
			if err := tx.Create(&logEntry).Error; err != nil {
				return fmt.Errorf("failed to log stock deduction for product ID %d: %w", itemID, err)
			}
		}

		// Mark invoice as deducted
		invoice.InventoryDeducted = true
		if err := txRepo.UpdateInvoice(&invoice); err != nil {
			return fmt.Errorf("failed to save invoice inventory deduction status: %w", err)
		}

		return nil
	})
}

// RevertInventory manually reverts warehouse stock levels for an issued B2B invoice's items.
func (s *B2BService) RevertInventory(id int64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)
		invoice, err := txRepo.GetInvoiceByID(id)
		if err != nil {
			return err
		}

		if invoice.Status != "ISSUED" {
			return fmt.Errorf("inventory can only be reverted for ISSUED invoices")
		}

		if !invoice.InventoryDeducted {
			return fmt.Errorf("inventory has not been deducted for this invoice")
		}

		locked, err := s.IsPeriodLocked(invoice.InvoiceDate)
		if err == nil && locked {
			return fmt.Errorf("cannot revert inventory in a locked GST filing period")
		}

		// Perform restocking for each item
		for _, item := range invoice.Items {
			var itemID int
			var found bool

			// Try to find the inventory item
			if item.ProductID != nil && *item.ProductID > 0 {
				var count int64
				if err := tx.Model(&inventoryEntity.InventoryItem{}).Where("id = ?", *item.ProductID).Count(&count).Error; err == nil && count > 0 {
					itemID = int(*item.ProductID)
					found = true
				}
			}

			if !found && item.SKU != nil && *item.SKU != "" {
				var invItem inventoryEntity.InventoryItem
				if err := tx.Where("mi_sku = ?", *item.SKU).First(&invItem).Error; err == nil {
					itemID = invItem.ID
					found = true
				}
			}

			if !found {
				return fmt.Errorf("could not find warehouse product for item %q (SKU: %v)", item.ItemDetails, item.SKU)
			}

			// Adjust stock: add quantity back
			qtyInt := int(item.Quantity)
			if qtyInt <= 0 {
				continue // skip zero or negative quantity items
			}

			if err := tx.Model(&inventoryEntity.InventoryItem{}).
				Where("id = ?", itemID).
				Update("current_stock", gorm.Expr("current_stock + ?", qtyInt)).Error; err != nil {
				return fmt.Errorf("failed to revert stock for product ID %d: %w", itemID, err)
			}

			// Write to inventory_logs
			logEntry := inventoryEntity.InventoryLog{
				InventoryItemID: itemID,
				Delta:           qtyInt,
				Reason:          "return",
				Platform:        "B2B",
				ExternalOrderID: invoice.InvoiceNumber,
				CreatedAt:       time.Now(),
			}
			if err := tx.Create(&logEntry).Error; err != nil {
				return fmt.Errorf("failed to log stock reversal for product ID %d: %w", itemID, err)
			}
		}

		// Mark invoice as not deducted
		invoice.InventoryDeducted = false
		if err := txRepo.UpdateInvoice(&invoice); err != nil {
			return fmt.Errorf("failed to save invoice inventory deduction status: %w", err)
		}

		return nil
	})
}

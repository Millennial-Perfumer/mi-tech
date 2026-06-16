package repository

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
	"mi-tech/internal/domain/b2b/entity"
)

type B2BRepository interface {
	WithTx(tx *gorm.DB) B2BRepository

	// Customers
	ListCustomers(search string) ([]entity.B2BCustomer, error)
	GetCustomerByID(id int64) (entity.B2BCustomer, error)
	GetCustomerByGSTIN(gstin string) (entity.B2BCustomer, error)
	CreateCustomer(cust *entity.B2BCustomer) error
	UpdateCustomer(cust *entity.B2BCustomer) error
	DeleteCustomer(id int64) error

	// Invoices
	ListInvoices(startDate, endDate string, status string) ([]entity.B2BInvoice, error)
	GetInvoiceByID(id int64) (entity.B2BInvoice, error)
	CreateInvoice(inv *entity.B2BInvoice) error
	UpdateInvoice(inv *entity.B2BInvoice) error
	DeleteInvoice(id int64) error
	GetNextSequenceForFY(fy string) (int, error)

	// Credit Notes
	ListCreditNotes(invoiceID int64) ([]entity.B2BCreditNote, error)
	GetCreditNoteByID(id int64) (entity.B2BCreditNote, error)
	CreateCreditNote(cn *entity.B2BCreditNote) error
	UpdateCreditNote(cn *entity.B2BCreditNote) error
	DeleteCreditNote(id int64) error
	GetNextCreditNoteSequenceForFY(fy string) (int, error)

	// Debit Notes
	ListDebitNotes(invoiceID int64) ([]entity.B2BDebitNote, error)
	GetDebitNoteByID(id int64) (entity.B2BDebitNote, error)
	CreateDebitNote(dn *entity.B2BDebitNote) error
	UpdateDebitNote(dn *entity.B2BDebitNote) error
	DeleteDebitNote(id int64) error
	GetNextDebitNoteSequenceForFY(fy string) (int, error)

	// GST Periods
	ListGSTPeriods() ([]entity.GSTPeriod, error)
	GetGSTPeriod(month, year int) (entity.GSTPeriod, error)
	SaveGSTPeriod(p *entity.GSTPeriod) error

	// Audit Logs
	SaveAuditLog(log *entity.B2BFinancialAuditLog) error
	ListAuditLogs(entityType string, entityID int64) ([]entity.B2BFinancialAuditLog, error)

	// Payment Terms
	ListPaymentTerms() ([]entity.B2BPaymentTerm, error)
	CreatePaymentTerm(term *entity.B2BPaymentTerm) error

	// Proformas
	ListProformas(startDate, endDate string, status string) ([]entity.B2BProformaInvoice, error)
	GetProformaByID(id int64) (entity.B2BProformaInvoice, error)
	CreateProforma(pf *entity.B2BProformaInvoice) error
	UpdateProforma(pf *entity.B2BProformaInvoice) error
	DeleteProforma(id int64) error
	GetNextProformaSequenceForFY(fy string) (int, error)
}

type gormB2BRepository struct {
	db *gorm.DB
}

func NewB2BRepository(db *gorm.DB) B2BRepository {
	return &gormB2BRepository{db: db}
}

func (r *gormB2BRepository) WithTx(tx *gorm.DB) B2BRepository {
	if tx == nil {
		return r
	}
	return &gormB2BRepository{db: tx}
}

// Customers implementation
func (r *gormB2BRepository) ListCustomers(search string) ([]entity.B2BCustomer, error) {
	var customers []entity.B2BCustomer
	query := r.db
	if search != "" {
		term := "%" + search + "%"
		query = query.Where("legal_name ILIKE ? OR trade_name ILIKE ? OR gstin ILIKE ?", term, term, term)
	}
	err := query.Order("legal_name ASC").Find(&customers).Error
	return customers, err
}

func (r *gormB2BRepository) GetCustomerByID(id int64) (entity.B2BCustomer, error) {
	var customer entity.B2BCustomer
	err := r.db.First(&customer, id).Error
	return customer, err
}

func (r *gormB2BRepository) GetCustomerByGSTIN(gstin string) (entity.B2BCustomer, error) {
	var customer entity.B2BCustomer
	err := r.db.Where("gstin = ?", gstin).First(&customer).Error
	return customer, err
}

func (r *gormB2BRepository) CreateCustomer(cust *entity.B2BCustomer) error {
	return r.db.Create(cust).Error
}

func (r *gormB2BRepository) UpdateCustomer(cust *entity.B2BCustomer) error {
	return r.db.Save(cust).Error
}

func (r *gormB2BRepository) DeleteCustomer(id int64) error {
	return r.db.Delete(&entity.B2BCustomer{}, id).Error
}

// Invoices implementation
func (r *gormB2BRepository) ListInvoices(startDate, endDate string, status string) ([]entity.B2BInvoice, error) {
	var invoices []entity.B2BInvoice
	query := r.db.Preload("Items")

	if startDate != "" {
		query = query.Where("invoice_date >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("invoice_date <= ?", endDate)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Order("invoice_date DESC, id DESC").Find(&invoices).Error
	return invoices, err
}

func (r *gormB2BRepository) GetInvoiceByID(id int64) (entity.B2BInvoice, error) {
	var invoice entity.B2BInvoice
	err := r.db.Preload("Items").First(&invoice, id).Error
	return invoice, err
}

func (r *gormB2BRepository) syncToOrdersTable(tx *gorm.DB, inv *entity.B2BInvoice) error {
	extID := fmt.Sprintf("B2B-%d", inv.ID)
	
	// Check if order already exists
	var orderID int64
	err := tx.Table("orders").Where("source_id = ? AND external_order_id = ?", "b2b", extID).Pluck("id", &orderID).Error
	if err != nil {
		return err
	}

	// Prepare order data
	financialStatus := "unpaid"
	lowerPaymentStatus := strings.ToLower(inv.PaymentStatus)
	if lowerPaymentStatus == "paid" {
		financialStatus = "paid"
	} else if lowerPaymentStatus == "partial" {
		financialStatus = "partially_paid"
	}

	totalTax := inv.CGSTAmount + inv.SGSTAmount + inv.IGSTAmount
	
	orderNumber := extID
	if inv.InvoiceNumber != nil && *inv.InvoiceNumber != "" {
		orderNumber = *inv.InvoiceNumber
	}

	var invoiceNumber *string
	if inv.InvoiceNumber != nil && *inv.InvoiceNumber != "" {
		invoiceNumber = inv.InvoiceNumber
	}

	orderData := map[string]interface{}{
		"source_id":          "b2b",
		"external_order_id":   extID,
		"order_number":       orderNumber,
		"invoice_number":     invoiceNumber,
		"total_price":        inv.TotalPrice,
		"subtotal_price":     inv.SubtotalPrice,
		"total_tax":          totalTax,
		"currency":           "INR",
		"financial_status":   financialStatus,
		"fulfillment_status": "fulfilled",
		"status":             strings.ToLower(inv.Status),
		"customer_name":      inv.CustomerName,
		"customer_phone":     inv.CustomerPhone,
		"customer_email":     inv.CustomerEmail,
		"customer_address1":  inv.CustomerAddress,
		"customer_state":     inv.CustomerState,
		"total_discount":     inv.DiscountAmount,
		"updated_at":         inv.UpdatedAt,
	}

	if orderID == 0 {
		// Insert order
		orderData["created_at"] = inv.InvoiceDate
		if err := tx.Table("orders").Create(&orderData).Error; err != nil {
			return err
		}
		// Fetch generated ID
		err = tx.Table("orders").Where("source_id = ? AND external_order_id = ?", "b2b", extID).Pluck("id", &orderID).Error
		if err != nil {
			return err
		}
	} else {
		// Update order
		if err := tx.Table("orders").Where("id = ?", orderID).Updates(orderData).Error; err != nil {
			return err
		}
	}

	// Delete existing line items
	if err := tx.Table("order_line_items").Where("order_id = ?", orderID).Delete(nil).Error; err != nil {
		return err
	}

	// Insert line items
	if len(inv.Items) > 0 {
		for i, item := range inv.Items {
			liID := fmt.Sprintf("b2b-%d-%d", inv.ID, item.ID)
			if item.ID == 0 {
				liID = fmt.Sprintf("b2b-%d-idx-%d", inv.ID, i)
			}
			
			var productIDStr *string
			if item.ProductID != nil {
				s := fmt.Sprintf("%d", *item.ProductID)
				productIDStr = &s
			}

			liData := map[string]interface{}{
				"id":             liID,
				"order_id":       orderID,
				"product_id":     productIDStr,
				"title":          item.ItemDetails,
				"sku":            item.SKU,
				"hs_code":        item.HSNCode,
				"quantity":       int(item.Quantity),
				"price":          item.Rate,
				"discount":       0.0,
				"order_discount": 0.0,
			}
			if err := tx.Table("order_line_items").Create(&liData).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *gormB2BRepository) CreateInvoice(inv *entity.B2BInvoice) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Items").Create(inv).Error; err != nil {
			return err
		}
		if len(inv.Items) > 0 {
			for i := range inv.Items {
				inv.Items[i].InvoiceID = inv.ID
			}
			if err := tx.Create(&inv.Items).Error; err != nil {
				return err
			}
		}
		if err := r.syncToOrdersTable(tx, inv); err != nil {
			return err
		}
		return nil
	})
}

func (r *gormB2BRepository) UpdateInvoice(inv *entity.B2BInvoice) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Update parent fields
		if err := tx.Omit("Items").Save(inv).Error; err != nil {
			return err
		}

		// Delete existing line items
		if err := tx.Where("invoice_id = ?", inv.ID).Delete(&entity.B2BInvoiceItem{}).Error; err != nil {
			return err
		}

		// Re-create items
		if len(inv.Items) > 0 {
			for i := range inv.Items {
				inv.Items[i].InvoiceID = inv.ID
				inv.Items[i].ID = 0 // Reset PK for clean insert
			}
			if err := tx.Create(&inv.Items).Error; err != nil {
				return err
			}
		}
		if err := r.syncToOrdersTable(tx, inv); err != nil {
			return err
		}
		return nil
	})
}

func (r *gormB2BRepository) DeleteInvoice(id int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Delete B2B Credit Note Items associated with this invoice's credit notes
		if err := tx.Exec(`
			DELETE FROM b2b_credit_note_items 
			WHERE credit_note_id IN (SELECT id FROM b2b_credit_notes WHERE invoice_id = ?)
		`, id).Error; err != nil {
			return err
		}
		// 2. Delete B2B Credit Notes associated with this invoice
		if err := tx.Where("invoice_id = ?", id).Delete(&entity.B2BCreditNote{}).Error; err != nil {
			return err
		}

		// 3. Delete B2B Debit Note Items associated with this invoice's debit notes
		if err := tx.Exec(`
			DELETE FROM b2b_debit_note_items 
			WHERE debit_note_id IN (SELECT id FROM b2b_debit_notes WHERE invoice_id = ?)
		`, id).Error; err != nil {
			return err
		}
		// 4. Delete B2B Debit Notes associated with this invoice
		if err := tx.Where("invoice_id = ?", id).Delete(&entity.B2BDebitNote{}).Error; err != nil {
			return err
		}

		// 5. Delete synced order line items and orders
		extID := fmt.Sprintf("B2B-%d", id)
		if err := tx.Exec(`
			DELETE FROM order_line_items 
			WHERE order_id IN (SELECT id FROM orders WHERE source_id = 'b2b' AND external_order_id = ?)
		`, extID).Error; err != nil {
			return err
		}
		if err := tx.Exec(`
			DELETE FROM orders WHERE source_id = 'b2b' AND external_order_id = ?
		`, extID).Error; err != nil {
			return err
		}

		// 6. Delete invoice items
		if err := tx.Where("invoice_id = ?", id).Delete(&entity.B2BInvoiceItem{}).Error; err != nil {
			return err
		}
		// 7. Delete invoice itself
		return tx.Delete(&entity.B2BInvoice{}, id).Error
	})
}

func (r *gormB2BRepository) GetNextSequenceForFY(fy string) (int, error) {
	var count int64
	err := r.db.Model(&entity.B2BInvoice{}).
		Where("financial_year = ? AND status = 'ISSUED'", fy).
		Count(&count).Error
	return int(count) + 1, err
}

// Payment Terms implementation
func (r *gormB2BRepository) ListPaymentTerms() ([]entity.B2BPaymentTerm, error) {
	var terms []entity.B2BPaymentTerm
	err := r.db.Order("due_days ASC, name ASC").Find(&terms).Error
	return terms, err
}

func (r *gormB2BRepository) CreatePaymentTerm(term *entity.B2BPaymentTerm) error {
	return r.db.Create(term).Error
}

// Credit Notes implementation
func (r *gormB2BRepository) ListCreditNotes(invoiceID int64) ([]entity.B2BCreditNote, error) {
	var notes []entity.B2BCreditNote
	query := r.db.Preload("Items")
	if invoiceID > 0 {
		query = query.Where("invoice_id = ?", invoiceID)
	}
	err := query.Order("note_date DESC, id DESC").Find(&notes).Error
	return notes, err
}

func (r *gormB2BRepository) GetCreditNoteByID(id int64) (entity.B2BCreditNote, error) {
	var note entity.B2BCreditNote
	err := r.db.Preload("Items").First(&note, id).Error
	return note, err
}

func (r *gormB2BRepository) CreateCreditNote(cn *entity.B2BCreditNote) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Items").Create(cn).Error; err != nil {
			return err
		}
		if len(cn.Items) > 0 {
			for i := range cn.Items {
				cn.Items[i].CreditNoteID = cn.ID
			}
			if err := tx.Create(&cn.Items).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *gormB2BRepository) UpdateCreditNote(cn *entity.B2BCreditNote) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Items").Save(cn).Error; err != nil {
			return err
		}
		if err := tx.Where("credit_note_id = ?", cn.ID).Delete(&entity.B2BCreditNoteItem{}).Error; err != nil {
			return err
		}
		if len(cn.Items) > 0 {
			for i := range cn.Items {
				cn.Items[i].CreditNoteID = cn.ID
				cn.Items[i].ID = 0
			}
			if err := tx.Create(&cn.Items).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *gormB2BRepository) DeleteCreditNote(id int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var status string
		err := tx.Model(&entity.B2BCreditNote{}).Where("id = ?", id).Pluck("status", &status).Error
		if err != nil {
			return err
		}
		if status != "DRAFT" {
			return fmt.Errorf("only DRAFT credit notes can be deleted; current status is %s", status)
		}
		if err := tx.Where("credit_note_id = ?", id).Delete(&entity.B2BCreditNoteItem{}).Error; err != nil {
			return err
		}
		return tx.Delete(&entity.B2BCreditNote{}, id).Error
	})
}

func (r *gormB2BRepository) GetNextCreditNoteSequenceForFY(fy string) (int, error) {
	var count int64
	err := r.db.Model(&entity.B2BCreditNote{}).
		Where("financial_year = ? AND status = 'ISSUED'", fy).
		Count(&count).Error
	return int(count) + 1, err
}

// Debit Notes implementation
func (r *gormB2BRepository) ListDebitNotes(invoiceID int64) ([]entity.B2BDebitNote, error) {
	var notes []entity.B2BDebitNote
	query := r.db.Preload("Items")
	if invoiceID > 0 {
		query = query.Where("invoice_id = ?", invoiceID)
	}
	err := query.Order("note_date DESC, id DESC").Find(&notes).Error
	return notes, err
}

func (r *gormB2BRepository) GetDebitNoteByID(id int64) (entity.B2BDebitNote, error) {
	var note entity.B2BDebitNote
	err := r.db.Preload("Items").First(&note, id).Error
	return note, err
}

func (r *gormB2BRepository) CreateDebitNote(dn *entity.B2BDebitNote) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Items").Create(dn).Error; err != nil {
			return err
		}
		if len(dn.Items) > 0 {
			for i := range dn.Items {
				dn.Items[i].DebitNoteID = dn.ID
			}
			if err := tx.Create(&dn.Items).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *gormB2BRepository) UpdateDebitNote(dn *entity.B2BDebitNote) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Items").Save(dn).Error; err != nil {
			return err
		}
		if err := tx.Where("debit_note_id = ?", dn.ID).Delete(&entity.B2BDebitNoteItem{}).Error; err != nil {
			return err
		}
		if len(dn.Items) > 0 {
			for i := range dn.Items {
				dn.Items[i].DebitNoteID = dn.ID
				dn.Items[i].ID = 0
			}
			if err := tx.Create(&dn.Items).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *gormB2BRepository) DeleteDebitNote(id int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var status string
		err := tx.Model(&entity.B2BDebitNote{}).Where("id = ?", id).Pluck("status", &status).Error
		if err != nil {
			return err
		}
		if status != "DRAFT" {
			return fmt.Errorf("only DRAFT debit notes can be deleted; current status is %s", status)
		}
		if err := tx.Where("debit_note_id = ?", id).Delete(&entity.B2BDebitNoteItem{}).Error; err != nil {
			return err
		}
		return tx.Delete(&entity.B2BDebitNote{}, id).Error
	})
}

func (r *gormB2BRepository) GetNextDebitNoteSequenceForFY(fy string) (int, error) {
	var count int64
	err := r.db.Model(&entity.B2BDebitNote{}).
		Where("financial_year = ? AND status = 'ISSUED'", fy).
		Count(&count).Error
	return int(count) + 1, err
}

// GST Periods implementation
func (r *gormB2BRepository) ListGSTPeriods() ([]entity.GSTPeriod, error) {
	var periods []entity.GSTPeriod
	err := r.db.Order("year DESC, month DESC").Find(&periods).Error
	return periods, err
}

func (r *gormB2BRepository) GetGSTPeriod(month, year int) (entity.GSTPeriod, error) {
	var period entity.GSTPeriod
	err := r.db.Where("month = ? AND year = ?", month, year).First(&period).Error
	return period, err
}

func (r *gormB2BRepository) SaveGSTPeriod(p *entity.GSTPeriod) error {
	return r.db.Save(p).Error
}

// Audit Logs implementation
func (r *gormB2BRepository) SaveAuditLog(log *entity.B2BFinancialAuditLog) error {
	return r.db.Create(log).Error
}

func (r *gormB2BRepository) ListAuditLogs(entityType string, entityID int64) ([]entity.B2BFinancialAuditLog, error) {
	var logs []entity.B2BFinancialAuditLog
	query := r.db
	if entityType != "" {
		query = query.Where("entity_type = ?", entityType)
	}
	if entityID > 0 {
		query = query.Where("entity_id = ?", entityID)
	}
	err := query.Order("created_at DESC").Find(&logs).Error
	return logs, err
}

// Proformas implementation
func (r *gormB2BRepository) ListProformas(startDate, endDate string, status string) ([]entity.B2BProformaInvoice, error) {
	var pfs []entity.B2BProformaInvoice
	query := r.db.Preload("Items")

	if startDate != "" {
		query = query.Where("note_date >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("note_date <= ?", endDate)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Order("note_date DESC, id DESC").Find(&pfs).Error
	return pfs, err
}

func (r *gormB2BRepository) GetProformaByID(id int64) (entity.B2BProformaInvoice, error) {
	var pf entity.B2BProformaInvoice
	err := r.db.Preload("Items").First(&pf, id).Error
	return pf, err
}

func (r *gormB2BRepository) CreateProforma(pf *entity.B2BProformaInvoice) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Items").Create(pf).Error; err != nil {
			return err
		}
		if len(pf.Items) > 0 {
			for i := range pf.Items {
				pf.Items[i].ProformaInvoiceID = pf.ID
			}
			if err := tx.Create(&pf.Items).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *gormB2BRepository) UpdateProforma(pf *entity.B2BProformaInvoice) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Items").Save(pf).Error; err != nil {
			return err
		}
		if err := tx.Where("proforma_invoice_id = ?", pf.ID).Delete(&entity.B2BProformaInvoiceItem{}).Error; err != nil {
			return err
		}
		if len(pf.Items) > 0 {
			for i := range pf.Items {
				pf.Items[i].ProformaInvoiceID = pf.ID
				pf.Items[i].ID = 0
			}
			if err := tx.Create(&pf.Items).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *gormB2BRepository) DeleteProforma(id int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var status string
		err := tx.Model(&entity.B2BProformaInvoice{}).Where("id = ?", id).Pluck("status", &status).Error
		if err != nil {
			return err
		}
		if status != "DRAFT" {
			return fmt.Errorf("only DRAFT proforma invoices can be deleted; current status is %s", status)
		}
		if err := tx.Where("proforma_invoice_id = ?", id).Delete(&entity.B2BProformaInvoiceItem{}).Error; err != nil {
			return err
		}
		return tx.Delete(&entity.B2BProformaInvoice{}, id).Error
	})
}

func (r *gormB2BRepository) GetNextProformaSequenceForFY(fy string) (int, error) {
	var count int64
	err := r.db.Model(&entity.B2BProformaInvoice{}).
		Where("financial_year = ? AND proforma_sequence IS NOT NULL", fy).
		Count(&count).Error
	return int(count) + 1, err
}



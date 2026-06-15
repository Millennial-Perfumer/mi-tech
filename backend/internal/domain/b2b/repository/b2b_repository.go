package repository

import (
	"fmt"

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

	// Payment Terms
	ListPaymentTerms() ([]entity.B2BPaymentTerm, error)
	CreatePaymentTerm(term *entity.B2BPaymentTerm) error
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
		return nil
	})
}

func (r *gormB2BRepository) DeleteInvoice(id int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Verify status is draft
		var status string
		err := tx.Model(&entity.B2BInvoice{}).Where("id = ?", id).Pluck("status", &status).Error
		if err != nil {
			return err
		}
		if status != "DRAFT" {
			return fmt.Errorf("only DRAFT invoices can be deleted; current status is %s", status)
		}

		if err := tx.Where("invoice_id = ?", id).Delete(&entity.B2BInvoiceItem{}).Error; err != nil {
			return err
		}
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

package entity

import (
	"time"
)

// B2BCreditNote represents a B2B Credit Note
type B2BCreditNote struct {
	ID                 int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	CreditNoteNumber   *string   `gorm:"column:credit_note_number;unique" json:"credit_note_number"`
	CreditNoteSequence *int      `gorm:"column:credit_note_sequence" json:"credit_note_sequence"`
	FinancialYear      *string   `gorm:"column:financial_year" json:"financial_year"`
	InvoiceID          *int64    `gorm:"column:invoice_id" json:"invoice_id"`
	InvoiceNumber      *string   `gorm:"column:invoice_number" json:"invoice_number"`
	NoteDate           time.Time `gorm:"column:note_date" json:"note_date"`
	Reason             *string   `gorm:"column:reason" json:"reason"`

	// Customer snapshot
	CustomerID        *int64  `gorm:"column:customer_id" json:"customer_id"`
	CustomerGSTIN     string  `gorm:"column:customer_gstin" json:"customer_gstin"`
	CustomerName      string  `gorm:"column:customer_name" json:"customer_name"`
	CustomerEmail     *string `gorm:"column:customer_email" json:"customer_email"`
	CustomerPhone     *string `gorm:"column:customer_phone" json:"customer_phone"`
	CustomerState     string  `gorm:"column:customer_state" json:"customer_state"`
	CustomerStateCode string  `gorm:"column:customer_state_code" json:"customer_state_code"`
	CustomerAddress   string  `gorm:"column:customer_address" json:"customer_address"`

	// Seller details snapshot
	SellerGSTIN     string `gorm:"column:seller_gstin" json:"seller_gstin"`
	SellerName      string `gorm:"column:seller_name" json:"seller_name"`
	SellerState     string `gorm:"column:seller_state" json:"seller_state"`
	SellerStateCode string `gorm:"column:seller_state_code" json:"seller_state_code"`
	SellerAddress   string `gorm:"column:seller_address" json:"seller_address"`

	// Financial sums
	SubtotalPrice   float64 `gorm:"column:subtotal_price" json:"subtotal_price"`
	DiscountPercent float64 `gorm:"column:discount_percent" json:"discount_percent"`
	DiscountAmount  float64 `gorm:"column:discount_amount" json:"discount_amount"`

	// GST details
	CGSTRate   float64 `gorm:"column:cgst_rate" json:"cgst_rate"`
	CGSTAmount float64 `gorm:"column:cgst_amount" json:"cgst_amount"`
	SGSTRate   float64 `gorm:"column:sgst_rate" json:"sgst_rate"`
	SGSTAmount float64 `gorm:"column:sgst_amount" json:"sgst_amount"`
	IGSTRate   float64 `gorm:"column:igst_rate" json:"igst_rate"`
	IGSTAmount float64 `gorm:"column:igst_amount" json:"igst_amount"`

	// Final Totals
	TotalPrice float64 `gorm:"column:total_price" json:"total_price"`

	// Lifecycle status
	Status    string    `gorm:"column:status;default:DRAFT" json:"status"` // DRAFT, ISSUED, CANCELLED
	CreatedAt time.Time `gorm:"column:created_at;default:NOW()" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;default:NOW()" json:"updated_at"`

	Items []B2BCreditNoteItem `gorm:"foreignKey:CreditNoteID" json:"items"`
}

func (B2BCreditNote) TableName() string {
	return "b2b_credit_notes"
}

// B2BCreditNoteItem represents a line item in a Credit Note
type B2BCreditNoteItem struct {
	ID           int64   `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	CreditNoteID int64   `gorm:"column:credit_note_id" json:"credit_note_id"`
	ProductID    *int64  `gorm:"column:product_id" json:"product_id"`
	ItemDetails  string  `gorm:"column:item_details" json:"item_details"`
	SKU          *string `gorm:"column:sku" json:"sku"`
	HSNCode      *string `gorm:"column:hsn_code" json:"hsn_code"`
	Quantity     float64 `gorm:"column:quantity" json:"quantity"`
	Rate         float64 `gorm:"column:rate" json:"rate"`
	Amount       float64 `gorm:"column:amount" json:"amount"`
}

func (B2BCreditNoteItem) TableName() string {
	return "b2b_credit_note_items"
}

// B2BDebitNote represents a B2B Debit Note
type B2BDebitNote struct {
	ID                int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	DebitNoteNumber   *string   `gorm:"column:debit_note_number;unique" json:"debit_note_number"`
	DebitNoteSequence *int      `gorm:"column:debit_note_sequence" json:"debit_note_sequence"`
	FinancialYear     *string   `gorm:"column:financial_year" json:"financial_year"`
	InvoiceID         *int64    `gorm:"column:invoice_id" json:"invoice_id"`
	InvoiceNumber     *string   `gorm:"column:invoice_number" json:"invoice_number"`
	NoteDate          time.Time `gorm:"column:note_date" json:"note_date"`
	Reason            *string   `gorm:"column:reason" json:"reason"`

	// Customer snapshot
	CustomerID        *int64  `gorm:"column:customer_id" json:"customer_id"`
	CustomerGSTIN     string  `gorm:"column:customer_gstin" json:"customer_gstin"`
	CustomerName      string  `gorm:"column:customer_name" json:"customer_name"`
	CustomerEmail     *string `gorm:"column:customer_email" json:"customer_email"`
	CustomerPhone     *string `gorm:"column:customer_phone" json:"customer_phone"`
	CustomerState     string  `gorm:"column:customer_state" json:"customer_state"`
	CustomerStateCode string  `gorm:"column:customer_state_code" json:"customer_state_code"`
	CustomerAddress   string  `gorm:"column:customer_address" json:"customer_address"`

	// Seller details snapshot
	SellerGSTIN     string `gorm:"column:seller_gstin" json:"seller_gstin"`
	SellerName      string `gorm:"column:seller_name" json:"seller_name"`
	SellerState     string `gorm:"column:seller_state" json:"seller_state"`
	SellerStateCode string `gorm:"column:seller_state_code" json:"seller_state_code"`
	SellerAddress   string `gorm:"column:seller_address" json:"seller_address"`

	// Financial pricing summaries
	SubtotalPrice   float64 `gorm:"column:subtotal_price" json:"subtotal_price"`
	DiscountPercent float64 `gorm:"column:discount_percent" json:"discount_percent"`
	DiscountAmount  float64 `gorm:"column:discount_amount" json:"discount_amount"`

	// GST details
	CGSTRate   float64 `gorm:"column:cgst_rate" json:"cgst_rate"`
	CGSTAmount float64 `gorm:"column:cgst_amount" json:"cgst_amount"`
	SGSTRate   float64 `gorm:"column:sgst_rate" json:"sgst_rate"`
	SGSTAmount float64 `gorm:"column:sgst_amount" json:"sgst_amount"`
	IGSTRate   float64 `gorm:"column:igst_rate" json:"igst_rate"`
	IGSTAmount float64 `gorm:"column:igst_amount" json:"igst_amount"`

	// Final Totals
	TotalPrice float64 `gorm:"column:total_price" json:"total_price"`

	// Lifecycle status
	Status    string    `gorm:"column:status;default:DRAFT" json:"status"` // DRAFT, ISSUED, CANCELLED
	CreatedAt time.Time `gorm:"column:created_at;default:NOW()" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;default:NOW()" json:"updated_at"`

	Items []B2BDebitNoteItem `gorm:"foreignKey:DebitNoteID" json:"items"`
}

func (B2BDebitNote) TableName() string {
	return "b2b_debit_notes"
}

// B2BDebitNoteItem represents a line item in a Debit Note
type B2BDebitNoteItem struct {
	ID          int64   `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	DebitNoteID int64   `gorm:"column:debit_note_id" json:"debit_note_id"`
	ProductID   *int64  `gorm:"column:product_id" json:"product_id"`
	ItemDetails string  `gorm:"column:item_details" json:"item_details"`
	SKU         *string `gorm:"column:sku" json:"sku"`
	HSNCode     *string `gorm:"column:hsn_code" json:"hsn_code"`
	Quantity    float64 `gorm:"column:quantity" json:"quantity"`
	Rate        float64 `gorm:"column:rate" json:"rate"`
	Amount      float64 `gorm:"column:amount" json:"amount"`
}

func (B2BDebitNoteItem) TableName() string {
	return "b2b_debit_note_items"
}

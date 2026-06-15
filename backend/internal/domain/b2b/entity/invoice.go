package entity

import (
	"time"
)

// B2BInvoice represents a tax-compliant B2B invoice
type B2BInvoice struct {
	ID                   int64            `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	InvoiceNumber        *string          `gorm:"column:invoice_number;unique" json:"invoice_number"`
	InvoiceSequence      *int             `gorm:"column:invoice_sequence" json:"invoice_sequence"`
	FinancialYear        *string          `gorm:"column:financial_year" json:"financial_year"`
	OrderNumber          *string          `gorm:"column:order_number" json:"order_number"`
	InvoiceDate          time.Time        `gorm:"column:invoice_date" json:"invoice_date"`
	Terms                *string          `gorm:"column:terms" json:"terms"`
	DueDate              *time.Time       `gorm:"column:due_date" json:"due_date"`
	Salesperson          *string          `gorm:"column:salesperson" json:"salesperson"`
	Subject              *string          `gorm:"column:subject" json:"subject"`

	// Customer snapshot fields (immutable historical details)
	CustomerID          *int64           `gorm:"column:customer_id" json:"customer_id"`
	CustomerGSTIN        string           `gorm:"column:customer_gstin" json:"customer_gstin"`
	CustomerName         string           `gorm:"column:customer_name" json:"customer_name"`
	CustomerEmail        *string          `gorm:"column:customer_email" json:"customer_email"`
	CustomerPhone        *string          `gorm:"column:customer_phone" json:"customer_phone"`
	CustomerState        string           `gorm:"column:customer_state" json:"customer_state"`
	CustomerStateCode    string           `gorm:"column:customer_state_code" json:"customer_state_code"`
	CustomerAddress      string           `gorm:"column:customer_address" json:"customer_address"`

	// Seller details snapshot
	SellerGSTIN          string           `gorm:"column:seller_gstin" json:"seller_gstin"`
	SellerName           string           `gorm:"column:seller_name" json:"seller_name"`
	SellerState          string           `gorm:"column:seller_state" json:"seller_state"`
	SellerStateCode      string           `gorm:"column:seller_state_code" json:"seller_state_code"`
	SellerAddress        string           `gorm:"column:seller_address" json:"seller_address"`

	// Financial sums
	SubtotalPrice        float64          `gorm:"column:subtotal_price" json:"subtotal_price"`
	DiscountPercent      float64          `gorm:"column:discount_percent" json:"discount_percent"`
	DiscountAmount       float64          `gorm:"column:discount_amount" json:"discount_amount"`

	// GST details
	CGSTRate             float64          `gorm:"column:cgst_rate" json:"cgst_rate"`
	CGSTAmount           float64          `gorm:"column:cgst_amount" json:"cgst_amount"`
	SGSTRate             float64          `gorm:"column:sgst_rate" json:"sgst_rate"`
	SGSTAmount           float64          `gorm:"column:sgst_amount" json:"sgst_amount"`
	IGSTRate             float64          `gorm:"column:igst_rate" json:"igst_rate"`
	IGSTAmount           float64          `gorm:"column:igst_amount" json:"igst_amount"`

	// Additional Charges
	TDSTCSType          string           `gorm:"column:tds_tcs_type;default:NONE" json:"tds_tcs_type"`
	TDSTCSRate          float64          `gorm:"column:tds_tcs_rate" json:"tds_tcs_rate"`
	TDSTCSAmount        float64          `gorm:"column:tds_tcs_amount" json:"tds_tcs_amount"`
	TransportationCharge float64          `gorm:"column:transportation_charge" json:"transportation_charge"`

	// Final Totals
	TotalPrice           float64          `gorm:"column:total_price" json:"total_price"`

	// Lifecycle statuses
	Status               string           `gorm:"column:status;default:DRAFT" json:"status"` // DRAFT, ISSUED, CANCELLED
	PaymentStatus        string           `gorm:"column:payment_status;default:UNPAID" json:"payment_status"` // UNPAID, PARTIAL, PAID
	PaidAmount           float64          `gorm:"column:paid_amount" json:"paid_amount"`
	BalanceAmount        float64          `gorm:"column:balance_amount" json:"balance_amount"`
	PaymentDate          *time.Time       `gorm:"column:payment_date" json:"payment_date"`
	PaymentMethod        *string          `gorm:"column:payment_method" json:"payment_method"`

	CustomerNotes        *string          `gorm:"column:customer_notes" json:"customer_notes"`
	CreatedAt            time.Time        `gorm:"column:created_at;default:NOW()" json:"created_at"`
	UpdatedAt            time.Time        `gorm:"column:updated_at;default:NOW()" json:"updated_at"`

	Items                []B2BInvoiceItem `gorm:"foreignKey:InvoiceID" json:"items"`
}

func (B2BInvoice) TableName() string {
	return "b2b_invoices"
}

// B2BInvoiceItem represents a line item in a B2B invoice
type B2BInvoiceItem struct {
	ID          int64   `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	InvoiceID   int64   `gorm:"column:invoice_id" json:"invoice_id"`
	ProductID   *int64  `gorm:"column:product_id" json:"product_id"`
	ItemDetails string  `gorm:"column:item_details" json:"item_details"`
	SKU         *string `gorm:"column:sku" json:"sku"`
	HSNCode     *string `gorm:"column:hsn_code" json:"hsn_code"`
	Quantity    float64 `gorm:"column:quantity" json:"quantity"`
	Rate        float64 `gorm:"column:rate" json:"rate"`
	Amount      float64 `gorm:"column:amount" json:"amount"`
}

func (B2BInvoiceItem) TableName() string {
	return "b2b_invoice_items"
}

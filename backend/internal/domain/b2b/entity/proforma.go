package entity

import (
	"encoding/json"
	"time"
)

// B2BProformaInvoice represents a commercial proforma invoice
type B2BProformaInvoice struct {
	ID               int64                  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ProformaNumber   *string                `gorm:"column:proforma_number;unique" json:"proforma_number"`
	ProformaSequence *int                   `gorm:"column:proforma_sequence" json:"proforma_sequence"`
	FinancialYear    *string                `gorm:"column:financial_year" json:"financial_year"`
	NoteDate         time.Time              `gorm:"column:note_date" json:"note_date"`
	ValidUntil       *time.Time             `gorm:"column:valid_until" json:"valid_until"`
	Status           string                 `gorm:"column:status;default:DRAFT" json:"status"` // DRAFT, SENT, ACCEPTED, CONVERTED_TO_INVOICE, REJECTED, EXPIRED, CANCELLED
	RevisionNumber   int                    `gorm:"column:revision_number;default:1" json:"revision_number"`
	ParentProformaID *int64                 `gorm:"column:parent_proforma_id" json:"parent_proforma_id"`

	// Customer snapshot fields (immutable historical details)
	CustomerID        *int64  `gorm:"column:customer_id" json:"customer_id"`
	CustomerGSTIN     string  `gorm:"column:customer_gstin" json:"customer_gstin"`
	CustomerName      string  `gorm:"column:customer_name" json:"customer_name"`
	CustomerEmail     *string `gorm:"column:customer_email" json:"customer_email"`
	CustomerPhone     *string `gorm:"column:customer_phone" json:"customer_phone"`
	CustomerState     string  `gorm:"column:customer_state" json:"customer_state"`
	CustomerStateCode string  `gorm:"column:customer_state_code" json:"customer_state_code"`
	CustomerAddress   string  `gorm:"column:customer_address" json:"customer_address"`
	CustomerShippingAddress string `gorm:"column:customer_shipping_address" json:"customer_shipping_address"`

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
	TotalPrice  float64   `gorm:"column:total_price" json:"total_price"`
	AdvancePaid float64   `gorm:"column:advance_paid" json:"advance_paid"`
	CreatedAt   time.Time `gorm:"column:created_at;default:NOW()" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;default:NOW()" json:"updated_at"`

	Items []B2BProformaInvoiceItem `gorm:"foreignKey:ProformaInvoiceID" json:"items"`
}

func (B2BProformaInvoice) TableName() string {
	return "b2b_proforma_invoices"
}

// B2BProformaInvoiceItem represents a line item in a Proforma invoice
type B2BProformaInvoiceItem struct {
	ID                int64   `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ProformaInvoiceID int64   `gorm:"column:proforma_invoice_id" json:"proforma_invoice_id"`
	ProductID         *int64  `gorm:"column:product_id" json:"product_id"`
	ItemDetails       string  `gorm:"column:item_details" json:"item_details"`
	SKU               *string `gorm:"column:sku" json:"sku"`
	HSNCode           *string `gorm:"column:hsn_code" json:"hsn_code"`
	Quantity          float64 `gorm:"column:quantity" json:"quantity"`
	Rate              float64 `gorm:"column:rate" json:"rate"`
	Amount            float64 `gorm:"column:amount" json:"amount"`
}

func (B2BProformaInvoiceItem) TableName() string {
	return "b2b_proforma_invoice_items"
}

// UnmarshalJSON customizes unmarshaling to support flexible date formats
func (pf *B2BProformaInvoice) UnmarshalJSON(data []byte) error {
	type Alias B2BProformaInvoice
	aux := &struct {
		NoteDate   string  `json:"note_date"`
		ValidUntil *string `json:"valid_until"`
		*Alias
	}{
		Alias: (*Alias)(pf),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Parse NoteDate
	if aux.NoteDate != "" {
		t, err := parseFlexDate(aux.NoteDate)
		if err != nil {
			return err
		}
		pf.NoteDate = t
	}

	// Parse ValidUntil
	if aux.ValidUntil != nil && *aux.ValidUntil != "" {
		t, err := parseFlexDate(*aux.ValidUntil)
		if err != nil {
			return err
		}
		pf.ValidUntil = &t
	} else {
		pf.ValidUntil = nil
	}

	return nil
}


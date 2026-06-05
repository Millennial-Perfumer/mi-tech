package dto

// LineItemCreateRequest represents a line item within a manual order creation request.
type LineItemCreateRequest struct {
	MISKU    string  `json:"mi_sku"`    // e.g. "mi-01"
	Title    string  `json:"title"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`     // unit price
	Discount float64 `json:"discount"`  // item-level discount amount
}

// OrderCreateRequest represents a manual order creation request (POS orders).
type OrderCreateRequest struct {
	TerminalCode      string                  `json:"terminal_code"`      // "POS1" (defaults if empty)
	TotalPrice        float64                 `json:"total_price"`
	TotalDiscount     float64                 `json:"total_discount"`
	FinancialStatus   string                  `json:"financial_status"`   // default: "paid"
	FulfillmentStatus string                  `json:"fulfillment_status"` // default: "fulfilled"
	CustomerName      string                  `json:"customer_name"`
	CustomerPhone     string                  `json:"customer_phone"`
	CustomerEmail     string                  `json:"customer_email"`
	CustomerAddress1  string                  `json:"customer_address1"`
	CustomerAddress2  string                  `json:"customer_address2"`
	CustomerCity      string                  `json:"customer_city"`
	CustomerState     string                  `json:"customer_state"`
	CustomerZip       string                  `json:"customer_zip"`
	CustomerCountry   string                  `json:"customer_country"`
	LineItems         []LineItemCreateRequest `json:"line_items"`
}

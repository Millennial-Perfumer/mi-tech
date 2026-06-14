package dto

// GSTSummaryResponse is the DTO for the GST report summary.
type GSTSummaryResponse struct {
	TotalOrders       int     `json:"total_orders"`
	CancelledOrders   int     `json:"cancelled_orders"`
	InvoicesGenerated int     `json:"invoices_generated"`
	TotalRevenue      float64 `json:"total_revenue"`
	TotalTaxableValue float64 `json:"total_taxable_value"`
	TotalGSTCollected float64 `json:"total_gst_collected"`
	TotalIGST         float64 `json:"total_igst"`
	TotalCGST         float64 `json:"total_cgst"`
	TotalSGST         float64 `json:"total_sgst"`
	FulfilledOrders   int     `json:"fulfilled_orders"`
	UnfulfilledOrders int     `json:"unfulfilled_orders"`
	PaidOrders        int     `json:"paid_orders"`
	RefundedOrders    int     `json:"refunded_orders"`
}

// StateSummaryRow is the DTO for a single state in the state-wise GST report.
type StateSummaryRow struct {
	State        string  `json:"state"`
	Orders       int     `json:"orders"`
	TaxableValue float64 `json:"taxable_value"`
	IGST         float64 `json:"igst"`
	CGST         float64 `json:"cgst"`
	SGST         float64 `json:"sgst"` // Note: on frontend B2C State-wise maps Samt as sgst or similar. The frontend uses fields from stateData: state, orders, taxable_value, igst, cgst, sgst, total_gst, revenue.
	TotalGST     float64 `json:"total_gst"`
	Revenue      float64 `json:"revenue"`
}

// HSNSummaryRow is the DTO for a single HSN code in the HSN-wise report.
type HSNSummaryRow struct {
	HSNCode      string  `json:"hsn_code"`
	ProductCount int     `json:"product_count"`
	QtySold      int     `json:"qty_sold"`
	TaxableValue float64 `json:"taxable_value"`
	IGST         float64 `json:"igst"`
	CGST         float64 `json:"cgst"`
	SGST         float64 `json:"sgst"`
	TotalGST     float64 `json:"total_gst"`
	Revenue      float64 `json:"revenue"`
}

// DocumentIssuedRow is the DTO for the documents-issued report.
type DocumentIssuedRow struct {
	DocumentType string `json:"document_type"`
	FromSerial   string `json:"from_serial"`
	ToSerial     string `json:"to_serial"`
	TotalIssued  int    `json:"total_issued"`
	Cancelled    int    `json:"cancelled"`
	NetIssued    int    `json:"net_issued"`
}

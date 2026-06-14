package dto

// MetricsResponse is the DTO for the dashboard metrics endpoint.
type MetricsResponse struct {
	Success bool             `json:"success"`
	Metrics DashboardMetrics `json:"metrics"`
}

// DashboardMetrics contains the computed dashboard values.
type DashboardMetrics struct {
	TotalRevenue      float64          `json:"total_revenue"`
	TotalInvoices     int              `json:"total_invoices"`
	TotalGSTCollected float64          `json:"total_gst_collected"`
	CGSTCollected     float64          `json:"cgst_collected"`
	SGSTCollected     float64          `json:"sgst_collected"`
	IGSTCollected     float64          `json:"igst_collected"`
	TotalOrders       int              `json:"total_orders"`
	CancelledOrders   int              `json:"cancelled_orders"`
	FulfilledOrders   int              `json:"fulfilled_orders"`
	UnfulfilledOrders int              `json:"unfulfilled_orders"`
	TotalDiscount     float64          `json:"total_discount"`
	DiscountPercent   float64          `json:"discount_percent"`
	ChannelBreakdown  []ChannelMetrics `json:"channel_breakdown"`
	PaymentBreakdown  PaymentHealth    `json:"payment_breakdown"`
}

type ChannelMetrics struct {
	SourceID string  `json:"source_id"`
	Revenue  float64 `json:"revenue"`
	Orders   int     `json:"orders"`
	AOV      float64 `json:"aov"`
}

type PaymentHealth struct {
	Paid      int `json:"paid"`
	Pending   int `json:"pending"`
	Partial   int `json:"partial"`
	Cancelled int `json:"cancelled"`
}

type TopProductRow struct {
	SKU      string  `json:"sku"`
	Title    string  `json:"title"`
	Quantity int     `json:"quantity"`
	Revenue  float64 `json:"revenue"`
}

type RevenueTrendRow struct {
	Date    string  `json:"date"`
	Revenue float64 `json:"revenue"`
	Orders  int     `json:"orders"`
}

type GeoDistributionRow struct {
	State   string  `json:"state"`
	Orders  int     `json:"orders"`
	Revenue float64 `json:"revenue"`
}


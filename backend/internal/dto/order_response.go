package dto

// OrderResponse is the DTO returned by the GET /api/orders endpoint.
// It matches what the frontend expects.
type OrderResponse struct {
	ID                string `json:"id"`
	OrderNumber       string `json:"order_number"`
	TotalPrice        string `json:"total_price"`
	SubtotalPrice     string `json:"subtotal_price"`
	TotalTax          string `json:"total_tax"`
	Currency          string `json:"currency"`
	FinancialStatus   string `json:"financial_status"`
	FulfillmentStatus string `json:"fulfillment_status"`
	DeliveryStatus    string `json:"delivery_status"`
	TrackingNumber    string `json:"tracking_number"`
	ShippingCompany   string `json:"shipping_company"`
	TrackingUrl       string `json:"tracking_url"`
	SourceID          string `json:"source_id"`
	Status            string `json:"status"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
	CancelledAt       string `json:"cancelled_at,omitempty"`
	CancelReason      string `json:"cancel_reason,omitempty"`
	CustomerName      string `json:"customer_name"`
	CustomerFirstName string `json:"customer_first_name"`
	CustomerLastName  string `json:"customer_last_name"`
	CustomerEmail     string `json:"customer_email"`
	CustomerPhone     string `json:"customer_phone"`
	CustomerCity      string `json:"customer_city"`
	CustomerState     string `json:"customer_state"`
	CustomerCountry   string `json:"customer_country"`
	CustomerAddress1  string `json:"customer_address1"`
	CustomerAddress2  string `json:"customer_address2"`
	CustomerZip       string `json:"customer_zip"`
	LineItems         []LineItemResponse `json:"line_items,omitempty"`
}

// LineItemResponse is the DTO for order line items in API responses.
type LineItemResponse struct {
	ID        string `json:"id"`
	ProductID string `json:"product_id"`
	VariantID string `json:"variant_id"`
	Title     string `json:"title"`
	SKU       string `json:"sku"`
	HSCode    string `json:"hs_code"`
	Quantity  int    `json:"quantity"`
	Price     string `json:"price"`
	Discount  string `json:"discount"`
}

// OrderListResponse wraps a paginated list of orders for the API.
type OrderListResponse struct {
	Success    bool            `json:"success"`
	Orders     []OrderResponse `json:"orders"`
	TotalCount int             `json:"total_count"`
	Page       int             `json:"page"`
	Limit      int             `json:"limit"`
}

// OrderUpdateRequest is the DTO for updating an order from the frontend.
type OrderUpdateRequest struct {
	CustomerFirstName string `json:"customer_first_name"`
	CustomerLastName  string `json:"customer_last_name"`
	CustomerEmail     string `json:"customer_email"`
	CustomerPhone     string `json:"customer_phone"`
	CustomerAddress1  string `json:"customer_address1"`
	CustomerAddress2  string `json:"customer_address2"`
	CustomerCity      string `json:"customer_city"`
	CustomerState     string `json:"customer_state"`
	CustomerZip       string `json:"customer_zip"`
	CustomerCountry   string `json:"customer_country"`
}

// WebhookEventResponse is the DTO for webhook event log entries.
type WebhookEventResponse struct {
	ID                int    `json:"id"`
	SourceID          string `json:"source_id"`
	OrderID           string `json:"order_id,omitempty"`
	Topic             string `json:"topic"`
	ExternalID        string `json:"external_id"`
	WebhookDeliveryID string `json:"webhook_delivery_id"`
	Processed         bool   `json:"processed"`
	CreatedAt         string `json:"created_at"`
}

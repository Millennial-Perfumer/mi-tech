package dto

// ShopifyWebhookOrder represents the REST payload from Shopify order webhooks.
type ShopifyWebhookOrder struct {
	ID                int64                `json:"id"`
	OrderNumber       int64                `json:"order_number"`
	Name              string               `json:"name"`
	Email             string               `json:"email"`
	TotalPrice        string               `json:"total_price"`
	SubtotalPrice     string               `json:"subtotal_price"`
	TotalTax          string               `json:"total_tax"`
	TotalDiscounts    string               `json:"total_discounts"`
	Currency          string               `json:"currency"`
	FinancialStatus   string               `json:"financial_status"`
	FulfillmentStatus string               `json:"fulfillment_status"`
	SourceName        string               `json:"source_name"`
	CreatedAt         string               `json:"created_at"`
	UpdatedAt         string               `json:"updated_at"`
	CancelledAt       *string              `json:"cancelled_at"`
	CancelReason      string               `json:"cancel_reason"`
	Customer          *ShopifyCustomer     `json:"customer"`
	BillingAddress    *ShopifyAddress      `json:"billing_address"`
	ShippingAddress   *ShopifyAddress      `json:"shipping_address"`
	LineItems         []ShopifyLineItem    `json:"line_items"`
	Fulfillments      []ShopifyFulfillment `json:"fulfillments"`
}

// ShopifyWebhookFulfillment represents the payload from Shopify fulfillment webhooks.
type ShopifyWebhookFulfillment struct {
	ID              int64   `json:"id"`
	OrderID         int64   `json:"order_id"`
	Status          string  `json:"status"`
	ShipmentStatus  *string `json:"shipment_status"`
	TrackingCompany string  `json:"tracking_company"`
	TrackingNumber  string  `json:"tracking_number"`
	TrackingUrl     string  `json:"tracking_url"`
}

// ShopifyFulfillment represents fulfillment data nested within an order webhook.
type ShopifyFulfillment struct {
	ID              int64   `json:"id"`
	Status          string  `json:"status"`
	DisplayStatus   string  `json:"display_status"`
	ShipmentStatus  *string `json:"shipment_status"`
	TrackingNumber  string  `json:"tracking_number"`
	TrackingCompany string  `json:"tracking_company"`
	TrackingUrl     string  `json:"tracking_url"`
}

// ShopifyCustomer represents customer data from a Shopify webhook payload.
type ShopifyCustomer struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
}

// ShopifyAddress represents an address from a Shopify webhook payload.
type ShopifyAddress struct {
	Name      string `json:"name"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
	Address1  string `json:"address1"`
	Address2  string `json:"address2"`
	City      string `json:"city"`
	Province  string `json:"province"`
	Country   string `json:"country"`
	Zip       string `json:"zip"`
}

// ShopifyLineItem represents a single line item from a Shopify webhook payload.
type ShopifyLineItem struct {
	ID              int64  `json:"id"`
	ProductID       int64  `json:"product_id"`
	VariantID       int64  `json:"variant_id"`
	Title           string `json:"title"`
	Quantity        int    `json:"quantity"`
	CurrentQuantity *int   `json:"current_quantity"`
	Price           string `json:"price"`
	SKU             string `json:"sku"`
	TaxLines        []struct {
		Title string  `json:"title"`
		Price string  `json:"price"`
		Rate  float64 `json:"rate"`
	} `json:"tax_lines"`
	TotalDiscount       string                     `json:"total_discount"`
	DiscountAllocations []ShopifyDiscountAllocation `json:"discount_allocations"`
}

// ShopifyDiscountAllocation represents a discount allocation in the REST webhook payload.
type ShopifyDiscountAllocation struct {
	Amount string `json:"amount"`
}

// ShopifyWebhookCustomer represents the REST payload from Shopify customer webhooks.
type ShopifyWebhookCustomer struct {
	ID             int64            `json:"id"`
	Email          string           `json:"email"`
	FirstName      string           `json:"first_name"`
	LastName       string           `json:"last_name"`
	Phone          string           `json:"phone"`
	State          string           `json:"state"`
	OrdersCount    int              `json:"orders_count"`
	TotalSpent     string           `json:"total_spent"`
	TaxExempt      bool             `json:"tax_exempt"`
	Tags           string           `json:"tags"`
	Currency       string           `json:"currency"`
	CreatedAt      string           `json:"created_at"`
	UpdatedAt      string           `json:"updated_at"`
	DefaultAddress *ShopifyAddress  `json:"default_address"`
	Addresses      []ShopifyAddress `json:"addresses"`
}

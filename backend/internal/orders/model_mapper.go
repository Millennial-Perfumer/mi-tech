package orders

import (
	"shopify-gst-app/internal/models"
	"strconv"
)

// ShopifyWebhookOrder represents the REST payload from Shopify webhooks
type ShopifyWebhookOrder struct {
	ID                int64             `json:"id"`
	OrderNumber       int64             `json:"order_number"`
	Email             string            `json:"email"`
	TotalPrice        string            `json:"total_price"`
	SubtotalPrice     string            `json:"subtotal_price"`
	TotalTax          string            `json:"total_tax"`
	Currency          string            `json:"currency"`
	FinancialStatus   string            `json:"financial_status"`
	FulfillmentStatus string            `json:"fulfillment_status"`
	CreatedAt         string            `json:"created_at"`
	UpdatedAt         string            `json:"updated_at"`
	CancelledAt       *string           `json:"cancelled_at"`
	CancelReason      string            `json:"cancel_reason"`
	Customer          *ShopifyCustomer  `json:"customer"`
	BillingAddress    *ShopifyAddress   `json:"billing_address"`
	ShippingAddress   *ShopifyAddress   `json:"shipping_address"`
	LineItems         []ShopifyLineItem `json:"line_items"`
}

type ShopifyCustomer struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
}

type ShopifyAddress struct {
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	City     string `json:"city"`
	Province string `json:"province"`
	Country  string `json:"country"`
}

type ShopifyLineItem struct {
	ID        int64  `json:"id"`
	ProductID int64  `json:"product_id"`
	VariantID int64  `json:"variant_id"`
	Title     string `json:"title"`
	Quantity  int    `json:"quantity"`
	Price     string `json:"price"`
	SKU       string `json:"sku"`
	TaxLines  []struct {
		Title string  `json:"title"`
		Price string  `json:"price"`
		Rate  float64 `json:"rate"`
	} `json:"tax_lines"`
	TotalDiscount string `json:"total_discount"`
}

func MapWebhookToOrder(payload ShopifyWebhookOrder) models.Order {
	customerName := ""
	if payload.Customer != nil {
		customerName = payload.Customer.FirstName + " " + payload.Customer.LastName
	} else if payload.ShippingAddress != nil {
		customerName = payload.ShippingAddress.Name
	}

	city, state, country := "", "", ""
	if payload.ShippingAddress != nil {
		city = payload.ShippingAddress.City
		state = payload.ShippingAddress.Province
		country = payload.ShippingAddress.Country
	}

	phone := ""
	if payload.ShippingAddress != nil && payload.ShippingAddress.Phone != "" {
		phone = payload.ShippingAddress.Phone
	} else if payload.Customer != nil {
		phone = payload.Customer.Phone
	}

	order := models.Order{
		ID:                strconv.FormatInt(payload.ID, 10),
		ShopifyOrderID:    strconv.FormatInt(payload.ID, 10),
		OrderNumber:       strconv.FormatInt(payload.OrderNumber, 10),
		TotalPrice:        payload.TotalPrice,
		SubtotalPrice:     payload.SubtotalPrice,
		TotalTax:          payload.TotalTax,
		Currency:          payload.Currency,
		FinancialStatus:   payload.FinancialStatus,
		FulfillmentStatus: payload.FulfillmentStatus,
		Status:            payload.FulfillmentStatus, // Fallback for existing status field
		CreatedAt:         payload.CreatedAt,
		UpdatedAt:         payload.UpdatedAt,
		CancelledAt:       payload.CancelledAt,
		CancelReason:      payload.CancelReason,
		CustomerName:      customerName,
		CustomerEmail:     payload.Email,
		CustomerPhone:     phone,
		CustomerCity:      city,
		CustomerState:     state,
		CustomerCountry:   country,
	}

	for _, li := range payload.LineItems {
		item := models.LineItem{
			ID:        strconv.FormatInt(li.ID, 10),
			ProductID: strconv.FormatInt(li.ProductID, 10),
			VariantID: strconv.FormatInt(li.VariantID, 10),
			Title:     li.Title,
			SKU:       li.SKU,
			Quantity:  li.Quantity,
			Price:     li.Price,
			Discount:  li.TotalDiscount,
		}

		// Attempt to extract HS code if title contains hints or use a default
		// In a real scenario, this would come from the product/variant resource
		// but since we only have the webhook payload, we might need to rely on previous data or a default.
		item.HSCode = "33029019" // Default HSN

		order.LineItems = append(order.LineItems, item)
	}

	return order
}

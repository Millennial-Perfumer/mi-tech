package orders

import (
	"encoding/json"
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

func MapWebhookToOrder(payload ShopifyWebhookOrder, rawPayload *json.RawMessage) models.Order {
	// Prioritize Billing Address for GST/Invoicing names and locations
	customerName := ""
	firstName, lastName := "", ""

	if payload.Customer != nil {
		firstName = payload.Customer.FirstName
		lastName = payload.Customer.LastName
	}

	if payload.BillingAddress != nil && payload.BillingAddress.Name != "" {
		customerName = payload.BillingAddress.Name
		if firstName == "" {
			firstName = payload.BillingAddress.FirstName
		}
		if lastName == "" {
			lastName = payload.BillingAddress.LastName
		}
	} else if payload.ShippingAddress != nil && payload.ShippingAddress.Name != "" {
		customerName = payload.ShippingAddress.Name
		if firstName == "" {
			firstName = payload.ShippingAddress.FirstName
		}
		if lastName == "" {
			lastName = payload.ShippingAddress.LastName
		}
	}

	if customerName == "" && (firstName != "" || lastName != "") {
		customerName = (firstName + " " + lastName)
	}

	city, state, country, addr1, addr2, zip := "", "", "", "", "", ""
	addr := payload.BillingAddress
	if addr == nil {
		addr = payload.ShippingAddress
	}

	if addr != nil {
		city = addr.City
		state = addr.Province
		country = addr.Country
		addr1 = addr.Address1
		addr2 = addr.Address2
		zip = addr.Zip
	}

	phone := ""
	if payload.BillingAddress != nil && payload.BillingAddress.Phone != "" {
		phone = payload.BillingAddress.Phone
	} else if payload.ShippingAddress != nil && payload.ShippingAddress.Phone != "" {
		phone = payload.ShippingAddress.Phone
	} else if payload.Customer != nil {
		phone = payload.Customer.Phone
	}

	email := payload.Email
	if email == "" && payload.Customer != nil {
		email = payload.Customer.Email
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
		CustomerFirstName: firstName,
		CustomerLastName:  lastName,
		CustomerEmail:     email,
		CustomerPhone:     phone,
		CustomerCity:      city,
		CustomerState:     state,
		CustomerCountry:   country,
		CustomerAddress1:  addr1,
		CustomerAddress2:  addr2,
		CustomerZip:       zip,
		RawPayload:        rawPayload,
	}

	for _, li := range payload.LineItems {
		price := li.Price
		if price == "" {
			price = "0.00"
		}
		discount := li.TotalDiscount
		if discount == "" {
			discount = "0.00"
		}
		item := models.LineItem{
			ID:        strconv.FormatInt(li.ID, 10),
			ProductID: strconv.FormatInt(li.ProductID, 10),
			VariantID: strconv.FormatInt(li.VariantID, 10),
			Title:     li.Title,
			SKU:       li.SKU,
			Quantity:  li.Quantity,
			Price:     price,
			Discount:  discount,
		}

		// Attempt to extract HS code if title contains hints or use a default
		// In a real scenario, this would come from the product/variant resource
		// but since we only have the webhook payload, we might need to rely on previous data or a default.
		item.HSCode = "33029019" // Default HSN

		order.LineItems = append(order.LineItems, item)
	}

	return order
}

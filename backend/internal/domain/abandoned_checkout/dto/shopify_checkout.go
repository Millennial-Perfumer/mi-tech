package dto

type ShopifyWebhookCheckout struct {
	ID                   int64                 `json:"id"`
	Token                string                `json:"token"`
	CartToken            string                `json:"cart_token"`
	Email                string                `json:"email"`
	Phone                string                `json:"phone"`
	TotalPrice           string                `json:"total_price"`
	Currency             string                `json:"currency"`
	AbandonedCheckoutURL string                `json:"abandoned_checkout_url"`
	BuyerAcceptsMarketing bool                 `json:"buyer_accepts_marketing"`
	SMSMarketingConsent  *SMSMarketingConsent  `json:"sms_marketing_consent"`
	Customer             *ShopifyCustomer      `json:"customer"`
	ShippingAddress      *ShopifyAddress       `json:"shipping_address"`
	BillingAddress       *ShopifyAddress       `json:"billing_address"`
	LineItems            []ShopifyCheckoutItem `json:"line_items"`
	AbandonedAt          *string               `json:"abandoned_at"`
	CreatedAt            string                `json:"created_at"`
	UpdatedAt            string                `json:"updated_at"`
}

type SMSMarketingConsent struct {
	State        string `json:"state"` // accepted, declined, pending
	OptInLevel   string `json:"opt_in_level"`
	ConsentCollectedFrom string `json:"consent_collected_from"`
}

type ShopifyCustomer struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
}

type ShopifyAddress struct {
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

type ShopifyCheckoutItem struct {
	ID         int64  `json:"id"`
	ProductID  int64  `json:"product_id"`
	VariantID  int64  `json:"variant_id"`
	Title      string `json:"title"`
	Quantity   int    `json:"quantity"`
	Price      string `json:"price"`
	SKU        string `json:"sku"`
}

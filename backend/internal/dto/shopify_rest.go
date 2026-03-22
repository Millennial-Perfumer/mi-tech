package dto

type ShopifyCustomerWrapper struct {
	Customer ShopifyRestCustomer `json:"customer"`
}

type ShopifyRestCustomer struct {
	ID        int64                `json:"id,omitempty"`
	FirstName string               `json:"first_name,omitempty"`
	LastName  string               `json:"last_name,omitempty"`
	Email     string               `json:"email,omitempty"`
	Phone     string               `json:"phone,omitempty"`
	Addresses []ShopifyRestAddress `json:"addresses,omitempty"`
}

type ShopifyRestAddress struct {
	ID       int64  `json:"id,omitempty"`
	Address1 string `json:"address1,omitempty"`
	Address2 string `json:"address2,omitempty"`
	City     string `json:"city,omitempty"`
	Province string `json:"province,omitempty"`
	Country  string `json:"country,omitempty"`
	Zip      string `json:"zip,omitempty"`
	Default  bool   `json:"default,omitempty"`
}

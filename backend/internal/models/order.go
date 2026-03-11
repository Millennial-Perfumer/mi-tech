package models

import "encoding/json"

// GraphQLOrderResponse represents the top-level JSON from the GraphQL endpoint
type GraphQLOrderResponse struct {
	Data struct {
		Orders struct {
			PageInfo struct {
				HasNextPage bool   `json:"hasNextPage"`
				EndCursor   string `json:"endCursor"`
			} `json:"pageInfo"`
			Edges []GraphQLOrderEdge `json:"edges"`
		} `json:"orders"`
	} `json:"data"`
	Errors []interface{} `json:"errors"`
}

type GraphQLOrderEdge struct {
	Node GraphQLOrderNode `json:"node"`
}

type GraphQLOrderNode struct {
	ID                       string               `json:"id"`
	Name                     string               `json:"name"`
	Email                    string               `json:"email"`
	CurrentTotalPriceSet     TotalPriceSet        `json:"currentTotalPriceSet"`
	CurrentSubtotalPriceSet  TotalPriceSet        `json:"currentSubtotalPriceSet"`
	CurrentTotalTaxSet       TotalPriceSet        `json:"currentTotalTaxSet"`
	TotalPriceSet            TotalPriceSet        `json:"totalPriceSet"`
	SourceName               string               `json:"sourceName"`
	ProcessedAt              string               `json:"processedAt"`
	CreatedAt                string               `json:"createdAt"`
	UpdatedAt                string               `json:"updatedAt"`
	DisplayFinancialStatus   string               `json:"displayFinancialStatus"`
	DisplayFulfillmentStatus string               `json:"displayFulfillmentStatus"`
	Customer                 *GraphQLCustomer     `json:"customer"`
	BillingAddress           *GraphQLAddress      `json:"billingAddress"`
	ShippingAddress          *GraphQLAddress      `json:"shippingAddress"`
	Fulfillments             []GraphQLFulfillment `json:"fulfillments"`
	LineItems                GraphQLWrapLine      `json:"lineItems"`
}

type GraphQLWrapLine struct {
	Edges []GraphQLLineEdge `json:"edges"`
}

type GraphQLLineEdge struct {
	Node GraphQLLineNode `json:"node"`
}

type GraphQLLineNode struct {
	ID                   string              `json:"id"`
	Title                string              `json:"title"`
	SKU                  string              `json:"sku"`
	Quantity             int                 `json:"quantity"`
	Vendor               string              `json:"vendor"`
	OriginalTotalSet     TotalPriceSet       `json:"originalTotalSet"`
	TotalDiscountSet     TotalPriceSet       `json:"totalDiscountSet"`
	OriginalUnitPriceSet TotalPriceSet       `json:"originalUnitPriceSet"`
	Variant              *GraphQLLineVariant `json:"variant"`
}

type GraphQLLineVariant struct {
	InventoryItem struct {
		HarmonizedSystemCode string `json:"harmonizedSystemCode"`
	} `json:"inventoryItem"`
}

type TotalPriceSet struct {
	ShopMoney ShopMoney `json:"shopMoney"`
}

type ShopMoney struct {
	Amount string `json:"amount"`
}

type GraphQLCustomer struct {
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	DisplayName string `json:"displayName"`
}

type GraphQLAddress struct {
	Name      string `json:"name"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Phone     string `json:"phone"`
	City      string `json:"city"`
	Province  string `json:"province"`
	Country   string `json:"country"`
}

type GraphQLFulfillment struct {
	Status        string                `json:"status"`
	DisplayStatus string                `json:"displayStatus"`
	CreatedAt     string                `json:"createdAt"`
	TrackingInfo  []GraphQLTrackingInfo `json:"trackingInfo"`
}

type GraphQLTrackingInfo struct {
	Number  string `json:"number"`
	Company string `json:"company"`
	Url     string `json:"url"`
}

// Order represents our internal schema matching the PostgreSQL database
type Order struct {
	ID                string           `json:"id"`
	StoreID           string           `json:"store_id"`
	ExternalOrderID   string           `json:"external_order_id"`
	SourceID          string           `json:"source_id"`
	OrderNumber       string           `json:"order_number"`
	TotalPrice        string           `json:"total_price"`
	SubtotalPrice     string           `json:"subtotal_price"`
	TotalTax          string           `json:"total_tax"`
	Currency          string           `json:"currency"`
	FinancialStatus   string           `json:"financial_status"`
	FulfillmentStatus string           `json:"fulfillment_status"`
	DeliveryStatus    string           `json:"delivery_status"`
	TrackingNumber    string           `json:"tracking_number"`
	ShippingCompany   string           `json:"shipping_company"`
	TrackingUrl       string           `json:"tracking_url"`
	Status            string           `json:"status"`
	CreatedAt         string           `json:"created_at"`
	UpdatedAt         string           `json:"updated_at"`
	CancelledAt       *string          `json:"cancelled_at"`
	CancelReason      string           `json:"cancel_reason"`
	CustomerName      string           `json:"customer_name"`
	CustomerFirstName string           `json:"customer_first_name"`
	CustomerLastName  string           `json:"customer_last_name"`
	CustomerEmail     string           `json:"customer_email"`
	CustomerPhone     string           `json:"customer_phone"`
	CustomerCity      string           `json:"customer_city"`
	CustomerState     string           `json:"customer_state"`
	CustomerCountry   string           `json:"customer_country"`
	CustomerAddress1  string           `json:"customer_address1"`
	CustomerAddress2  string           `json:"customer_address2"`
	CustomerZip       string           `json:"customer_zip"`
	LineItems         []LineItem       `json:"line_items,omitempty"`
	RawPayload        *json.RawMessage `json:"raw_payload,omitempty"`
}

type LineItem struct {
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

package dto

// GraphQLOrderResponse represents the top-level JSON from the Shopify GraphQL Admin API.
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

// GraphQLOrderEdge wraps a single order node from the GraphQL response.
type GraphQLOrderEdge struct {
	Node GraphQLOrderNode `json:"node"`
}

// GraphQLOrderNode represents a single order from the Shopify GraphQL response.
type GraphQLOrderNode struct {
	ID                       string               `json:"id"`
	Name                     string               `json:"name"`
	Email                    string               `json:"email"`
	CurrentTotalPriceSet     MoneySet             `json:"currentTotalPriceSet"`
	CurrentSubtotalPriceSet  MoneySet             `json:"currentSubtotalPriceSet"`
	CurrentTotalTaxSet       MoneySet             `json:"currentTotalTaxSet"`
	CurrentTotalDiscountsSet MoneySet             `json:"currentTotalDiscountsSet"`
	TotalPriceSet            MoneySet             `json:"totalPriceSet"`
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
	LineItems                GraphQLLineItemWrap  `json:"lineItems"`
	CancelledAt              string               `json:"cancelledAt"`
	CancelReason             string               `json:"cancelReason"`
}

// MoneySet wraps a shopMoney amount (replaces the old TotalPriceSet).
type MoneySet struct {
	ShopMoney ShopMoney `json:"shopMoney"`
}

// ShopMoney holds the actual amount string.
type ShopMoney struct {
	Amount string `json:"amount"`
}

// GraphQLCustomer represents customer data from the GraphQL response.
type GraphQLCustomer struct {
	ID          string `json:"id"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	DisplayName string `json:"displayName"`
	Phone       string `json:"phone"`
}

// GraphQLAddress represents an address from the GraphQL response.
type GraphQLAddress struct {
	Name      string `json:"name"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Phone     string `json:"phone"`
	City      string `json:"city"`
	Province  string `json:"province"`
	Country   string `json:"country"`
	Address1  string `json:"address1"`
	Address2  string `json:"address2"`
	Zip       string `json:"zip"`
}

// GraphQLFulfillment contains fulfillment & tracking data.
type GraphQLFulfillment struct {
	Status        string                   `json:"status"`
	DisplayStatus string                   `json:"displayStatus"`
	CreatedAt     string                   `json:"createdAt"`
	TrackingInfo  []GraphQLTrackingInfo    `json:"trackingInfo"`
	Events        GraphQLFulfillmentEvents `json:"events"`
}

// GraphQLFulfillmentEvents wraps the edges array for fulfillment events.
type GraphQLFulfillmentEvents struct {
	Edges []GraphQLFulfillmentEventEdge `json:"edges"`
}

// GraphQLFulfillmentEventEdge wraps a single fulfillment event node.
type GraphQLFulfillmentEventEdge struct {
	Node GraphQLFulfillmentEventNode `json:"node"`
}

// GraphQLFulfillmentEventNode represents a single fulfillment event from the Shopify GraphQL response.
type GraphQLFulfillmentEventNode struct {
	Status     string `json:"status"`
	HappenedAt string `json:"happenedAt"`
}

// GraphQLTrackingInfo holds a single tracking entry.
type GraphQLTrackingInfo struct {
	Number  string `json:"number"`
	Company string `json:"company"`
	Url     string `json:"url"`
}

// GraphQLLineItemWrap wraps the edges array for line items.
type GraphQLLineItemWrap struct {
	Edges []GraphQLLineItemEdge `json:"edges"`
}

// GraphQLLineItemEdge wraps a single line item node.
type GraphQLLineItemEdge struct {
	Node GraphQLLineItemNode `json:"node"`
}

// GraphQLLineItemNode represents a single line item from the GraphQL response.
type GraphQLLineItemNode struct {
	ID                   string              `json:"id"`
	Title                string              `json:"title"`
	SKU                  string              `json:"sku"`
	Quantity             int                 `json:"quantity"`
	CurrentQuantity      *int                `json:"currentQuantity"`
	Vendor               string              `json:"vendor"`
	OriginalTotalSet     MoneySet            `json:"originalTotalSet"`
	TotalDiscountSet     MoneySet            `json:"totalDiscountSet"`
	OriginalUnitPriceSet MoneySet            `json:"originalUnitPriceSet"`
	Variant              *GraphQLLineVariant `json:"variant"`
	DiscountAllocations  []DiscountAllocation `json:"discountAllocations"`
}

// DiscountAllocation represents an allocated portion of a discount (e.g., from a coupon).
type DiscountAllocation struct {
	AllocatedAmount MoneySet `json:"allocatedAmount"`
}

// GraphQLLineVariant holds variant-level data such as HS codes.
type GraphQLLineVariant struct {
	InventoryItem struct {
		HarmonizedSystemCode string `json:"harmonizedSystemCode"`
	} `json:"inventoryItem"`
}

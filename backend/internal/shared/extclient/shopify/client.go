package shopify

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"mi-tech/internal/shared/config"
)

type Client struct {
	settings   *config.SettingsProvider
	httpClient *http.Client
}

func NewClient(settings *config.SettingsProvider) *Client {
	return &Client{
		settings: settings,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// baselineDate is the earliest date we care about for orders (2026-01-01).
var baselineDate = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

// FetchOrders fetches orders from Shopify using the GraphQL Admin API, extracting specific location vectors.
func (c *Client) FetchOrders(since time.Time, to time.Time) ([]GraphQLOrderNode, error) {
	shopifyURL := c.settings.GetShopifyStoreURL()
	accessToken := c.settings.GetShopifyAccessToken()

	if shopifyURL == "" || accessToken == "" {
		return nil, fmt.Errorf("shopify credentials are not configured in DB")
	}

	var allOrders []GraphQLOrderNode
	apiURL := fmt.Sprintf("https://%s/admin/api/%s/graphql.json", shopifyURL, c.settings.GetShopifyAPIVersion())

	// Build the search query dynamically.
	searchQuery := fmt.Sprintf("updated_at:>'%s' AND updated_at:<='%s'", since.Format(time.RFC3339), to.Format(time.RFC3339))

	// If since is the default/zero value, we still enforce the 2026-01-01 baseline
	if since.Before(baselineDate) {
		searchQuery = fmt.Sprintf("created_at:>='%s' AND updated_at:>'%s' AND updated_at:<='%s'",
			baselineDate.Format(time.RFC3339), since.Format(time.RFC3339), to.Format(time.RFC3339))
	}

	queryTemplate := `
	query getOrders($cursor: String, $query: String!) {
		orders(first: 250, after: $cursor, query: $query) {
			pageInfo {
				hasNextPage
				endCursor
			}
			edges {
				node {
					id
					name
					processedAt
					createdAt
					updatedAt
					displayFinancialStatus
					displayFulfillmentStatus
					currentTotalPriceSet {
						shopMoney {
							amount
						}
					}
					currentSubtotalPriceSet {
						shopMoney {
							amount
						}
					}
					currentTotalTaxSet {
						shopMoney {
							amount
						}
					}
					currentTotalDiscountsSet {
						shopMoney {
							amount
						}
					}
					totalPriceSet {
						shopMoney {
							amount
						}
					}
					sourceName
					billingAddress {
						city
						province
						country
					}
					shippingAddress {
						city
						province
						country
					}
					fulfillments {
						status
						displayStatus
						createdAt
						trackingInfo {
							number
							company
							url
						}
						events(first: 10) {
							edges {
								node {
									status
									happenedAt
								}
							}
						}
					}
					cancelledAt
					cancelReason
					lineItems(first: 50) {
						edges {
							node {
								id
								title
								sku
								quantity
								totalDiscountSet {
									shopMoney {
										amount
									}
								}
								originalTotalSet {
									shopMoney {
										amount
									}
								}
								originalUnitPriceSet {
									shopMoney {
										amount
									}
								}
								currentQuantity
								discountAllocations {
									allocatedAmount {
										amount
									}
								}
								variant {
									inventoryItem {
										harmonizedSystemCode
									}
								}
								vendor
							}
						}
					}
				}
			}
		}
	}
	`

	var cursor *string

	for {
		// Prepare GraphQL payload
		payload := map[string]interface{}{
			"query": queryTemplate,
			"variables": map[string]interface{}{
				"query":  searchQuery,
				"cursor": cursor,
			},
		}

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}

		req, err := http.NewRequest("POST", apiURL, strings.NewReader(string(payloadBytes)))
		if err != nil {
			return nil, err
		}

		req.Header.Add("X-Shopify-Access-Token", c.settings.GetShopifyAccessToken())
		req.Header.Add("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}

		body, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("shopify graphql api error: %s - %s", resp.Status, string(body))
		}

		var result GraphQLOrderResponse
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal graphql response: %w", err)
		}

		// Log GraphQL errors but continue if we have data
		if len(result.Errors) > 0 {
			log.Printf("Shopify GraphQL semi-successful with errors: %v", result.Errors)
		}

		// Extract edges into nodes
		edges := result.Data.Orders.Edges
		for _, edge := range edges {
			allOrders = append(allOrders, edge.Node)
		}

		// Handle pagination
		if result.Data.Orders.PageInfo.HasNextPage && result.Data.Orders.PageInfo.EndCursor != "" {
			endCursor := result.Data.Orders.PageInfo.EndCursor
			cursor = &endCursor
			// To respect rate limits
			time.Sleep(500 * time.Millisecond)
		} else {
			break
		}
	}

	return allOrders, nil
}

// FetchOrderByID fetches a single order from Shopify using GraphQL.
func (c *Client) FetchOrderByID(id string) (*GraphQLOrderNode, error) {
	shopifyURL := c.settings.GetShopifyStoreURL()
	accessToken := c.settings.GetShopifyAccessToken()

	if shopifyURL == "" || accessToken == "" {
		return nil, fmt.Errorf("shopify credentials are not configured in DB")
	}

	apiURL := fmt.Sprintf("https://%s/admin/api/%s/graphql.json", shopifyURL, c.settings.GetShopifyAPIVersion())

	// Ensure the ID is in the correct GID format
	gid := id
	if !strings.HasPrefix(gid, "gid://shopify/Order/") {
		gid = "gid://shopify/Order/" + id
	}

	query := `
	query getOrder($id: ID!) {
		order(id: $id) {
			id
			name
			processedAt
			createdAt
			updatedAt
			displayFinancialStatus
			displayFulfillmentStatus
			currentTotalPriceSet { shopMoney { amount } }
			currentSubtotalPriceSet { shopMoney { amount } }
			currentTotalTaxSet { shopMoney { amount } }
			currentTotalDiscountsSet { shopMoney { amount } }
			sourceName
			billingAddress {
				city
				province
				country
			}
			shippingAddress {
				city
				province
				country
			}
			fulfillments {
				id
				status
				displayStatus
				createdAt
				trackingInfo {
					number
					company
					url
				}
				events(first: 10) {
					edges {
						node {
							status
							happenedAt
						}
					}
				}
			}
			cancelledAt
			cancelReason
			lineItems(first: 50) {
				edges {
					node {
						id
						title
						sku
						quantity
						totalDiscountSet { shopMoney { amount } }
						originalTotalSet { shopMoney { amount } }
						currentQuantity
						discountAllocations {
							allocatedAmount {
								amount
							}
						}
						variant {
							inventoryItem {
								harmonizedSystemCode
							}
						}
					}
				}
			}
		}
	}
	`

	payload := map[string]interface{}{
		"query": query,
		"variables": map[string]interface{}{
			"id": gid,
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return nil, err
	}

	req.Header.Add("X-Shopify-Access-Token", c.settings.GetShopifyAccessToken())
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("shopify graphql api error: %s - %s", resp.Status, string(body))
	}

	var result struct {
		Data struct {
			Order *GraphQLOrderNode `json:"order"`
		} `json:"data"`
		Errors []interface{} `json:"errors"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal graphql response: %w", err)
	}

	if len(result.Errors) > 0 {
		return nil, fmt.Errorf("shopify graphql error: %v", result.Errors)
	}

	return result.Data.Order, nil
}

// CreateCustomer creates a customer in Shopify using the REST API.
func (c *Client) CreateCustomer(customer ShopifyRestCustomer) (*ShopifyRestCustomer, error) {
	shopifyURL := c.settings.GetShopifyStoreURL()
	accessToken := c.settings.GetShopifyAccessToken()
	apiVersion := c.settings.GetShopifyAPIVersion()

	if shopifyURL == "" || accessToken == "" {
		return nil, fmt.Errorf("shopify credentials not configured")
	}

	apiURL := fmt.Sprintf("https://%s/admin/api/%s/customers.json", shopifyURL, apiVersion)
	wrapper := ShopifyCustomerWrapper{Customer: customer}
	payload, _ := json.Marshal(wrapper)

	req, _ := http.NewRequest("POST", apiURL, strings.NewReader(string(payload)))
	req.Header.Add("X-Shopify-Access-Token", accessToken)
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("shopify error (%d): %s", resp.StatusCode, string(body))
	}

	var result ShopifyCustomerWrapper
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result.Customer, nil
}

// UpdateCustomer updates an existing customer in Shopify using the REST API.
func (c *Client) UpdateCustomer(id int64, customer ShopifyRestCustomer) (*ShopifyRestCustomer, error) {
	shopifyURL := c.settings.GetShopifyStoreURL()
	accessToken := c.settings.GetShopifyAccessToken()
	apiVersion := c.settings.GetShopifyAPIVersion()

	if shopifyURL == "" || accessToken == "" {
		return nil, fmt.Errorf("shopify credentials not configured")
	}

	apiURL := fmt.Sprintf("https://%s/admin/api/%s/customers/%d.json", shopifyURL, apiVersion, id)
	wrapper := ShopifyCustomerWrapper{Customer: customer}
	payload, _ := json.Marshal(wrapper)

	req, _ := http.NewRequest("PUT", apiURL, strings.NewReader(string(payload)))
	req.Header.Add("X-Shopify-Access-Token", accessToken)
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("shopify error (%d): %s", resp.StatusCode, string(body))
	}

	var result ShopifyCustomerWrapper
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result.Customer, nil
}

// GetCustomer fetches a single customer from Shopify using the REST API.
func (c *Client) GetCustomer(id int64) (*ShopifyRestCustomer, error) {
	shopifyURL := c.settings.GetShopifyStoreURL()
	accessToken := c.settings.GetShopifyAccessToken()
	apiVersion := c.settings.GetShopifyAPIVersion()

	if shopifyURL == "" || accessToken == "" {
		return nil, fmt.Errorf("shopify credentials not configured")
	}

	apiURL := fmt.Sprintf("https://%s/admin/api/%s/customers/%d.json", shopifyURL, apiVersion, id)
	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Add("X-Shopify-Access-Token", accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("shopify error (%d): %s", resp.StatusCode, string(body))
	}

	var result ShopifyCustomerWrapper
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result.Customer, nil
}

// DeleteCustomer deletes a customer from Shopify using the REST API.
func (c *Client) DeleteCustomer(id int64) error {
	shopifyURL := c.settings.GetShopifyStoreURL()
	accessToken := c.settings.GetShopifyAccessToken()
	apiVersion := c.settings.GetShopifyAPIVersion()

	if shopifyURL == "" || accessToken == "" {
		return fmt.Errorf("shopify credentials not configured")
	}

	apiURL := fmt.Sprintf("https://%s/admin/api/%s/customers/%d.json", shopifyURL, apiVersion, id)
	req, _ := http.NewRequest("DELETE", apiURL, nil)
	req.Header.Add("X-Shopify-Access-Token", accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("shopify error (%d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// UpdateOrder updates an existing order in Shopify using the REST API.
// It primarily supports updating the shipping address and customer email.
func (c *Client) UpdateOrder(externalID string, updateData map[string]interface{}) error {
	shopifyURL := c.settings.GetShopifyStoreURL()
	accessToken := c.settings.GetShopifyAccessToken()
	apiVersion := c.settings.GetShopifyAPIVersion()

	if shopifyURL == "" || accessToken == "" {
		return fmt.Errorf("shopify credentials not configured")
	}

	// Extract numeric ID if it's a GraphQL GID
	numericID := c.extractNumericID(externalID)

	apiURL := fmt.Sprintf("https://%s/admin/api/%s/orders/%s.json", shopifyURL, apiVersion, numericID)

	// Wrap in "order" key as required by Shopify REST API
	payload := map[string]interface{}{
		"order": updateData,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal update payload: %w", err)
	}

	req, err := http.NewRequest("PUT", apiURL, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return err
	}

	req.Header.Add("X-Shopify-Access-Token", accessToken)
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute shopify request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("shopify update error (%d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// FetchProducts fetches products and their variants from Shopify using GraphQL.
func (c *Client) FetchProducts() ([]GraphQLProductNode, error) {
	shopifyURL := c.settings.GetShopifyStoreURL()
	accessToken := c.settings.GetShopifyAccessToken()

	if shopifyURL == "" || accessToken == "" {
		return nil, fmt.Errorf("shopify credentials are not configured in DB")
	}

	var allProducts []GraphQLProductNode
	apiURL := fmt.Sprintf("https://%s/admin/api/%s/graphql.json", shopifyURL, c.settings.GetShopifyAPIVersion())

	queryTemplate := `
	query getProducts($cursor: String) {
		products(first: 50, after: $cursor) {
			pageInfo {
				hasNextPage
				endCursor
			}
			edges {
				node {
					id
					title
					descriptionHtml
					handle
					descriptionMetafield: metafield(namespace: "custom", key: "product_description") {
						value
					}
					specificationMetafield: metafield(namespace: "custom", key: "product_specification") {
						value
					}
					variants(first: 50) {
						edges {
							node {
								id
								title
								sku
								price
								inventoryItem {
									id
									inventoryLevels(first: 10) {
										edges {
											node {
												location { id }
												quantities(names: ["available"]) {
													name
													quantity
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	`

	var cursor *string

	for {
		payload := map[string]interface{}{
			"query": queryTemplate,
			"variables": map[string]interface{}{
				"cursor": cursor,
			},
		}

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}

		req, err := http.NewRequest("POST", apiURL, strings.NewReader(string(payloadBytes)))
		if err != nil {
			return nil, err
		}

		req.Header.Add("X-Shopify-Access-Token", accessToken)
		req.Header.Add("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("shopify graphql api error: %s - %s", resp.Status, string(body))
		}

		log.Printf("[DEBUG] Shopify Response Body: %s", string(body))

		var result GraphQLProductResponse
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal graphql response: %w", err)
		}

		if len(result.Errors) > 0 {
			log.Printf("Shopify Product GraphQL semi-successful with errors: %v", result.Errors)
		}

		log.Printf("[DEBUG] Fetched %d product edges", len(result.Data.Products.Edges))

		edges := result.Data.Products.Edges
		for _, edge := range edges {
			allProducts = append(allProducts, edge.Node)
		}

		if result.Data.Products.PageInfo.HasNextPage && result.Data.Products.PageInfo.EndCursor != "" {
			endCursor := result.Data.Products.PageInfo.EndCursor
			cursor = &endCursor
			time.Sleep(500 * time.Millisecond) // Rate limiting
		} else {
			break
		}
	}

	return allProducts, nil
}

// extractNumericID converts a GID like "gid://shopify/Order/12345" to "12345"
func (c *Client) extractNumericID(id string) string {
	if strings.HasPrefix(id, "gid://shopify/") {
		parts := strings.Split(id, "/")
		return parts[len(parts)-1]
	}
	return id
}

// GetLocationID retrieves the configured Shopify location ID from settings.
func (c *Client) GetLocationID() string {
	return c.settings.GetShopifyLocationID()
}

// DiscoverPrimaryLocationID fetches all locations and finds the best match.
func (c *Client) DiscoverPrimaryLocationID(ctx context.Context) (string, error) {
	shopifyURL := c.settings.GetShopifyStoreURL()
	accessToken := c.settings.GetShopifyAccessToken()

	if shopifyURL == "" || accessToken == "" {
		return "", fmt.Errorf("shopify credentials not configured")
	}

	apiURL := fmt.Sprintf("https://%s/admin/api/%s/graphql.json", shopifyURL, c.settings.GetShopifyAPIVersion())

	query := `
	query {
		locations(first: 20) {
			edges {
				node {
					id
					name
					isPrimary
				}
			}
		}
	}
	`

	payload := map[string]interface{}{
		"query": query,
	}

	payloadBytes, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", apiURL, strings.NewReader(string(payloadBytes)))
	req.Header.Add("X-Shopify-Access-Token", accessToken)
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("shopify graphql error (%d): %s", resp.StatusCode, string(body))
	}

	var result GraphQLLocationResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if len(result.Errors) > 0 {
		return "", fmt.Errorf("shopify graphql error: %v", result.Errors)
	}

	edges := result.Data.Locations.Edges
	if len(edges) == 0 {
		return "", fmt.Errorf("no locations found in shopify store")
	}

	// 1. Priority Match: "Millennial Perfumer - WH"
	for _, edge := range edges {
		if strings.TrimSpace(strings.ToLower(edge.Node.Name)) == "millennial perfumer - wh" {
			log.Printf("Shopify Discovery: Matched target location by name: %s", edge.Node.Name)
			return edge.Node.ID, nil
		}
	}

	// 2. Secondary Match: isPrimary
	for _, edge := range edges {
		if edge.Node.IsPrimary {
			log.Printf("Shopify Discovery: Using primary location: %s", edge.Node.Name)
			return edge.Node.ID, nil
		}
	}

	// 3. Last Resort: First one
	log.Printf("Shopify Discovery: Using first available location: %s", edges[0].Node.Name)
	return edges[0].Node.ID, nil
}

// AdjustInventoryLevel updates the stock level for a specific inventory item at a specific location.
func (c *Client) AdjustInventoryLevel(inventoryItemID, locationID string, availableQuantity int) error {
	accessToken := c.settings.GetShopifyAccessToken()
	shopifyURL := c.settings.GetShopifyStoreURL()

	if accessToken == "" || shopifyURL == "" {
		return fmt.Errorf("shopify credentials not configured")
	}

	apiURL := fmt.Sprintf("https://%s/admin/api/%s/graphql.json", shopifyURL, c.settings.GetShopifyAPIVersion())

	// Ensure IDs are in GID format
	if !strings.HasPrefix(inventoryItemID, "gid://") {
		inventoryItemID = "gid://shopify/InventoryItem/" + inventoryItemID
	}
	if !strings.HasPrefix(locationID, "gid://") {
		locationID = "gid://shopify/Location/" + locationID
	}

	mutation := `
	mutation inventorySet($inventoryItemId: ID!, $locationId: ID!, $available: Int!) {
		inventorySetQuantities(input: {
			name: "available",
			reason: "correction",
			ignoreCompareQuantity: true,
			quantities: [
				{
					inventoryItemId: $inventoryItemId,
					locationId: $locationId,
					quantity: $available
				}
			]
		}) {
			inventoryAdjustmentGroup {
				createdAt
			}
			userErrors {
				field
				message
			}
		}
	}
	`

	payload := map[string]interface{}{
		"query": mutation,
		"variables": map[string]interface{}{
			"inventoryItemId": inventoryItemID,
			"locationId":      locationID,
			"available":       availableQuantity,
		},
	}

	payloadBytes, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", apiURL, strings.NewReader(string(payloadBytes)))
	req.Header.Add("X-Shopify-Access-Token", accessToken)
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("shopify graphql error (%d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data struct {
			InventorySetQuantities struct {
				UserErrors []struct {
					Field   []string `json:"field"`
					Message string   `json:"message"`
				} `json:"userErrors"`
			} `json:"inventorySetQuantities"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	if len(result.Data.InventorySetQuantities.UserErrors) > 0 {
		return fmt.Errorf("shopify mutation error: %s", result.Data.InventorySetQuantities.UserErrors[0].Message)
	}

	return nil
}

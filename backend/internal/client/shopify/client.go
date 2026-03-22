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

	"golang.org/x/time/rate"
	"mi-tech/internal/config"
	"mi-tech/internal/dto"
)

type Client struct {
	settings   *config.SettingsProvider
	httpClient *http.Client
	testURL    string // For testing
	limiter    *rate.Limiter
}

func NewClient(settings *config.SettingsProvider) *Client {
	limit := settings.GetShopifyRateLimit()
	burst := settings.GetShopifyRateBurst()

	var rLimit rate.Limit = rate.Inf // Default to no limit
	if limit > 0 {
		rLimit = rate.Limit(limit)
	}

	return &Client{
		settings: settings,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		// Shopify's GraphQL API has a leaky bucket limit.
		limiter: rate.NewLimiter(rLimit, burst),
	}
}

func (c *Client) getAPIURL() string {
	if c.testURL != "" {
		return c.testURL
	}
	shopifyURL := c.settings.GetShopifyStoreURL()
	return fmt.Sprintf("https://%s/admin/api/%s/graphql.json", shopifyURL, c.settings.GetShopifyAPIVersion())
}

// baselineDate is the earliest date we care about for orders (2026-01-01).
var baselineDate = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

// FetchOrders fetches orders from Shopify using the GraphQL Admin API, extracting specific location vectors.
func (c *Client) FetchOrders(ctx context.Context, since time.Time, to time.Time) ([]dto.GraphQLOrderNode, error) {
	shopifyURL := c.settings.GetShopifyStoreURL()
	accessToken := c.settings.GetShopifyAccessToken()

	if shopifyURL == "" || accessToken == "" {
		return nil, fmt.Errorf("shopify credentials are not configured in DB")
	}

	var allOrders []dto.GraphQLOrderNode
	apiURL := c.getAPIURL()

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

		req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(string(payloadBytes)))
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

		var result dto.GraphQLOrderResponse
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
			// Wait for rate limit before next page
			if err := c.limiter.Wait(ctx); err != nil {
				return nil, err
			}
		} else {
			break
		}
	}

	return allOrders, nil
}

// FetchOrderByID fetches a single order from Shopify using GraphQL.
func (c *Client) FetchOrderByID(ctx context.Context, id string) (*dto.GraphQLOrderNode, error) {
	// Respect rate limits
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	shopifyURL := c.settings.GetShopifyStoreURL()
	accessToken := c.settings.GetShopifyAccessToken()

	if shopifyURL == "" || accessToken == "" {
		return nil, fmt.Errorf("shopify credentials are not configured in DB")
	}

	apiURL := c.getAPIURL()

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

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(string(payloadBytes)))
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
			Order *dto.GraphQLOrderNode `json:"order"`
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

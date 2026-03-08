package shopify

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"shopify-gst-app/internal/config"
	"shopify-gst-app/internal/models"
)

type Client struct {
	config     *config.Config
	httpClient *http.Client
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// FetchOrders fetches orders from Shopify using the GraphQL Admin API, extracting specific location vectors.
func (c *Client) FetchOrders(since time.Time) ([]models.GraphQLOrderNode, error) {
	if c.config.ShopifyStoreURL == "" || c.config.ShopifyAccessToken == "" {
		return nil, fmt.Errorf("shopify credentials are not configured")
	}

	var allOrders []models.GraphQLOrderNode
	apiURL := fmt.Sprintf("https://%s/admin/api/2024-01/graphql.json", c.config.ShopifyStoreURL)

	// Build the search query dynamically.
	// We enforce that the order MUST have been created on or after March 1st, 2026,
	// AND it was updated after the given 'since' threshold.
	searchQuery := fmt.Sprintf("created_at:>='2026-03-01T00:00:00Z' AND updated_at:>'%s'", since.Format(time.RFC3339))

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
					email
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
					customer {
						firstName
						lastName
						displayName
					}
					billingAddress {
						name
						firstName
						lastName
					}
					shippingAddress {
						name
						firstName
						lastName
						phone
						city
						province
						country
					}
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
								originalUnitPriceSet {
									shopMoney {
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
		}
	}`

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

		req.Header.Add("X-Shopify-Access-Token", c.config.ShopifyAccessToken)
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

		var result models.GraphQLOrderResponse
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

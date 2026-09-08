package amazon

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mi-tech/internal/shared/config"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client is the Selling Partner API client.
type Client struct {
	settings     *config.SettingsProvider
	tokenManager *TokenManager
	stsSigner    *STSSigner
	httpClient   *http.Client
	endpoint     string
}

func NewClient(settings *config.SettingsProvider) *Client {
	endpoint := "https://sellingpartnerapi-eu.amazon.com"

	return &Client{
		settings: settings,
		tokenManager: NewTokenManager(
			settings.GetAmazonLWAClientID(),
			settings.GetAmazonLWAClientSecret(),
			settings.GetAmazonLWARefreshToken(),
		),
		stsSigner: NewSTSSigner(
			settings.GetAmazonAWSAccessKey(),
			settings.GetAmazonAWSSecretKey(),
			settings.GetAmazonAWSRegion(),
			settings.GetAmazonAWSRoleARN(),
		),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		endpoint: endpoint,
	}
}

// Do executes a signed request to the SP-API.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	// 1. Get LWA Access Token
	accessToken, err := c.tokenManager.GetAccessToken()
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-amz-access-token", accessToken)

	// 2. Get Temporary AWS Credentials via STS
	accessKey, secretKey, sessionToken, err := c.stsSigner.AssumeRole()
	if err != nil {
		return nil, fmt.Errorf("sts assume role failed: %w", err)
	}
	if sessionToken != "" {
		req.Header.Set("X-Amz-Security-Token", sessionToken)
	}

	// 3. Sign Request with SigV4
	err = SignRequest(req, accessKey, secretKey, c.settings.GetAmazonAWSRegion(), "execute-api")
	if err != nil {
		return nil, fmt.Errorf("sigv4 signing failed: %w", err)
	}

	return c.httpClient.Do(req)
}

// GetOrders fetches orders from Amazon India using either a creation or update timestamp.
// If useLastUpdated is true, it uses LastUpdatedAfter; otherwise, it uses CreatedAfter.
func (c *Client) GetOrders(after, before time.Time, useLastUpdated bool) ([]map[string]interface{}, error) {
	params := url.Values{}
	params.Set("MarketplaceIds", c.settings.GetAmazonMarketplaceID())

	key := "CreatedAfter"
	if useLastUpdated {
		key = "LastUpdatedAfter"
	}
	params.Set(key, after.UTC().Format(time.RFC3339))

	if !before.IsZero() {
		beforeKey := "CreatedBefore"
		if useLastUpdated {
			beforeKey = "LastUpdatedBefore"
		}
		params.Set(beforeKey, before.UTC().Format(time.RFC3339))
	}

	u := fmt.Sprintf("%s/orders/v0/orders?%s", c.endpoint, params.Encode())

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("amazon api error (%d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Payload struct {
			Orders []map[string]interface{} `json:"Orders"`
		} `json:"payload"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result.Payload.Orders, nil
}

// GetOrderItems fetches line items for a specific Amazon order.
func (c *Client) GetOrderItems(orderID string) ([]map[string]interface{}, error) {
	u := fmt.Sprintf("%s/orders/v0/orders/%s/orderItems", c.endpoint, orderID)

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("amazon api error (%d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Payload struct {
			OrderItems []map[string]interface{} `json:"OrderItems"`
		} `json:"payload"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result.Payload.OrderItems, nil
}

// ListingsItemsQuery contains the optional filters supported by
// searchListingsItems. Seller and marketplace IDs always come from the
// configured Amazon connection.
type ListingsItemsQuery struct {
	PageToken    string
	PageSize     int
	IncludedData []string
	IssueLocale  string
}

// ListingsItemsPage is one live page returned by Amazon's Listings Items API.
// NumberOfResults is the total count for the current filter, not just the
// number of items in this page.
type ListingsItemsPage struct {
	NumberOfResults int `json:"numberOfResults"`
	Pagination      struct {
		NextToken string `json:"nextToken"`
	} `json:"pagination"`
	Items []map[string]interface{} `json:"items"`
}

// SearchListingsItems fetches a live, read-only page of the seller's Amazon
// listings using the configured Seller ID and Marketplace ID.
func (c *Client) SearchListingsItems(query ListingsItemsQuery) (ListingsItemsPage, error) {
	var result ListingsItemsPage

	sellerID := strings.TrimSpace(c.settings.GetAmazonSellerID())
	if sellerID == "" {
		return result, fmt.Errorf("AmazonSellerID is missing in config and environment")
	}

	marketplaceID := strings.TrimSpace(c.settings.GetAmazonMarketplaceID())
	if marketplaceID == "" {
		return result, fmt.Errorf("AmazonMarketplaceID is missing in config and environment")
	}

	if query.PageSize < 0 || query.PageSize > 20 {
		return result, fmt.Errorf("page size must be between 1 and 20")
	}

	params := url.Values{}
	params.Set("marketplaceIds", marketplaceID)
	if query.PageToken != "" {
		params.Set("pageToken", query.PageToken)
	}
	if query.PageSize > 0 {
		params.Set("pageSize", fmt.Sprintf("%d", query.PageSize))
	}
	includedData := query.IncludedData
	if len(includedData) == 0 {
		includedData = []string{"summaries"}
	}
	params.Set("includedData", strings.Join(includedData, ","))
	if query.IssueLocale != "" {
		params.Set("issueLocale", query.IssueLocale)
	}

	u := fmt.Sprintf("%s/listings/2021-08-01/items/%s?%s",
		c.endpoint, url.PathEscape(sellerID), params.Encode())

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return result, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("amazon listings api error (%d): %s", resp.StatusCode, string(body))
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return result, err
	}

	return result, nil
}

// CreateReport requests a new report from Amazon.
func (c *Client) CreateReport(reportType string, startTime, endTime *time.Time) (string, error) {
	u := fmt.Sprintf("%s/reports/2021-06-30/reports", c.endpoint)

	payload := map[string]interface{}{
		"reportType":     reportType,
		"marketplaceIds": []string{c.settings.GetAmazonMarketplaceID()},
	}
	if startTime != nil {
		payload["dataStartTime"] = startTime.UTC().Format(time.RFC3339)
	}
	if endTime != nil {
		payload["dataEndTime"] = endTime.UTC().Format(time.RFC3339)
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", u, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("amazon create report error (%d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		ReportID string `json:"reportId"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", err
	}

	return result.ReportID, nil
}

// GetReport fetches the status and details of a report.
func (c *Client) GetReport(reportID string) (map[string]interface{}, error) {
	u := fmt.Sprintf("%s/reports/2021-06-30/reports/%s", c.endpoint, reportID)

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("amazon get report error (%d): %s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// GetReportDocument fetches the download URL for a report document.
func (c *Client) GetReportDocument(documentID string) (string, error) {
	u := fmt.Sprintf("%s/reports/2021-06-30/documents/%s", c.endpoint, documentID)

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return "", err
	}

	resp, err := c.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("amazon get document error (%d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", err
	}

	return result.URL, nil
}

// UpdateInventory pushes stock updates to Amazon using the Listings Items API.
func (c *Client) UpdateInventory(sku string, quantity int) error {
	sellerID := c.settings.GetAmazonSellerID()

	if sellerID == "" {
		return fmt.Errorf("AmazonSellerID is missing in config and environment")
	}

	u := fmt.Sprintf("%s/listings/2021-08-01/items/%s/%s?marketplaceIds=%s",
		c.endpoint, sellerID, sku, c.settings.GetAmazonMarketplaceID())

	patch := map[string]interface{}{
		"productType": "PRODUCT",
		"patches": []map[string]interface{}{
			{
				"op":   "replace",
				"path": "/attributes/fulfillment_availability",
				"value": []map[string]interface{}{
					{
						"fulfillment_channel_code": "DEFAULT",
						"quantity":                 quantity,
					},
				},
			},
		},
	}

	payload, _ := json.Marshal(patch)
	req, err := http.NewRequest("PATCH", u, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json-patch+json")

	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("amazon update inventory error (%d): %s", resp.StatusCode, string(body))
	}

	return nil
}

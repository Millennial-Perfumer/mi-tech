package amazon

import (
	"io"
	"mi-tech/internal/shared/config"
	"net/http"
	"strings"
	"testing"
	"time"
)

type mockConfigRepo struct{}

func (m *mockConfigRepo) Get(key string) (string, error) {
	if key == "amazon.seller_id" {
		return "A1B2C3D4E5", nil
	}
	return "mock_value", nil
}

type mockClientTransport struct{}

func (m *mockClientTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if strings.Contains(req.URL.Host, "api.amazon.com") && strings.Contains(req.URL.Path, "token") {
		respBody := `{"access_token": "mocked_token", "expires_in": 3600}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(respBody)),
			Header:     make(http.Header),
		}, nil
	}

	if strings.Contains(req.URL.Host, "sts") {
		respBody := `
		<AssumeRoleResponse>
		  <AssumeRoleResult>
			<Credentials>
			  <AccessKeyId>ASIA_TEST</AccessKeyId>
			  <SecretAccessKey>SECRET_TEST</SecretAccessKey>
			  <SessionToken>SESSION_TEST</SessionToken>
			</Credentials>
		  </AssumeRoleResult>
		</AssumeRoleResponse>`
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(strings.TrimSpace(respBody))),
			Header:     make(http.Header),
		}, nil
	}

	if strings.Contains(req.URL.Host, "sellingpartnerapi-eu.amazon.com") && strings.Contains(req.URL.Path, "orders/v0/orders") && !strings.Contains(req.URL.Path, "orderItems") {
		respBody := `{
			"payload": {
				"Orders": [
					{"AmazonOrderId": "123", "OrderStatus": "Unshipped"}
				]
			}
		}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(respBody)),
			Header:     make(http.Header),
		}, nil
	}

	if strings.Contains(req.URL.Host, "sellingpartnerapi-eu.amazon.com") && strings.Contains(req.URL.Path, "orderItems") {
		respBody := `{
			"payload": {
				"OrderItems": [
					{"OrderItemId": "item123", "Title": "Test Product"}
				]
			}
		}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(respBody)),
			Header:     make(http.Header),
		}, nil
	}

	if strings.Contains(req.URL.Host, "sellingpartnerapi-eu.amazon.com") && strings.Contains(req.URL.Path, "reports/2021-06-30/reports") && req.Method == "POST" {
		respBody := `{"reportId": "report123"}`
		return &http.Response{
			StatusCode: http.StatusAccepted,
			Body:       io.NopCloser(strings.NewReader(respBody)),
			Header:     make(http.Header),
		}, nil
	}

	if strings.Contains(req.URL.Host, "sellingpartnerapi-eu.amazon.com") && strings.Contains(req.URL.Path, "reports/2021-06-30/reports/") && req.Method == "GET" {
		respBody := `{"reportId": "report123", "processingStatus": "DONE"}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(respBody)),
			Header:     make(http.Header),
		}, nil
	}

	if strings.Contains(req.URL.Host, "sellingpartnerapi-eu.amazon.com") && strings.Contains(req.URL.Path, "reports/2021-06-30/documents/") && req.Method == "GET" {
		respBody := `{"url": "https://example.com/document"}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(respBody)),
			Header:     make(http.Header),
		}, nil
	}

	if strings.Contains(req.URL.Host, "sellingpartnerapi-eu.amazon.com") && strings.Contains(req.URL.Path, "items") && req.Method == "PATCH" {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{}`)),
			Header:     make(http.Header),
		}, nil
	}

	return &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(strings.NewReader("Not Found")),
	}, nil
}

func TestClient_GetOrders(t *testing.T) {
	settings := config.NewSettingsProvider(&mockConfigRepo{})
	client := NewClient(settings)

	// Avoid global state modification by injecting directly into the relevant http clients
	mockTransport := &mockClientTransport{}
	client.httpClient.Transport = mockTransport
	client.tokenManager.httpClient.Transport = mockTransport
	client.stsSigner.httpClient.Transport = mockTransport

	orders, err := client.GetOrders(time.Now().Add(-24*time.Hour), time.Now(), false)
	if err != nil {
		t.Fatalf("GetOrders failed: %v", err)
	}

	if len(orders) != 1 || orders[0]["AmazonOrderId"] != "123" {
		t.Errorf("Unexpected orders response: %v", orders)
	}
}

func TestClient_GetOrderItems(t *testing.T) {
	settings := config.NewSettingsProvider(&mockConfigRepo{})
	client := NewClient(settings)

	mockTransport := &mockClientTransport{}
	client.httpClient.Transport = mockTransport
	client.tokenManager.httpClient.Transport = mockTransport
	client.stsSigner.httpClient.Transport = mockTransport

	items, err := client.GetOrderItems("123")
	if err != nil {
		t.Fatalf("GetOrderItems failed: %v", err)
	}

	if len(items) != 1 || items[0]["OrderItemId"] != "item123" {
		t.Errorf("Unexpected order items response: %v", items)
	}
}

func TestClient_CreateReport(t *testing.T) {
	settings := config.NewSettingsProvider(&mockConfigRepo{})
	client := NewClient(settings)

	mockTransport := &mockClientTransport{}
	client.httpClient.Transport = mockTransport
	client.tokenManager.httpClient.Transport = mockTransport
	client.stsSigner.httpClient.Transport = mockTransport

	start := time.Now().Add(-24 * time.Hour)
	end := time.Now()
	reportID, err := client.CreateReport("GET_V2_SETTLEMENT_REPORT_DATA_FLAT_FILE_V2", &start, &end)
	if err != nil {
		t.Fatalf("CreateReport failed: %v", err)
	}

	if reportID != "report123" {
		t.Errorf("Expected report123, got %s", reportID)
	}
}

func TestClient_GetReport(t *testing.T) {
	settings := config.NewSettingsProvider(&mockConfigRepo{})
	client := NewClient(settings)

	mockTransport := &mockClientTransport{}
	client.httpClient.Transport = mockTransport
	client.tokenManager.httpClient.Transport = mockTransport
	client.stsSigner.httpClient.Transport = mockTransport

	report, err := client.GetReport("report123")
	if err != nil {
		t.Fatalf("GetReport failed: %v", err)
	}

	if report["processingStatus"] != "DONE" {
		t.Errorf("Unexpected report status: %v", report)
	}
}

func TestClient_GetReportDocument(t *testing.T) {
	settings := config.NewSettingsProvider(&mockConfigRepo{})
	client := NewClient(settings)

	mockTransport := &mockClientTransport{}
	client.httpClient.Transport = mockTransport
	client.tokenManager.httpClient.Transport = mockTransport
	client.stsSigner.httpClient.Transport = mockTransport

	urlStr, err := client.GetReportDocument("doc123")
	if err != nil {
		t.Fatalf("GetReportDocument failed: %v", err)
	}

	if urlStr != "https://example.com/document" {
		t.Errorf("Expected https://example.com/document, got %s", urlStr)
	}
}

func TestClient_UpdateInventory(t *testing.T) {
	settings := config.NewSettingsProvider(&mockConfigRepo{})
	client := NewClient(settings)

	mockTransport := &mockClientTransport{}
	client.httpClient.Transport = mockTransport
	client.tokenManager.httpClient.Transport = mockTransport
	client.stsSigner.httpClient.Transport = mockTransport

	err := client.UpdateInventory("SKU123", 10)
	if err != nil {
		t.Fatalf("UpdateInventory failed: %v", err)
	}
}

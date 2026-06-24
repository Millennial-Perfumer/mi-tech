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

	if strings.Contains(req.URL.Host, "sellingpartnerapi-eu.amazon.com") && strings.Contains(req.URL.Path, "orders/v0/orders") {
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

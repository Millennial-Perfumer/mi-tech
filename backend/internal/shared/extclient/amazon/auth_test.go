package amazon

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

type mockAuthTransport struct {
	tokenResponse string
	statusCode    int
	reqCount      int
}

func (m *mockAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	m.reqCount++
	if m.statusCode == 0 {
		m.statusCode = http.StatusOK
	}

	if req.URL.Host == "api.amazon.com" && req.URL.Path == "/auth/o2/token" {
		return &http.Response{
			StatusCode: m.statusCode,
			Body:       io.NopCloser(strings.NewReader(m.tokenResponse)),
			Header:     make(http.Header),
		}, nil
	}
	return &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(strings.NewReader("Not Found")),
	}, nil
}

func TestTokenManager_GetAccessToken(t *testing.T) {
	mockTransport := &mockAuthTransport{
		tokenResponse: `{"access_token": "mocked_token", "expires_in": 3600}`,
	}
	tm := NewTokenManager("client_id", "client_secret", "refresh_token")
	tm.httpClient.Transport = mockTransport

	token, err := tm.GetAccessToken()
	if err != nil {
		t.Fatalf("GetAccessToken failed: %v", err)
	}
	if token != "mocked_token" {
		t.Errorf("Expected mocked_token, got %s", token)
	}

	// Test caching
	token2, err := tm.GetAccessToken()
	if err != nil {
		t.Fatalf("GetAccessToken failed: %v", err)
	}
	if token2 != "mocked_token" {
		t.Errorf("Expected mocked_token, got %s", token2)
	}
	if mockTransport.reqCount != 1 {
		t.Errorf("Expected only 1 HTTP request due to caching, got %d", mockTransport.reqCount)
	}
}

func TestTokenManager_RefreshError(t *testing.T) {
	mockTransport := &mockAuthTransport{
		statusCode:    http.StatusBadRequest,
		tokenResponse: `{"error": "invalid_grant"}`,
	}
	tm := NewTokenManager("client_id", "client_secret", "refresh_token")
	tm.httpClient.Transport = mockTransport

	_, err := tm.GetAccessToken()
	if err == nil {
		t.Error("Expected error for bad request")
	}
}

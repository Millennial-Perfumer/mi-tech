package amazon

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/url"

	"testing"
	"time"
)

func TestSignV4(t *testing.T) {
	signTime := time.Date(2015, 8, 30, 12, 36, 0, 0, time.UTC)

	tests := []struct {
		name          string
		method        string
		urlStr        string
		body          []byte
		exactExpected string
	}{
		{
			name:          "GET request no body",
			method:        "GET",
			urlStr:        "https://example.amazonaws.com/",
			body:          nil,
			exactExpected: "AWS4-HMAC-SHA256 Credential=AKIDEXAMPLE/20150830/us-east-1/service/aws4_request, SignedHeaders=host;x-amz-date, Signature=5fa00fa31553b73ebf1942676e86291e8372ff2a2260956d9b8aae1d763fbf31",
		},
		{
			name:          "POST request with body",
			method:        "POST",
			urlStr:        "https://example.amazonaws.com/",
			body:          []byte(`{"test":"body"}`),
			exactExpected: "AWS4-HMAC-SHA256 Credential=AKIDEXAMPLE/20150830/us-east-1/service/aws4_request, SignedHeaders=host;x-amz-date, Signature=72fe8531a9f9918483947f9bd1faca8bef05861e1e786dbc265ef25b6f2e3833",
		},
		{
			name:          "GET request with query params",
			method:        "GET",
			urlStr:        "https://example.amazonaws.com/?Param1=Value1&Param2=Value2",
			body:          nil,
			exactExpected: "AWS4-HMAC-SHA256 Credential=AKIDEXAMPLE/20150830/us-east-1/service/aws4_request, SignedHeaders=host;x-amz-date, Signature=9db2b1c2412a767b643ad7e026212de27519a82f49c72c339b93a865b9e4f3a5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsedURL, err := url.Parse(tt.urlStr)
			if err != nil {
				t.Fatalf("Failed to parse URL: %v", err)
			}

			req := &http.Request{
				Method: tt.method,
				URL:    parsedURL,
				Header: make(http.Header),
			}

			err = SignV4(req, tt.body, "AKIDEXAMPLE", "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY", "us-east-1", "service", signTime)
			if err != nil {
				t.Fatalf("SignV4 failed: %v", err)
			}

			auth := req.Header.Get("Authorization")
			if auth == "" {
				t.Error("Expected Authorization header")
			}

			if auth != tt.exactExpected {
				t.Errorf("\nExpected: %s\nGot:      %s", tt.exactExpected, auth)
			}

			amzDate := req.Header.Get("X-Amz-Date")
			expectedAmzDate := "20150830T123600Z"
			if amzDate != expectedAmzDate {
				t.Errorf("Expected X-Amz-Date to be %s, got %s", expectedAmzDate, amzDate)
			}
		})
	}
}

func TestSignRequest(t *testing.T) {
	bodyContent := "test body"
	parsedURL, _ := url.Parse("https://example.amazonaws.com/")
	req := &http.Request{
		Method: "POST",
		URL:    parsedURL,
		Header: make(http.Header),
		Body:   io.NopCloser(bytes.NewBufferString(bodyContent)),
	}

	err := SignRequest(req, "AKIDEXAMPLE", "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY", "us-east-1", "service")
	if err != nil {
		t.Fatalf("SignRequest failed: %v", err)
	}

	// Verify body is restored
	restoredBody, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("Failed to read restored body: %v", err)
	}
	if string(restoredBody) != bodyContent {
		t.Errorf("Expected body %q, got %q", bodyContent, restoredBody)
	}

	auth := req.Header.Get("Authorization")
	if auth == "" {
		t.Error("Expected Authorization header")
	}
}

type errorReader struct{}

func (e errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("simulated read error")
}

func TestSignV4_CoverageEdges(t *testing.T) {
	signTime := time.Date(2015, 8, 30, 12, 36, 0, 0, time.UTC)

	t.Run("Empty path becomes slash", func(t *testing.T) {
		req := &http.Request{
			Method: "GET",
			URL:    &url.URL{Host: "example.amazonaws.com", Path: ""},
			Header: make(http.Header),
		}
		err := SignV4(req, nil, "AKID", "SECRET", "us-east-1", "service", signTime)
		if err != nil {
			t.Fatalf("SignV4 failed: %v", err)
		}
	})

	t.Run("req.Host fallback", func(t *testing.T) {
		req := &http.Request{
			Method: "GET",
			URL:    &url.URL{Path: "/"},
			Host:   "example.amazonaws.com",
			Header: make(http.Header),
		}
		err := SignV4(req, nil, "AKID", "SECRET", "us-east-1", "service", signTime)
		if err != nil {
			t.Fatalf("SignV4 failed: %v", err)
		}
		if req.Header.Get("Host") != "example.amazonaws.com" {
			t.Errorf("Expected Host header to be set from req.Host")
		}
	})

	t.Run("Skip authorization header", func(t *testing.T) {
		req := &http.Request{
			Method: "GET",
			URL:    &url.URL{Host: "example.amazonaws.com", Path: "/"},
			Header: make(http.Header),
		}
		req.Header.Set("Authorization", "Some-Old-Auth")
		err := SignV4(req, nil, "AKID", "SECRET", "us-east-1", "service", signTime)
		if err != nil {
			t.Fatalf("SignV4 failed: %v", err)
		}
		// Authorization should be overwritten by SignV4
		if req.Header.Get("Authorization") == "Some-Old-Auth" {
			t.Errorf("Expected Authorization header to be overwritten")
		}
	})

	t.Run("uriEscapePath escaping", func(t *testing.T) {
		req := &http.Request{
			Method: "GET",
			URL:    &url.URL{Host: "example.amazonaws.com", Path: "/a b c/!"},
			Header: make(http.Header),
		}
		err := SignV4(req, nil, "AKID", "SECRET", "us-east-1", "service", signTime)
		if err != nil {
			t.Fatalf("SignV4 failed: %v", err)
		}
	})
}

func TestSignRequest_ReadError(t *testing.T) {
	parsedURL, _ := url.Parse("https://example.amazonaws.com/")
	req := &http.Request{
		Method: "POST",
		URL:    parsedURL,
		Header: make(http.Header),
		Body:   io.NopCloser(errorReader{}),
	}
	err := SignRequest(req, "AKID", "SECRET", "us-east-1", "service")
	if err == nil {
		t.Error("Expected error from reading body")
	}
}

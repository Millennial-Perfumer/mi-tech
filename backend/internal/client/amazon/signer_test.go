package amazon

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestSignV4(t *testing.T) {
	signTime := time.Date(2015, 8, 30, 12, 36, 0, 0, time.UTC)

	tests := []struct {
		name           string
		method         string
		urlStr         string
		body           []byte
		expectedPrefix string
	}{
		{
			name:           "GET request no body",
			method:         "GET",
			urlStr:         "https://example.amazonaws.com/",
			body:           nil,
			expectedPrefix: "AWS4-HMAC-SHA256 Credential=AKIDEXAMPLE/20150830/us-east-1/service/aws4_request, SignedHeaders=host;x-amz-date, Signature=",
		},
		{
			name:           "POST request with body",
			method:         "POST",
			urlStr:         "https://example.amazonaws.com/",
			body:           []byte(`{"test":"body"}`),
			expectedPrefix: "AWS4-HMAC-SHA256 Credential=AKIDEXAMPLE/20150830/us-east-1/service/aws4_request, SignedHeaders=host;x-amz-date, Signature=",
		},
		{
			name:           "GET request with query params",
			method:         "GET",
			urlStr:         "https://example.amazonaws.com/?Param1=Value1&Param2=Value2",
			body:           nil,
			expectedPrefix: "AWS4-HMAC-SHA256 Credential=AKIDEXAMPLE/20150830/us-east-1/service/aws4_request, SignedHeaders=host;x-amz-date, Signature=",
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

			if !strings.HasPrefix(auth, tt.expectedPrefix) {
				t.Errorf("Unexpected Authorization prefix: got %s, want prefix %s", auth, tt.expectedPrefix)
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

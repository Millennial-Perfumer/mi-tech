package amazon

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("simulated read error")
}

func (e *errorReader) Close() error {
	return nil
}

func TestSignRequest(t *testing.T) {
	req, err := http.NewRequest("GET", "https://example.com/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	err = SignRequest(req, "ACCESS_KEY", "SECRET_KEY", "us-east-1", "s3")
	if err != nil {
		t.Fatalf("SignRequest failed: %v", err)
	}

	auth := req.Header.Get("Authorization")
	if auth == "" {
		t.Error("Expected Authorization header to be set")
	}

	reqWithBody, err := http.NewRequest("POST", "https://example.com/test", bytes.NewReader([]byte("test body")))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	err = SignRequest(reqWithBody, "ACCESS_KEY", "SECRET_KEY", "us-east-1", "s3")
	if err != nil {
		t.Fatalf("SignRequest failed: %v", err)
	}

	authWithBody := reqWithBody.Header.Get("Authorization")
	if authWithBody == "" {
		t.Error("Expected Authorization header to be set")
	}

	// Verify that body can still be read after signing
	bodyBytes, err := io.ReadAll(reqWithBody.Body)
	if err != nil {
		t.Fatalf("Failed to read body after signing: %v", err)
	}
	if string(bodyBytes) != "test body" {
		t.Errorf("Expected body 'test body', got '%s'", string(bodyBytes))
	}

	reqError := &http.Request{
		Method: "POST",
		URL:    &url.URL{Scheme: "https", Host: "example.com", Path: "/test"},
		Body:   &errorReader{},
	}
	err = SignRequest(reqError, "ACCESS_KEY", "SECRET_KEY", "us-east-1", "s3")
	if err == nil {
		t.Error("Expected SignRequest to fail with errorReader")
	}
}

func TestSignV4Edges(t *testing.T) {
	t.Run("Empty path becomes slash", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "https://example.com", nil)
		req.URL.Path = "" // Force empty path
		err := SignV4(req, nil, "AKID", "SECRET", "us-east-1", "s3", time.Now())
		if err != nil {
			t.Fatalf("SignV4 failed: %v", err)
		}
		if req.Header.Get("Authorization") == "" {
			t.Error("Expected Authorization header")
		}
	})

	t.Run("req.Host fallback", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "https://example.com", nil)
		req.Host = "custom-host.com"
		err := SignV4(req, nil, "AKID", "SECRET", "us-east-1", "s3", time.Now())
		if err != nil {
			t.Fatalf("SignV4 failed: %v", err)
		}
		if req.Header.Get("Host") != "custom-host.com" {
			t.Errorf("Expected Host to be set to custom-host.com, got %s", req.Header.Get("Host"))
		}
	})

	t.Run("Skip Authorization header", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "https://example.com", nil)
		req.Header.Set("Authorization", "ExistingAuth")
		err := SignV4(req, nil, "AKID", "SECRET", "us-east-1", "s3", time.Now())
		if err != nil {
			t.Fatalf("SignV4 failed: %v", err)
		}

		auth := req.Header.Get("Authorization")
		if !strings.Contains(auth, "AWS4-HMAC-SHA256") {
			t.Errorf("Expected Authorization header to be replaced with AWS4 signature, got %s", auth)
		}
	})

	t.Run("req.URL.Host fallback", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "https://example-url-host.com", nil)
		// Request automatically parses "https://example-url-host.com" into req.URL.Host
		req.Host = ""
		err := SignV4(req, nil, "AKID", "SECRET", "us-east-1", "s3", time.Now())
		if err != nil {
			t.Fatalf("SignV4 failed: %v", err)
		}
		if req.Header.Get("Host") != "example-url-host.com" {
			t.Errorf("Expected Host to be set to example-url-host.com, got %s", req.Header.Get("Host"))
		}
	})

	t.Run("Path with characters to escape", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "https://example.com/test path/!chars", nil)
		err := SignV4(req, nil, "AKID", "SECRET", "us-east-1", "s3", time.Now())
		if err != nil {
			t.Fatalf("SignV4 failed: %v", err)
		}
		auth := req.Header.Get("Authorization")
		if auth == "" {
			t.Error("Expected Authorization header")
		}
	})
}

func TestUriEscapePath(t *testing.T) {
	s := "test path/+with space"
	escaped := uriEscapePath(s)
	if escaped != "test%20path/%2Bwith%20space" {
		t.Errorf("Unexpected uriEscapePath output: %s", escaped)
	}
}

func TestSignV4HostHeader(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://api.amazon.com/test", nil)
	req.Host = "" // Force empty req.Host so req.URL.Host is used
	req.Header.Del("Host")

	err := SignV4(req, nil, "AKID", "SECRET", "us-east-1", "s3", time.Now())
	if err != nil {
		t.Fatalf("SignV4 failed: %v", err)
	}
	if req.Header.Get("Host") != "api.amazon.com" {
		t.Errorf("Expected Host to be set to api.amazon.com, got %s", req.Header.Get("Host"))
	}
}

func TestSignV4_Simple(t *testing.T) {
	req, err := http.NewRequest("GET", "https://example.amazonaws.com/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	signTime := time.Date(2015, 8, 30, 12, 36, 0, 0, time.UTC)
	err = SignV4(req, nil, "AKIDEXAMPLE", "SECRETEXAMPLE", "us-east-1", "service", signTime)
	if err != nil {
		t.Fatalf("SignV4 failed: %v", err)
	}

	auth := req.Header.Get("Authorization")
	if auth == "" {
		t.Error("Expected Authorization header to be set")
	}

	if !strings.HasPrefix(auth, "AWS4-HMAC-SHA256 Credential=AKIDEXAMPLE/20150830/us-east-1/service/aws4_request") {
		t.Errorf("Unexpected Authorization header format: %s", auth)
	}
}

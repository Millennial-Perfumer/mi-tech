package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestCORSMiddleware_Wildcard(t *testing.T) {
	os.Setenv("ALLOWED_ORIGINS", "*")
	defer os.Unsetenv("ALLOWED_ORIGINS")

	handler := CORSMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/api/health", nil)
	req.Header.Set("Origin", "http://evil.com")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	originHeader := res.Header().Get("Access-Control-Allow-Origin")
	if originHeader != "*" {
		t.Errorf("Expected Origin header '*', got %q", originHeader)
	}

	credsHeader := res.Header().Get("Access-Control-Allow-Credentials")
	if credsHeader != "" {
		t.Errorf("Expected Access-Control-Allow-Credentials to be empty for wildcard, got %q", credsHeader)
	}
}

func TestCORSMiddleware_Specific(t *testing.T) {
	os.Setenv("ALLOWED_ORIGINS", "http://trusted.com")
	defer os.Unsetenv("ALLOWED_ORIGINS")

	handler := CORSMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/api/health", nil)
	req.Header.Set("Origin", "http://trusted.com")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	originHeader := res.Header().Get("Access-Control-Allow-Origin")
	if originHeader != "http://trusted.com" {
		t.Errorf("Expected Origin header 'http://trusted.com', got %q", originHeader)
	}

	credsHeader := res.Header().Get("Access-Control-Allow-Credentials")
	if credsHeader != "true" {
		t.Errorf("Expected Access-Control-Allow-Credentials to be 'true' for specific origin, got %q", credsHeader)
	}
}

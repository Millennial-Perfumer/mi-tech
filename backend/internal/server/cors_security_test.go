package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSMiddleware_WildcardSecurity(t *testing.T) {
	// Setup allowed origins with wildcard
	t.Setenv("ALLOWED_ORIGINS", "*")

	handler := CORSMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/api/health", nil)
	req.Header.Set("Origin", "http://evil.com")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	originHeader := res.Header().Get("Access-Control-Allow-Origin")
	credsHeader := res.Header().Get("Access-Control-Allow-Credentials")

	if originHeader == "http://evil.com" && credsHeader == "true" {
		t.Errorf("SECURITY VULNERABILITY: Wildcard ALLOWED_ORIGINS allows arbitrary origin %q with credentials", originHeader)
	}
}

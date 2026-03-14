package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLogin_InvalidJSON(t *testing.T) {
	// authService is not used when JSON decoding fails
	h := &AuthHandler{authService: nil}

	// Malformed JSON (trailing comma)
	reqBody := `{"username": "admin",}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(reqBody))
	rr := httptest.NewRecorder()

	h.Login(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status code %d, got %d", http.StatusBadRequest, rr.Code)
	}

	expected := "invalid request\n"
	if rr.Body.String() != expected {
		t.Errorf("expected body %q, got %q", expected, rr.Body.String())
	}
}

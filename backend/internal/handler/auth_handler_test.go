package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"mi-tech/internal/service"
)

func TestAuthHandler_Login_InvalidRequest(t *testing.T) {
	// We can test the handler with an invalid request payload without needing a real DB.
	// We just pass nil for the DB since it won't be reached if the request fails decoding.
	authService := service.NewAuthService(nil, "secret")
	handler := NewAuthHandler(authService)

	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBufferString("{invalid_json}"))
	w := httptest.NewRecorder()

	handler.Login(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status code 400, got %d", resp.StatusCode)
	}
}

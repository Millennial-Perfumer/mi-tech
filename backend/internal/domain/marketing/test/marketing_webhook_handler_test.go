package test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"mi-tech/internal/domain/marketing/handler"
	"mi-tech/internal/domain/marketing/service"
	"mi-tech/internal/shared/config"

	"github.com/stretchr/testify/assert"
)

func TestMarketingWebhookHandler_handleNotification(t *testing.T) {
	// Set up dependencies
	metaClient := &service.MetaMarketingClient{}
	settings := &config.SettingsProvider{}

	marketingHandler := handler.NewMarketingWebhookHandler(metaClient, settings)

	// Helper to generate a valid signature
	generateSignature := func(secret, body string) string {
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte(body))
		return "sha256=" + hex.EncodeToString(mac.Sum(nil))
	}

	t.Run("Missing Signature", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/webhooks/meta", bytes.NewBufferString(`{"test": true}`))
		rr := httptest.NewRecorder()

		marketingHandler.MetaWebhook(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "Unauthorized")
	})

	t.Run("Missing App Secret", func(t *testing.T) {
		// App secret is not set via environment variable
		t.Setenv("META_APP_SECRET", "") // Ensure it's empty

		bodyStr := `{"test": true}`
		sig := generateSignature("dummy_secret", bodyStr)

		req, _ := http.NewRequest(http.MethodPost, "/webhooks/meta", bytes.NewBufferString(bodyStr))
		req.Header.Set("X-Hub-Signature-256", sig)
		rr := httptest.NewRecorder()

		marketingHandler.MetaWebhook(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "Internal Server Error")
	})

	t.Run("Invalid Signature", func(t *testing.T) {
		t.Setenv("META_APP_SECRET", "my_secret_key")

		req, _ := http.NewRequest(http.MethodPost, "/webhooks/meta", bytes.NewBufferString(`{"test": true}`))
		req.Header.Set("X-Hub-Signature-256", "sha256=invalid_signature_hex_here")
		rr := httptest.NewRecorder()

		marketingHandler.MetaWebhook(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "Unauthorized")
	})

	t.Run("Invalid Signature Format", func(t *testing.T) {
		t.Setenv("META_APP_SECRET", "my_secret_key")

		req, _ := http.NewRequest(http.MethodPost, "/webhooks/meta", bytes.NewBufferString(`{"test": true}`))
		req.Header.Set("X-Hub-Signature-256", "invalid_signature_hex_here")
		rr := httptest.NewRecorder()

		marketingHandler.MetaWebhook(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "Unauthorized")
	})

	t.Run("Valid Signature", func(t *testing.T) {
		secret := "my_secret_key"
		t.Setenv("META_APP_SECRET", secret)

		bodyStr := `{"test": true}`
		sig := generateSignature(secret, bodyStr)

		req, _ := http.NewRequest(http.MethodPost, "/webhooks/meta", bytes.NewBufferString(bodyStr))
		req.Header.Set("X-Hub-Signature-256", sig)
		rr := httptest.NewRecorder()

		marketingHandler.MetaWebhook(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "Acknowledged")
	})
}

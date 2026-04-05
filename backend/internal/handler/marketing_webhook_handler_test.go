package handler

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"mi-tech/internal/config"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockConfigsRepo struct {
	values map[string]string
}

func (m *mockConfigsRepo) Get(key string) (string, error) {
	return m.values[key], nil
}

func TestMarketingWebhookHandler_MetaWebhook(t *testing.T) {
	mockRepo := &mockConfigsRepo{
		values: map[string]string{
			"meta_marketing_webhook_verify_token": "test-token",
		},
	}
	settings := config.NewSettingsProvider(mockRepo)
	handler := NewMarketingWebhookHandler(nil, settings)

	t.Run("GET Verification - Success", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/marketing/meta/webhook?hub.mode=subscribe&hub.verify_token=test-token&hub.challenge=12345", nil)
		w := httptest.NewRecorder()

		handler.MetaWebhook(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "12345", w.Body.String())
	})

	t.Run("POST Notification - No Signature (Unauthorized)", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/marketing/meta/webhook", bytes.NewBufferString(`{"test":"payload"}`))
		w := httptest.NewRecorder()

		handler.MetaWebhook(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("POST Notification - Valid Signature", func(t *testing.T) {
		secret := "test-app-secret"
		mockRepo.values["meta_app_secret"] = secret
		body := `{"test":"payload"}`

		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte(body))
		signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

		req := httptest.NewRequest("POST", "/api/marketing/meta/webhook", bytes.NewBufferString(body))
		req.Header.Set("X-Hub-Signature-256", signature)
		w := httptest.NewRecorder()

		handler.MetaWebhook(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "Acknowledged", w.Body.String())
	})

	t.Run("POST Notification - Invalid Signature", func(t *testing.T) {
		secret := "test-app-secret"
		mockRepo.values["meta_app_secret"] = secret
		body := `{"test":"payload"}`

		req := httptest.NewRequest("POST", "/api/marketing/meta/webhook", bytes.NewBufferString(body))
		req.Header.Set("X-Hub-Signature-256", "sha256=invalid")
		w := httptest.NewRecorder()

		handler.MetaWebhook(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

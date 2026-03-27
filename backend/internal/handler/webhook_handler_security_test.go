package handler

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http/httptest"
	"testing"

	"mi-tech/internal/config"
	"mi-tech/internal/repository"
	"mi-tech/internal/testutil"

	"github.com/stretchr/testify/assert"
)

func TestWebhookHandler_VerifyWebhook_Security(t *testing.T) {
	db, err := testutil.SetupTestDB()
	if err != nil {
		t.Skip("DB not available")
	}
	defer testutil.CleanupTestDB(db)

	configsRepo := repository.NewConfigsRepository(db)
	settingsProvider := config.NewSettingsProvider(configsRepo)
	h := &WebhookHandler{
		settings: settingsProvider,
	}

	body := []byte(`{"id": 123456}`)

	t.Run("Should fail when secret is missing", func(t *testing.T) {
		// Ensure secret is empty in DB
		db.Exec("DELETE FROM app_configs WHERE key = 'shopify_webhook_secret'")

		req := httptest.NewRequest("POST", "/api/webhooks/shopify", bytes.NewBuffer(body))
		req.Header.Set("X-Shopify-Hmac-Sha256", "some-hmac")

		assert.False(t, h.verifyWebhook(req, body), "Should return false when secret is missing")
	})

	t.Run("Should fail when HMAC is invalid", func(t *testing.T) {
		secret := "test_secret"
		configsRepo.Set("shopify_webhook_secret", secret)

		req := httptest.NewRequest("POST", "/api/webhooks/shopify", bytes.NewBuffer(body))
		req.Header.Set("X-Shopify-Hmac-Sha256", "invalid-hmac")

		assert.False(t, h.verifyWebhook(req, body), "Should return false for invalid HMAC")
	})

	t.Run("Should succeed when HMAC is valid", func(t *testing.T) {
		secret := "test_secret"
		configsRepo.Set("shopify_webhook_secret", secret)

		hash := hmac.New(sha256.New, []byte(secret))
		hash.Write(body)
		expectedHmac := base64.StdEncoding.EncodeToString(hash.Sum(nil))

		req := httptest.NewRequest("POST", "/api/webhooks/shopify", bytes.NewBuffer(body))
		req.Header.Set("X-Shopify-Hmac-Sha256", expectedHmac)

		assert.True(t, h.verifyWebhook(req, body), "Should return true for valid HMAC")
	})
}

package handler

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"mi-tech/internal/automation/whatsapp"
	"mi-tech/internal/config"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockConfigsRepo struct {
	configs map[string]string
}

func (m *mockConfigsRepo) Get(key string) (string, error) {
	return m.configs[key], nil
}

func setupTestHandler(shopifySecret, whatsappVerifyToken string) (*WebhookHandler, *whatsapp.AutomationHandler) {
	m := &mockConfigsRepo{
		configs: make(map[string]string),
	}
	if shopifySecret != "" {
		m.configs["shopify_webhook_secret"] = shopifySecret
	}
	if whatsappVerifyToken != "" {
		m.configs["whatsapp_webhook_verify_token"] = whatsappVerifyToken
	}

	settingsProvider := config.NewSettingsProvider(m)

	// Pass nil for webhookService to test initial verification without triggering async background logic
	h := NewWebhookHandler(nil, nil, settingsProvider)
	ah := whatsapp.NewAutomationHandler(nil, nil, nil, nil, nil, settingsProvider)

	return h, ah
}

func TestShopifyWebhook_FailClosed(t *testing.T) {
	// No secret configured
	h, _ := setupTestHandler("", "")

	req := httptest.NewRequest("POST", "/api/webhooks/shopify", bytes.NewBuffer([]byte(`{}`)))
	req.Header.Set("X-Shopify-Hmac-Sha256", "some-hmac")
	w := httptest.NewRecorder()

	h.ShopifyWebhookHandler(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestShopifyWebhook_VerifyHMAC_Success(t *testing.T) {
	secret := "test-secret"
	h, _ := setupTestHandler(secret, "")

	body := []byte(`{"id": 123}`)
	hash := hmac.New(sha256.New, []byte(secret))
	hash.Write(body)
	expectedHmac := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	req := httptest.NewRequest("POST", "/api/webhooks/shopify", bytes.NewBuffer(body))
	req.Header.Set("X-Shopify-Hmac-Sha256", expectedHmac)
	req.Header.Set("X-Shopify-Topic", "orders/create")
	w := httptest.NewRecorder()

	h.ShopifyWebhookHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// Give a tiny bit of time for the goroutine to run and exit (since webhookService is nil)
	time.Sleep(10 * time.Millisecond)
}

func TestShopifyWebhook_MaxBodySize(t *testing.T) {
	h, _ := setupTestHandler("secret", "")

	// 2MB body
	largeBody := make([]byte, 2*1024*1024)
	req := httptest.NewRequest("POST", "/api/webhooks/shopify", bytes.NewBuffer(largeBody))
	w := httptest.NewRecorder()

	h.ShopifyWebhookHandler(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to read body")
}

func TestWhatsAppWebhook_VerifyToken_Success(t *testing.T) {
	verifyToken := "my-verify-token"
	_, ah := setupTestHandler("", verifyToken)

	challenge := "123456"
	req := httptest.NewRequest("GET", "/api/automation/whatsapp/webhook?hub.mode=subscribe&hub.verify_token="+verifyToken+"&hub.challenge="+challenge, nil)
	w := httptest.NewRecorder()

	ah.WhatsAppWebhook(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, challenge, w.Body.String())
}

func TestWhatsAppWebhook_VerifyToken_Failure(t *testing.T) {
	verifyToken := "my-verify-token"
	_, ah := setupTestHandler("", verifyToken)

	req := httptest.NewRequest("GET", "/api/automation/whatsapp/webhook?hub.mode=subscribe&hub.verify_token=wrong-token", nil)
	w := httptest.NewRecorder()

	ah.WhatsAppWebhook(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestWhatsAppWebhook_VerifyToken_FailClosed(t *testing.T) {
	// No token configured
	_, ah := setupTestHandler("", "")

	req := httptest.NewRequest("GET", "/api/automation/whatsapp/webhook?hub.mode=subscribe&hub.verify_token=anything", nil)
	w := httptest.NewRecorder()

	ah.WhatsAppWebhook(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestWhatsAppWebhook_MaxBodySize(t *testing.T) {
	_, ah := setupTestHandler("", "token")

	// 2MB body
	largeBody := make([]byte, 2*1024*1024)
	req := httptest.NewRequest("POST", "/api/automation/whatsapp/webhook", bytes.NewBuffer(largeBody))
	w := httptest.NewRecorder()

	ah.WhatsAppWebhook(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to read body")
}

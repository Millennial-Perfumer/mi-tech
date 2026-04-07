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
	"github.com/stretchr/testify/mock"
)

type MockConfigGetter struct {
	mock.Mock
}

func (m *MockConfigGetter) Get(key string) (string, error) {
	args := m.Called(key)
	return args.String(0), args.Error(1)
}

func TestMarketingWebhookHandler_MetaWebhook_GET(t *testing.T) {
	mockRepo := new(MockConfigGetter)
	settings := config.NewSettingsProvider(mockRepo)

	verifyToken := "test_verify_token"
	mockRepo.On("Get", "meta_marketing_webhook_verify_token").Return(verifyToken, nil)

	h := NewMarketingWebhookHandler(nil, settings)

	// 1. Success case
	req := httptest.NewRequest("GET", "/api/marketing/meta/webhook?hub.mode=subscribe&hub.verify_token="+verifyToken+"&hub.challenge=test_challenge", nil)
	w := httptest.NewRecorder()
	h.MetaWebhook(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "test_challenge", w.Body.String())

	// 2. Failure case (wrong token)
	req = httptest.NewRequest("GET", "/api/marketing/meta/webhook?hub.mode=subscribe&hub.verify_token=wrong_token&hub.challenge=test_challenge", nil)
	w = httptest.NewRecorder()
	h.MetaWebhook(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestMarketingWebhookHandler_MetaWebhook_POST(t *testing.T) {
	mockRepo := new(MockConfigGetter)
	settings := config.NewSettingsProvider(mockRepo)

	appSecret := "test_app_secret"
	mockRepo.On("Get", "meta_app_secret").Return(appSecret, nil)

	h := NewMarketingWebhookHandler(nil, settings)

	payload := []byte(`{"entry": [{"id": "123", "time": 123456789}]}`)

	// Compute HMAC signature
	mac := hmac.New(sha256.New, []byte(appSecret))
	mac.Write(payload)
	signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	// 1. Success case
	req := httptest.NewRequest("POST", "/api/marketing/meta/webhook", bytes.NewBuffer(payload))
	req.Header.Set("X-Hub-Signature-256", signature)
	w := httptest.NewRecorder()
	h.MetaWebhook(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Acknowledged", w.Body.String())

	// 2. Failure case (invalid signature)
	req = httptest.NewRequest("POST", "/api/marketing/meta/webhook", bytes.NewBuffer(payload))
	req.Header.Set("X-Hub-Signature-256", "sha256=invalid_signature")
	w = httptest.NewRecorder()
	h.MetaWebhook(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// 3. Failure case (body too large)
	largePayload := make([]byte, 2*1024*1024) // 2MB
	req = httptest.NewRequest("POST", "/api/marketing/meta/webhook", bytes.NewBuffer(largePayload))
	w = httptest.NewRecorder()
	h.MetaWebhook(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

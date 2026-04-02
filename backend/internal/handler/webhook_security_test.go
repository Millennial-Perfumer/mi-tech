package handler

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"mi-tech/internal/automation/whatsapp"
	"mi-tech/internal/config"
	"mi-tech/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockConfigsRepository is a mock of the ConfigsRepository for testing security settings.
type MockConfigsRepository struct {
	mock.Mock
}

func (m *MockConfigsRepository) TableName() string { return "app_configs" }

func (m *MockConfigsRepository) Get(key string) (string, error) {
	args := m.Called(key)
	return args.String(0), args.Error(1)
}

func (m *MockConfigsRepository) GetAll() ([]repository.AppConfig, error) {
	args := m.Called()
	return args.Get(0).([]repository.AppConfig), args.Error(1)
}

func (m *MockConfigsRepository) GetAllRevealed() ([]repository.AppConfig, error) {
	args := m.Called()
	return args.Get(0).([]repository.AppConfig), args.Error(1)
}

func (m *MockConfigsRepository) Set(key, value string) error {
	args := m.Called(key, value)
	return args.Error(0)
}

func TestWebhookSecurity_Shopify(t *testing.T) {
	mockRepo := new(MockConfigsRepository)
	settings := config.NewSettingsProvider(mockRepo)

	// We need a dummy WebhookService to avoid nil pointer in background goroutines
	// even if they are triggered asynchronously after the response.
	handler := NewWebhookHandler(nil, nil, settings)

	t.Run("Fail closed when secret is missing", func(t *testing.T) {
		mockRepo.On("Get", "shopify_webhook_secret").Return("", nil).Once()

		req := httptest.NewRequest("POST", "/api/webhooks/shopify", bytes.NewBuffer([]byte("{}")))
		req.Header.Set("X-Shopify-Hmac-Sha256", "some-hmac")
		w := httptest.NewRecorder()

		handler.ShopifyWebhookHandler(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Reject invalid HMAC", func(t *testing.T) {
		secret := "test-secret"
		mockRepo.On("Get", "shopify_webhook_secret").Return(secret, nil).Once()

		req := httptest.NewRequest("POST", "/api/webhooks/shopify", bytes.NewBuffer([]byte("{}")))
		req.Header.Set("X-Shopify-Hmac-Sha256", "invalid-hmac")
		w := httptest.NewRecorder()

		handler.ShopifyWebhookHandler(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Accept valid HMAC", func(t *testing.T) {
		secret := "test-secret"
		body := []byte("{\"test\": true}")
		mockRepo.On("Get", "shopify_webhook_secret").Return(secret, nil).Once()

		hash := hmac.New(sha256.New, []byte(secret))
		hash.Write(body)
		expectedHmac := base64.StdEncoding.EncodeToString(hash.Sum(nil))

		req := httptest.NewRequest("POST", "/api/webhooks/shopify", bytes.NewBuffer(body))
		req.Header.Set("X-Shopify-Hmac-Sha256", expectedHmac)
		w := httptest.NewRecorder()

		// We need a mock webhook service to avoid nil pointer in the background goroutine if we actually reach it
		// But verifyWebhook is called before background processing.
		// However, ShopifyWebhookHandler returns 200 before background processing if verifyWebhook passes.

		handler.ShopifyWebhookHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Reject oversized body", func(t *testing.T) {
		mockRepo.On("Get", "shopify_webhook_secret").Return("test", nil).Maybe()
		// 1MB + 1 byte
		oversized := make([]byte, 1024*1024+1)
		req := httptest.NewRequest("POST", "/api/webhooks/shopify", bytes.NewBuffer(oversized))
		w := httptest.NewRecorder()

		handler.ShopifyWebhookHandler(w, req)

		// http.MaxBytesReader will cause ReadAll to return an error, which the handler catches
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestWebhookSecurity_WhatsApp(t *testing.T) {
	mockRepo := new(MockConfigsRepository)
	settings := config.NewSettingsProvider(mockRepo)
	handler := whatsapp.NewAutomationHandler(nil, nil, nil, nil, nil, settings)

	t.Run("GET challenge requires valid verify token", func(t *testing.T) {
		verifyToken := "my-secret-token"
		mockRepo.On("Get", "whatsapp_webhook_verify_token").Return(verifyToken, nil).Once()

		url := fmt.Sprintf("/api/automation/whatsapp/webhook?hub.mode=subscribe&hub.verify_token=%s&hub.challenge=1234", verifyToken)
		req := httptest.NewRequest("GET", url, nil)
		w := httptest.NewRecorder()

		handler.WhatsAppWebhook(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "1234", w.Body.String())
	})

	t.Run("GET challenge fails with invalid mode", func(t *testing.T) {
		verifyToken := "my-secret-token"
		mockRepo.On("Get", "whatsapp_webhook_verify_token").Return(verifyToken, nil).Once()

		url := fmt.Sprintf("/api/automation/whatsapp/webhook?hub.mode=not-subscribe&hub.verify_token=%s&hub.challenge=1234", verifyToken)
		req := httptest.NewRequest("GET", url, nil)
		w := httptest.NewRecorder()

		handler.WhatsAppWebhook(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("GET challenge fails with invalid token", func(t *testing.T) {
		verifyToken := "my-secret-token"
		mockRepo.On("Get", "whatsapp_webhook_verify_token").Return(verifyToken, nil).Once()

		url := "/api/automation/whatsapp/webhook?hub.mode=subscribe&hub.verify_token=wrong&hub.challenge=1234"
		req := httptest.NewRequest("GET", url, nil)
		w := httptest.NewRecorder()

		handler.WhatsAppWebhook(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("POST reject oversized body", func(t *testing.T) {
		oversized := make([]byte, 1024*1024+1)
		req := httptest.NewRequest("POST", "/api/automation/whatsapp/webhook", bytes.NewBuffer(oversized))
		w := httptest.NewRecorder()

		handler.WhatsAppWebhook(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code) // Handler currently returns 500 on ReadAll error
	})
}

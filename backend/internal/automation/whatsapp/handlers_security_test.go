package whatsapp

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"mi-tech/internal/config"
	"mi-tech/internal/service/mocks"

	"github.com/stretchr/testify/assert"
)

func TestWhatsAppWebhookVerification_Success(t *testing.T) {
	mockRepo := new(mocks.MockConfigsRepository)
	mockRepo.On("Get", "whatsapp_webhook_verify_token").Return("my_secure_token", nil)

	settings := config.NewSettingsProvider(mockRepo)
	handler := NewAutomationHandler(nil, nil, nil, nil, nil, settings)

	req, _ := http.NewRequest("GET", "/api/automation/whatsapp/webhook?hub.mode=subscribe&hub.verify_token=my_secure_token&hub.challenge=12345", nil)
	rr := httptest.NewRecorder()

	handler.WhatsAppWebhook(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "12345", rr.Body.String())
}

func TestWhatsAppWebhookVerification_InvalidToken(t *testing.T) {
	mockRepo := new(mocks.MockConfigsRepository)
	mockRepo.On("Get", "whatsapp_webhook_verify_token").Return("my_secure_token", nil)

	settings := config.NewSettingsProvider(mockRepo)
	handler := NewAutomationHandler(nil, nil, nil, nil, nil, settings)

	req, _ := http.NewRequest("GET", "/api/automation/whatsapp/webhook?hub.mode=subscribe&hub.verify_token=wrong_token&hub.challenge=12345", nil)
	rr := httptest.NewRecorder()

	handler.WhatsAppWebhook(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.Contains(t, rr.Body.String(), "Forbidden")
}

func TestWhatsAppWebhookVerification_InvalidMode(t *testing.T) {
	mockRepo := new(mocks.MockConfigsRepository)
	mockRepo.On("Get", "whatsapp_webhook_verify_token").Return("my_secure_token", nil)

	settings := config.NewSettingsProvider(mockRepo)
	handler := NewAutomationHandler(nil, nil, nil, nil, nil, settings)

	req, _ := http.NewRequest("GET", "/api/automation/whatsapp/webhook?hub.mode=unsubscribe&hub.verify_token=my_secure_token&hub.challenge=12345", nil)
	rr := httptest.NewRecorder()

	handler.WhatsAppWebhook(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.Contains(t, rr.Body.String(), "Forbidden")
}

func TestWhatsAppWebhookVerification_MissingParams(t *testing.T) {
	mockRepo := new(mocks.MockConfigsRepository)

	settings := config.NewSettingsProvider(mockRepo)
	handler := NewAutomationHandler(nil, nil, nil, nil, nil, settings)

	req, _ := http.NewRequest("GET", "/api/automation/whatsapp/webhook?hub.challenge=12345", nil)
	rr := httptest.NewRecorder()

	handler.WhatsAppWebhook(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Missing hub.mode or hub.verify_token")
}

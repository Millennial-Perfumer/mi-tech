package whatsapp

import (
	"mi-tech/internal/config"
	"mi-tech/internal/repository"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestWhatsAppWebhookVerification(t *testing.T) {
	// Setup DB for ConfigsRepository
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&repository.AppConfig{})

	configsRepo := repository.NewConfigsRepository(db)
	settings := config.NewSettingsProvider(configsRepo)

	// Set the verify token in DB
	configsRepo.Set("whatsapp_webhook_verify_token", "secret_token_123")

	handler := &AutomationHandler{
		settings: settings,
	}

	tests := []struct {
		name           string
		query          string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Valid Verification",
			query:          "?hub.mode=subscribe&hub.verify_token=secret_token_123&hub.challenge=test_challenge",
			expectedStatus: http.StatusOK,
			expectedBody:   "test_challenge",
		},
		{
			name:           "Invalid Token",
			query:          "?hub.mode=subscribe&hub.verify_token=wrong_token&hub.challenge=test_challenge",
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Forbidden\n",
		},
		{
			name:           "Invalid Mode",
			query:          "?hub.mode=unsubscribe&hub.verify_token=secret_token_123&hub.challenge=test_challenge",
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Forbidden\n",
		},
		{
			name:           "Missing Parameters",
			query:          "?hub.challenge=test_challenge",
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Forbidden\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "/api/automation/whatsapp/webhook"+tt.query, nil)
			rr := httptest.NewRecorder()

			handler.WhatsAppWebhook(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, tt.expectedBody, rr.Body.String())
		})
	}
}

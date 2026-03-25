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
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&repository.AppConfig{})
	configsRepo := repository.NewConfigsRepository(db)

	expectedToken := "my_secure_token"
	configsRepo.Set("whatsapp_webhook_verify_token", expectedToken)
	handler := &AutomationHandler{settings: config.NewSettingsProvider(configsRepo)}

	cases := []struct {
		name, query string
		expCode     int
		expBody     string
	}{
		{"Valid", "hub.mode=subscribe&hub.verify_token=" + expectedToken + "&hub.challenge=1234", 200, "1234"},
		{"Invalid token", "hub.mode=subscribe&hub.verify_token=wrong&hub.challenge=1234", 403, ""},
		{"Invalid mode", "hub.mode=wrong&hub.verify_token=" + expectedToken + "&hub.challenge=1234", 403, ""},
		{"Missing params", "hub.challenge=1234", 403, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/automation/whatsapp/webhook?"+tc.query, nil)
			rr := httptest.NewRecorder()
			handler.WhatsAppWebhook(rr, req)
			assert.Equal(t, tc.expCode, rr.Code)
			if tc.expBody != "" {
				assert.Equal(t, tc.expBody, rr.Body.String())
			}
		})
	}
}

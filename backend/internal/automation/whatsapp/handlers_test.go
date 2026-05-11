package whatsapp

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"mi-tech/internal/config"
	"mi-tech/internal/repository"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestWhatsAppWebhook_Verification(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.Exec("CREATE TABLE app_configs (key TEXT PRIMARY KEY, value TEXT)")

	configsRepo := repository.NewConfigsRepository(db)
	settings := config.NewSettingsProvider(configsRepo)

	// Set verify token
	db.Exec("INSERT INTO app_configs (key, value) VALUES (?, ?)", "whatsapp_webhook_verify_token", "mytoken")

	handler := &AutomationHandler{
		settings: settings,
	}

	t.Run("Successful Verification", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/webhook?hub.mode=subscribe&hub.verify_token=mytoken&hub.challenge=1234", nil)
		rr := httptest.NewRecorder()

		handler.WhatsAppWebhook(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "1234", rr.Body.String())
	})

	t.Run("Failed Verification - Wrong Token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/webhook?hub.mode=subscribe&hub.verify_token=wrong&hub.challenge=1234", nil)
		rr := httptest.NewRecorder()

		handler.WhatsAppWebhook(rr, req)

		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("Failed Verification - Missing Mode", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/webhook?hub.verify_token=mytoken&hub.challenge=1234", nil)
		rr := httptest.NewRecorder()

		handler.WhatsAppWebhook(rr, req)

		assert.Equal(t, http.StatusForbidden, rr.Code)
	})
}

func TestWhatsAppWebhook_SignatureValidation(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.Exec("CREATE TABLE app_configs (key TEXT PRIMARY KEY, value TEXT)")

	configsRepo := repository.NewConfigsRepository(db)
	settings := config.NewSettingsProvider(configsRepo)

	// Set app secret
	db.Exec("INSERT INTO app_configs (key, value) VALUES (?, ?)", "whatsapp_app_secret", "mysecret")

	handler := &AutomationHandler{
		settings: settings,
	}

	body := []byte("{\"object\":\"whatsapp\"}")
	mac := hmac.New(sha256.New, []byte("mysecret"))
	mac.Write(body)
	signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	t.Run("Valid Signature", func(t *testing.T) {
		isValid := handler.validateWhatsAppSignature(body, signature)
		assert.True(t, isValid)
	})

	t.Run("Invalid Signature", func(t *testing.T) {
		isValid := handler.validateWhatsAppSignature(body, "sha256=wrong")
		assert.False(t, isValid)
	})

	t.Run("Missing Signature", func(t *testing.T) {
		isValid := handler.validateWhatsAppSignature(body, "")
		assert.False(t, isValid)
	})
}

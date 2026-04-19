package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"io"
	"log"
	"mi-tech/internal/config"
	"mi-tech/internal/marketing"
	"net/http"
	"strings"
)

type MarketingWebhookHandler struct {
	metaClient *marketing.MetaMarketingClient
	settings   *config.SettingsProvider
}

func NewMarketingWebhookHandler(metaClient *marketing.MetaMarketingClient, settings *config.SettingsProvider) *MarketingWebhookHandler {
	return &MarketingWebhookHandler{
		metaClient: metaClient,
		settings:   settings,
	}
}

// MetaWebhook handles both the verification (GET) and data (POST) from Meta.
func (h *MarketingWebhookHandler) MetaWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.verifyWebhook(w, r)
		return
	}

	if r.Method == http.MethodPost {
		h.handleNotification(w, r)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func (h *MarketingWebhookHandler) verifyWebhook(w http.ResponseWriter, r *http.Request) {
	// The 'hub.verify_token' from Meta must match our configured one.
	query := r.URL.Query()
	mode := query.Get("hub.mode")
	token := query.Get("hub.verify_token")
	challenge := query.Get("hub.challenge")

	expectedToken := h.settings.GetMetaMarketingWebhookVerifyToken()
	if expectedToken == "" {
		log.Printf("Meta Marketing Webhook Error: No meta_marketing_webhook_verify_token configured")
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Security: Use constant-time comparison to prevent timing attacks
	if mode == "subscribe" && subtle.ConstantTimeCompare([]byte(token), []byte(expectedToken)) == 1 {
		log.Printf("Meta Marketing Webhook verified successfully!")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(challenge))
		return
	}

	log.Printf("Meta Marketing Webhook verification failed")
	http.Error(w, "Verification failed", http.StatusForbidden)
}

func (h *MarketingWebhookHandler) handleNotification(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Meta Marketing Webhook Error: Failed to read body: %v", err)
		http.Error(w, "Failed to read body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// 1. Security: Validate X-Hub-Signature-256
	signature := r.Header.Get("X-Hub-Signature-256")
	if !h.validateMetaSignature(body, signature) {
		log.Printf("Meta Marketing Webhook Error: Invalid signature")
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	log.Printf("Received Meta Marketing Notification: %s", string(body))

	// For now, we just acknowledge.
	// In a real implementation, we'd parse the 'ads_management' or 'ads_insights' payload.
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Acknowledged"))
}

func (h *MarketingWebhookHandler) validateMetaSignature(body []byte, signature string) bool {
	if signature == "" {
		return false
	}

	// Signature format: sha256=HEX_DIGEST
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}
	actualHash := signature[7:]

	appSecret := h.settings.GetMetaAppSecret()
	if appSecret == "" {
		log.Printf("Meta Marketing Webhook Warning: No meta_app_secret configured for validation")
		return false // Fail-closed
	}

	mac := hmac.New(sha256.New, []byte(appSecret))
	mac.Write(body)
	expectedHash := hex.EncodeToString(mac.Sum(nil))

	return subtle.ConstantTimeCompare([]byte(actualHash), []byte(expectedHash)) == 1
}

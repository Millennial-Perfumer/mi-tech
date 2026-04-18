package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
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

	if mode == "subscribe" && subtle.ConstantTimeCompare([]byte(token), []byte(expectedToken)) == 1 {
		fmt.Printf("Meta Marketing Webhook verified successfully!\n")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(challenge))
		return
	}

	http.Error(w, "Verification failed", http.StatusForbidden)
}

func (h *MarketingWebhookHandler) handleNotification(w http.ResponseWriter, r *http.Request) {
	// 1. Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading Meta marketing webhook body: %v", err)
		http.Error(w, "Failed to read body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// 2. Validate HMAC Signature
	signature := r.Header.Get("X-Hub-Signature-256")
	if !h.validateMetaSignature(body, signature) {
		log.Printf("Invalid Meta Marketing X-Hub-Signature-256 received")
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	// For now, we just acknowledge and log.
	// In a real implementation, we'd parse the 'ads_management' or 'ads_insights' payload.
	fmt.Printf("Received Verified Meta Marketing Notification\n")
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

	secret := h.settings.GetMetaAppSecret()
	if secret == "" {
		log.Printf("Meta Marketing Webhook Warning: No meta_app_secret configured")
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expectedHash := hex.EncodeToString(mac.Sum(nil))

	return subtle.ConstantTimeCompare([]byte(actualHash), []byte(expectedHash)) == 1
}

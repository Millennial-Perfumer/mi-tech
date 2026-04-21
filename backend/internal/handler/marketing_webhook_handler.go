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

	// Security: Use constant-time comparison to prevent timing attacks and ensure token is configured
	if mode == "subscribe" && expectedToken != "" && subtle.ConstantTimeCompare([]byte(token), []byte(expectedToken)) == 1 {
		log.Printf("Meta Marketing Webhook verified successfully!")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(challenge))
		return
	}

	log.Printf("Meta Marketing Webhook verification failed")
	http.Error(w, "Verification failed", http.StatusForbidden)
}

func (h *MarketingWebhookHandler) handleNotification(w http.ResponseWriter, r *http.Request) {
	// Security: Limit request body to 1MB to prevent DoS
	limitedReader := io.LimitReader(r.Body, 1<<20)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		log.Printf("Marketing Webhook Error: Failed to read body: %v", err)
		http.Error(w, "Failed to read body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Security: Validate X-Hub-Signature-256
	signature := r.Header.Get("X-Hub-Signature-256")
	if !h.validateSignature(body, signature) {
		log.Printf("Invalid Meta Marketing signature received")
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	// For now, we just acknowledge and log metadata (not full body to avoid PII leak).
	log.Printf("Received Meta Marketing Notification (Size: %d bytes)", len(body))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Acknowledged"))
}

func (h *MarketingWebhookHandler) validateSignature(body []byte, signature string) bool {
	if signature == "" {
		return false
	}

	// Signature format: sha256=HEX_DIGEST
	if len(signature) <= 7 || signature[:7] != "sha256=" {
		return false
	}
	actualHash := signature[7:]

	secret := h.settings.GetMetaAppSecret()
	if secret == "" {
		log.Printf("Marketing Webhook Warning: No meta_app_secret configured. Failing closed.")
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expectedHash := hex.EncodeToString(mac.Sum(nil))

	return subtle.ConstantTimeCompare([]byte(actualHash), []byte(expectedHash)) == 1
}

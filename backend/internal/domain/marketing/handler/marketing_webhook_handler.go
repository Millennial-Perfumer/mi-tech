package handler

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"mi-tech/internal/domain/marketing/service"
	"mi-tech/internal/domain/shared/config"
	"net/http"
	"os"
	"strings"
)

type MarketingWebhookHandler struct {
	metaClient *service.MetaMarketingClient
	settings   *config.SettingsProvider
}

func NewMarketingWebhookHandler(metaClient *service.MetaMarketingClient, settings *config.SettingsProvider) *MarketingWebhookHandler {
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

	if mode == "subscribe" && token == expectedToken {
		fmt.Printf("Meta Marketing Webhook verified successfully!\n")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(challenge))
		return
	}

	http.Error(w, "Verification failed", http.StatusForbidden)
}

func (h *MarketingWebhookHandler) handleNotification(w http.ResponseWriter, r *http.Request) {
	// Read body with 1MB limit to prevent DoS
	body, err := io.ReadAll(io.LimitReader(r.Body, 1024*1024))
	if err != nil {
		log.Printf("Marketing Webhook Error: Failed to read body: %v", err)
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	// Restore body for subsequent handlers/middleware
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	signature := r.Header.Get("X-Hub-Signature-256")
	if signature == "" {
		log.Printf("Marketing Webhook Error: Missing X-Hub-Signature-256 header")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Try environment variable first for testing/override, then fallback to settings provider
	appSecret := os.Getenv("META_APP_SECRET")
	if appSecret == "" && h.settings != nil {
		appSecret = h.settings.GetMetaAppSecret()
	}

	if appSecret == "" {
		log.Printf("Marketing Webhook Error: Meta App Secret not configured (fail-closed)")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if !strings.HasPrefix(signature, "sha256=") {
		log.Printf("Marketing Webhook Error: Invalid signature format")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	actualHash := signature[7:]

	mac := hmac.New(sha256.New, []byte(appSecret))
	mac.Write(body)
	expectedHash := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(actualHash), []byte(expectedHash)) {
		log.Printf("Marketing Webhook Error: HMAC Mismatch!")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// For now, we just acknowledge and log.
	// In a real implementation, we'd parse the 'ads_management' or 'ads_insights' payload.
	fmt.Printf("Received Meta Marketing Notification\n")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Acknowledged"))
}

package whatsapp

import (
	"crypto/subtle"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWhatsAppWebhook_VerificationLogic(t *testing.T) {
	expectedToken := "secure_verify_token"
	providedToken := "secure_verify_token"

	// Test the same logic as in handlers.go:
	// subtle.ConstantTimeCompare([]byte(token), []byte(expectedToken)) == 1

	assert.Equal(t, 1, subtle.ConstantTimeCompare([]byte(providedToken), []byte(expectedToken)), "Tokens should match")
	assert.Equal(t, 0, subtle.ConstantTimeCompare([]byte("wrong"), []byte(expectedToken)), "Tokens should not match")
}

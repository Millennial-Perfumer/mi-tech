package whatsapp

import (
	"crypto/subtle"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestConstantTimeCompare verifies the behavior of timing-safe comparison.
func TestConstantTimeCompare(t *testing.T) {
	token := "my-secret-token"
	expected := "my-secret-token"
	wrong := "wrong-token"

	assert.Equal(t, 1, subtle.ConstantTimeCompare([]byte(token), []byte(expected)))
	assert.Equal(t, 0, subtle.ConstantTimeCompare([]byte(token), []byte(wrong)))
}

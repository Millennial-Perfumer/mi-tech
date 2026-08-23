package test

import (
	"testing"

	"mi-tech/internal/domain/communication/service"

	"github.com/stretchr/testify/assert"
)

func TestRenderTemplateMessageText(t *testing.T) {
	components := []interface{}{
		map[string]interface{}{
			"type": "header",
			"parameters": []map[string]string{
				{"type": "text", "text": "Hi Alice"},
			},
		},
		map[string]interface{}{
			"type": "body",
			"parameters": []map[string]string{
				{"type": "text", "text": "Alice"},
				{"type": "text", "text": "#1001"},
				{"type": "text", "text": "https://feedback-form.example.in/?order_id=42&phone=919000000000"},
			},
		},
		map[string]interface{}{
			"type":     "button",
			"sub_type": "url",
			"index":    "0",
			"parameters": []map[string]interface{}{
				{"type": "text", "text": "https://example.in/checkout"},
			},
		},
	}

	got := service.RenderTemplateMessageText(
		"Hello {{1}}, your order {{2}} has been delivered. Please share your feedback: {{3}}",
		"{{1}}",
		components,
	)

	want := "Hi Alice\n\nHello Alice, your order #1001 has been delivered. Please share your feedback: https://feedback-form.example.in/?order_id=42&phone=919000000000\nhttps://example.in/checkout"
	assert.Equal(t, want, got)
}

func TestRenderTemplateMessageTextEmptyParams(t *testing.T) {
	// No components -> placeholders stay as-is, empty result for body without placeholders.
	got := service.RenderTemplateMessageText("", "", nil)
	assert.Equal(t, "", got)
}
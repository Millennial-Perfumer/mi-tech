package whatsapp

import (
	"testing"

	"mi-tech/internal/entity"

	"github.com/stretchr/testify/assert"
)

func TestResolveVariable(t *testing.T) {
	order := entity.Order{
		OrderNumber:       "#1001",
		TotalPrice:        500.50,
		CustomerFirstName: entity.StrPtr("Alice"),
		CustomerCity:      entity.StrPtr("Mumbai"),
	}

	assert.Equal(t, "Alice", resolveVariable("customer_name", order))
	assert.Equal(t, "1001", resolveVariable("order_id", order))
	assert.Equal(t, "500.50", resolveVariable("order_total", order))
	assert.Equal(t, "Mumbai", resolveVariable("customer_city", order))
	assert.Equal(t, "", resolveVariable("unknown", order))
}

func TestSanitizePhoneNumber(t *testing.T) {
	assert.Equal(t, "919876543210", sanitizePhoneNumber("+91 98765-43210"))
	assert.Equal(t, "916383173716", sanitizePhoneNumber("555-555-SHIP"))
}

func TestCountRequiredParams(t *testing.T) {
	service := &WebhookMappingService{}
	assert.Equal(t, 2, service.countRequiredParams("Hello {{1}}, your order {{2}} is ready"))
	assert.Equal(t, 0, service.countRequiredParams("Hello world"))
	assert.Equal(t, 5, service.countRequiredParams("{{5}} {{1}}"))
}

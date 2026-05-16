package whatsapp

import (
	"testing"

	"mi-tech/internal/entity"
	"mi-tech/internal/service"

	"github.com/stretchr/testify/assert"
)

func TestResolveVariable(t *testing.T) {
	order := entity.Order{
		OrderNumber:       "#1001",
		TotalPrice:        500.50,
		CustomerFirstName: entity.StrPtr("Alice"),
		CustomerCity:      entity.StrPtr("Mumbai"),
	}

	service := &WebhookMappingService{
		invoiceService: &service.InvoiceService{},
	}

	// 1. Test basic fields
	assert.Equal(t, "Alice", service.resolveVariable("customer_name", order, nil))
	assert.Equal(t, "1001", service.resolveVariable("order_id", order, nil))
	assert.Equal(t, "500.50", service.resolveVariable("order_total", order, nil))
	assert.Equal(t, "Mumbai", service.resolveVariable("customer_city", order, nil))

	// 2. Test pricing variables with LineItems
	order.LineItems = []entity.LineItem{
		{Price: 100, Quantity: 1},
	}
	// Calculate totals: Gross=100, Taxable=84.75, Tax=15.25, Grand=100
	totals := service.invoiceService.CalculateInvoiceTotals(order.LineItems)

	// Test with nil totals (lazy load)
	assert.Equal(t, "100.00", service.resolveVariable("order_total", order, nil))
	assert.Equal(t, "100.00", service.resolveVariable("order_subtotal", order, nil))

	// Test with pre-calculated totals (optimized path)
	assert.Equal(t, "100.00", service.resolveVariable("order_total", order, &totals))
	assert.Equal(t, "15.25", service.resolveVariable("order_tax", order, &totals))

	assert.Equal(t, "", service.resolveVariable("unknown", order, nil))
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

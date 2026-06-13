package test

import (
	"testing"

	"mi-tech/internal/domain/communication/service"
	"mi-tech/internal/domain/order/entity"
	orderServicePkg "mi-tech/internal/domain/order/service"
	"mi-tech/internal/shared/util"

	"github.com/stretchr/testify/assert"
)

func TestResolveVariable(t *testing.T) {
	order := entity.Order{
		OrderNumber:       "#1001",
		TotalPrice:        500.50,
		CustomerFirstName: util.StrPtr("Alice"),
		CustomerCity:      util.StrPtr("Mumbai"),
	}

	// Use constructor to inject the invoiceService dependency
	mappingService := service.NewWebhookMappingService(nil, nil, &orderServicePkg.InvoiceService{}, nil, nil, nil, nil)

	// 1. Test basic fields
	assert.Equal(t, "Alice", mappingService.ResolveVariable("customer_name", order, nil))
	assert.Equal(t, "1001", mappingService.ResolveVariable("order_id", order, nil))
	assert.Equal(t, "500.50", mappingService.ResolveVariable("order_total", order, nil))
	assert.Equal(t, "Mumbai", mappingService.ResolveVariable("customer_city", order, nil))

	// 2. Test pricing variables with LineItems
	order.LineItems = []entity.LineItem{
		{Price: 100, Quantity: 1},
	}
	// Calculate totals: Gross=100, Taxable=84.75, Tax=15.25, Grand=100
	// Wait, we need to get invoiceService from mappingService or instantiate it directly.
	// Since we can just call CalculateInvoiceTotals on a separate instance of InvoiceService:
	invoiceService := &orderServicePkg.InvoiceService{}
	totals := invoiceService.CalculateInvoiceTotals(order.LineItems)

	// Test with nil totals (lazy load)
	assert.Equal(t, "100.00", mappingService.ResolveVariable("order_total", order, nil))
	assert.Equal(t, "100.00", mappingService.ResolveVariable("order_subtotal", order, nil))

	// Test with pre-calculated totals (optimized path)
	assert.Equal(t, "100.00", mappingService.ResolveVariable("order_total", order, &totals))
	assert.Equal(t, "15.25", mappingService.ResolveVariable("order_tax", order, &totals))

	assert.Equal(t, "", mappingService.ResolveVariable("unknown", order, nil))
}

func TestSanitizePhoneNumber(t *testing.T) {
	assert.Equal(t, "919876543210", service.SanitizePhoneNumber("+91 98765-43210"))
	assert.Equal(t, "916383173716", service.SanitizePhoneNumber("555-555-SHIP"))
}

func TestCountRequiredParams(t *testing.T) {
	mappingService := service.NewWebhookMappingService(nil, nil, nil, nil, nil, nil, nil)
	assert.Equal(t, 2, mappingService.CountRequiredParams("Hello {{1}}, your order {{2}} is ready"))
	assert.Equal(t, 0, mappingService.CountRequiredParams("Hello world"))
	assert.Equal(t, 5, mappingService.CountRequiredParams("{{5}} {{1}}"))
}

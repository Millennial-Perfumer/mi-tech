package test

import (
	"bytes"
	"testing"
	"time"

	orderEntity "mi-tech/internal/domain/order/entity"
	orderService "mi-tech/internal/domain/order/service"
	configRepoPkg "mi-tech/internal/shared/config/repository"
	util "mi-tech/internal/shared/util"

	"github.com/stretchr/testify/assert"
)

type mockSettingsRepo struct {
	configRepoPkg.SettingsRepository
	settings map[string]string
}

func (m *mockSettingsRepo) Get(key string) (string, error) {
	return m.settings[key], nil
}

func TestInvoiceService_GeneratePDF(t *testing.T) {
	mockRepo := &mockSettingsRepo{
		settings: map[string]string{
			"business_name":          "Test Brand",
			"business_gstin":         "1234567890",
			"business_address_line1": "Test Street",
			"business_address_line2": "Test City",
			"business_phone":         "9999999999",
		},
	}
	service := orderService.NewInvoiceService(mockRepo)

	order := orderEntity.Order{
		OrderNumber:  "1001",
		CreatedAt:    time.Now(),
		CustomerName: util.StrPtr("John Doe"),
	}
	items := []orderEntity.LineItem{
		{ID: "L1", Title: util.StrPtr("Product 1"), Quantity: 1, Price: 100.0},
	}

	var buf bytes.Buffer
	err := service.GeneratePDF(order, items, &buf)

	assert.NoError(t, err)
	assert.True(t, buf.Len() > 0)
}

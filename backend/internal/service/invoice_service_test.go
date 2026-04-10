package service

import (
	"bytes"
	"testing"
	"time"

	"mi-tech/internal/entity"
	"mi-tech/internal/repository"

	"github.com/stretchr/testify/assert"
)

type mockSettingsRepo struct {
	repository.ISettingsRepository
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
	service := NewInvoiceService(mockRepo)

	order := entity.Order{
		OrderNumber:  "1001",
		CreatedAt:    time.Now(),
		CustomerName: entity.StrPtr("John Doe"),
	}
	items := []entity.LineItem{
		{ID: "L1", Title: entity.StrPtr("Product 1"), Quantity: 1, Price: 100.0},
	}

	var buf bytes.Buffer
	err := service.GeneratePDF(order, items, &buf)

	assert.NoError(t, err)
	assert.True(t, buf.Len() > 0)
}

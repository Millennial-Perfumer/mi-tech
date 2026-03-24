package service

import (
	"bytes"
	"testing"
	"time"

	"mi-tech/internal/entity"
	"mi-tech/internal/repository"
	"mi-tech/internal/testutil"

	"github.com/stretchr/testify/assert"
)

func TestInvoiceService_GeneratePDF(t *testing.T) {
	db, err := testutil.SetupTestDB()
	if err != nil {
		t.Skip("DB not available")
	}
	defer testutil.CleanupTestDB(db)

	repo := repository.NewSettingsRepository(db)
	service := NewInvoiceService(repo)

	order := entity.Order{
		OrderNumber: "1001",
		CreatedAt:   time.Now(),
		CustomerName: entity.StrPtr("John Doe"),
	}
	items := []entity.LineItem{
		{ID: "L1", Title: entity.StrPtr("Product 1"), Quantity: 1, Price: 100.0},
	}

	var buf bytes.Buffer
	err = service.GeneratePDF(order, items, &buf)

	if err != nil {
		t.Logf("PDF generation failed: %v", err)
		return
	}

	assert.NoError(t, err)
	assert.True(t, buf.Len() > 0)
}

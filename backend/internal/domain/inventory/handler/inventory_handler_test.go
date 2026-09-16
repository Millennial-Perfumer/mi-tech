package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"mi-tech/internal/domain/inventory/service"
	inventorytest "mi-tech/internal/domain/inventory/test"
)

type recordingStockOrchestrator struct {
	itemID          int
	delta           int
	sourcePlatform  string
	reason          string
	externalOrderID *string
}

func (o *recordingStockOrchestrator) AdjustStock(_ context.Context, itemID, delta int, sourcePlatform, reason string, externalOrderID *string) error {
	o.itemID = itemID
	o.delta = delta
	o.sourcePlatform = sourcePlatform
	o.reason = reason
	o.externalOrderID = externalOrderID
	return nil
}

func (o *recordingStockOrchestrator) UpdateStock(context.Context, int, int, string, string) error {
	return nil
}

func newTestInventoryHandler(orchestrator service.StockOrchestrator) *InventoryHandler {
	repo := &inventorytest.MockInventoryRepository{}
	inventoryService := service.NewInventoryService(repo, nil, orchestrator, nil, nil)
	return NewInventoryHandler(inventoryService)
}

func TestAdjustStockPassesAuditMetadata(t *testing.T) {
	orchestrator := &recordingStockOrchestrator{}
	handler := newTestInventoryHandler(orchestrator)
	req := httptest.NewRequest(http.MethodPost, "/api/inventory/adjust?id=66&delta=-1&reason=sale&platform=Flipkart&external_order_id=FK-123", nil)
	recorder := httptest.NewRecorder()

	handler.AdjustStock(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, 66, orchestrator.itemID)
	require.Equal(t, -1, orchestrator.delta)
	require.Equal(t, "flipkart", orchestrator.sourcePlatform)
	require.Equal(t, "sale", orchestrator.reason)
	require.NotNil(t, orchestrator.externalOrderID)
	require.Equal(t, "FK-123", *orchestrator.externalOrderID)
}

func TestAdjustStockUsesLegacyMetadataDefaults(t *testing.T) {
	orchestrator := &recordingStockOrchestrator{}
	handler := newTestInventoryHandler(orchestrator)
	req := httptest.NewRequest(http.MethodPost, "/api/inventory/adjust?id=66&delta=1", nil)
	recorder := httptest.NewRecorder()

	handler.AdjustStock(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "internal", orchestrator.sourcePlatform)
	require.Equal(t, "manual_adjustment", orchestrator.reason)
	require.Nil(t, orchestrator.externalOrderID)
}

func TestAdjustStockRejectsZeroDelta(t *testing.T) {
	orchestrator := &recordingStockOrchestrator{}
	handler := newTestInventoryHandler(orchestrator)
	req := httptest.NewRequest(http.MethodPost, "/api/inventory/adjust?id=66&delta=0", nil)
	recorder := httptest.NewRecorder()

	handler.AdjustStock(recorder, req)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

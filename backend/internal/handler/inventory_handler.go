package handler

import (
	"encoding/json"
	"mi-tech/internal/entity"
	"mi-tech/internal/service"
	"net/http"
	"strconv"
)

type InventoryHandler struct {
	service *service.InventoryService
}

func NewInventoryHandler(service *service.InventoryService) *InventoryHandler {
	return &InventoryHandler{service: service}
}

// GetDashboard returns all inventory items with their mappings.
func (h *InventoryHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	items, err := h.service.GetInventoryDashboard(search)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

// GetNextSKU returns the suggested mi-XX SKU.
func (h *InventoryHandler) GetNextSKU(w http.ResponseWriter, r *http.Request) {
	next, err := h.service.SuggestNextSKU()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"next_sku": next})
}

// CreateItem handles manual product creation.
func (h *InventoryHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	var item entity.InventoryItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.CreateItem(r.Context(), &item); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
}

// SyncShopify returns a list of staged products from Shopify for mapping.
func (h *InventoryHandler) SyncShopify(w http.ResponseWriter, r *http.Request) {
	staged, err := h.service.SyncShopifyProducts(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(staged)
}

// CreateMapping creates a link between an external and internal SKU.
func (h *InventoryHandler) CreateMapping(w http.ResponseWriter, r *http.Request) {
	var req struct {
		InternalItemID int    `json:"internal_item_id"`
		Platform       string `json:"platform"`
		ExternalSKU    string `json:"external_sku"`
		VariantID      string `json:"variant_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.MapProduct(r.Context(), req.InternalItemID, req.Platform, req.ExternalSKU, req.VariantID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// AdjustStock handles manual stock updates.
func (h *InventoryHandler) AdjustStock(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	deltaStr := r.URL.Query().Get("delta")

	id, _ := strconv.Atoi(idStr)
	delta, _ := strconv.Atoi(deltaStr)

	if err := h.service.AdjustStock(id, delta); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

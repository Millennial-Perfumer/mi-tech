package handler

import (
	"encoding/json"
	"log"
	"mi-tech/internal/entity"
	"mi-tech/internal/service"
	"net/http"
	"strconv"
	"time"
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
		log.Printf("InventoryHandler.GetDashboard error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

// GetNextSKU returns the suggested mi-XX SKU.
func (h *InventoryHandler) GetNextSKU(w http.ResponseWriter, r *http.Request) {
	next, err := h.service.SuggestNextSKU()
	if err != nil {
		log.Printf("InventoryHandler.GetNextSKU error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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
		log.Printf("InventoryHandler.CreateItem error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
}

// BulkCreate handles bulk product creation.
func (h *InventoryHandler) BulkCreate(w http.ResponseWriter, r *http.Request) {
	var items []entity.InventoryItem
	if err := json.NewDecoder(r.Body).Decode(&items); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.BulkImport(r.Context(), items); err != nil {
		log.Printf("InventoryHandler.BulkCreate error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// Clear handles warehouse reset.
func (h *InventoryHandler) Clear(w http.ResponseWriter, r *http.Request) {
	if err := h.service.ClearAll(r.Context()); err != nil {
		log.Printf("InventoryHandler.Clear error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// SyncShopify returns a list of staged products from Shopify for mapping.
func (h *InventoryHandler) SyncShopify(w http.ResponseWriter, r *http.Request) {
	staged, err := h.service.SyncShopifyProducts(r.Context())
	if err != nil {
		log.Printf("InventoryHandler.SyncShopify error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(staged)
}

// SyncPrices calls the shopify api and gets the price of each product and updates the local db.
func (h *InventoryHandler) SyncPrices(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.SyncShopifyPrices(r.Context())
	if err != nil {
		log.Printf("InventoryHandler.SyncPrices error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
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
		log.Printf("InventoryHandler.CreateMapping error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// DeleteMapping removes a link between an external and internal SKU by ID.
func (h *InventoryHandler) DeleteMapping(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid mapping ID", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteMapping(r.Context(), id); err != nil {
		log.Printf("InventoryHandler.DeleteMapping error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}


// AdjustStock handles manual stock updates.
func (h *InventoryHandler) AdjustStock(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	deltaStr := r.URL.Query().Get("delta")

	id, _ := strconv.Atoi(idStr)
	delta, _ := strconv.Atoi(deltaStr)

	if err := h.service.AdjustStock(id, delta); err != nil {
		log.Printf("InventoryHandler.AdjustStock error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// UpdateStock handles absolute manual stock overrides.
func (h *InventoryHandler) UpdateStock(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	valStr := r.URL.Query().Get("val")

	id, _ := strconv.Atoi(idStr)
	val, _ := strconv.Atoi(valStr)

	if err := h.service.UpdateStockCount(id, val); err != nil {
		log.Printf("InventoryHandler.UpdateStock error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetLogs returns the stock movement history for an item.
func (h *InventoryHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)

	logs, err := h.service.GetLogs(id)
	if err != nil {
		log.Printf("InventoryHandler.GetLogs error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

// SyncAmazon triggers an immediate poll of Amazon orders.
func (h *InventoryHandler) SyncAmazon(w http.ResponseWriter, r *http.Request) {
	var req struct {
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
	}
	
	// Optional body
	json.NewDecoder(r.Body).Decode(&req)

	var start, end *time.Time
	if req.StartDate != "" {
		t, err := time.Parse("2006-01-02", req.StartDate)
		if err == nil {
			start = &t
		}
	}
	if req.EndDate != "" {
		t, err := time.Parse("2006-01-02", req.EndDate)
		if err == nil {
			// Set end to 23:59:59 to include the entire end date
			t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			end = &t
		}
	}

	h.service.SyncAmazonOrders(r.Context(), start, end)
	w.WriteHeader(http.StatusAccepted)
}

// UpdateItem handles partial updates to a product.
func (h *InventoryHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	var item entity.InventoryItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateItem(r.Context(), &item); err != nil {
		log.Printf("InventoryHandler.UpdateItem error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

package handler

import (
	"encoding/json"
	"mi-tech/internal/entity"
	"mi-tech/internal/service"
	"net/http"
	"strconv"
)

type OilInventoryHandler struct {
	service *service.OilInventoryService
}

func NewOilInventoryHandler(service *service.OilInventoryService) *OilInventoryHandler {
	return &OilInventoryHandler{service: service}
}

func (h *OilInventoryHandler) ListOils(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	oils, err := h.service.ListOils(search)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(oils)
}

func (h *OilInventoryHandler) CreateOil(w http.ResponseWriter, r *http.Request) {
	var oil entity.OilInventory
	if err := json.NewDecoder(r.Body).Decode(&oil); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.service.CreateOil(&oil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(oil)
}

func (h *OilInventoryHandler) UpdateOil(w http.ResponseWriter, r *http.Request) {
	var oil entity.OilInventory
	if err := json.NewDecoder(r.Body).Decode(&oil); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.service.UpdateOil(&oil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(oil)
}

func (h *OilInventoryHandler) DeleteOil(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	if err := h.service.DeleteOil(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *OilInventoryHandler) BulkDeleteOils(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDs []int `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.BulkDeleteOils(req.IDs); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

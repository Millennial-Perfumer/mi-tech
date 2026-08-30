package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"mi-tech/internal/domain/production/entity"
	"mi-tech/internal/domain/production/service"
)

type OilInventoryHandler struct {
	service *service.OilInventoryService
}

type oilPageResponse struct {
	Items []entity.OilInventory `json:"items"`
	Total int64                 `json:"total"`
	Page  int                   `json:"page"`
	Limit int                   `json:"limit"`
}

func NewOilInventoryHandler(service *service.OilInventoryService) *OilInventoryHandler {
	return &OilInventoryHandler{service: service}
}

func (h *OilInventoryHandler) ListOils(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	pageParam := r.URL.Query().Get("page")
	limitParam := r.URL.Query().Get("limit")
	if pageParam != "" || limitParam != "" {
		page, _ := strconv.Atoi(pageParam)
		limit, _ := strconv.Atoi(limitParam)
		if page < 1 {
			page = 1
		}
		if limit < 1 || limit > 100 {
			limit = 10
		}

		oils, total, err := h.service.ListOilsPage(search, r.URL.Query().Get("sort"), page, limit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(oilPageResponse{Items: oils, Total: total, Page: page, Limit: limit})
		return
	}

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
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)
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
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)
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
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)
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

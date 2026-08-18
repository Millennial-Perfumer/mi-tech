package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"mi-tech/internal/domain/production/entity"
	"mi-tech/internal/domain/production/service"
)

type ManufacturingHandler struct {
	svc *service.ManufacturingService
}

type manufacturingPageResponse struct {
	Items []entity.ManufacturingRecord `json:"items"`
	Total int64                        `json:"total"`
	Page  int                          `json:"page"`
	Limit int                          `json:"limit"`
}

func NewManufacturingHandler(svc *service.ManufacturingService) *ManufacturingHandler {
	return &ManufacturingHandler{svc: svc}
}

func (h *ManufacturingHandler) List(w http.ResponseWriter, r *http.Request) {
	if pageParam := r.URL.Query().Get("page"); pageParam != "" {
		page, _ := strconv.Atoi(pageParam)
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		records, total, err := h.svc.ListPage(page, limit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(manufacturingPageResponse{Items: records, Total: total, Page: page, Limit: limit})
		return
	}

	if limitParam := r.URL.Query().Get("limit"); limitParam != "" {
		limit, _ := strconv.Atoi(limitParam)
		records, err := h.svc.ListRecent(limit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(records)
		return
	}

	records, err := h.svc.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(records)
}

func (h *ManufacturingHandler) Create(w http.ResponseWriter, r *http.Request) {
	var record entity.ManufacturingRecord
	if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.svc.Create(r.Context(), &record); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(record)
}

func (h *ManufacturingHandler) Update(w http.ResponseWriter, r *http.Request) {
	var record entity.ManufacturingRecord
	if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.svc.Update(r.Context(), &record); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(record)
}

func (h *ManufacturingHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

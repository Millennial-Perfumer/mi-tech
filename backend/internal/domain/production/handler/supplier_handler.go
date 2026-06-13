package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"mi-tech/internal/domain/production/entity"
	"mi-tech/internal/domain/production/service"
)

type SupplierHandler struct {
	service *service.SupplierService
}

func NewSupplierHandler(service *service.SupplierService) *SupplierHandler {
	return &SupplierHandler{service: service}
}

func (h *SupplierHandler) ListSuppliers(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	suppliers, err := h.service.ListSuppliers(search)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(suppliers)
}

func (h *SupplierHandler) CreateSupplier(w http.ResponseWriter, r *http.Request) {
	var supplier entity.Supplier
	if err := json.NewDecoder(r.Body).Decode(&supplier); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.service.CreateSupplier(&supplier); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(supplier)
}

func (h *SupplierHandler) UpdateSupplier(w http.ResponseWriter, r *http.Request) {
	var supplier entity.Supplier
	if err := json.NewDecoder(r.Body).Decode(&supplier); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.service.UpdateSupplier(&supplier); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(supplier)
}

func (h *SupplierHandler) DeleteSupplier(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	if err := h.service.DeleteSupplier(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

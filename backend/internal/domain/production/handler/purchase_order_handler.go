package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"mi-tech/internal/domain/production/entity"
	"mi-tech/internal/domain/production/service"
)

type PurchaseOrderHandler struct {
	svc *service.PurchaseOrderService
}

type purchaseOrderPageResponse struct {
	Items      []entity.PurchaseOrder `json:"items"`
	TotalDates int64                  `json:"total_dates"`
	Page       int                    `json:"page"`
	Days       int                    `json:"days"`
}

func NewPurchaseOrderHandler(svc *service.PurchaseOrderService) *PurchaseOrderHandler {
	return &PurchaseOrderHandler{svc: svc}
}

func (h *PurchaseOrderHandler) List(w http.ResponseWriter, r *http.Request) {
	if daysParam := r.URL.Query().Get("days"); daysParam != "" {
		days, _ := strconv.Atoi(daysParam)
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page < 1 {
			page = 1
		}
		pos, totalDates, err := h.svc.ListRecentDays(days, page)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(purchaseOrderPageResponse{Items: pos, TotalDates: totalDates, Page: page, Days: days})
		return
	}

	if limitParam := r.URL.Query().Get("limit"); limitParam != "" {
		limit, _ := strconv.Atoi(limitParam)
		pos, err := h.svc.ListRecent(limit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(pos)
		return
	}

	pos, err := h.svc.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(pos)
}

func (h *PurchaseOrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var po entity.PurchaseOrder
	if err := json.NewDecoder(r.Body).Decode(&po); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.svc.Create(&po); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(po)
}

func (h *PurchaseOrderHandler) BulkCreate(w http.ResponseWriter, r *http.Request) {
	var pos []entity.PurchaseOrder
	if err := json.NewDecoder(r.Body).Decode(&pos); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.svc.BulkCreate(pos); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(pos)
}

func (h *PurchaseOrderHandler) Update(w http.ResponseWriter, r *http.Request) {
	var po entity.PurchaseOrder
	if err := json.NewDecoder(r.Body).Decode(&po); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.svc.Update(&po); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(po)
}

func (h *PurchaseOrderHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := h.svc.Delete(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

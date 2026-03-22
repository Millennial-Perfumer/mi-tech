package handler

import (
	"encoding/json"
	"mi-tech/internal/service"
	"net/http"
	"strconv"
)

type CustomerHandler struct {
	service *service.CustomerService
}

func NewCustomerHandler(service *service.CustomerService) *CustomerHandler {
	return &CustomerHandler{service: service}
}

func (h *CustomerHandler) ListCustomers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	pageSize, _ := strconv.Atoi(q.Get("pageSize"))
	minSpent, _ := strconv.ParseFloat(q.Get("min_spent"), 64)
	maxSpent, _ := strconv.ParseFloat(q.Get("max_spent"), 64)
	minOrders, _ := strconv.Atoi(q.Get("min_orders"))

	filter := service.CustomerFilter{
		Search:    q.Get("search"),
		SortBy:    q.Get("sortBy"),
		SortOrder: q.Get("sortOrder"),
		SourceID:  q.Get("source_id"),
		MinSpent:  minSpent,
		MaxSpent:  maxSpent,
		MinOrders: minOrders,
		City:      q.Get("city"),
		State:     q.Get("state"),
		Page:      page,
		PageSize:  pageSize,
	}

	customers, total, err := h.service.ListCustomers(r.Context(), filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"customers": customers,
		"total":     total,
	})
}

func (h *CustomerHandler) ImportCSV(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 10MB max
	r.ParseMultipartForm(10 << 20)
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file from request", http.StatusBadRequest)
		return
	}
	defer file.Close()

	sourceID := r.FormValue("source_id")

	err = h.service.ImportFromCSV(r.Context(), file, sourceID)
	if err != nil {
		http.Error(w, "Failed to import CSV: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Import successful"})
}

func (h *CustomerHandler) DeleteAllCustomers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := h.service.DeleteAllCustomers(r.Context())
	if err != nil {
		http.Error(w, "Failed to delete customers: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "All customers cleared successfully"})
}

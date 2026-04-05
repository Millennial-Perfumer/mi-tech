package handler

import (
	"encoding/json"
	"mi-tech/internal/entity"
	"mi-tech/internal/service"
	"net/http"
	"strconv"
	"strings"
)

type CustomerHandler struct {
	service *service.CustomerService
}

func NewCustomerHandler(service *service.CustomerService) *CustomerHandler {
	return &CustomerHandler{service: service}
}

// ListCustomers handles GET /api/customers.
// @Summary List customers
// @Description Retrieve a paginated list of customers with advanced filtering.
// @Tags customers
// @Security Bearer
// @Produce json
// @Param page query int false "Page number"
// @Param pageSize query int false "Items per page"
// @Param search query string false "Search"
// @Param sortBy query string false "Sort field"
// @Param sortOrder query string false "Sort order"
// @Param city query string false "City filter"
// @Param state query string false "State filter"
// @Success 200 {object} map[string]interface{}
// @Router /customers [get]
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

// ImportCSV handles POST /api/customers/import.
// @Summary Import customers CSV
// @Description Upload a CSV file and import customer data.
// @Tags customers
// @Security Bearer
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "CSV file"
// @Param source_id formData string false "Source ID"
// @Success 200 {object} map[string]string
// @Router /customers/import [post]
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

// DeleteAllCustomers handles DELETE /api/customers.
// @Summary Delete all customers
// @Description Wipe the entire customers table. Required 'admin' role.
// @Tags customers
// @Security Bearer
// @Success 200 {object} map[string]string
// @Router /customers [delete]
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
// CreateCustomer handles POST /api/customers.
// @Summary Create customer
// @Description Manually create a new customer profile.
// @Tags customers
// @Security Bearer
// @Accept json
// @Produce json
// @Param body body entity.Customer true "Customer data"
// @Success 201 {object} entity.Customer
// @Router /customers [post]
func (h *CustomerHandler) CreateCustomer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		entity.Customer
		SyncToShopify bool `json:"sync_to_shopify"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.service.CreateCustomer(r.Context(), &req.Customer, req.SyncToShopify)
	if err != nil {
		http.Error(w, "Failed to create customer: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(req.Customer)
}

// UpdateCustomer handles PUT /api/customers/{id}.
// @Summary Update customer
// @Description Update an existing customer profile.
// @Tags customers
// @Security Bearer
// @Accept json
// @Produce json
// @Param id path int true "Customer ID"
// @Param body body entity.Customer true "Updated data"
// @Success 200 {object} entity.Customer
// @Router /customers/{id} [put]
func (h *CustomerHandler) UpdateCustomer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/customers/")
	idStr = strings.TrimSuffix(idStr, "/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid customer ID: "+err.Error(), http.StatusBadRequest)
		return
	}

	var req struct {
		entity.Customer
		SyncToShopify bool `json:"sync_to_shopify"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	req.Customer.ID = id
	err = h.service.UpdateCustomer(r.Context(), &req.Customer, req.SyncToShopify)
	if err != nil {
		http.Error(w, "Failed to update customer: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(req.Customer)
}

// DeleteCustomer handles DELETE /api/customers/{id}.
// @Summary Delete customer
// @Description Remove a specific customer.
// @Tags customers
// @Security Bearer
// @Param id path int true "Customer ID"
// @Success 204 "No Content"
// @Router /customers/{id} [delete]
func (h *CustomerHandler) DeleteCustomer(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/customers/")
	idStr = strings.TrimSuffix(idStr, "/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid customer ID", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteCustomer(r.Context(), id)
	if err != nil {
		http.Error(w, "Failed to delete customer: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// BulkDeleteCustomers handles POST /api/customers/bulk-delete.
// @Summary Bulk delete customers
// @Description Delete multiple customers by their IDs.
// @Tags customers
// @Security Bearer
// @Accept json
// @Produce json
// @Param body body object true "List of IDs"
// @Success 200 {object} map[string]string
// @Router /customers/bulk-delete [post]
func (h *CustomerHandler) BulkDeleteCustomers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		IDs []int64 `json:"ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.service.BulkDeleteCustomers(r.Context(), req.IDs)
	if err != nil {
		http.Error(w, "Failed to delete customers: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Bulk deletion completed"})
}

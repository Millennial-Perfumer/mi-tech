package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"mi-tech/internal/service"
)

// OrderHandler is a thin HTTP adapter for order operations.
type OrderHandler struct {
	orderService   *service.OrderService
	invoiceService *service.InvoiceService
}

// NewOrderHandler creates a new OrderHandler.
func NewOrderHandler(orderService *service.OrderService, invoiceService *service.InvoiceService) *OrderHandler {
	return &OrderHandler{
		orderService:   orderService,
		invoiceService: invoiceService,
	}
}

// GetOrders handles GET /api/orders with pagination and date filters.
// @Summary List orders
// @Description Get a paginated list of orders with filters
// @Tags orders
// @Accept json
// @Produce json
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Param search query string false "Search term"
// @Param source query string false "Order source"
// @Param financial_status query string false "Financial status"
// @Param fulfillment_status query string false "Fulfillment status"
// @Param sort_by query string false "Sort by field"
// @Param sort_order query string false "Sort order (asc/desc)"
// @Success 200 {object} map[string]interface{}
// @Router /orders [get]
func (h *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	search := r.URL.Query().Get("search")
	source := r.URL.Query().Get("source")
	finStatus := r.URL.Query().Get("financial_status")
	fulStatus := r.URL.Query().Get("fulfillment_status")
	sortBy := r.URL.Query().Get("sort_by")
	sortOrder := r.URL.Query().Get("sort_order")

	orders, totalCount, err := h.orderService.ListOrders(startDate, endDate, page, limit, search, source, finStatus, fulStatus, sortBy, sortOrder)
	if err != nil {
		http.Error(w, "Failed to retrieve orders", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"orders":      orders,
		"total_count": totalCount,
		"page":        page,
		"limit":       limit,
	})
}

// UpdateStatusRequest represents the body for status update.
type UpdateStatusRequest struct {
	Status string `json:"status"`
}

// UpdateOrderStatus handles PUT /api/orders/status.
// @Summary Update order status
// @Description Update the internal status of an order
// @Tags orders
// @Accept json
// @Produce json
// @Param id query string true "Order ID (Internal or External)"
// @Param body body UpdateStatusRequest true "New status"
// @Success 200 {object} map[string]interface{}
// @Router /orders/status [put]
func (h *OrderHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodOptions {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing order id", http.StatusBadRequest)
		return
	}

	var reqBody UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	order, err := h.orderService.GetOrderFlexible(id)
	if err != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	rowsAffected, err := h.orderService.UpdateOrderStatus(order.ID, reqBody.Status)
	if err != nil {
		http.Error(w, "Failed to update database", http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Status updated successfully",
	})
}

// GenerateInvoice handles GET /api/orders/invoice and streams a PDF.
// @Summary Generate invoice PDF
// @Description Generate and download a GST invoice PDF for an order
// @Tags orders
// @Produce application/pdf
// @Param id query string true "Order ID (Internal or External)"
// @Success 200 {file} file
// @Router /orders/invoice [get]
func (h *OrderHandler) GenerateInvoice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Missing order id", http.StatusBadRequest)
		return
	}

	// Fetch order entity via service
	order, err := h.orderService.GetOrderFlexible(idStr)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Order not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	items, err := h.orderService.GetLineItems(order.ID)
	if err != nil {
		http.Error(w, "Database error fetching items", http.StatusInternalServerError)
		return
	}

	// Set response headers and generate PDF
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=invoice-%s.pdf", order.OrderNumber))

	if err := h.invoiceService.GeneratePDF(order, items, w); err != nil {
		http.Error(w, "Failed to generate PDF", http.StatusInternalServerError)
	}
}

package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"shopify-gst-app/internal/service"
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

// UpdateOrderStatus handles PUT /api/orders/status.
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

	var reqBody struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	rowsAffected, err := h.orderService.UpdateOrderStatus(id, reqBody.Status)
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
func (h *OrderHandler) GenerateInvoice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing order id", http.StatusBadRequest)
		return
	}

	// Fetch order entity + line items via service
	order, err := h.orderService.GetOrderEntity(id)
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

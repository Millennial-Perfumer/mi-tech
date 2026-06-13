package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	communicationServicePkg "mi-tech/internal/domain/communication/service"
	"mi-tech/internal/domain/order/dto"
	"mi-tech/internal/domain/order/entity"
	"mi-tech/internal/domain/order/service"
	"mi-tech/internal/domain/shared/config"
	legacyService "mi-tech/internal/service"
)

// OrderHandler is a thin HTTP adapter for order operations.
type OrderHandler struct {
	orderService   *service.OrderService
	invoiceService *legacyService.InvoiceService
	mappingService *communicationServicePkg.WebhookMappingService
}

// NewOrderHandler creates a new OrderHandler.
func NewOrderHandler(orderService *service.OrderService, invoiceService *legacyService.InvoiceService, mappingService *communicationServicePkg.WebhookMappingService) *OrderHandler {
	return &OrderHandler{
		orderService:   orderService,
		invoiceService: invoiceService,
		mappingService: mappingService,
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
	status := r.URL.Query().Get("status")
	sortBy := r.URL.Query().Get("sort_by")
	sortOrder := r.URL.Query().Get("sort_order")

	orders, totalCount, err := h.orderService.ListOrders(startDate, endDate, page, limit, search, source, finStatus, fulStatus, status, sortBy, sortOrder)
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

	// Use UpdateOrderStatusWithEntity to avoid redundant DB lookup if status is 'cancelled'
	rowsAffected, err := h.orderService.UpdateOrderStatusWithEntity(order.ID, reqBody.Status, &order)
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

// UpdatePaymentStatus handles PUT /api/orders/payment-status.
// @Summary Update order payment status
// @Description Update the financial status of an order manually
// @Tags orders
// @Accept json
// @Produce json
// @Param id query string true "Order ID"
// @Param body body UpdateStatusRequest true "New payment status"
// @Success 200 {object} map[string]interface{}
// @Router /orders/payment-status [put]
func (h *OrderHandler) UpdatePaymentStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodOptions {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Missing order id", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid order id", http.StatusBadRequest)
		return
	}

	var reqBody UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.orderService.UpdateOrderPaymentStatus(id, reqBody.Status); err != nil {
		http.Error(w, "Failed to update payment status: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Payment status updated successfully",
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

	// Fetch line items - note: we could optimize this further by including items in GetOrderFlexible
	// but GenerateInvoice requires items separately to pass to the invoice service.
	items, err := h.orderService.GetLineItems(order.ID)
	if err != nil {
		http.Error(w, "Database error fetching items", http.StatusInternalServerError)
		return
	}

	// Set response headers and generate PDF
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=invoice-%s.pdf", order.OrderNumber))

	// Note: We need to cast domain/order/entity.Order and LineItem to the legacy types that InvoiceService expects,
	// OR we can wait. Wait, does InvoiceService import legacy entity.Order and LineItem?
	// Yes, legacy InvoiceService uses legacy entity.Order and legacy entity.LineItem.
	// Since order and items are of domain types, we need to convert them or update InvoiceService imports.
	// Wait, we will update InvoiceService imports and all other consumers in the next step!
	// Yes, when we update imports of mi-tech/internal/entity, we will redirect it to domain/order/entity.
	// Thus, InvoiceService will naturally accept our new domain/order/entity types! This is perfect.
	if err := h.invoiceService.GeneratePDF(order, items, w); err != nil {
		http.Error(w, "Failed to generate PDF", http.StatusInternalServerError)
	}
}

// GetSources returns all unique order sources.
// @Summary List order sources
// @Description Retrieve a list of all unique platforms from which orders originated (e.g., Shopify, Amazon, POS).
// @Tags orders
// @Security Bearer
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /orders/sources [get]
func (h *OrderHandler) GetSources(w http.ResponseWriter, r *http.Request) {
	sources, err := h.orderService.ListSources()
	if err != nil {
		log.Printf("Error fetching sources: %v", err)
		http.Error(w, "Failed to fetch sources", http.StatusInternalServerError)
		return
	}

	log.Printf("GetSources: returning %d sources", len(sources))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"sources": sources,
	})
}

// GetOrder handles GET /api/orders/:id or GET /api/orders?id=
// GetOrder handles GET /api/orders/{id}.
// @Summary Get order details
// @Description Retrieve full details for a specific order, including line items.
// @Tags orders
// @Security Bearer
// @Produce json
// @Param id query string true "Order ID (Internal or External)"
// @Success 200 {object} map[string]interface{}
// @Router /orders/detail [get]
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		// Try to extract from path if using /api/orders/{id}
		idStr = strings.TrimPrefix(r.URL.Path, "/api/orders/")
		idStr = strings.TrimSuffix(idStr, "/")
	}

	if idStr == "" || idStr == "orders" {
		http.Error(w, "Missing order id", http.StatusBadRequest)
		return
	}

	// Consistently handle both internal and external IDs in a single service call
	// This reduces database lookups by avoiding separate resolution and retrieval steps
	orderResp, err := h.orderService.GetOrderResponseFlexible(idStr)
	if err != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"order":   orderResp,
	})
}

// UpdateOrder handles PUT /api/orders?id=
// UpdateOrder handles PUT /api/orders/{id}.
// @Summary Update order
// @Description Update order details and synchronize improvements back to Shopify.
// @Tags orders
// @Security Bearer
// @Accept json
// @Produce json
// @Param id query string true "Order ID"
// @Param body body dto.OrderUpdateRequest true "Updated order data"
// @Success 200 {object} map[string]interface{}
// @Router /orders [put]
func (h *OrderHandler) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		idStr = strings.TrimPrefix(r.URL.Path, "/api/orders/")
		idStr = strings.TrimSuffix(idStr, "/")
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid order id", http.StatusBadRequest)
		return
	}

	var req dto.OrderUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.orderService.UpdateOrder(id, req); err != nil {
		log.Printf("Error updating order %d: %v", id, err)
		http.Error(w, "Failed to update order: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Order updated and synchronized with Shopify successfully",
	})
}

// MarkAsDelivered handles PUT /api/orders/delivered.
// @Summary Mark order as delivered
// @Description Set delivery_status = 'delivered' and stamp delivered_at time. Trigger immediate WhatsApp notification.
// @Tags orders
// @Security Bearer
// @Produce json
// @Param id query string true "Order ID"
// @Success 200 {object} map[string]interface{}
// @Router /orders/delivered [put]
func (h *OrderHandler) MarkAsDelivered(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPost && r.Method != http.MethodOptions {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	var id int64
	var err error
	var order entity.Order

	if idStr == "" && r.Method == http.MethodPost {
		var body struct {
			ID int64 `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
			id = body.ID
		}
	}

	if id == 0 && idStr != "" {
		id, err = strconv.ParseInt(idStr, 10, 64)
	}

	if id == 0 && idStr != "" && err != nil {
		// Try flexible look up if it's an external ID
		order, err = h.orderService.GetOrderFlexible(idStr)
		if err != nil {
			http.Error(w, "Order not found", http.StatusNotFound)
			return
		}
		id = order.ID
	}

	if id == 0 {
		http.Error(w, "Missing or invalid order id", http.StatusBadRequest)
		return
	}

	if err := h.orderService.MarkAsDelivered(id); err != nil {
		log.Printf("Error marking order %d as delivered: %v", id, err)
		http.Error(w, "Failed to update delivery status", http.StatusInternalServerError)
		return
	}

	// Trigger "Order Delivered" WhatsApp notification
	// We already have the order entity from the flexible lookup above if it was an external ID,
	// or we can fetch it once here. The previous code was fetching it even if 'id' was already known.
	// To minimize lookups, we use the 'order' variable which was either populated by GetOrderFlexible
	// or remains empty if id was parsed from int.
	var orderToNotify entity.Order
	if order.ID != 0 {
		orderToNotify = order
	} else {
		orderToNotify, _ = h.orderService.GetOrderEntity(id)
	}

	if orderToNotify.ID != 0 {
		// Use "orders/delivered" topic as per existing conventions in WebhookMappingService
		go func() {
			// Use standardized Store ID
			if err := h.mappingService.ExecuteMapping(config.StoreIDShopify, "orders/delivered", orderToNotify); err != nil {
				log.Printf("Failed to send delivery notification for order %d: %v", id, err)
			}
		}()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Order marked as delivered. Customer notification triggered.",
	})
}

// CreateOrder handles POST /api/orders.
// @Summary Create a manual/POS order
// @Description Create a new POS order directly in the system. Triggers inventory deduction and external synchronization.
// @Tags orders
// @Security Bearer
// @Accept json
// @Produce json
// @Param order body dto.OrderCreateRequest true "Order Details"
// @Success 201 {object} dto.OrderResponse
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /orders [post]
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req dto.OrderCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Basic validation
	if len(req.LineItems) == 0 {
		http.Error(w, "At least one line item is required", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.CustomerName) == "" {
		http.Error(w, "Customer name is required", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.CustomerPhone) == "" {
		http.Error(w, "Customer phone number is required", http.StatusBadRequest)
		return
	}

	orderResponse, err := h.orderService.CreateManualOrder(r.Context(), req)
	if err != nil {
		log.Printf("Error creating manual order: %v", err)
		http.Error(w, "Failed to create order: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(orderResponse)
}

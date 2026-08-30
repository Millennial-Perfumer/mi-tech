package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"mi-tech/internal/domain/order/repository"
)

// HistoryHandler exposes immutable order and customer audit timelines.
// Current-state reads remain on the existing orders/customers endpoints.
type HistoryHandler struct {
	events repository.EventRepository
}

func NewHistoryHandler(events repository.EventRepository) *HistoryHandler {
	return &HistoryHandler{events: events}
}

// GetOrderHistory handles GET /api/orders/history. Use id for one order or
// search to find historical values such as an AWB that is no longer current.
func (h *HistoryHandler) GetOrderHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	q := r.URL.Query()
	filter := repository.EventFilter{
		ExternalOrderID: q.Get("external_order_id"),
		Search:          q.Get("search"),
		EventType:       q.Get("event_type"),
		StartDate:       q.Get("start_date"),
		EndDate:         q.Get("end_date"),
		Page:            parsePositiveInt(q.Get("page")),
		Limit:           parsePositiveInt(q.Get("limit")),
	}
	filter.OrderID = parseInt64(q.Get("id"))
	if filter.OrderID == 0 {
		filter.OrderID = parseInt64(q.Get("order_id"))
	}

	events, total, err := h.events.ListOrderEvents(r.Context(), filter)
	if err != nil {
		http.Error(w, "Failed to retrieve order history", http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]any{
		"success":      true,
		"order_events": events,
		"total_count":  total,
		"page":         normalizedPage(filter.Page),
		"limit":        normalizedLimit(filter.Limit),
	})
}

// GetCustomerHistory handles GET /api/customers/history. It supports a
// customer ID, order context, phone number, or historical value search.
func (h *HistoryHandler) GetCustomerHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	q := r.URL.Query()
	filter := repository.EventFilter{
		CustomerID:    parseInt64(q.Get("id")),
		CustomerPhone: q.Get("phone"),
		Search:        q.Get("search"),
		EventType:     q.Get("event_type"),
		StartDate:     q.Get("start_date"),
		EndDate:       q.Get("end_date"),
		Page:          parsePositiveInt(q.Get("page")),
		Limit:         parsePositiveInt(q.Get("limit")),
	}
	if filter.CustomerID == 0 {
		filter.CustomerID = parseInt64(q.Get("customer_id"))
	}
	filter.OrderID = parseInt64(q.Get("order_id"))

	events, total, err := h.events.ListCustomerEvents(r.Context(), filter)
	if err != nil {
		http.Error(w, "Failed to retrieve customer history", http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]any{
		"success":         true,
		"customer_events": events,
		"total_count":     total,
		"page":            normalizedPage(filter.Page),
		"limit":           normalizedLimit(filter.Limit),
	})
}

func parseInt64(value string) int64 {
	parsed, _ := strconv.ParseInt(value, 10, 64)
	return parsed
}

func parsePositiveInt(value string) int {
	if value == "" {
		return 0
	}
	parsed, _ := strconv.Atoi(value)
	if parsed < 1 {
		return 0
	}
	return parsed
}

func normalizedPage(value int) int {
	if value < 1 {
		return 1
	}
	return value
}

func normalizedLimit(value int) int {
	if value < 1 || value > 100 {
		return 50
	}
	return value
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}

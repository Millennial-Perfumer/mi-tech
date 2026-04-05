package handler

import (
	"net/http"
	"mi-tech/internal/repository"
	"strings"
)

type RedirectHandler struct {
	orderRepo repository.OrderRepository
}

func NewRedirectHandler(orderRepo repository.OrderRepository) *RedirectHandler {
	return &RedirectHandler{orderRepo: orderRepo}
}

// RedirectTracking handles GET /t/{id}.
// @Summary Order tracking redirect
// @Description Redirects the user to the carrier's tracking page based on the order ID.
// @Tags system
// @Param id path string true "Order ID or Number"
// @Success 307 {string} string "Redirect"
// @Router /t/{id} [get]
func (h *RedirectHandler) RedirectTracking(w http.ResponseWriter, r *http.Request) {
	// Expected URL: /t/ORDER_ID
	id := strings.TrimPrefix(r.URL.Path, "/t/")
	if id == "" {
		http.Error(w, "Order ID required", http.StatusBadRequest)
		return
	}

	order, err := h.orderRepo.GetByFlexibleID(id)
	if err != nil {
		http.Redirect(w, r, "https://millennialperfumer.com", http.StatusTemporaryRedirect)
		return
	}

	trackingURL := ""
	if order.TrackingUrl != nil {
		trackingURL = *order.TrackingUrl
	}

	if trackingURL == "" {
		// Fallback to Shopify order page if no tracking URL exists
		http.Redirect(w, r, "https://millennialperfumer.com", http.StatusTemporaryRedirect)
		return
	}

	// Double check protocol
	if !strings.HasPrefix(trackingURL, "http") {
		trackingURL = "https://" + trackingURL
	}

	http.Redirect(w, r, trackingURL, http.StatusTemporaryRedirect)
}

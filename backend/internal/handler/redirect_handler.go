package handler

import (
	"mi-tech/internal/repository"
	"net/http"
	"net/url"
	"strings"
)

// allowedTrackingDomains contains a list of allowed root domains for courier tracking links
// to prevent Open Redirect vulnerabilities while supporting major providers.
var allowedTrackingDomains = map[string]bool{
	"tracking.com":     true,
	"delhivery.com":    true,
	"shiprocket.in":    true,
	"bluedart.com":     true,
	"dhl.com":          true,
	"fedex.com":        true,
	"ups.com":          true,
	"usps.com":         true,
	"amazon.in":        true,
	"amazon.com":       true,
	"ecomexpress.in":   true,
	"xpressbees.com":   true,
	"indiapost.gov.in": true,
}

// isAllowedDomain checks if the parsed hostname matches an allowed domain or its subdomains.
func isAllowedDomain(hostname string) bool {
	hostname = strings.ToLower(hostname)
	if allowedTrackingDomains[hostname] {
		return true
	}

	// Check for subdomains (e.g., track.delhivery.com)
	for domain := range allowedTrackingDomains {
		if strings.HasSuffix(hostname, "."+domain) {
			return true
		}
	}
	return false
}

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

	// Validate tracking URL domain to prevent Open Redirect
	parsedURL, err := url.Parse(trackingURL)
	if err != nil || !isAllowedDomain(parsedURL.Hostname()) {
		http.Redirect(w, r, "https://millennialperfumer.com", http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, trackingURL, http.StatusTemporaryRedirect)
}

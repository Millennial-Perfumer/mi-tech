package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"mi-tech/internal/shared/extclient/amazon"
	"net/http"
	"strconv"
	"strings"
)

// AmazonHandler serves read-only Amazon integration endpoints.
type AmazonHandler struct {
	client listingsClient
}

type listingsClient interface {
	SearchListingsItems(amazon.ListingsItemsQuery) (amazon.ListingsItemsPage, error)
}

func NewAmazonHandler(client listingsClient) *AmazonHandler {
	return &AmazonHandler{client: client}
}

type listingsResponse struct {
	NumberOfResults int                      `json:"number_of_results"`
	Items           []map[string]interface{} `json:"items"`
	NextToken       string                   `json:"next_token,omitempty"`
}

// ListListings returns a live page of Amazon listings and the exact count for
// the current filter. Use next_token as page_token to fetch another page.
func (h *AmazonHandler) ListListings(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	pageSize, err := parsePageSize(query.Get("page_size"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	page, err := h.client.SearchListingsItems(amazon.ListingsItemsQuery{
		PageToken:    query.Get("page_token"),
		PageSize:     pageSize,
		IncludedData: splitCSV(query.Get("included_data")),
		IssueLocale:  query.Get("issue_locale"),
	})
	if err != nil {
		log.Printf("AmazonHandler.ListListings error: %v", err)
		http.Error(w, "Amazon listings unavailable", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(listingsResponse{
		NumberOfResults: page.NumberOfResults,
		Items:           page.Items,
		NextToken:       page.Pagination.NextToken,
	})
}

func parsePageSize(value string) (int, error) {
	if value == "" {
		return 0, nil
	}

	pageSize, err := strconv.Atoi(value)
	if err != nil || pageSize < 1 || pageSize > 20 {
		return 0, fmt.Errorf("page_size must be between 1 and 20")
	}
	return pageSize, nil
}

func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			items = append(items, part)
		}
	}
	return items
}

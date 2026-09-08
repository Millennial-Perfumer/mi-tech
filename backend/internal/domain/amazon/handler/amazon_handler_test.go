package handler

import (
	"encoding/json"
	"mi-tech/internal/shared/extclient/amazon"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockListingsClient struct {
	query amazon.ListingsItemsQuery
}

func (m *mockListingsClient) SearchListingsItems(query amazon.ListingsItemsQuery) (amazon.ListingsItemsPage, error) {
	m.query = query
	var page amazon.ListingsItemsPage
	page.NumberOfResults = 37
	page.Pagination.NextToken = "next-page"
	page.Items = []map[string]interface{}{{"sku": "sku-1"}}
	return page, nil
}

func TestListListings(t *testing.T) {
	client := &mockListingsClient{}
	handler := NewAmazonHandler(client)
	req := httptest.NewRequest(http.MethodGet, "/api/amazon/listings?page_token=previous-page&page_size=20&included_data=summaries%2Cattributes&issue_locale=en_IN", nil)
	recorder := httptest.NewRecorder()

	handler.ListListings(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if client.query.PageToken != "previous-page" || client.query.PageSize != 20 || client.query.IssueLocale != "en_IN" {
		t.Fatalf("unexpected client query: %+v", client.query)
	}
	if len(client.query.IncludedData) != 2 || client.query.IncludedData[1] != "attributes" {
		t.Fatalf("unexpected included data: %+v", client.query.IncludedData)
	}

	var response struct {
		NumberOfResults int                      `json:"number_of_results"`
		Items           []map[string]interface{} `json:"items"`
		NextToken       string                   `json:"next_token"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.NumberOfResults != 37 || response.NextToken != "next-page" || len(response.Items) != 1 {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func TestListListingsRejectsInvalidPageSize(t *testing.T) {
	handler := NewAmazonHandler(&mockListingsClient{})
	req := httptest.NewRequest(http.MethodGet, "/api/amazon/listings?page_size=21", nil)
	recorder := httptest.NewRecorder()

	handler.ListListings(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}

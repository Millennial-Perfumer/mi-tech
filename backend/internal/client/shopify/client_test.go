package shopify

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"mi-tech/internal/config"
	"mi-tech/internal/dto"
	"mi-tech/internal/repository"
)

func setupTestClient(b *testing.B, apiURL string) *Client {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		b.Fatalf("failed to open sqlite: %v", err)
	}
	db.AutoMigrate(&repository.AppConfig{})

	repo := repository.NewConfigsRepository(db)
	repo.Set("shopify_store_url", "mock.myshopify.com")
	repo.Set("shopify_access_token", "test-token")
	repo.Set("shopify_api_version", "2024-01")

	settings := config.NewSettingsProvider(repo)
	client := NewClient(settings)
	client.testURL = apiURL
	return client
}

type mockShopifyHandler struct {
	count    int32
	maxPages int32
}

func (h *mockShopifyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	current := atomic.AddInt32(&h.count, 1)

	resp := dto.GraphQLOrderResponse{}
	if current < h.maxPages {
		resp.Data.Orders.PageInfo.HasNextPage = true
		resp.Data.Orders.PageInfo.EndCursor = "cursor"
	} else {
		resp.Data.Orders.PageInfo.HasNextPage = false
	}

	// Add mock data to simulate processing overhead
	resp.Data.Orders.Edges = []dto.GraphQLOrderEdge{
		{Node: dto.GraphQLOrderNode{ID: "gid://shopify/Order/1", Name: "#1001"}},
		{Node: dto.GraphQLOrderNode{ID: "gid://shopify/Order/2", Name: "#1002"}},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func BenchmarkFetchOrdersOptimized(b *testing.B) {
	handler := &mockShopifyHandler{maxPages: 5}
	server := httptest.NewServer(handler)
	defer server.Close()

	client := setupTestClient(b, server.URL)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		atomic.StoreInt32(&handler.count, 0)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		_, err := client.FetchOrders(ctx, time.Now().Add(-24*time.Hour), time.Now())
		cancel()
		if err != nil {
			b.Fatalf("FetchOrders failed: %v", err)
		}
	}
}

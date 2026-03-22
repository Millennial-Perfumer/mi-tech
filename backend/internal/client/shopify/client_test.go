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

func BenchmarkFetchOrdersOptimized(b *testing.B) {
	var count int32
	maxPages := int32(5)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := atomic.AddInt32(&count, 1)

		resp := dto.GraphQLOrderResponse{}
		if current < maxPages {
			resp.Data.Orders.PageInfo.HasNextPage = true
			resp.Data.Orders.PageInfo.EndCursor = "cursor"
		} else {
			resp.Data.Orders.PageInfo.HasNextPage = false
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := setupTestClient(b, server.URL)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		atomic.StoreInt32(&count, 0)
		_, err := client.FetchOrders(context.Background(), time.Now().Add(-24*time.Hour), time.Now())
		if err != nil {
			b.Fatalf("FetchOrders failed: %v", err)
		}
	}
}

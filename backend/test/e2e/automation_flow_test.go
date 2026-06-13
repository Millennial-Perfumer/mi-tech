package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	communicationEntity "mi-tech/internal/domain/communication/entity"
	communicationRepoPkg "mi-tech/internal/domain/communication/repository"
	communicationServicePkg "mi-tech/internal/domain/communication/service"
	orderDto "mi-tech/internal/domain/order/dto"
	orderRepoPkg "mi-tech/internal/domain/order/repository"
	orderServicePkg "mi-tech/internal/domain/order/service"
	webhookHandlerPkg "mi-tech/internal/domain/webhook/handler"
	webhookRepoPkg "mi-tech/internal/domain/webhook/repository"
	webhookServicePkg "mi-tech/internal/domain/webhook/service"
	"mi-tech/internal/shared/config"
	configRepoPkg "mi-tech/internal/shared/config/repository"
	"mi-tech/internal/shared/extclient/shopify"
	"mi-tech/internal/shared/testutil"

	"github.com/stretchr/testify/assert"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestEndToEnd_OrderCreationAutomation(t *testing.T) {
	db, err := testutil.SetupTestDB()
	if err != nil {
		t.Logf("Database connection error: %v", err)
		t.Skip("E2E skipping: Database not available")
	}
	defer testutil.CleanupTestDB(db)

	sqlDB, _ := db.DB()

	// 0. Setup Mock Meta Server
	mockMsgID := fmt.Sprintf("wa_test_id_%d_%d", time.Now().UnixNano(), rand.Intn(1000))
	mockMetaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Respond with success to ensure transaction commits
		fmt.Fprintf(w, `{"messages":[{"id":"%s"}]}`, mockMsgID)
	}))
	defer mockMetaServer.Close()

	// Intercept graph.facebook.com
	originalTransport := http.DefaultClient.Transport
	http.DefaultClient.Transport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Host == "graph.facebook.com" {
			u, _ := url.Parse(mockMetaServer.URL)
			req.URL.Scheme = u.Scheme
			req.URL.Host = u.Host
		}
		return http.DefaultTransport.RoundTrip(req)
	})
	defer func() { http.DefaultClient.Transport = originalTransport }()

	// 1. Setup repos
	orderRepo := orderRepoPkg.NewOrderRepository(db)
	lineItemRepo := orderRepoPkg.NewLineItemRepository(db)
	customerRepo := orderRepoPkg.NewCustomerRepository(db)
	webhookEventRepo := webhookRepoPkg.NewWebhookEventRepository(db)
	webhookStatusRepo := webhookRepoPkg.NewWebhookStatusRepository(db)
	configsRepo := configRepoPkg.NewConfigsRepository(db)
	settingsRepo := configRepoPkg.NewSettingsRepository(db)
	templatesRepo := communicationRepoPkg.NewTemplatesRepository(sqlDB)
	messagesRepo := communicationRepoPkg.NewMessagesRepository(sqlDB)

	// 2. Setup services
	settingsProvider := config.NewSettingsProvider(configsRepo)
	customerService := orderServicePkg.NewCustomerService(customerRepo, orderRepo, nil)
	orderService := orderServicePkg.NewOrderService(orderRepo, lineItemRepo, customerService, nil, nil)
	invoiceService := orderServicePkg.NewInvoiceService(settingsRepo)

	messagesService := communicationServicePkg.NewMessagesService(messagesRepo, settingsProvider, customerRepo, nil)
	mappingService := communicationServicePkg.NewWebhookMappingService(templatesRepo, messagesService, invoiceService, settingsRepo, lineItemRepo, settingsProvider, orderRepo)

	shopifyClient := shopify.NewClient(settingsProvider)
	webhookService := webhookServicePkg.NewWebhookService(orderService, shopifyClient, webhookEventRepo, webhookStatusRepo)

	webhookHandler := webhookHandlerPkg.NewWebhookHandler(webhookService, mappingService, settingsProvider)

	// Pre-cleanup: Delete existing triggers for orders/create to avoid conflicts
	db.Exec("DELETE FROM automation_triggers WHERE webhook_topic = 'orders/create'")

	// 2. Seed a template and a trigger for orders/create
	templateName := fmt.Sprintf("order_conf_%d", rand.Intn(1000))
	templateID, err := templatesRepo.SaveTemplate(communicationEntity.AutomationTemplate{
		StoreID:      config.StoreIDShopify,
		TemplateName: templateName,
		Body:         "Hello {{1}}, your order {{2}} is confirmed!",
		Language:     "en",
		Category:     "UTILITY",
		Status:       "APPROVED",
	})
	assert.NoError(t, err)

	err = templatesRepo.SaveTrigger(communicationEntity.Trigger{
		StoreID:      config.StoreIDShopify,
		WebhookTopic: "orders/create",
		TemplateID:   templateID,
		Enabled:      true,
	})
	assert.NoError(t, err)

	db.Exec("INSERT INTO webhook_status (id, topic, status, last_received) VALUES (1, 'initial', 'idle', NOW()) ON CONFLICT DO NOTHING")

	// 3. Construct Shopify Webhook Payload
	orderID := int64(rand.Intn(1000000) + 30000)
	extIDStr := fmt.Sprintf("%d", orderID)
	deliveryID := fmt.Sprintf("del_%d", time.Now().UnixNano())

	payload := orderDto.ShopifyWebhookOrder{
		ID:          orderID,
		OrderNumber: orderID,
		Name:        fmt.Sprintf("#%d", orderID),
		TotalPrice:  "150.00",
		Currency:    "INR",
		CreatedAt:   time.Now().Format(time.RFC3339),
		Customer: &orderDto.ShopifyCustomer{
			ID:        orderID + 10,
			FirstName: "John",
			LastName:  "Doe",
			Phone:     "+919876543210",
		},
		LineItems: []orderDto.ShopifyLineItem{
			{ID: orderID + 20, Title: "Perfume", Quantity: 1, Price: "150.00"},
		},
	}
	body, _ := json.Marshal(payload)

	// 4. Call Webhook Handler
	req := httptest.NewRequest("POST", "/api/webhooks/shopify", bytes.NewBuffer(body))
	req.Header.Set("X-Shopify-Topic", "orders/create")
	req.Header.Set("X-Shopify-Webhook-Id", deliveryID)

	w := httptest.NewRecorder()
	webhookHandler.ShopifyWebhookHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// 5. Wait for async processing
	time.Sleep(5 * time.Second)

	// 6. Verify order was created
	order, err := orderRepo.GetByExternalID(extIDStr)
	assert.NoError(t, err)
	assert.NotEqual(t, int64(0), order.ID, "Order ID should not be zero")
	assert.Equal(t, fmt.Sprintf("#%d", orderID), order.OrderNumber)

	// 7. Verify automation message was recorded as 'sent'
	messages, err := messagesRepo.GetMessagesByOrderID(order.ID)
	assert.NoError(t, err)

	if len(messages) == 0 {
		t.Errorf("No automation messages found for Order ID %d (External %s). Trigger may have failed silently.", order.ID, extIDStr)
	}

	assert.GreaterOrEqual(t, len(messages), 1, "Automation message should have been recorded for the new order")

	found := false
	for _, m := range messages {
		if m.TemplateID == templateID {
			found = true
			assert.Equal(t, "sent", m.Status)
			assert.Equal(t, mockMsgID, m.MessageID)
			break
		}
	}
	assert.True(t, found, "Expected automation message for our seeded template (ID: %d)", templateID)
}

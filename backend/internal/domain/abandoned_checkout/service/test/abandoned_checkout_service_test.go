package test

import (
	"context"
	"errors"
	dtoPkg "mi-tech/internal/domain/abandoned_checkout/dto"
	entityPkg "mi-tech/internal/domain/abandoned_checkout/entity"
	mapperPkg "mi-tech/internal/domain/abandoned_checkout/mapper"
	servicePkg "mi-tech/internal/domain/abandoned_checkout/service"
	communicationEntity "mi-tech/internal/domain/communication/entity"
	configPkg "mi-tech/internal/shared/config"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAbandonedCheckoutRepository struct {
	mock.Mock
}

func (m *MockAbandonedCheckoutRepository) Upsert(ctx context.Context, ac *entityPkg.AbandonedCheckout) error {
	args := m.Called(ctx, ac)
	return args.Error(0)
}

func (m *MockAbandonedCheckoutRepository) MarkCompleted(ctx context.Context, checkoutToken, checkoutID, orderID string) error {
	args := m.Called(ctx, checkoutToken, checkoutID, orderID)
	return args.Error(0)
}

func (m *MockAbandonedCheckoutRepository) GetPendingForRecovery(ctx context.Context, threshold time.Time, limit int) ([]entityPkg.AbandonedCheckout, error) {
	args := m.Called(ctx, threshold, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entityPkg.AbandonedCheckout), args.Error(1)
}

func (m *MockAbandonedCheckoutRepository) UpdateRecoveryStatus(ctx context.Context, id int, status string, attempts int, lastError string, sentAt *time.Time) error {
	args := m.Called(ctx, id, status, attempts, lastError, sentAt)
	return args.Error(0)
}

func (m *MockAbandonedCheckoutRepository) List(ctx context.Context, storeID string, page, limit int, search, status, startDate, endDate string) ([]entityPkg.AbandonedCheckout, int64, error) {
	args := m.Called(ctx, storeID, page, limit, search, status, startDate, endDate)
	var checkouts []entityPkg.AbandonedCheckout
	if args.Get(0) != nil {
		checkouts = args.Get(0).([]entityPkg.AbandonedCheckout)
	}
	return checkouts, int64(args.Int(1)), args.Error(2)
}

func (m *MockAbandonedCheckoutRepository) GetByID(ctx context.Context, storeID string, id int) (*entityPkg.AbandonedCheckout, error) {
	args := m.Called(ctx, storeID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entityPkg.AbandonedCheckout), args.Error(1)
}

func (m *MockAbandonedCheckoutRepository) GetAnalytics(ctx context.Context, storeID string, startDate, endDate string) (*dtoPkg.AbandonedCheckoutAnalyticsResponse, error) {
	args := m.Called(ctx, storeID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dtoPkg.AbandonedCheckoutAnalyticsResponse), args.Error(1)
}

func (m *MockAbandonedCheckoutRepository) CheckRecentOrders(ctx context.Context, phone, email string, since time.Time) (bool, error) {
	args := m.Called(ctx, phone, email, since)
	return args.Bool(0), args.Error(1)
}

func (m *MockAbandonedCheckoutRepository) Delete(ctx context.Context, storeID string, id int) error {
	args := m.Called(ctx, storeID, id)
	return args.Error(0)
}

func (m *MockAbandonedCheckoutRepository) UpdateStatus(ctx context.Context, storeID string, id int, status string, completed bool) error {
	args := m.Called(ctx, storeID, id, status, completed)
	return args.Error(0)
}

type MockTemplatesRepository struct {
	mock.Mock
}

func (m *MockTemplatesRepository) SaveTemplate(t communicationEntity.AutomationTemplate) (int, error) {
	args := m.Called(t)
	return args.Int(0), args.Error(1)
}

func (m *MockTemplatesRepository) GetTemplates(storeID string, startDate, endDate *time.Time) ([]communicationEntity.AutomationTemplate, error) {
	args := m.Called(storeID, startDate, endDate)
	return args.Get(0).([]communicationEntity.AutomationTemplate), args.Error(1)
}

func (m *MockTemplatesRepository) UpdateStatus(templateName, status string) error {
	args := m.Called(templateName, status)
	return args.Error(0)
}

func (m *MockTemplatesRepository) SaveTrigger(tr communicationEntity.Trigger) error {
	args := m.Called(tr)
	return args.Error(0)
}

func (m *MockTemplatesRepository) GetTriggerByTopic(storeID, topic string) (*communicationEntity.Trigger, error) {
	args := m.Called(storeID, topic)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*communicationEntity.Trigger), args.Error(1)
}

func (m *MockTemplatesRepository) GetTemplateByName(storeID, name string) (*communicationEntity.AutomationTemplate, error) {
	args := m.Called(storeID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*communicationEntity.AutomationTemplate), args.Error(1)
}

func (m *MockTemplatesRepository) GetTriggers(storeID string) ([]communicationEntity.Trigger, error) {
	args := m.Called(storeID)
	return args.Get(0).([]communicationEntity.Trigger), args.Error(1)
}

func (m *MockTemplatesRepository) GetTemplateByID(id int) (*communicationEntity.AutomationTemplate, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*communicationEntity.AutomationTemplate), args.Error(1)
}

func (m *MockTemplatesRepository) UpdateTemplate(t communicationEntity.AutomationTemplate) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *MockTemplatesRepository) DeleteTemplate(id int, storeID string) error {
	args := m.Called(id, storeID)
	return args.Error(0)
}

func (m *MockTemplatesRepository) UpdateTrigger(id int, storeID string, enabled bool) error {
	args := m.Called(id, storeID, enabled)
	return args.Error(0)
}

func (m *MockTemplatesRepository) DeleteTrigger(id int, storeID string) error {
	args := m.Called(id, storeID)
	return args.Error(0)
}

func (m *MockTemplatesRepository) DeleteTriggersByTemplateID(templateID int, storeID string) error {
	args := m.Called(templateID, storeID)
	return args.Error(0)
}

func (m *MockTemplatesRepository) UpsertMetaTemplate(t communicationEntity.AutomationTemplate) (int, error) {
	args := m.Called(t)
	return args.Int(0), args.Error(1)
}

func (m *MockTemplatesRepository) GetEvents() ([]communicationEntity.AutomationEvent, error) {
	args := m.Called()
	return args.Get(0).([]communicationEntity.AutomationEvent), args.Error(1)
}

func (m *MockTemplatesRepository) SaveEvent(e communicationEntity.AutomationEvent) error {
	args := m.Called(e)
	return args.Error(0)
}

func (m *MockTemplatesRepository) DeleteEvent(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func TestAbandonedCheckoutService_ProcessWebhook(t *testing.T) {
	mockRepo := new(MockAbandonedCheckoutRepository)
	mockTemplatesRepo := new(MockTemplatesRepository)
	dummySettings := &configPkg.SettingsProvider{}

	srv := servicePkg.NewAbandonedCheckoutService(mockRepo, mockTemplatesRepo, nil, dummySettings)

	payload := dtoPkg.ShopifyWebhookCheckout{
		ID:                   12345,
		Token:                "test_token",
		CartToken:            "test_cart",
		Email:                "buyer@example.com",
		Phone:                "+919999999999",
		TotalPrice:           "125.50",
		Currency:             "INR",
		AbandonedCheckoutURL: "http://checkout.shopify.com/12345",
		Customer: &dtoPkg.ShopifyCustomer{
			FirstName: "Test",
			LastName:  "Buyer",
		},
	}

	entity := mapperPkg.WebhookCheckoutToEntity(payload)

	mockRepo.On("Upsert", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		ac := args.Get(1).(*entityPkg.AbandonedCheckout)
		assert.Equal(t, "12345", ac.CheckoutID)
		assert.Equal(t, "test_token", ac.CheckoutToken)
		assert.Equal(t, "+919999999999", ac.Phone)
		assert.Equal(t, "Test Buyer", ac.CustomerName)
		assert.Equal(t, 125.50, ac.TotalPrice)
		assert.False(t, ac.Completed)
		assert.Equal(t, "PENDING", ac.RecoveryStatus)
	})

	err := srv.ProcessCheckoutWebhook(context.Background(), "default", entity)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAbandonedCheckoutService_MarkCompleted(t *testing.T) {
	mockRepo := new(MockAbandonedCheckoutRepository)
	mockTemplatesRepo := new(MockTemplatesRepository)
	dummySettings := &configPkg.SettingsProvider{}

	srv := servicePkg.NewAbandonedCheckoutService(mockRepo, mockTemplatesRepo, nil, dummySettings)

	mockRepo.On("MarkCompleted", mock.Anything, "test_token", "12345", "999").Return(nil)

	err := srv.MarkCheckoutCompleted(context.Background(), "test_token", "12345", "999")
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAbandonedCheckoutService_ProcessRecoveryQueue_NoTriggers(t *testing.T) {
	mockRepo := new(MockAbandonedCheckoutRepository)
	mockTemplatesRepo := new(MockTemplatesRepository)
	dummySettings := &configPkg.SettingsProvider{}

	srv := servicePkg.NewAbandonedCheckoutService(mockRepo, mockTemplatesRepo, nil, dummySettings)

	dummyCheckouts := []entityPkg.AbandonedCheckout{
		{
			ID:            1,
			StoreID:       "default",
			CheckoutID:    "12345",
			CheckoutToken: "test_token",
			Phone:         "+919999999999",
			CustomerName:  "Test Buyer",
			CheckoutURL:   "http://checkout.url",
		},
	}

	mockRepo.On("GetPendingForRecovery", mock.Anything, mock.Anything, 50).Return(dummyCheckouts, nil)
	mockRepo.On("UpdateRecoveryStatus", mock.Anything, 1, "PROCESSING", 1, "", mock.Anything).Return(nil)
	mockTemplatesRepo.On("GetTriggerByTopic", "default", "checkouts/abandoned").Return((*communicationEntity.Trigger)(nil), errors.New("no trigger"))
	mockRepo.On("UpdateRecoveryStatus", mock.Anything, 1, "CANCELLED", 0, "No active trigger configured", mock.Anything).Return(nil)

	err := srv.ProcessRecoveryQueue(context.Background())
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockTemplatesRepo.AssertExpectations(t)
}

func (m *MockTemplatesRepository) BulkUpdateStatuses(updates map[string]string) error {
	args := m.Called(updates)
	return args.Error(0)
}

package service

import (
	"mi-tech/internal/domain/feedback/dto"
	"mi-tech/internal/domain/feedback/entity"
	"mi-tech/internal/domain/feedback/repository"
	orderEntity "mi-tech/internal/domain/order/entity"
	orderServicePkg "mi-tech/internal/domain/order/service"
)

// FeedbackService orchestrates the customer feedback business logic.
type FeedbackService struct {
	repo         repository.FeedbackRepository
	orderService *orderServicePkg.OrderService
}

// NewFeedbackService constructs a new FeedbackService.
func NewFeedbackService(repo repository.FeedbackRepository, orderService *orderServicePkg.OrderService) *FeedbackService {
	return &FeedbackService{
		repo:         repo,
		orderService: orderService,
	}
}

// SaveCustomerFeedback saves a feedback record.
func (s *FeedbackService) SaveCustomerFeedback(feedback entity.CustomerFeedback) error {
	return s.repo.SaveCustomerFeedback(feedback)
}

// GetCustomerFeedback retrieves enriched customer feedback.
func (s *FeedbackService) GetCustomerFeedback() ([]dto.FeedbackResponse, error) {
	return s.repo.GetCustomerFeedback()
}

// UpdateFeedbackAdminComment updates the admin comment on a feedback.
func (s *FeedbackService) UpdateFeedbackAdminComment(id int, comment string) error {
	return s.repo.UpdateFeedbackAdminComment(id, comment)
}

// GetOrdersForFeedback delegates to OrderService.
func (s *FeedbackService) GetOrdersForFeedback(delayMinutes int) ([]orderEntity.Order, error) {
	return s.orderService.GetOrdersForFeedback(delayMinutes)
}

// UpdateFeedbackStatus delegates to OrderService.
func (s *FeedbackService) UpdateFeedbackStatus(id int64, statusID int) error {
	return s.orderService.UpdateFeedbackStatus(id, statusID)
}

// GetOrderEntity delegates to OrderService.
func (s *FeedbackService) GetOrderEntity(orderID int64) (orderEntity.Order, error) {
	return s.orderService.GetOrderEntity(orderID)
}

// ValidateFeedback delegates to OrderService.
func (s *FeedbackService) ValidateFeedback(orderID int64, phone string) (bool, error) {
	return s.orderService.ValidateFeedback(orderID, phone)
}

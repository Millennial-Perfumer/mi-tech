package service

import (
	"context"
	"fmt"
	"log"
	"strings"

	"mi-tech/internal/domain/feedback/dto"
	"mi-tech/internal/domain/feedback/entity"
	"mi-tech/internal/domain/feedback/repository"
	marketingDto "mi-tech/internal/domain/marketing/dto"
	marketingServicePkg "mi-tech/internal/domain/marketing/service"
	orderEntity "mi-tech/internal/domain/order/entity"
	orderServicePkg "mi-tech/internal/domain/order/service"
)

// FeedbackService orchestrates the customer feedback business logic.
type FeedbackService struct {
	repo           repository.FeedbackRepository
	orderService   *orderServicePkg.OrderService
	judgeMeService *marketingServicePkg.JudgeMeService
}

// NewFeedbackService constructs a new FeedbackService.
func NewFeedbackService(
	repo repository.FeedbackRepository,
	orderService *orderServicePkg.OrderService,
	judgeMeService *marketingServicePkg.JudgeMeService,
) *FeedbackService {
	return &FeedbackService{
		repo:           repo,
		orderService:   orderService,
		judgeMeService: judgeMeService,
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

// PostJudgeMeReviewsForFeedback posts Judge.me product reviews for every product in a feedback's order.
func (s *FeedbackService) PostJudgeMeReviewsForFeedback(ctx context.Context, feedbackID int, customEmail string) error {
	feedback, err := s.repo.GetFeedbackByID(feedbackID)
	if err != nil {
		return fmt.Errorf("feedback record not found: %w", err)
	}

	if feedback.JudgeMePosted {
		return fmt.Errorf("review has already been posted previously")
	}

	order, err := s.orderService.GetOrder(feedback.OrderID)
	if err != nil {
		return fmt.Errorf("failed to fetch order details for order %d: %v", feedback.OrderID, err)
	}

	if len(order.LineItems) == 0 {
		return fmt.Errorf("order %d has no product line items", feedback.OrderID)
	}

	reviewerName := order.CustomerName
	if reviewerName == "" {
		reviewerName = "Verified Customer"
	}

	reviewerEmail := strings.TrimSpace(customEmail)
	if reviewerEmail == "" {
		reviewerEmail = "hari.crze.101@gmail.com"
	}

	reviewTitle := ""
	reviewBody := feedback.Message

	var reviewsToSubmit []marketingDto.GeneratedReviewDTO
	for idx, item := range order.LineItems {
		prodID := item.ProductID
		if prodID == "" {
			prodID = item.Title
		}
		reviewsToSubmit = append(reviewsToSubmit, marketingDto.GeneratedReviewDTO{
			ID:           fmt.Sprintf("feedback_%d_item_%d", feedbackID, idx+1),
			ProductID:    prodID,
			ProductTitle: item.Title,
			ReviewerName: reviewerName,
			Email:        reviewerEmail,
			Rating:       feedback.Rating,
			Title:        reviewTitle,
			Body:         reviewBody,
			ShopDomain:   "4296fb-8e.myshopify.com",
		})
	}

	if s.judgeMeService == nil {
		return fmt.Errorf("JudgeMeService is not configured")
	}

	submitReq := marketingDto.SubmitReviewsRequest{
		Reviews: reviewsToSubmit,
		DelayMs: 500,
		DryRun:  false,
	}

	res, err := s.judgeMeService.SubmitReviews(ctx, submitReq)
	if err != nil {
		return fmt.Errorf("Judge.me review submission failed: %w", err)
	}

	if res.Successful == 0 {
		log.Printf("Warning: 0 reviews succeeded out of %d attempted for feedback %d", res.TotalProcessed, feedbackID)
		if len(res.Results) > 0 {
			return fmt.Errorf("submission failed: %s", res.Results[0].ResponseBody)
		}
		return fmt.Errorf("failed to publish reviews to Judge.me")
	}

	if err := s.repo.MarkJudgeMePosted(feedbackID); err != nil {
		log.Printf("Warning: Reviews posted but failed to mark feedback %d as posted: %v", feedbackID, err)
	}

	return nil
}

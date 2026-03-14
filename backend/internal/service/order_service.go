package service

import (
	"mi-tech/internal/dto"
	"mi-tech/internal/entity"
	"mi-tech/internal/mapper"
	"mi-tech/internal/repository"
)

// OrderService handles order CRUD business logic.
type OrderService struct {
	orderRepo    repository.OrderRepository
	lineItemRepo repository.LineItemRepository
}

// NewOrderService creates a new OrderService.
func NewOrderService(orderRepo repository.OrderRepository, lineItemRepo repository.LineItemRepository) *OrderService {
	return &OrderService{
		orderRepo:    orderRepo,
		lineItemRepo: lineItemRepo,
	}
}

// ListOrders retrieves a paginated list of orders and converts them to DTOs.
func (s *OrderService) ListOrders(startDate, endDate string, page, limit int, search, source, finStatus, fulStatus, sortBy, sortOrder string) ([]dto.OrderResponse, int, error) {
	filter := repository.OrderFilter{
		StartDate:         startDate,
		EndDate:           endDate,
		Page:              page,
		Limit:             limit,
		Search:            search,
		Source:            source,
		FinancialStatus:   finStatus,
		FulfillmentStatus: fulStatus,
		SortBy:            sortBy,
		SortOrder:         sortOrder,
	}

	entities, totalCount, err := s.orderRepo.List(filter)
	if err != nil {
		return nil, 0, err
	}

	responses := mapper.OrderEntitiesToResponses(entities)
	if responses == nil {
		responses = []dto.OrderResponse{} // return empty list instead of null
	}
	return responses, totalCount, nil
}

// GetOrder retrieves a single order by ID with its line items.
func (s *OrderService) GetOrder(id string) (dto.OrderResponse, error) {
	e, err := s.orderRepo.GetByID(id)
	if err != nil {
		return dto.OrderResponse{}, err
	}

	items, err := s.lineItemRepo.GetByOrderID(e.ID)
	if err != nil {
		return dto.OrderResponse{}, err
	}
	e.LineItems = items

	return mapper.OrderEntityToResponse(e), nil
}

// GetOrderEntity retrieves the raw entity for internal use (e.g., invoice generation).
func (s *OrderService) GetOrderEntity(id string) (entity.Order, error) {
	return s.orderRepo.GetByID(id)
}

// GetOrderByExternalID retrieves an order by external (Shopify) ID.
func (s *OrderService) GetOrderByExternalID(externalID string) (entity.Order, error) {
	return s.orderRepo.GetByExternalID(externalID)
}

// GetLineItems retrieves line items for an order.
func (s *OrderService) GetLineItems(orderID string) ([]entity.LineItem, error) {
	return s.lineItemRepo.GetByOrderID(orderID)
}

// UpdateOrderStatus updates the status field of an order.
func (s *OrderService) UpdateOrderStatus(id string, status string) (int64, error) {
	return s.orderRepo.UpdateOrderStatus(id, status)
}

// UpsertOrder inserts or updates a single order (used by webhooks).
func (s *OrderService) UpsertOrder(order entity.Order) error {
	return s.orderRepo.Upsert(order)
}

// UpdatePaymentStatus updates the financial status of an order.
func (s *OrderService) UpdatePaymentStatus(externalOrderID string, status string) error {
	return s.orderRepo.UpdateStatus(externalOrderID, status, "")
}

// UpdateFulfillmentStatus updates the fulfillment status of an order.
func (s *OrderService) UpdateFulfillmentStatus(externalOrderID string, status string) error {
	return s.orderRepo.UpdateStatus(externalOrderID, "", status)
}

// UpdateTrackingInfo updates the tracking details of an order.
func (s *OrderService) UpdateTrackingInfo(externalOrderID string, trackingNumber, shippingCompany, trackingUrl, deliveryStatus string) error {
	return s.orderRepo.UpdateTrackingInfo(externalOrderID, trackingNumber, shippingCompany, trackingUrl, deliveryStatus)
}

// CancelOrder marks an order as cancelled.
func (s *OrderService) CancelOrder(externalOrderID string, cancelledAt *string, reason string) error {
	return s.orderRepo.CancelOrder(externalOrderID, cancelledAt, reason)
}

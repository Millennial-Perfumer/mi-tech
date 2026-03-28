package service

import (
	"context"
	"fmt"
	"mi-tech/internal/client/shopify"
	"mi-tech/internal/dto"
	"mi-tech/internal/entity"
	"mi-tech/internal/mapper"
	"mi-tech/internal/repository"
)

// OrderService handles order CRUD business logic.
type OrderService struct {
	orderRepo       repository.OrderRepository
	lineItemRepo    repository.LineItemRepository
	customerService *CustomerService
	shopifyClient   *shopify.Client
}

// NewOrderService creates a new OrderService.
func NewOrderService(orderRepo repository.OrderRepository, lineItemRepo repository.LineItemRepository, customerService *CustomerService, shopifyClient *shopify.Client) *OrderService {
	return &OrderService{
		orderRepo:       orderRepo,
		lineItemRepo:    lineItemRepo,
		customerService: customerService,
		shopifyClient:   shopifyClient,
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
func (s *OrderService) GetOrder(id int64) (dto.OrderResponse, error) {
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
func (s *OrderService) GetOrderEntity(id int64) (entity.Order, error) {
	return s.orderRepo.GetByID(id)
}

// GetOrderFlexible retrieves an order by either internal ID or external ID.
func (s *OrderService) GetOrderFlexible(id string) (entity.Order, error) {
	return s.orderRepo.GetByFlexibleID(id)
}

// GetOrderByExternalID retrieves an order by external (Shopify) ID.
func (s *OrderService) GetOrderByExternalID(externalID string) (entity.Order, error) {
	return s.orderRepo.GetByExternalID(externalID)
}

// GetLineItems retrieves line items for an order.
func (s *OrderService) GetLineItems(orderID int64) ([]entity.LineItem, error) {
	return s.lineItemRepo.GetByOrderID(orderID)
}

// UpdateOrderStatus updates the status field of an order.
func (s *OrderService) UpdateOrderStatus(id int64, status string) (int64, error) {
	return s.orderRepo.UpdateOrderStatus(id, status)
}

// UpsertOrder inserts or updates a single order (used by webhooks).
func (s *OrderService) UpsertOrder(order entity.Order) error {
	if err := s.orderRepo.Upsert(order); err != nil {
		return err
	}
	if s.customerService != nil {
		_ = s.customerService.UpdateFromOrder(context.Background(), &order)
	}
	return nil
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
	if err := s.orderRepo.CancelOrder(externalOrderID, cancelledAt, reason); err != nil {
		return err
	}
	// Fetch full order to update customer stats
	order, err := s.orderRepo.GetByExternalID(externalOrderID)
	if err == nil && s.customerService != nil {
		_ = s.customerService.UpdateFromOrder(context.Background(), &order)
	}
	return nil
}

func (s *OrderService) ListSources() ([]entity.Source, error) {
	return s.orderRepo.ListSources()
}

// UpdateOrder updates an order locally and in Shopify.
func (s *OrderService) UpdateOrder(id int64, req dto.OrderUpdateRequest) error {
	// 1. Fetch current order to get external ID
	order, err := s.orderRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("order not found: %w", err)
	}

	// 2. Prepare entity updates
	order.CustomerFirstName = entity.StrPtr(req.CustomerFirstName)
	order.CustomerLastName = entity.StrPtr(req.CustomerLastName)
	order.CustomerEmail = entity.StrPtr(req.CustomerEmail)
	order.CustomerPhone = entity.StrPtr(req.CustomerPhone)
	order.CustomerAddress1 = entity.StrPtr(req.CustomerAddress1)
	order.CustomerAddress2 = entity.StrPtr(req.CustomerAddress2)
	order.CustomerCity = entity.StrPtr(req.CustomerCity)
	order.CustomerState = entity.StrPtr(req.CustomerState)
	order.CustomerZip = entity.StrPtr(req.CustomerZip)
	order.CustomerCountry = entity.StrPtr(req.CustomerCountry)
	
	fullName := req.CustomerFirstName
	if req.CustomerLastName != "" {
		fullName += " " + req.CustomerLastName
	}
	order.CustomerName = entity.StrPtr(fullName)

	// 3. Update local database
	if err := s.orderRepo.UpdateOrderDetails(id, order); err != nil {
		return fmt.Errorf("failed to update local database: %w", err)
	}

	// 4. Sync to Shopify if it's a Shopify order
	if s.shopifyClient != nil && order.SourceID == "shopify" && order.ExternalOrderID != "" {
		shopifyData := map[string]interface{}{
			"email": req.CustomerEmail,
			"shipping_address": map[string]interface{}{
				"first_name": req.CustomerFirstName,
				"last_name":  req.CustomerLastName,
				"address1":   req.CustomerAddress1,
				"address2":   req.CustomerAddress2,
				"city":       req.CustomerCity,
				"province":   req.CustomerState,
				"zip":        req.CustomerZip,
				"country":    req.CustomerCountry,
				"phone":      req.CustomerPhone,
			},
		}

		if err := s.shopifyClient.UpdateOrder(order.ExternalOrderID, shopifyData); err != nil {
			return fmt.Errorf("failed to sync with Shopify: %w", err)
		}
	}

	return nil
}

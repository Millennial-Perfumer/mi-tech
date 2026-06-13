package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"mi-tech/internal/client/shopify"
	"mi-tech/internal/domain/order/dto"
	"mi-tech/internal/domain/order/entity"
	"mi-tech/internal/domain/order/mapper"
	"mi-tech/internal/domain/order/repository"
	"mi-tech/internal/domain/shared/util"
)

// SyncOrchestrator defines the contract for synchronizing stock updates across platforms.
type SyncOrchestrator interface {
	AdjustStockByPlatformSKU(ctx context.Context, platform string, sku string, delta int, reason string, externalOrderID *string) error
	GlobalSync(ctx context.Context, itemID int, sourcePlatform string) error
}

// OrderService handles order CRUD business logic.
type OrderService struct {
	orderRepo       repository.OrderRepository
	lineItemRepo    repository.LineItemRepository
	customerService *CustomerService
	shopifyClient   *shopify.Client
	orchestrator    SyncOrchestrator
}

// NewOrderService creates a new OrderService.
func NewOrderService(orderRepo repository.OrderRepository, lineItemRepo repository.LineItemRepository, customerService *CustomerService, shopifyClient *shopify.Client, orchestrator SyncOrchestrator) *OrderService {
	return &OrderService{
		orderRepo:       orderRepo,
		lineItemRepo:    lineItemRepo,
		customerService: customerService,
		shopifyClient:   shopifyClient,
		orchestrator:    orchestrator,
	}
}

// CustomerService returns the nested CustomerService.
func (s *OrderService) CustomerService() *CustomerService {
	return s.customerService
}

// ListOrders retrieves a paginated list of orders and converts them to DTOs.
func (s *OrderService) ListOrders(startDate, endDate string, page, limit int, search, source, finStatus, fulStatus, status, sortBy, sortOrder string) ([]dto.OrderResponse, int, error) {
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
		Status:            status,
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

// GetOrderResponseFlexible retrieves an order by either internal ID or external ID with line items, returning a DTO.
func (s *OrderService) GetOrderResponseFlexible(id string) (dto.OrderResponse, error) {
	e, err := s.orderRepo.GetByFlexibleID(id)
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

// GetOrderByExternalID retrieves an order by external (Shopify) ID.
func (s *OrderService) GetOrderByExternalID(externalID string) (entity.Order, error) {
	return s.orderRepo.GetByExternalID(externalID)
}

// GetLineItems retrieves line items for an order.
func (s *OrderService) GetLineItems(orderID int64) ([]entity.LineItem, error) {
	return s.lineItemRepo.GetByOrderID(orderID)
}

// UpdateOrderStatus updates the status field of an order and handles side effects like inventory reversal on cancellation.
func (s *OrderService) UpdateOrderStatus(id int64, status string) (int64, error) {
	return s.UpdateOrderStatusWithEntity(id, status, nil)
}

// UpdateOrderStatusWithEntity updates the status field, optionally using an already-fetched entity to avoid redundant DB lookups.
func (s *OrderService) UpdateOrderStatusWithEntity(id int64, status string, orderPtr *entity.Order) (int64, error) {
	// 1. If cancelling, handle inventory reversal
	if status == "cancelled" {
		var order entity.Order
		var err error
		if orderPtr != nil {
			order = *orderPtr
		} else {
			order, err = s.orderRepo.GetByID(id)
		}
		if err == nil && order.InventoryDeducted {
			// Trigger reversal for each item
			items, err := s.lineItemRepo.GetByOrderID(order.ID)
			if err == nil {
				for _, item := range items {
					if item.SKU != nil && s.orchestrator != nil {
						// For Amazon, we need to map the SKU back to the inventory item
						// Actually, our orchestrator Handles deduction based on the Platform SKU
						// but AdjustStock usually takes the InventoryItemID.
						// The poller uses inventoryRepo.GetItemByPlatformSKU.

						// To keep it simple and consistent with the poller:
						// We'll perform the same reversal logic here for Amazon orders.
						if order.SourceID == "amazon" {
							_ = s.orchestrator.AdjustStockByPlatformSKU(context.Background(), "amazon", *item.SKU, item.Quantity, "manual_cancellation", &order.ExternalOrderID)
						}
					}
				}
				// Mark as not deducted anymore
				order.InventoryDeducted = false
				s.orderRepo.UpdateOrderDetails(order.ID, order)
			}
		}
	}

	return s.orderRepo.UpdateOrderStatus(id, status)
}

// UpsertOrder inserts or updates a single order (used by webhooks).
func (s *OrderService) UpsertOrder(order entity.Order) error {
	affectedIDs, err := s.orderRepo.Upsert(order)
	if err != nil {
		return err
	}

	// Trigger global sync for all affected inventory items
	for _, id := range affectedIDs {
		_ = s.orchestrator.GlobalSync(context.Background(), id, order.SourceID)
	}

	if s.customerService != nil {
		_ = s.customerService.UpdateFromOrder(context.Background(), &order)
	}
	return nil
}

// UpdateOrderPaymentStatus updates the financial status of an order using its internal ID.
func (s *OrderService) UpdateOrderPaymentStatus(id int64, status string) error {
	return s.orderRepo.UpdateFinancialStatus(id, status)
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
	order.CustomerFirstName = util.StrPtr(req.CustomerFirstName)
	order.CustomerLastName = util.StrPtr(req.CustomerLastName)
	order.CustomerEmail = util.StrPtr(req.CustomerEmail)
	order.CustomerPhone = util.StrPtr(req.CustomerPhone)
	order.CustomerAddress1 = util.StrPtr(req.CustomerAddress1)
	order.CustomerAddress2 = util.StrPtr(req.CustomerAddress2)
	order.CustomerCity = util.StrPtr(req.CustomerCity)
	order.CustomerState = util.StrPtr(req.CustomerState)
	order.CustomerZip = util.StrPtr(req.CustomerZip)
	order.CustomerCountry = util.StrPtr(req.CustomerCountry)

	fullName := req.CustomerFirstName
	if req.CustomerLastName != "" {
		fullName += " " + req.CustomerLastName
	}
	order.CustomerName = util.StrPtr(fullName)

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

func (s *OrderService) GetOrdersForFeedback(delayMinutes int) ([]entity.Order, error) {
	return s.orderRepo.GetOrdersForFeedback(delayMinutes)
}

func (s *OrderService) MarkAsDelivered(id int64) error {
	return s.orderRepo.MarkAsDelivered(id)
}

func (s *OrderService) UpdateFeedbackStatus(id int64, statusID int) error {
	return s.orderRepo.UpdateFeedbackStatus(id, statusID)
}

func (s *OrderService) ValidateFeedback(orderID int64, phone string) (bool, error) {
	_, err := s.orderRepo.GetByIDAndPhone(orderID, phone)
	if err != nil {
		return false, err
	}
	return true, nil
}

// CreateManualOrder generates a new sequential POS order, updates inventory, and syncs to other platforms.
func (s *OrderService) CreateManualOrder(ctx context.Context, req dto.OrderCreateRequest) (dto.OrderResponse, error) {
	// 1. Generate sequential order number
	if req.TerminalCode == "" {
		req.TerminalCode = "POS1"
	}
	orderNumber, err := s.orderRepo.GetNextPOSSequence(req.TerminalCode)
	if err != nil {
		return dto.OrderResponse{}, fmt.Errorf("failed to generate POS order number: %w", err)
	}

	fullName := strings.TrimSpace(req.CustomerName)
	firstName := fullName
	lastName := ""
	if parts := strings.SplitN(fullName, " ", 2); len(parts) == 2 {
		firstName = parts[0]
		lastName = parts[1]
	}

	financialStatus := req.FinancialStatus
	if financialStatus == "" {
		financialStatus = "paid"
	}
	fulfillmentStatus := req.FulfillmentStatus
	if fulfillmentStatus == "" {
		fulfillmentStatus = "fulfilled"
	}

	// Calculate subtotal price and tax
	// TotalPrice is inclusive of GST. Inclusive tax calculation (18% inclusive GST):
	totalTax := req.TotalPrice - (req.TotalPrice / 1.18)
	subtotalPrice := req.TotalPrice - totalTax

	order := entity.Order{
		SourceID:          "pos",
		ExternalOrderID:   "pos-" + orderNumber,
		OrderNumber:       orderNumber,
		TotalPrice:        req.TotalPrice,
		SubtotalPrice:     util.Float64Ptr(subtotalPrice),
		TotalTax:          util.Float64Ptr(totalTax),
		TotalDiscount:     req.TotalDiscount,
		Currency:          util.StrPtr("INR"),
		FinancialStatus:   util.StrPtr(financialStatus),
		FulfillmentStatus: util.StrPtr(fulfillmentStatus),
		CustomerName:      util.StrPtr(fullName),
		CustomerFirstName: util.StrPtr(firstName),
		CustomerLastName:  util.StrPtr(lastName),
		CustomerEmail:     util.StrPtr(req.CustomerEmail),
		CustomerPhone:     util.StrPtr(req.CustomerPhone),
		CustomerAddress1:  util.StrPtr(req.CustomerAddress1),
		CustomerAddress2:  util.StrPtr(req.CustomerAddress2),
		CustomerCity:      util.StrPtr(req.CustomerCity),
		CustomerState:     util.StrPtr(req.CustomerState),
		CustomerZip:       util.StrPtr(req.CustomerZip),
		CustomerCountry:   util.StrPtr(req.CustomerCountry),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	for i, li := range req.LineItems {
		order.LineItems = append(order.LineItems, entity.LineItem{
			ID:       fmt.Sprintf("pos-%s-%d", orderNumber, i),
			SKU:      util.StrPtr(li.MISKU),
			Title:    util.StrPtr(li.Title),
			Quantity: li.Quantity,
			Price:    li.Price,
			Discount: li.Discount,
		})
	}

	// 4. Upsert → syncInventoryDeltas → deducts current_stock
	affectedIDs, err := s.orderRepo.Upsert(order)
	if err != nil {
		return dto.OrderResponse{}, fmt.Errorf("failed to save manual order: %w", err)
	}

	// Fetch fully saved order with generated ID
	savedOrder, err := s.orderRepo.GetByExternalID(order.ExternalOrderID)
	if err != nil {
		return dto.OrderResponse{}, fmt.Errorf("failed to fetch saved manual order: %w", err)
	}

	// Fetch or assign line items
	if s.lineItemRepo != nil {
		lineItems, err := s.lineItemRepo.GetByOrderID(savedOrder.ID)
		if err == nil {
			savedOrder.LineItems = lineItems
		} else {
			savedOrder.LineItems = order.LineItems
		}
	} else {
		savedOrder.LineItems = order.LineItems
	}

	// 5. GlobalSync → push updated stock to Shopify + Amazon
	if s.orchestrator != nil {
		for _, id := range affectedIDs {
			_ = s.orchestrator.GlobalSync(ctx, id, "pos")
		}
	}

	// 6. Create/update customer profile
	if s.customerService != nil {
		_ = s.customerService.UpdateFromOrder(ctx, &savedOrder)
	}

	return mapper.OrderEntityToResponse(savedOrder), nil
}

package mapper

import (
	"fmt"
	"time"

	"mi-tech/internal/dto"
	"mi-tech/internal/entity"
)

// OrderEntityToResponse converts a DB entity to an API response DTO.
func OrderEntityToResponse(e entity.Order) dto.OrderResponse {
	resp := dto.OrderResponse{
		ID:                fmt.Sprintf("%d", e.ID),
		OrderNumber:       e.OrderNumber,
		TotalPrice:        fmt.Sprintf("%.2f", e.TotalPrice),
		SubtotalPrice:     ptrFloat64ToStr(e.SubtotalPrice),
		TotalTax:          ptrFloat64ToStr(e.TotalTax),
		Currency:          deref(e.Currency),
		FinancialStatus:   deref(e.FinancialStatus),
		FulfillmentStatus: deref(e.FulfillmentStatus),
		DeliveryStatus:    deref(e.DeliveryStatus),
		TrackingNumber:    deref(e.TrackingNumber),
		ShippingCompany:   deref(e.ShippingCompany),
		TrackingUrl:       deref(e.TrackingUrl),
		SourceID:          e.SourceID,
		Status:            deref(e.Status),
		CreatedAt:         e.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         e.UpdatedAt.Format(time.RFC3339),
		CancelReason:      deref(e.CancelReason),
		CustomerName:      deref(e.CustomerName),
		CustomerFirstName: deref(e.CustomerFirstName),
		CustomerLastName:  deref(e.CustomerLastName),
		CustomerEmail:     deref(e.CustomerEmail),
		CustomerPhone:     deref(e.CustomerPhone),
		CustomerCity:      deref(e.CustomerCity),
		CustomerState:     deref(e.CustomerState),
		CustomerCountry:   deref(e.CustomerCountry),
		CustomerAddress1:  deref(e.CustomerAddress1),
		CustomerAddress2:  deref(e.CustomerAddress2),
		CustomerZip:       deref(e.CustomerZip),
		FeedbackStatusID:  e.FeedbackStatusID,
	}

	if e.FeedbackSentAt != nil {
		sentAtStr := e.FeedbackSentAt.Format(time.RFC3339)
		resp.FeedbackSentAt = &sentAtStr
	}

	if e.CancelledAt != nil {
		resp.CancelledAt = e.CancelledAt.Format(time.RFC3339)
	}

	for _, li := range e.LineItems {
		resp.LineItems = append(resp.LineItems, LineItemEntityToResponse(li))
	}

	return resp
}

// OrderEntitiesToResponses converts a slice of entities to response DTOs.
func OrderEntitiesToResponses(entities []entity.Order) []dto.OrderResponse {
	results := make([]dto.OrderResponse, 0, len(entities))
	for _, e := range entities {
		results = append(results, OrderEntityToResponse(e))
	}
	return results
}

// LineItemEntityToResponse converts a line item entity to a response DTO.
func LineItemEntityToResponse(e entity.LineItem) dto.LineItemResponse {
	return dto.LineItemResponse{
		ID:        e.ID,
		ProductID: deref(e.ProductID),
		VariantID: deref(e.VariantID),
		Title:     deref(e.Title),
		SKU:       deref(e.SKU),
		HSCode:    deref(e.HSCode),
		Quantity:  e.Quantity,
		Price:     fmt.Sprintf("%.2f", e.Price),
		Discount:  fmt.Sprintf("%.2f", e.Discount),
	}
}

// --- Helper functions for pointer to plain string conversions ---

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func ptrFloat64ToStr(f *float64) string {
	if f == nil {
		return "0.00"
	}
	return fmt.Sprintf("%.2f", *f)
}

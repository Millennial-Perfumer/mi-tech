package mapper

import (
	"database/sql"
	"fmt"
	"time"

	"shopify-gst-app/internal/dto"
	"shopify-gst-app/internal/entity"
)

// OrderEntityToResponse converts a DB entity to an API response DTO.
func OrderEntityToResponse(e entity.Order) dto.OrderResponse {
	resp := dto.OrderResponse{
		ID:                e.ID,
		OrderNumber:       e.OrderNumber,
		TotalPrice:        fmt.Sprintf("%.2f", e.TotalPrice),
		SubtotalPrice:     nullFloat64ToStr(e.SubtotalPrice),
		TotalTax:          nullFloat64ToStr(e.TotalTax),
		Currency:          nullStr(e.Currency),
		FinancialStatus:   nullStr(e.FinancialStatus),
		FulfillmentStatus: nullStr(e.FulfillmentStatus),
		DeliveryStatus:    nullStr(e.DeliveryStatus),
		TrackingNumber:    nullStr(e.TrackingNumber),
		ShippingCompany:   nullStr(e.ShippingCompany),
		TrackingUrl:       nullStr(e.TrackingUrl),
		SourceID:          e.SourceID,
		Status:            nullStr(e.Status),
		CreatedAt:         e.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         e.UpdatedAt.Format(time.RFC3339),
		CancelReason:      nullStr(e.CancelReason),
		CustomerName:      nullStr(e.CustomerName),
		CustomerFirstName: nullStr(e.CustomerFirstName),
		CustomerLastName:  nullStr(e.CustomerLastName),
		CustomerEmail:     nullStr(e.CustomerEmail),
		CustomerPhone:     nullStr(e.CustomerPhone),
		CustomerCity:      nullStr(e.CustomerCity),
		CustomerState:     nullStr(e.CustomerState),
		CustomerCountry:   nullStr(e.CustomerCountry),
		CustomerAddress1:  nullStr(e.CustomerAddress1),
		CustomerAddress2:  nullStr(e.CustomerAddress2),
		CustomerZip:       nullStr(e.CustomerZip),
	}

	if e.CancelledAt.Valid {
		resp.CancelledAt = e.CancelledAt.Time.Format(time.RFC3339)
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
		ProductID: nullStr(e.ProductID),
		VariantID: nullStr(e.VariantID),
		Title:     nullStr(e.Title),
		SKU:       nullStr(e.SKU),
		HSCode:    nullStr(e.HSCode),
		Quantity:  e.Quantity,
		Price:     fmt.Sprintf("%.2f", e.Price),
		Discount:  fmt.Sprintf("%.2f", e.Discount),
	}
}

// --- Helper functions for sql.Null* to plain string conversions ---

func nullStr(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func nullFloat64ToStr(nf sql.NullFloat64) string {
	if nf.Valid {
		return fmt.Sprintf("%.2f", nf.Float64)
	}
	return "0.00"
}

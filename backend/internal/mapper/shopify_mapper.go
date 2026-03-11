package mapper

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"

	"shopify-gst-app/internal/dto"
	"shopify-gst-app/internal/entity"
)

// GraphQLOrderToEntity converts a Shopify GraphQL order node into a DB entity.
// It handles customer name resolution, address fallback, fulfillment/tracking extraction,
// and source_id mapping.
func GraphQLOrderToEntity(so dto.GraphQLOrderNode) entity.Order {
	var custName, custEmail, custPhone, custCity, custState, custCountry string

	custEmail = so.Email
	if so.Customer != nil {
		custName = strings.TrimSpace(so.Customer.DisplayName)
		if custName == "" {
			custName = strings.TrimSpace(so.Customer.FirstName + " " + so.Customer.LastName)
		}
	}

	if so.ShippingAddress != nil {
		if custName == "" {
			custName = strings.TrimSpace(so.ShippingAddress.Name)
		}
		custPhone = so.ShippingAddress.Phone
		custCity = so.ShippingAddress.City
		custState = so.ShippingAddress.Province
		custCountry = so.ShippingAddress.Country
	} else if so.BillingAddress != nil {
		if custName == "" {
			custName = strings.TrimSpace(so.BillingAddress.Name)
		}
		custPhone = so.BillingAddress.Phone
		custCity = so.BillingAddress.City
		custState = so.BillingAddress.Province
		custCountry = so.BillingAddress.Country
	}

	if custName == "" {
		custName = "Valued Customer"
	}

	// Determine status
	status := "unfulfilled"
	if so.DisplayFulfillmentStatus == "FULFILLED" {
		status = "fulfilled"
	} else if so.DisplayFinancialStatus == "PAID" {
		status = "paid"
	}

	financialStatus := strings.ToLower(so.DisplayFinancialStatus)
	fulfillmentStatus := strings.ToLower(so.DisplayFulfillmentStatus)
	deliveryStatus := "pending"
	trackingNumber := ""
	shippingCompany := ""
	trackingUrl := ""

	if len(so.Fulfillments) > 0 {
		f := so.Fulfillments[0]
		if f.DisplayStatus != "" {
			deliveryStatus = strings.ToLower(strings.ReplaceAll(f.DisplayStatus, "_", " "))
		} else {
			deliveryStatus = strings.ToLower(strings.ReplaceAll(f.Status, "_", " "))
		}
		if len(f.TrackingInfo) > 0 {
			trackingNumber = f.TrackingInfo[0].Number
			shippingCompany = f.TrackingInfo[0].Company
			trackingUrl = f.TrackingInfo[0].Url
		}
	}

	// Map sourceName to internal source_id
	sourceID := "shopify"
	switch strings.ToLower(so.SourceName) {
	case "amazon":
		sourceID = "amazon"
	case "pos":
		sourceID = "pos"
	}

	idStr := strings.TrimPrefix(so.ID, "gid://shopify/Order/")

	return entity.Order{
		ID:                idStr,
		SourceID:          sourceID,
		ExternalOrderID:   idStr,
		OrderNumber:       so.Name,
		TotalPrice:        parseFloat(so.CurrentTotalPriceSet.ShopMoney.Amount),
		SubtotalPrice:     toNullFloat64(so.CurrentSubtotalPriceSet.ShopMoney.Amount),
		TotalTax:          toNullFloat64(so.CurrentTotalTaxSet.ShopMoney.Amount),
		FinancialStatus:   toNullString(financialStatus),
		FulfillmentStatus: toNullString(fulfillmentStatus),
		DeliveryStatus:    toNullString(deliveryStatus),
		TrackingNumber:    toNullString(trackingNumber),
		ShippingCompany:   toNullString(shippingCompany),
		TrackingUrl:       toNullString(trackingUrl),
		Status:            toNullString(status),
		CustomerName:      toNullString(custName),
		CustomerEmail:     toNullString(custEmail),
		CustomerPhone:     toNullString(custPhone),
		CustomerCity:      toNullString(custCity),
		CustomerState:     toNullString(custState),
		CustomerCountry:   toNullString(custCountry),
	}
}

// GraphQLLineItemsToEntities converts GraphQL line items into DB entities.
func GraphQLLineItemsToEntities(orderID string, items dto.GraphQLLineItemWrap) []entity.LineItem {
	var result []entity.LineItem
	for _, edge := range items.Edges {
		li := edge.Node
		hsCode := ""
		if li.Variant != nil {
			hsCode = li.Variant.InventoryItem.HarmonizedSystemCode
		}

		itemID := strings.TrimPrefix(li.ID, "gid://shopify/LineItem/")
		result = append(result, entity.LineItem{
			ID:       itemID,
			OrderID:  orderID,
			Title:    toNullString(li.Title),
			SKU:      toNullString(li.SKU),
			HSCode:   toNullString(hsCode),
			Quantity: li.Quantity,
			Price:    parseFloat(li.OriginalTotalSet.ShopMoney.Amount),
			Discount: parseFloat(li.TotalDiscountSet.ShopMoney.Amount),
		})
	}
	return result
}

// WebhookOrderToEntity converts a Shopify REST webhook payload into a DB entity.
func WebhookOrderToEntity(payload dto.ShopifyWebhookOrder, rawPayload *json.RawMessage) entity.Order {
	customerName := ""
	firstName, lastName := "", ""

	if payload.Customer != nil {
		firstName = payload.Customer.FirstName
		lastName = payload.Customer.LastName
	}

	if payload.BillingAddress != nil && payload.BillingAddress.Name != "" {
		customerName = payload.BillingAddress.Name
		if firstName == "" {
			firstName = payload.BillingAddress.FirstName
		}
		if lastName == "" {
			lastName = payload.BillingAddress.LastName
		}
	} else if payload.ShippingAddress != nil && payload.ShippingAddress.Name != "" {
		customerName = payload.ShippingAddress.Name
		if firstName == "" {
			firstName = payload.ShippingAddress.FirstName
		}
		if lastName == "" {
			lastName = payload.ShippingAddress.LastName
		}
	}

	if customerName == "" && (firstName != "" || lastName != "") {
		customerName = firstName + " " + lastName
	}

	city, state, country, addr1, addr2, zip := "", "", "", "", "", ""
	addr := payload.BillingAddress
	if addr == nil {
		addr = payload.ShippingAddress
	}
	if addr != nil {
		city = addr.City
		state = addr.Province
		country = addr.Country
		addr1 = addr.Address1
		addr2 = addr.Address2
		zip = addr.Zip
	}

	phone := ""
	if payload.BillingAddress != nil && payload.BillingAddress.Phone != "" {
		phone = payload.BillingAddress.Phone
	} else if payload.ShippingAddress != nil && payload.ShippingAddress.Phone != "" {
		phone = payload.ShippingAddress.Phone
	} else if payload.Customer != nil {
		phone = payload.Customer.Phone
	}

	email := payload.Email
	if email == "" && payload.Customer != nil {
		email = payload.Customer.Email
	}

	// Map source_name to source_id
	sourceID := "shopify"
	switch strings.ToLower(payload.SourceName) {
	case "amazon":
		sourceID = "amazon"
	case "pos":
		sourceID = "pos"
	}

	idStr := strconv.FormatInt(payload.ID, 10)
	order := entity.Order{
		ID:                idStr,
		ExternalOrderID:   idStr,
		SourceID:          sourceID,
		OrderNumber:       strconv.FormatInt(payload.OrderNumber, 10),
		TotalPrice:        parseFloat(payload.TotalPrice),
		SubtotalPrice:     toNullFloat64(payload.SubtotalPrice),
		TotalTax:          toNullFloat64(payload.TotalTax),
		Currency:          toNullString(payload.Currency),
		FinancialStatus:   toNullString(payload.FinancialStatus),
		FulfillmentStatus: toNullString(payload.FulfillmentStatus),
		DeliveryStatus:    toNullString("pending"),
		Status:            toNullString(payload.FulfillmentStatus),
		CustomerName:      toNullString(customerName),
		CustomerFirstName: toNullString(firstName),
		CustomerLastName:  toNullString(lastName),
		CustomerEmail:     toNullString(email),
		CustomerPhone:     toNullString(phone),
		CustomerCity:      toNullString(city),
		CustomerState:     toNullString(state),
		CustomerCountry:   toNullString(country),
		CustomerAddress1:  toNullString(addr1),
		CustomerAddress2:  toNullString(addr2),
		CustomerZip:       toNullString(zip),
		RawPayload:        rawPayload,
	}

	if payload.CancelledAt != nil {
		order.CancelledAt = sql.NullTime{Valid: false} // Will be parsed by service if needed
		order.CancelReason = toNullString(payload.CancelReason)
	}

	for _, li := range payload.LineItems {
		price := li.Price
		if price == "" {
			price = "0.00"
		}
		discount := li.TotalDiscount
		if discount == "" {
			discount = "0.00"
		}
		order.LineItems = append(order.LineItems, entity.LineItem{
			ID:        strconv.FormatInt(li.ID, 10),
			OrderID:   idStr,
			ProductID: toNullString(strconv.FormatInt(li.ProductID, 10)),
			VariantID: toNullString(strconv.FormatInt(li.VariantID, 10)),
			Title:     toNullString(li.Title),
			SKU:       toNullString(li.SKU),
			HSCode:    toNullString("33029019"), // Default HSN
			Quantity:  li.Quantity,
			Price:     parseFloat(price),
			Discount:  parseFloat(discount),
		})
	}

	return order
}

// --- Helpers ---

func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

func toNullFloat64(s string) sql.NullFloat64 {
	if s == "" {
		return sql.NullFloat64{Valid: false}
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return sql.NullFloat64{Valid: false}
	}
	return sql.NullFloat64{Float64: v, Valid: true}
}

func parseFloat(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

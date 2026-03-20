package mapper

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"mi-tech/internal/dto"
	"mi-tech/internal/entity"
)

// GraphQLOrderToEntity converts a Shopify GraphQL order node into a DB entity.
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
	deliveryStatus := "pending"

	if so.CancelledAt != "" {
		status = "CANCELLED"
		deliveryStatus = ""
	} else if so.DisplayFulfillmentStatus == "FULFILLED" {
		status = "fulfilled"
	} else if so.DisplayFinancialStatus == "PAID" {
		status = "paid"
	}

	financialStatus := strings.ToLower(so.DisplayFinancialStatus)
	fulfillmentStatus := strings.ToLower(so.DisplayFulfillmentStatus)
	trackingNumber := ""
	shippingCompany := ""
	trackingUrl := ""

	if len(so.Fulfillments) > 0 {
		// Iterate to find the best active fulfillment (ignore cancelled ones if possible)
		var bestFulfillment *dto.GraphQLFulfillment
		for i := len(so.Fulfillments) - 1; i >= 0; i-- {
			f := so.Fulfillments[i]
			if strings.ToLower(f.Status) != "cancelled" {
				bestFulfillment = &so.Fulfillments[i]
				break
			}
		}

		// Fallback to the last one if all are cancelled
		if bestFulfillment == nil {
			bestFulfillment = &so.Fulfillments[len(so.Fulfillments)-1]
		}

		f := bestFulfillment
		
		// Prioritize: 1. Latest event status, 2. Display status, 3. Raw status
		if len(f.Events.Edges) > 0 {
			lastEvent := f.Events.Edges[len(f.Events.Edges)-1].Node
			deliveryStatus = strings.ToLower(strings.ReplaceAll(lastEvent.Status, "_", " "))
		} else if f.DisplayStatus != "" {
			deliveryStatus = strings.ToLower(strings.ReplaceAll(f.DisplayStatus, "_", " "))
		} else if f.Status != "" {
			deliveryStatus = strings.ToLower(strings.ReplaceAll(f.Status, "_", " "))
		}

		if len(f.TrackingInfo) > 0 {
			trackingNumber = f.TrackingInfo[0].Number
			shippingCompany = f.TrackingInfo[0].Company
			trackingUrl = f.TrackingInfo[0].Url
		}

		// If no carrier events exist but we possess a tracking number, assume Confirmed
		if (deliveryStatus == "fulfilled" || deliveryStatus == "success") && trackingNumber != "" {
			deliveryStatus = "confirmed"
		}
	}

	// Map sourceName to internal source_id
	sourceID := "shopify"
	sourceName := strings.ToLower(so.SourceName)
	if strings.Contains(sourceName, "amazon") {
		sourceID = "amazon"
	} else if strings.Contains(sourceName, "pos") {
		sourceID = "pos"
	}

	idStr := strings.TrimPrefix(so.ID, "gid://shopify/Order/")

	createdAt := parseTime(so.ProcessedAt)
	if createdAt.IsZero() {
		createdAt = parseTime(so.CreatedAt)
	}
	updatedAt := parseTime(so.UpdatedAt)

	totalPrice := parseFloat(so.CurrentTotalPriceSet.ShopMoney.Amount)
	taxableValue := totalPrice / 1.18
	totalTax := totalPrice - taxableValue

	inr := "INR"
	return entity.Order{
		ExternalOrderID:   idStr,
		SourceID:          sourceID,
		OrderNumber:       so.Name,
		TotalPrice:        totalPrice,
		SubtotalPrice:     &taxableValue,
		TotalTax:          &totalTax,
		Currency:          &inr,
		CreatedAt:         createdAt,
		UpdatedAt:         updatedAt,
		FinancialStatus:   strPtr(financialStatus),
		FulfillmentStatus: strPtr(fulfillmentStatus),
		DeliveryStatus:    strPtr(deliveryStatus),
		TrackingNumber:    strPtr(trackingNumber),
		ShippingCompany:   strPtr(shippingCompany),
		TrackingUrl:       strPtr(trackingUrl),
		Status:            strPtr(status),
		CancelledAt:       parseTimePtr(&so.CancelledAt),
		CancelReason:      strPtr(so.CancelReason),
		CustomerName:      strPtr(custName),
		CustomerEmail:     strPtr(custEmail),
		CustomerPhone:     strPtr(custPhone),
		CustomerCity:      strPtr(custCity),
		CustomerState:     strPtr(custState),
		CustomerCountry:   strPtr(custCountry),
	}
}

// GraphQLLineItemsToEntities converts GraphQL line items into DB entities.
func GraphQLLineItemsToEntities(orderID int64, items dto.GraphQLLineItemWrap) []entity.LineItem {
	var result []entity.LineItem
	defaultHSN := "33029019"
	for _, edge := range items.Edges {
		li := edge.Node
		hsCode := ""
		if li.Variant != nil {
			hsCode = li.Variant.InventoryItem.HarmonizedSystemCode
		}

		itemID := strings.TrimPrefix(li.ID, "gid://shopify/LineItem/")
		qty := li.Quantity
		if li.CurrentQuantity != nil {
			qty = *li.CurrentQuantity
		}

		if li.CurrentQuantity != nil && *li.CurrentQuantity <= 0 {
			continue
		}

		result = append(result, entity.LineItem{
			ID:       itemID,
			OrderID:  orderID,
			Title:    strPtr(li.Title),
			SKU:      strPtr(li.SKU),
			HSCode:   strPtrOr(hsCode, defaultHSN),
			Quantity: qty,
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
	sourceName := strings.ToLower(payload.SourceName)
	if strings.Contains(sourceName, "amazon") {
		sourceID = "amazon"
	} else if strings.Contains(sourceName, "pos") {
		sourceID = "pos"
	}

	totalPrice := parseFloat(payload.TotalPrice)
	taxableValue := totalPrice / 1.18
	totalTax := totalPrice - taxableValue

	idStr := strconv.FormatInt(payload.ID, 10)
	defaultHSN := "33029019"

	financialStatus := strings.ToLower(payload.FinancialStatus)
	fulfillmentStatus := strings.ToLower(payload.FulfillmentStatus)
	deliveryStatus := "pending"
	trackingNumber := ""
	shippingCompany := ""
	trackingUrl := ""

	if len(payload.Fulfillments) > 0 {
		var bestFulfillment *dto.ShopifyFulfillment
		for i := len(payload.Fulfillments) - 1; i >= 0; i-- {
			f := payload.Fulfillments[i]
			if strings.ToLower(f.Status) != "cancelled" {
				bestFulfillment = &payload.Fulfillments[i]
				break
			}
		}

		if bestFulfillment == nil {
			bestFulfillment = &payload.Fulfillments[len(payload.Fulfillments)-1]
		}

		f := bestFulfillment
		
		// Determine delivery status from shipment_status, display_status, or status
		if f.ShipmentStatus != nil && *f.ShipmentStatus != "" {
			deliveryStatus = strings.ToLower(strings.ReplaceAll(*f.ShipmentStatus, "_", " "))
		} else if f.DisplayStatus != "" {
			deliveryStatus = strings.ToLower(strings.ReplaceAll(f.DisplayStatus, "_", " "))
		} else if f.Status != "" {
			deliveryStatus = strings.ToLower(strings.ReplaceAll(f.Status, "_", " "))
		}
		
		trackingNumber = f.TrackingNumber
		shippingCompany = f.TrackingCompany
		trackingUrl = f.TrackingUrl

		if (deliveryStatus == "fulfilled" || deliveryStatus == "success") && trackingNumber != "" {
			deliveryStatus = "confirmed"
		}

	}

	orderNumber := payload.Name
	if orderNumber == "" {
		orderNumber = strconv.FormatInt(payload.OrderNumber, 10)
	}

	order := entity.Order{
		ExternalOrderID:   idStr,
		SourceID:          sourceID,
		OrderNumber:       orderNumber,
		TotalPrice:        totalPrice,
		SubtotalPrice:     &taxableValue,
		TotalTax:          &totalTax,
		Currency:          strPtr(payload.Currency),
		FinancialStatus:   strPtr(financialStatus),
		FulfillmentStatus: strPtr(fulfillmentStatus),
		DeliveryStatus:    strPtr(deliveryStatus),
		TrackingNumber:    strPtr(trackingNumber),
		ShippingCompany:   strPtr(shippingCompany),
		TrackingUrl:       strPtr(trackingUrl),
		Status:            strPtr(fulfillmentStatus),
		CustomerName:      strPtr(customerName),
		CustomerFirstName: strPtr(firstName),
		CustomerLastName:  strPtr(lastName),
		CustomerEmail:     strPtr(email),
		CustomerPhone:     strPtr(phone),
		CustomerCity:      strPtr(city),
		CustomerState:     strPtr(state),
		CustomerCountry:   strPtr(country),
		CustomerAddress1:  strPtr(addr1),
		CustomerAddress2:  strPtr(addr2),
		CustomerZip:       strPtr(zip),
		CreatedAt:         parseTime(payload.CreatedAt),
		UpdatedAt:         parseTime(payload.UpdatedAt),
		RawPayload:        rawPayload,
	}

	if payload.CancelledAt != nil {
		order.CancelledAt = parseTimePtr(payload.CancelledAt)
		order.CancelReason = strPtr(payload.CancelReason)
		order.Status = strPtr("CANCELLED")
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
		qty := li.Quantity
		if li.CurrentQuantity != nil {
			qty = *li.CurrentQuantity
		}

		if li.CurrentQuantity != nil && *li.CurrentQuantity <= 0 {
			continue
		}

		order.LineItems = append(order.LineItems, entity.LineItem{
			ID:        strconv.FormatInt(li.ID, 10),
			OrderID:   0, // Will be linked in repository
			ProductID: strPtr(strconv.FormatInt(li.ProductID, 10)),
			VariantID: strPtr(strconv.FormatInt(li.VariantID, 10)),
			Title:     strPtr(li.Title),
			SKU:       strPtr(li.SKU),
			HSCode:    &defaultHSN,
			Quantity:  qty,
			Price:     parseFloat(price),
			Discount:  parseFloat(discount),
		})
	}

	return order
}

// --- Helpers ---

// strPtr returns a pointer to the string, or nil if empty.
func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// strPtrOr returns a pointer to the string, or a pointer to the fallback if empty.
func strPtrOr(s string, fallback string) *string {
	if s == "" {
		return &fallback
	}
	return &s
}

func parseFloat(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

func parseTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, s)
	if err == nil {
		return t
	}
	t, err = time.Parse("2006-01-02T15:04:05.000Z", s)
	if err == nil {
		return t
	}
	return time.Time{}
}

func parseTimePtr(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	t := parseTime(*s)
	if t.IsZero() {
		return nil
	}
	return &t
}

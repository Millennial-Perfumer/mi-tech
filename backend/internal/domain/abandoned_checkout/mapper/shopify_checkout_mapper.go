package mapper

import (
	"encoding/json"
	"strconv"
	"time"

	checkoutDto "mi-tech/internal/domain/abandoned_checkout/dto"
	checkoutEntity "mi-tech/internal/domain/abandoned_checkout/entity"
)

func WebhookCheckoutToEntity(payload checkoutDto.ShopifyWebhookCheckout) checkoutEntity.AbandonedCheckout {
	totalPrice, _ := strconv.ParseFloat(payload.TotalPrice, 64)

	// Pick phone
	phone := payload.Phone
	if phone == "" && payload.Customer != nil {
		phone = payload.Customer.Phone
	}
	if phone == "" && payload.ShippingAddress != nil {
		phone = payload.ShippingAddress.Phone
	}
	if phone == "" && payload.BillingAddress != nil {
		phone = payload.BillingAddress.Phone
	}

	// Pick name
	name := ""
	if payload.ShippingAddress != nil {
		name = payload.ShippingAddress.FirstName
		if payload.ShippingAddress.LastName != "" {
			if name != "" {
				name += " "
			}
			name += payload.ShippingAddress.LastName
		}
	}
	if name == "" && payload.Customer != nil {
		name = payload.Customer.FirstName
		if payload.Customer.LastName != "" {
			if name != "" {
				name += " "
			}
			name += payload.Customer.LastName
		}
	}

	// Consent
	marketingConsent := payload.BuyerAcceptsMarketing
	smsConsent := false
	if payload.SMSMarketingConsent != nil && payload.SMSMarketingConsent.State == "accepted" {
		smsConsent = true
	}

	// Times
	abandonedAt := time.Now()
	if payload.AbandonedAt != nil {
		if t, err := time.Parse(time.RFC3339, *payload.AbandonedAt); err == nil {
			abandonedAt = t
		}
	} else if payload.UpdatedAt != "" {
		if t, err := time.Parse(time.RFC3339, payload.UpdatedAt); err == nil {
			abandonedAt = t
		}
	}

	createdAt := time.Now()
	if t, err := time.Parse(time.RFC3339, payload.CreatedAt); err == nil {
		createdAt = t
	}

	// Line items JSON
	var lineItemsJSON *json.RawMessage
	if payload.LineItems != nil {
		if bytes, err := json.Marshal(payload.LineItems); err == nil {
			rawJSON := json.RawMessage(bytes)
			lineItemsJSON = &rawJSON
		}
	}

	// Extract location details
	city := ""
	province := ""
	country := ""
	zip := ""
	if payload.ShippingAddress != nil {
		city = payload.ShippingAddress.City
		province = payload.ShippingAddress.Province
		country = payload.ShippingAddress.Country
		zip = payload.ShippingAddress.Zip
	} else if payload.BillingAddress != nil {
		city = payload.BillingAddress.City
		province = payload.BillingAddress.Province
		country = payload.BillingAddress.Country
		zip = payload.BillingAddress.Zip
	}

	return checkoutEntity.AbandonedCheckout{
		CheckoutID:       strconv.FormatInt(payload.ID, 10),
		CheckoutToken:    payload.Token,
		CartToken:        payload.CartToken,
		Email:            payload.Email,
		Phone:            phone,
		CustomerName:     name,
		CheckoutURL:      payload.AbandonedCheckoutURL,
		LineItems:        lineItemsJSON,
		TotalPrice:       totalPrice,
		Currency:         payload.Currency,
		MarketingConsent: marketingConsent,
		SMSConsent:       smsConsent,
		City:             city,
		Province:         province,
		Country:          country,
		Zip:              zip,
		AbandonedAt:      abandonedAt,
		CreatedAt:        createdAt,
	}
}

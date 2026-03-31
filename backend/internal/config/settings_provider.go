package config

import (
	"log"
	"os"
)

type ConfigsRepository interface {
	Get(key string) (string, error)
}

type SettingsProvider struct {
	configsRepo ConfigsRepository
}

func NewSettingsProvider(configsRepo ConfigsRepository) *SettingsProvider {
	return &SettingsProvider{configsRepo: configsRepo}
}

func (p *SettingsProvider) get(key string) string {
	val, err := p.configsRepo.Get(key)
	if err != nil {
		log.Printf("ERROR: failed to get config for key %s: %v", key, err)
	}
	return val
}

func (p *SettingsProvider) GetShopifyStoreURL() string {
	return p.get("shopify_store_url")
}

func (p *SettingsProvider) GetShopifyAccessToken() string {
	return p.get("shopify_access_token")
}

func (p *SettingsProvider) GetShopifyWebhookSecret() string {
	return p.get("shopify_webhook_secret")
}

func (p *SettingsProvider) GetWhatsAppPhoneNumberID() string {
	return p.get("whatsapp_phone_number_id")
}

func (p *SettingsProvider) GetWhatsAppAccessToken() string {
	return p.get("whatsapp_access_token")
}

func (p *SettingsProvider) GetWhatsAppWABAID() string {
	return p.get("whatsapp_waba_id")
}

func (p *SettingsProvider) GetWhatsAppAppID() string {
	return p.get("whatsapp_app_id")
}

func (p *SettingsProvider) GetWhatsAppAppSecret() string {
	return p.get("whatsapp_app_secret")
}

func (p *SettingsProvider) GetJWTSecret() string {
	// Fallback to environment variable if not in DB
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		return secret
	}
	return p.get("jwt_secret")
}

func (p *SettingsProvider) GetShopifyAPIVersion() string {
	v := p.get("shopify_api_version")
	if v == "" {
		return "2026-01"
	}
	return v
}

func (p *SettingsProvider) GetWhatsAppWebhookVerifyToken() string {
	return p.get("whatsapp_webhook_verify_token")
}

func (p *SettingsProvider) IsPIIProtectionEnabled() bool {
	return p.get("pii_protection") == "true"
}

func (p *SettingsProvider) ShouldSendInvoice() bool {
	return p.get("send_invoice") == "true"
}

func (p *SettingsProvider) GetBusinessName() string {
	return p.get("business_name")
}

func (p *SettingsProvider) GetBusinessGSTIN() string {
	return p.get("business_gstin")
}

func (p *SettingsProvider) GetBusinessAddressLine1() string {
	return p.get("business_address_line1")
}

func (p *SettingsProvider) GetBusinessAddressLine2() string {
	return p.get("business_address_line2")
}

func (p *SettingsProvider) GetBusinessPhone() string {
	return p.get("business_phone")
}

func (p *SettingsProvider) GetBulkTemplateSuffix() string {
	s := p.get("bulk_template_suffix")
	if s == "" {
		return "_marketing"
	}
	return s
}

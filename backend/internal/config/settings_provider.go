package config

import (
	"log"
	"os"
	"mi-tech/internal/repository"
)

type SettingsProvider struct {
	configsRepo *repository.ConfigsRepository
}

func NewSettingsProvider(configsRepo *repository.ConfigsRepository) *SettingsProvider {
	return &SettingsProvider{configsRepo: configsRepo}
}

func (p *SettingsProvider) Get(key string) string {
	val, err := p.configsRepo.Get(key)
	if err != nil {
		log.Printf("ERROR: failed to get config for key %s: %v", key, err)
	}
	return val
}

func (p *SettingsProvider) GetShopifyStoreURL() string {
	return p.Get("shopify_store_url")
}

func (p *SettingsProvider) GetShopifyAccessToken() string {
	return p.Get("shopify_access_token")
}

func (p *SettingsProvider) GetShopifyWebhookSecret() string {
	return p.Get("shopify_webhook_secret")
}

func (p *SettingsProvider) GetWhatsAppPhoneNumberID() string {
	return p.Get("whatsapp_phone_number_id")
}

func (p *SettingsProvider) GetWhatsAppAccessToken() string {
	return p.GetMetaSystemUserToken()
}

func (p *SettingsProvider) GetWhatsAppWABAID() string {
	return p.Get("whatsapp_waba_id")
}

func (p *SettingsProvider) GetWhatsAppAppID() string {
	return p.Get("whatsapp_app_id")
}

func (p *SettingsProvider) GetWhatsAppAppSecret() string {
	return p.Get("whatsapp_app_secret")
}

func (p *SettingsProvider) GetJWTSecret() string {
	// Fallback to environment variable if not in DB
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		return secret
	}
	return p.Get("jwt_secret")
}

func (p *SettingsProvider) GetShopifyAPIVersion() string {
	v := p.Get("shopify_api_version")
	if v == "" {
		return "2026-01"
	}
	return v
}

func (p *SettingsProvider) GetWhatsAppWebhookVerifyToken() string {
	return p.Get("whatsapp_webhook_verify_token")
}

func (p *SettingsProvider) IsPIIProtectionEnabled() bool {
	return p.Get("pii_protection") == "true"
}

func (p *SettingsProvider) ShouldSendInvoice() bool {
	return p.Get("send_invoice") == "true"
}

func (p *SettingsProvider) GetBusinessName() string {
	return p.Get("business_name")
}

func (p *SettingsProvider) GetBusinessGSTIN() string {
	return p.Get("business_gstin")
}

func (p *SettingsProvider) GetBusinessAddressLine1() string {
	return p.Get("business_address_line1")
}

func (p *SettingsProvider) GetBusinessAddressLine2() string {
	return p.Get("business_address_line2")
}

func (p *SettingsProvider) GetBusinessPhone() string {
	return p.Get("business_phone")
}

func (p *SettingsProvider) GetBulkTemplateSuffix() string {
	s := p.Get("bulk_template_suffix")
	if s == "" {
		return "_marketing"
	}
	return s
}

func (p *SettingsProvider) GetMetaSystemUserToken() string {
	return p.Get("meta_system_user_token")
}

func (p *SettingsProvider) GetMetaMarketingAccessToken() string {
	return p.GetMetaSystemUserToken()
}

func (p *SettingsProvider) GetMetaMarketingAdAccountID() string {
	return p.Get("meta_marketing_ad_account_id")
}

func (p *SettingsProvider) GetMetaMarketingWebhookVerifyToken() string {
	return p.Get("meta_marketing_webhook_verify_token")
}

func (p *SettingsProvider) GetMetaAppID() string {
	return p.Get("meta_app_id")
}

func (p *SettingsProvider) GetMetaAppSecret() string {
	return p.Get("meta_app_secret")
}

func (p *SettingsProvider) GetFacebookPageID() string {
	return p.Get("facebook_page_id")
}

func (p *SettingsProvider) GetInstagramBusinessID() string {
	return p.Get("instagram_business_id")
}

func (p *SettingsProvider) GetThreadsUserID() string {
	return p.Get("threads_user_id")
}

func (p *SettingsProvider) GetFeedbackBaseURL() string {
	val := p.Get("feedback_base_url")
	if val == "" {
		return "https://feedback-form.millennialperfumer.in"
	}
	return val
}

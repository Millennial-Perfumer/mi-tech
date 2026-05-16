package config

import (
	"fmt"
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
	if p == nil || p.configsRepo == nil {
		return ""
	}
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

func (p *SettingsProvider) GetShopifyLocationID() string {
	return p.Get("shopify_location_id")
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

func (p *SettingsProvider) GetFeedbackExpiryMinutes() int {
	val := p.Get("feedback_link_expiry_minutes")
	if val == "" {
		return 2880 // 48 hours
	}
	var mins int
	fmt.Sscanf(val, "%d", &mins)
	if mins <= 0 {
		return 2880
	}
	return mins
}

func (p *SettingsProvider) GetFeedbackAutomationDelayMinutes() int {
	val := p.Get("feedback_automation_delay_minutes")
	if val == "" {
		return 60 // 1 hour
	}
	var mins int
	fmt.Sscanf(val, "%d", &mins)
	if mins <= 0 {
		return 60
	}
	return mins
}

func (p *SettingsProvider) GetAmazonLWAClientID() string {
	val := p.Get("amazon_lwa_client_id")
	if val == "" {
		return os.Getenv("AMAZON_LWA_CLIENT_ID")
	}
	return val
}

func (p *SettingsProvider) GetAmazonLWAClientSecret() string {
	val := p.Get("amazon_lwa_client_secret")
	if val == "" {
		return os.Getenv("AMAZON_LWA_CLIENT_SECRET")
	}
	return val
}

func (p *SettingsProvider) GetAmazonLWARefreshToken() string {
	val := p.Get("amazon_lwa_refresh_token")
	if val == "" {
		return os.Getenv("AMAZON_LWA_REFRESH_TOKEN")
	}
	return val
}

func (p *SettingsProvider) GetAmazonAWSAccessKey() string {
	val := p.Get("amazon_aws_access_key")
	if val == "" {
		return os.Getenv("AMAZON_AWS_ACCESS_KEY")
	}
	return val
}

func (p *SettingsProvider) GetAmazonAWSSecretKey() string {
	val := p.Get("amazon_aws_secret_key")
	if val == "" {
		return os.Getenv("AMAZON_AWS_SECRET_KEY")
	}
	return val
}

func (p *SettingsProvider) GetAmazonAWSRegion() string {
	val := p.Get("amazon_aws_region")
	if val == "" {
		val = os.Getenv("AMAZON_AWS_REGION")
	}
	if val == "" {
		return "eu-west-1"
	}
	return val
}

func (p *SettingsProvider) GetAmazonAWSRoleARN() string {
	val := p.Get("amazon_aws_role_arn")
	if val == "" {
		return os.Getenv("AMAZON_AWS_ROLE_ARN")
	}
	return val
}

func (p *SettingsProvider) GetAmazonMarketplaceID() string {
	val := p.Get("amazon_marketplace_id")
	if val == "" {
		val = os.Getenv("AMAZON_MARKETPLACE_ID")
	}
	if val == "" {
		return "A21TJRUUN4KGV"
	}
	return val
}

func (p *SettingsProvider) GetAmazonSellerID() string {
	val := p.Get("amazon_seller_id")
	if val == "" {
		return os.Getenv("AMAZON_SELLER_ID")
	}
	return val
}

func (p *SettingsProvider) IsInventorySyncEnabled() bool {
	val := p.Get("enable_inventory_sync")
	if val == "" {
		return true // Default to true if not explicitly set
	}
	return val == "true"
}

func (p *SettingsProvider) GetAIProvider() string {
	val := p.Get("ai_provider")
	if val == "" {
		return "cloud"
	}
	return val
}

func (p *SettingsProvider) IsAIEnabled() bool {
	return p.Get("ai_enabled") == "true"
}

func (p *SettingsProvider) GetOpenAIAPIKey() string {
	return p.Get("openai_api_key")
}

func (p *SettingsProvider) GetAICloudModel() string {
	val := p.Get("ai_cloud_model")
	if val == "" {
		return "gpt-5.4-nano"
	}
	return val
}

func (p *SettingsProvider) GetAILocalURL() string {
	val := p.Get("ai_local_url")
	if val == "" {
		return "http://localhost:11434"
	}
	return val
}

func (p *SettingsProvider) GetAILocalModel() string {
	val := p.Get("ai_local_model")
	if val == "" {
		return "gemma4"
	}
	return val
}

package marketing

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mi-tech/internal/config"
	"net/http"
	"net/url"
	"strings"
)

type MetaMarketingClient struct {
	settings *config.SettingsProvider
	baseURL  string
	version  string
}

func NewMetaMarketingClient(settings *config.SettingsProvider) *MetaMarketingClient {
	return &MetaMarketingClient{
		settings: settings,
		baseURL:  "https://graph.facebook.com",
		version:  "v19.0",
	}
}

type AdAccount struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Campaign struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Status          string `json:"status"`
	EffectiveStatus string `json:"effective_status"`
	Objective       string `json:"objective"`
	DailyBudget     string `json:"daily_budget"`
	LifetimeBudget  string `json:"lifetime_budget"`
}

type Insight struct {
	CampaignID     string        `json:"campaign_id"`
	AdSetID        string        `json:"adset_id"`
	AdID           string        `json:"ad_id"`
	Spend          string        `json:"spend"`
	Reach          string        `json:"reach"`
	Impressions    string        `json:"impressions"`
	Clicks         string        `json:"inline_link_clicks"`
	CTR            string        `json:"ctr"`
	CPC            string        `json:"cpc"`
	CPM            string        `json:"cpm"`
	ROAS           []ActionRate  `json:"purchase_roas"`
	Frequency      string        `json:"frequency"`
	Actions        []ActionValue `json:"actions"`
	ActionValues   []ActionValue `json:"action_values"`
	PurchaseValue  string        `json:"purchase_value"`
	Conversions    string        `json:"conversions"`
	AggregatedROAS string        `json:"purchase_roas_val"`

	// New Categorized Metrics (Meta returns these as action arrays frequently)
	QualityRanking          string `json:"quality_ranking"`
	EngagementRateRanking  string `json:"engagement_rate_ranking"`
	ConversionRateRanking string `json:"conversion_rate_ranking"`

	// Video (Raw Arrays)
	VideoP25Raw    []ActionValue `json:"video_p25_watched_actions"`
	VideoP50Raw    []ActionValue `json:"video_p50_watched_actions"`
	VideoP75Raw    []ActionValue `json:"video_p75_watched_actions"`
	VideoP100Raw   []ActionValue `json:"video_p100_watched_actions"`
	VideoAvgTimeRaw []ActionValue `json:"video_avg_time_watched_actions"`
	
	// Video (Flat values for frontend)
	VideoP25    string `json:"video_p25_val"`
	VideoP50    string `json:"video_p50_val"`
	VideoP75    string `json:"video_p75_val"`
	VideoP100   string `json:"video_p100_val"`
	VideoAvgTime string `json:"video_avg_time_val"`

	// Social Engagement
	PostEngagementRaw []ActionValue `json:"post_engagement"`
	PostReactionsRaw  []ActionValue `json:"post_reaction"`
	PostCommentsRaw   []ActionValue `json:"post_comment"`
	PostSharesRaw     []ActionValue `json:"post_share"`
	PageLikesRaw      []ActionValue `json:"page_like"`

	// Social Engagement (Flat values for frontend)
	PostEngagement string `json:"post_engagement_val"`
	PostReactions  string `json:"post_reaction_val"`
	PostComments   string `json:"post_comment_val"`
	PostShares     string `json:"post_share_val"`
	PageLikes      string `json:"page_like_val"`
}

type ActionValue struct {
	ActionType string `json:"action_type"`
	Value      string `json:"value"`
}

type ActionRate struct {
	ActionType string `json:"action_type"`
	Value      string `json:"value"`
}

type AdSet struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Status          string `json:"status"`
	EffectiveStatus string `json:"effective_status"`
	DailyBudget     string `json:"daily_budget"`
	LifetimeBudget  string `json:"lifetime_budget"`
	BidStrategy     string `json:"bid_strategy"`
}

type Ad struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	Status          string     `json:"status"`
	EffectiveStatus string     `json:"effective_status"`
	AdSetID         string     `json:"adset_id"`
	Creative        AdCreative `json:"creative"`
}

type AdCreative struct {
	ID           string `json:"id"`
	ThumbnailURL string `json:"thumbnail_url"`
}

func (c *MetaMarketingClient) getToken() (string, bool) {
	token := c.settings.GetMetaMarketingAccessToken()
	return token, token != ""
}

func (c *MetaMarketingClient) GetConfiguredAdAccountID() string {
	return c.settings.GetMetaMarketingAdAccountID()
}

func (c *MetaMarketingClient) normalizeAdAccountID(id string) string {
	if len(id) > 4 && id[:4] == "act_" {
		return id
	}
	return "act_" + id
}

func (c *MetaMarketingClient) FetchAdAccounts() ([]AdAccount, error) {
	token, ok := c.getToken()
	if !ok {
		return nil, fmt.Errorf("Meta Marketing access token not configured")
	}

	u := fmt.Sprintf("%s/%s/me/adaccounts?access_token=%s&fields=id,name", c.baseURL, c.version, token)
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Meta API error (%d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data []AdAccount `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

func (c *MetaMarketingClient) FetchCampaigns(adAccountID string) ([]Campaign, error) {
	token, ok := c.getToken()
	if !ok {
		return nil, fmt.Errorf("Meta Marketing access token not configured")
	}

	normalizedID := c.normalizeAdAccountID(adAccountID)
	u := fmt.Sprintf("%s/%s/%s/campaigns?access_token=%s&fields=id,name,status,effective_status,objective,daily_budget,lifetime_budget", c.baseURL, c.version, normalizedID, token)
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Printf("DEBUG: Meta API Campaigns Response for %s: %s", normalizedID, string(body))

	var result struct {
		Data []Campaign `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

func (c *MetaMarketingClient) FetchAdSets(campaignID string) ([]AdSet, error) {
	token, ok := c.getToken()
	if !ok {
		return nil, fmt.Errorf("Meta Marketing access token not configured")
	}

	u := fmt.Sprintf("%s/%s/%s/adsets?access_token=%s&fields=id,name,status,effective_status,daily_budget,lifetime_budget,bid_strategy", c.baseURL, c.version, campaignID, token)
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Printf("DEBUG: Meta API AdSets Response for %s: %s", campaignID, string(body))

	var result struct {
		Data []AdSet `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

func (c *MetaMarketingClient) FetchAds(adSetID string) ([]Ad, error) {
	token, ok := c.getToken()
	if !ok {
		return nil, fmt.Errorf("Meta Marketing access token not configured")
	}

	u := fmt.Sprintf("%s/%s/%s/ads?access_token=%s&fields=id,name,status,effective_status,adset_id,creative{id,thumbnail_url}", c.baseURL, c.version, adSetID, token)
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Printf("DEBUG: Meta API Ads Response for %s: %s", adSetID, string(body))

	var result struct {
		Data []Ad `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

func (c *MetaMarketingClient) FetchInsights(id string, level string, startDate, endDate string) ([]Insight, error) {
	token, ok := c.getToken()
	if !ok {
		return nil, fmt.Errorf("Meta Marketing access token not configured")
	}

	params := url.Values{}
	params.Add("access_token", token)
	params.Add("level", level)
	
	if startDate != "" && endDate != "" {
		timeRange := fmt.Sprintf("{\"since\":\"%s\",\"until\":\"%s\"}", startDate, endDate)
		params.Add("time_range", timeRange)
	} else {
		params.Add("date_preset", "last_30d")
	}

	params.Add("fields", "campaign_id,adset_id,ad_id,spend,reach,impressions,inline_link_clicks,ctr,cpc,cpm,purchase_roas,frequency,action_values,actions,quality_ranking,engagement_rate_ranking,conversion_rate_ranking,video_p25_watched_actions,video_p50_watched_actions,video_p75_watched_actions,video_p100_watched_actions,video_avg_time_watched_actions")

	// Normalize ID if it's an ad account ID (entirely digits or missing act_)
	normalizedID := id
	if level == "account" || level == "campaign" || level == "adset" || level == "ad" {
		// If we are calling insights on an account node, it must have 'act_'
		// We can check if the ID provided is the configured account ID
		if id == c.GetConfiguredAdAccountID() || !strings.HasPrefix(id, "act_") {
			// Only prefix if it looks like it might be an account ID (long digit string)
			// Actually, just use the helper
			if level == "campaign" && id == c.GetConfiguredAdAccountID() {
				normalizedID = c.normalizeAdAccountID(id)
			}
		}
	}

	u := fmt.Sprintf("%s/%s/%s/insights?%s", c.baseURL, c.version, normalizedID, params.Encode())
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		log.Printf("ERROR: Meta API Insights (%s) for %s returned %d: %s", level, normalizedID, resp.StatusCode, string(body))
		return nil, fmt.Errorf("Meta API error (%d): %s", resp.StatusCode, string(body))
	}
	
	log.Printf("DEBUG: Meta API Insights (%s) for %s: %s", level, normalizedID, string(body))

	var result struct {
		Data []Insight `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	// Post-process to extract purchase values
	for i := range result.Data {
		// Extract Conversions
		for _, action := range result.Data[i].Actions {
			if action.ActionType == "purchase" || action.ActionType == "omni_purchase" || action.ActionType == "offsite_conversion.fb_pixel_purchase" {
				result.Data[i].Conversions = action.Value
				break
			}
		}
		// Extract ROAS
		for _, roas := range result.Data[i].ROAS {
			if roas.ActionType == "purchase" || roas.ActionType == "omni_purchase" || roas.ActionType == "offsite_conversion.fb_pixel_purchase" {
				result.Data[i].AggregatedROAS = roas.Value
				break
			}
		}
		// Extract Purchase Value
		for _, v := range result.Data[i].ActionValues {
			if v.ActionType == "purchase" || v.ActionType == "omni_purchase" || v.ActionType == "offsite_conversion.fb_pixel_purchase" {
				result.Data[i].PurchaseValue = v.Value
				break
			}
		}

		// Extract Video Funnel
		if len(result.Data[i].VideoP25Raw) > 0 {
			result.Data[i].VideoP25 = result.Data[i].VideoP25Raw[0].Value
		}
		if len(result.Data[i].VideoP50Raw) > 0 {
			result.Data[i].VideoP50 = result.Data[i].VideoP50Raw[0].Value
		}
		if len(result.Data[i].VideoP100Raw) > 0 {
			result.Data[i].VideoP100 = result.Data[i].VideoP100Raw[0].Value
		}
		if len(result.Data[i].VideoAvgTimeRaw) > 0 {
			result.Data[i].VideoAvgTime = result.Data[i].VideoAvgTimeRaw[0].Value
		}

		// Extract Social Engagement
		if len(result.Data[i].PostEngagementRaw) > 0 {
			result.Data[i].PostEngagement = result.Data[i].PostEngagementRaw[0].Value
		}
		if len(result.Data[i].PostReactionsRaw) > 0 {
			result.Data[i].PostReactions = result.Data[i].PostReactionsRaw[0].Value
		}
		if len(result.Data[i].PostCommentsRaw) > 0 {
			result.Data[i].PostComments = result.Data[i].PostCommentsRaw[0].Value
		}
	}

	return result.Data, nil
}

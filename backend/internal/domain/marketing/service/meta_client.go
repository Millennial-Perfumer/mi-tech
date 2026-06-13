package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mi-tech/internal/shared/config"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type MetaMarketingClient struct {
	settings             *config.SettingsProvider
	baseURL              string
	version              string
	pageAccessTokenCache map[string]string
	cacheMu              sync.RWMutex
	client               *http.Client
}

type MetaError struct {
	Message      string `json:"message"`
	Type         string `json:"type"`
	Code         int    `json:"code"`
	ErrorSubcode int    `json:"error_subcode"`
	FBTraceID    string `json:"fbtrace_id"`
}

func (e *MetaError) Error() string {
	return fmt.Sprintf("Meta API Error (%d): %s [Trace: %s]", e.Code, e.Message, e.FBTraceID)
}

type metaResponse struct {
	Data  json.RawMessage `json:"data"`
	Error *MetaError      `json:"error"`
}

func decodeResponse(resp *http.Response, target interface{}) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var mResp metaResponse
	if err := json.Unmarshal(body, &mResp); err != nil {
		return fmt.Errorf("failed to unmarshal meta response: %w", err)
	}

	if mResp.Error != nil {
		log.Printf("LOUD META ERROR: %s", string(body))
		return mResp.Error
	}

	if target != nil {
		if err := json.Unmarshal(mResp.Data, target); err != nil {
			// If it's not a 'data' array, it might be a direct object (like insights)
			// Try to unmarshal the whole body into target
			return json.Unmarshal(body, target)
		}
	}

	return nil
}

func NewMetaMarketingClient(settings *config.SettingsProvider) *MetaMarketingClient {
	return &MetaMarketingClient{
		settings:             settings,
		baseURL:              "https://graph.facebook.com",
		version:              "v22.0", // Upgraded to v22.0 for latest SMM/Threads support
		pageAccessTokenCache: make(map[string]string),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type Page struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type InstagramUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type ThreadsUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type SocialPost struct {
	ID           string `json:"id"`
	Caption      string `json:"caption,omitempty"`
	Message      string `json:"message,omitempty"`
	Permalink    string `json:"permalink"`
	Timestamp    string `json:"timestamp"`
	MediaType    string `json:"media_type"`
	MediaURL     string `json:"media_url,omitempty"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
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
	QualityRanking        string `json:"quality_ranking"`
	EngagementRateRanking string `json:"engagement_rate_ranking"`
	ConversionRateRanking string `json:"conversion_rate_ranking"`

	// Video (Raw Arrays)
	VideoP25Raw     []ActionValue `json:"video_p25_watched_actions"`
	VideoP50Raw     []ActionValue `json:"video_p50_watched_actions"`
	VideoP75Raw     []ActionValue `json:"video_p75_watched_actions"`
	VideoP100Raw    []ActionValue `json:"video_p100_watched_actions"`
	VideoAvgTimeRaw []ActionValue `json:"video_avg_time_watched_actions"`

	// Video (Flat values for frontend)
	VideoP25     string `json:"video_p25_val"`
	VideoP50     string `json:"video_p50_val"`
	VideoP75     string `json:"video_p75_val"`
	VideoP100    string `json:"video_p100_val"`
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

func (c *MetaMarketingClient) GetPageAccessToken(pageID string) (string, error) {
	// 1. Check Cache
	c.cacheMu.RLock()
	cached, exists := c.pageAccessTokenCache[pageID]
	c.cacheMu.RUnlock()
	if exists {
		return cached, nil
	}

	// 2. Fetch from Meta
	masterToken, ok := c.getToken()
	if !ok {
		return "", fmt.Errorf("master Meta access token not configured")
	}

	// We use /me/accounts to get all pages and their specific tokens
	u := fmt.Sprintf("%s/%s/me/accounts?access_token=%s&fields=id,access_token", c.baseURL, c.version, masterToken)
	resp, err := c.client.Get(u)
	if err != nil {
		return "", fmt.Errorf("failed to fetch page accounts: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("meta error fetching accounts: status %d", resp.StatusCode)
	}

	var result struct {
		Data []struct {
			ID          string `json:"id"`
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode page accounts: %w", err)
	}

	// 3. Find and Cache
	var pageToken string
	c.cacheMu.Lock()
	for _, p := range result.Data {
		c.pageAccessTokenCache[p.ID] = p.AccessToken
		if p.ID == pageID {
			pageToken = p.AccessToken
		}
	}
	c.cacheMu.Unlock()

	if pageToken == "" {
		// DIAGNOSTIC LOGGING: This helps the user identify which Pages DO have tokens
		var foundIDs []string
		for _, p := range result.Data {
			foundIDs = append(foundIDs, p.ID)
		}
		log.Printf("ERROR: Page ID %s not found in user authorized accounts. Available Page IDs: %v", pageID, foundIDs)
		return "", fmt.Errorf("page ID %s not found in user accounts. Check META_FACEBOOK_PAGE_ID in .env", pageID)
	}

	return pageToken, nil
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

func (c *MetaMarketingClient) GetConfiguredPageID() string {
	return c.settings.GetFacebookPageID()
}

func (c *MetaMarketingClient) GetConfiguredIGID() string {
	return c.settings.GetInstagramBusinessID()
}

func (c *MetaMarketingClient) GetConfiguredThreadsID() string {
	return c.settings.GetThreadsUserID()
}

func (c *MetaMarketingClient) FetchAdAccounts() ([]AdAccount, error) {
	token, ok := c.getToken()
	if !ok {
		return nil, fmt.Errorf("Meta Marketing access token not configured")
	}

	u := fmt.Sprintf("%s/%s/me/adaccounts?access_token=%s&fields=id,name", c.baseURL, c.version, token)
	resp, err := c.client.Get(u)
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
	resp, err := c.client.Get(u)
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
	resp, err := c.client.Get(u)
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
	resp, err := c.client.Get(u)
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
	resp, err := c.client.Get(u)
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

// --- Social Media Management (SMM) Extensions ---

func (c *MetaMarketingClient) FetchPages() ([]Page, error) {
	token, ok := c.getToken()
	if !ok {
		return nil, fmt.Errorf("Meta access token not configured")
	}

	u := fmt.Sprintf("%s/%s/me/accounts?access_token=%s&fields=id,name", c.baseURL, c.version, token)
	resp, err := c.client.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data []Page `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

func (c *MetaMarketingClient) FetchInstagramAccounts(pageID string) ([]InstagramUser, error) {
	token, ok := c.getToken()
	if !ok {
		return nil, fmt.Errorf("Meta access token not configured")
	}

	u := fmt.Sprintf("%s/%s/%s?fields=instagram_business_account{id,username}&access_token=%s", c.baseURL, c.version, pageID, token)
	resp, err := c.client.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		InstagramBusinessAccount InstagramUser `json:"instagram_business_account"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if result.InstagramBusinessAccount.ID == "" {
		return []InstagramUser{}, nil
	}
	return []InstagramUser{result.InstagramBusinessAccount}, nil
}

func (c *MetaMarketingClient) PostToFacebookPage(pageID, message string) (string, error) {
	token, ok := c.getToken()
	if !ok {
		return "", fmt.Errorf("Meta access token not configured")
	}

	u := fmt.Sprintf("%s/%s/%s/feed", c.baseURL, c.version, pageID)
	params := url.Values{}
	params.Add("message", message)
	params.Add("access_token", token)

	resp, err := http.PostForm(u, params)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.ID, nil
}

// PostToInstagram implements the 2-step IG posting flow
func (c *MetaMarketingClient) PostToInstagram(igUserID, imageURL, caption string) (string, error) {
	token, ok := c.getToken()
	if !ok {
		return "", fmt.Errorf("Meta access token not configured")
	}

	// Step 1: Create Media Container
	u1 := fmt.Sprintf("%s/%s/%s/media", c.baseURL, c.version, igUserID)
	params1 := url.Values{}
	params1.Add("image_url", imageURL)
	params1.Add("caption", caption)
	params1.Add("access_token", token)

	resp1, err := c.client.PostForm(u1, params1)
	if err != nil {
		return "", err
	}
	defer resp1.Body.Close()

	var res1 struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp1.Body).Decode(&res1); err != nil {
		return "", err
	}

	// Step 2: Publish Container
	u2 := fmt.Sprintf("%s/%s/%s/media_publish", c.baseURL, c.version, igUserID)
	params2 := url.Values{}
	params2.Add("creation_id", res1.ID)
	params2.Add("access_token", token)

	resp2, err := c.client.PostForm(u2, params2)
	if err != nil {
		return "", err
	}
	defer resp2.Body.Close()

	var res2 struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp2.Body).Decode(&res2); err != nil {
		return "", err
	}

	return res2.ID, nil
}

func (c *MetaMarketingClient) PostToThreads(threadsUserID, text string) (string, error) {
	token, ok := c.getToken()
	if !ok {
		return "", fmt.Errorf("Meta access token not configured")
	}

	// Step 1: Create Container
	u1 := fmt.Sprintf("%s/%s/%s/threads", c.baseURL, c.version, threadsUserID)
	params1 := url.Values{}
	params1.Add("media_type", "TEXT")
	params1.Add("text", text)
	params1.Add("access_token", token)

	resp1, err := c.client.PostForm(u1, params1)
	if err != nil {
		return "", err
	}
	defer resp1.Body.Close()

	var res1 struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp1.Body).Decode(&res1); err != nil {
		return "", err
	}

	// Step 2: Publish
	u2 := fmt.Sprintf("%s/%s/%s/threads_publish", c.baseURL, c.version, threadsUserID)
	params2 := url.Values{}
	params2.Add("creation_id", res1.ID)
	params2.Add("access_token", token)

	resp2, err := c.client.PostForm(u2, params2)
	if err != nil {
		return "", err
	}
	defer resp2.Body.Close()

	var res2 struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp2.Body).Decode(&res2); err != nil {
		return "", err
	}

	return res2.ID, nil
}

func (c *MetaMarketingClient) FetchInstagramInsights(igUserID string, since, until string) (map[string]int, error) {
	token, ok := c.getToken()
	if !ok {
		return nil, fmt.Errorf("Meta access token not configured")
	}

	u := fmt.Sprintf("%s/%s/%s/insights?metric=reach,impressions,views,total_interactions&metric_type=total_value&access_token=%s", c.baseURL, c.version, igUserID, token)
	if since != "" && until != "" {
		sTime, _ := time.Parse("2006-01-02", since)
		uTime, _ := time.Parse("2006-01-02", until)
		// Meta insights since/until are exclusive on until.
		// Add 24 hours to include the entire 'until' day (Graph API v22.0)
		u += fmt.Sprintf("&since=%d&until=%d", sTime.Unix(), uTime.Add(24*time.Hour).Unix())
	}

	resp, err := c.client.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result []struct {
		Name   string `json:"name"`
		Values []struct {
			Value int `json:"value"`
		} `json:"values"`
		TotalValue struct {
			Value int `json:"value"`
		} `json:"total_value"`
	}
	if err := decodeResponse(resp, &result); err != nil {
		return nil, err
	}

	metrics := make(map[string]int)
	for _, d := range result {
		total := 0
		if d.TotalValue.Value > 0 {
			total = d.TotalValue.Value
		} else {
			for _, v := range d.Values {
				total += v.Value
			}
		}
		metrics[d.Name] = total
	}

	// v22.0 Fallback Logic: Impressions vs Views
	if metrics["views"] == 0 && metrics["impressions"] > 0 {
		metrics["views"] = metrics["impressions"]
	}

	return metrics, nil
}

func (c *MetaMarketingClient) FetchInstagramMedia(igUserID string) ([]SocialPost, error) {
	token, ok := c.getToken()
	if !ok {
		return nil, fmt.Errorf("Meta access token not configured")
	}

	// Limit 100 to ensure we capture all in range for most users
	u := fmt.Sprintf("%s/%s/%s/media?fields=id,caption,permalink,timestamp,media_type,media_url,thumbnail_url&limit=100&access_token=%s", c.baseURL, c.version, igUserID, token)
	resp, err := c.client.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result []SocialPost
	if err := decodeResponse(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *MetaMarketingClient) FetchFacebookPageMedia(pageID string) ([]SocialPost, error) {
	token, err := c.GetPageAccessToken(pageID)
	if err != nil {
		log.Printf("WARNING: GetPageAccessToken failed for %s, falling back to master: %v", pageID, err)
		master, ok := c.getToken()
		if !ok {
			return nil, fmt.Errorf("no tokens available")
		}
		token = master
	}

	// Fetch page posts with standard fields
	u := fmt.Sprintf("%s/%s/%s/posts?fields=id,message,permalink_url,created_time,full_picture&limit=100&access_token=%s", c.baseURL, c.version, pageID, token)
	resp, err := c.client.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result []struct {
		ID          string `json:"id"`
		Message     string `json:"message"`
		Permalink   string `json:"permalink_url"`
		CreatedTime string `json:"created_time"`
		FullPicture string `json:"full_picture"`
	}

	if err := decodeResponse(resp, &result); err != nil {
		return nil, err
	}

	var posts []SocialPost
	for _, p := range result {
		posts = append(posts, SocialPost{
			ID:        p.ID,
			Message:   p.Message,
			MediaURL:  p.FullPicture,
			Permalink: p.Permalink,
			Timestamp: p.CreatedTime,
			MediaType: "FB_POST",
		})
	}
	return posts, nil
}

func (c *MetaMarketingClient) FetchFacebookPostInsights(postID string) (map[string]int, error) {
	// For post-level insights, we MUST use the Page ID to get a Page Access Token
	parts := strings.Split(postID, "_")
	var token string
	if len(parts) > 0 {
		pageID := parts[0]
		pageToken, err := c.GetPageAccessToken(pageID)
		if err == nil {
			token = pageToken
		} else {
			log.Printf("ERROR: FetchFacebookPostInsights could not get Page Token for %s: %v", pageID, err)
		}
	}

	if token == "" {
		master, ok := c.getToken()
		if !ok {
			return nil, fmt.Errorf("Meta access token not configured")
		}
		token = master
	}

	// v22.0 Migration: Facebook 'New Pages Experience' deprecates legacy reach metrics.
	// We iteratively try metric candidates until one succeeds.
	candidates := []string{
		"post_reach,post_engagements",                // Modern v22.0 standard
		"post_impressions_unique,post_engaged_users", // Classic (fallback)
		"impressions,reach",                          // Basic Metadata (last resort)
	}

	var lastErr error
	var finalMetrics map[string]int

	for _, metricSet := range candidates {
		u := fmt.Sprintf("%s/%s/%s/insights?metric=%s&access_token=%s", c.baseURL, c.version, postID, metricSet, token)
		resp, err := c.client.Get(u)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		if strings.Contains(string(body), "\"error\"") {
			// If it's a #100 'invalid metric' error, try the next set
			if strings.Contains(string(body), "\"code\":100") {
				log.Printf("DEBUG: Metric set [%s] rejected for %s, trying next candidate...", metricSet, postID)
				continue
			}
			lastErr = fmt.Errorf("Meta API Error: %s", string(body))
			continue
		}

		// Success! Parse results
		var result struct {
			Data []struct {
				Name   string `json:"name"`
				Values []struct {
					Value int `json:"value"`
				} `json:"values"`
			} `json:"data"`
		}
		if err := json.Unmarshal(body, &result); err != nil {
			lastErr = err
			continue
		}

		log.Printf("DEBUG: Metric set [%s] successful for Facebook Post %s", metricSet, postID)
		finalMetrics = make(map[string]int)
		for _, d := range result.Data {
			if len(d.Values) > 0 {
				name := d.Name
				// Map modern names to dashboard expectations
				if name == "post_reach" || name == "reach" {
					name = "post_impressions_unique"
				}
				if name == "post_engagements" {
					name = "post_engaged_users"
				}
				finalMetrics[name] = d.Values[0].Value
			}
		}
		break
	}

	if finalMetrics == nil {
		return nil, fmt.Errorf("all metric sets failed for Facebook post %s: %v", postID, lastErr)
	}
	return finalMetrics, nil
}

func (c *MetaMarketingClient) FetchMediaInsights(mediaID string, mediaType string) (map[string]int, error) {
	token, ok := c.getToken()
	if !ok {
		return nil, fmt.Errorf("Meta access token not configured")
	}

	// v22.0 Migration: Request character-perfect supported metrics
	// images/videos use total_interactions; reels use reach/plays.
	metricsStr := "reach,saved,total_interactions"
	if mediaType == "VIDEO" {
		metricsStr = "reach,saved,total_interactions,video_views"
	}

	u := fmt.Sprintf("%s/%s/%s/insights?metric=%s&access_token=%s", c.baseURL, c.version, mediaID, metricsStr, token)
	resp, err := c.client.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// v22.0 Bugfix: Structure MUST have 'data' wrapper
	var res struct {
		Data []struct {
			Name   string `json:"name"`
			Values []struct {
				Value int `json:"value"`
			} `json:"values"`
		} `json:"data"`
	}

	if err := decodeResponse(resp, &res); err != nil {
		return nil, err
	}

	metrics := make(map[string]int)
	for _, d := range res.Data {
		if len(d.Values) > 0 {
			name := d.Name
			// Map v22.0 'total_interactions' to service 'engagement'
			if name == "total_interactions" {
				name = "engagement"
			}
			if name == "video_views" {
				name = "views"
			}
			metrics[name] = d.Values[0].Value
		}
	}
	return metrics, nil
}

func (c *MetaMarketingClient) FetchPageInsights(pageID string, since, until string) (map[string]int, error) {
	token, err := c.GetPageAccessToken(pageID)
	if err != nil {
		log.Printf("ERROR: Facebook Insights node restricted: %v", err)
		// Facebook Insight metrics REQUIRE a page token. A User token will return 'subcode 33'.
		// We still try master to check if the user has direct 'pages_read_engagement' scope but log it as error
		master, ok := c.getToken()
		if !ok {
			return nil, fmt.Errorf("no tokens available")
		}
		token = master
	}

	// Fetch Daily Insights (Impressions, Engagements, Views)
	u := fmt.Sprintf("%s/%s/%s/insights?metric=page_impressions,page_post_engagements,page_views_total&period=day&access_token=%s", c.baseURL, c.version, pageID, token)
	if since != "" && until != "" {
		sTime, _ := time.Parse("2006-01-02", since)
		uTime, _ := time.Parse("2006-01-02", until)
		// Meta insights are exclusive on until. Add 24 hours to include the entire 'until' day.
		u += fmt.Sprintf("&since=%d&until=%d", sTime.Unix(), uTime.Add(24*time.Hour).Unix())
	}

	resp, err := c.client.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	// DISCOVERY MODE: Log the raw response and available metrics for this Page node
	// If the user sees 0s, these logs will provide the definitive list of supported metric strings.
	log.Printf("DEBUG: Facebook Page Insights v22.0 Raw Response: %s", string(body))

	if strings.Contains(string(body), "\"error\"") {
		// Attempt a discovery call to list available metrics for this specific Page Node
		uDiscovery := fmt.Sprintf("%s/%s/%s/insights?access_token=%s", c.baseURL, c.version, pageID, token)
		if respDisc, err := c.client.Get(uDiscovery); err == nil {
			defer respDisc.Body.Close()
			discBody, _ := io.ReadAll(respDisc.Body)
			log.Printf("DISCOVERY: Supported metrics for Page %s: %s", pageID, string(discBody))
		}
	}

	var result []struct {
		Name   string `json:"name"`
		Values []struct {
			Value int `json:"value"`
		} `json:"values"`
	}

	// Reset reader for decoding if needed, or just Unmarshal
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("ERROR: JSON Unmarshal failed for Page Insights: %v", err)
	}

	metrics := make(map[string]int)
	for _, d := range result {
		total := 0
		for _, v := range d.Values {
			total += v.Value
		}
		metrics[d.Name] = total
	}

	// Fetch Lifetime Fans (Follower Count)
	uFans := fmt.Sprintf("%s/%s/%s/insights?metric=page_fans&period=lifetime&access_token=%s", c.baseURL, c.version, pageID, token)
	respF, err := c.client.Get(uFans)
	if err == nil {
		defer respF.Body.Close()
		var resF []struct {
			Values []struct {
				Value int `json:"value"`
			} `json:"values"`
		}
		if decodeResponse(respF, &resF) == nil && len(resF) > 0 && len(resF[0].Values) > 0 {
			metrics["page_fans"] = resF[0].Values[len(resF[0].Values)-1].Value
		}
	}

	// v22.0 Fallback Logic: Page Views vs Impressions
	if metrics["page_views_total"] == 0 && metrics["page_impressions"] > 0 {
		metrics["page_views_total"] = metrics["page_impressions"]
	}

	return metrics, nil
}

// AccountInsights represents high-level metrics for an Instagram business account.
type AccountInsights struct {
	FollowerCount int            `json:"follower_count"`
	Reach         int            `json:"reach"`
	Views         int            `json:"views"`
	Breakdowns    map[string]int `json:"breakdowns"`
}

type DetailedInsights struct {
	Metrics    map[string]int            `json:"metrics"`
	Breakdowns map[string]map[string]int `json:"breakdowns"`
}

func (c *MetaMarketingClient) FetchDetailedMediaInsights(mediaID string, mediaType string) (*DetailedInsights, error) {
	token, ok := c.getToken()
	if !ok {
		return nil, fmt.Errorf("Meta access token not configured")
	}

	// 1. Fetch standard metrics
	standardMetrics := "reach,views,likes,comments,shares,saved,total_interactions"
	u := fmt.Sprintf("%s/%s/%s/insights?metric=%s&access_token=%s", c.baseURL, c.version, mediaID, standardMetrics, token)
	resp, err := c.client.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result []struct {
		Name   string `json:"name"`
		Values []struct {
			Value int `json:"value"`
		} `json:"values"`
	}
	if err := decodeResponse(resp, &result); err != nil {
		return nil, err
	}

	insights := &DetailedInsights{
		Metrics:    make(map[string]int),
		Breakdowns: make(map[string]map[string]int),
	}

	for _, d := range result {
		if len(d.Values) > 0 {
			insights.Metrics[d.Name] = d.Values[0].Value
		}
	}

	// 2. Try follower breakdowns
	// v22.0 Migration: 'reach' and 'views' are the primary breakdown targets.
	// Since 'reach' can take 24-48h to calculate, we prioritize it, but fall back to 'views'
	// to ensure some data is shown when reach is 0.
	breakdownMetrics := []string{"reach", "views", "total_interactions", "likes"}
	var bodyB []byte

	for _, metric := range breakdownMetrics {
		// Optimization: If reach is 0 in standard metrics, skip it in breakdown and try others
		if metric == "reach" && insights.Metrics["reach"] == 0 {
			continue
		}

		u2 := fmt.Sprintf("%s/%s/%s/insights?metric=%s&breakdown=follow_type&metric_type=total_value&access_token=%s",
			c.baseURL, c.version, mediaID, metric, token)
		resp2, err := http.Get(u2)
		if err != nil {
			continue
		}
		defer resp2.Body.Close()

		b, _ := io.ReadAll(resp2.Body)
		log.Printf("DEBUG: Breakdown response metric=%s for %s: %s", metric, mediaID, string(b))
		bodyB = b

		// If Meta returned an error for this metric, try the next one
		if strings.Contains(string(b), "\"error\"") {
			continue
		}

		// We got a potentially valid response
		// Check if it's the modern structure (total_value -> breakdowns -> results)
		if strings.Contains(string(b), "\"total_value\"") && strings.Contains(string(b), "\"breakdowns\"") {
			break
		}
	}

	if bodyB != nil {
		followBreakdown := make(map[string]int)

		// Pass 1: Total Value Structure (Nested Results) - Correct for modern Media Insights
		var totalValueResult struct {
			Data []struct {
				TotalValue struct {
					Breakdowns []struct {
						Results []struct {
							DimensionValues []string `json:"dimension_values"`
							Value           int      `json:"value"`
						} `json:"results"`
					} `json:"breakdowns"`
				} `json:"total_value"`
			} `json:"data"`
		}
		if err := json.Unmarshal(bodyB, &totalValueResult); err == nil && len(totalValueResult.Data) > 0 {
			for _, b := range totalValueResult.Data[0].TotalValue.Breakdowns {
				for _, r := range b.Results {
					if len(r.DimensionValues) > 0 {
						key := strings.ToUpper(r.DimensionValues[0])
						followBreakdown[key] = r.Value
					}
				}
			}
		}

		// Pass 2: Legacy Values-based Structure (Fallback)
		if len(followBreakdown) == 0 {
			var bResult struct {
				Data []struct {
					Values []struct {
						Value      int                 `json:"value"`
						Breakdowns []map[string]string `json:"breakdowns"`
					} `json:"values"`
				} `json:"data"`
			}
			if err := json.Unmarshal(bodyB, &bResult); err == nil && len(bResult.Data) > 0 {
				for _, v := range bResult.Data[0].Values {
					for _, br := range v.Breakdowns {
						if fType, ok := br["follow_type"]; ok {
							followBreakdown[strings.ToUpper(fType)] = v.Value
						}
					}
				}
			}
		}

		// Pass 3: Map-based Breakdowns (Alternative Fallback)
		if len(followBreakdown) == 0 {
			var bResultMap struct {
				Data []struct {
					Values []struct {
						Value      int               `json:"value"`
						Breakdowns map[string]string `json:"breakdowns"`
					} `json:"values"`
				} `json:"data"`
			}
			if err := json.Unmarshal(bodyB, &bResultMap); err == nil && len(bResultMap.Data) > 0 {
				for _, v := range bResultMap.Data[0].Values {
					if fType, ok := v.Breakdowns["follow_type"]; ok {
						followBreakdown[strings.ToUpper(fType)] = v.Value
					}
				}
			}
		}

		if len(followBreakdown) > 0 {
			insights.Breakdowns["follow_type"] = followBreakdown
		}
	}

	return insights, nil
}

type AssetHealth struct {
	PageAuthorized      bool   `json:"page_authorized"`
	InstagramLinked     bool   `json:"instagram_linked"`
	InstagramAuthorized bool   `json:"instagram_authorized"`
	Error               string `json:"error,omitempty"`
}

func (c *MetaMarketingClient) CheckAssetAlignment() (*AssetHealth, error) {
	token, ok := c.getToken()
	if !ok {
		return nil, fmt.Errorf("Meta access token not configured")
	}

	pageID := c.settings.GetFacebookPageID()
	igID := c.settings.GetInstagramBusinessID()
	health := &AssetHealth{}

	// 1. Check Page Authorization
	u1 := fmt.Sprintf("%s/%s/me/accounts?access_token=%s&fields=id", c.baseURL, c.version, token)
	resp1, err := c.client.Get(u1)
	if err == nil {
		defer resp1.Body.Close()
		var res1 struct {
			Data []struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		if json.NewDecoder(resp1.Body).Decode(&res1) == nil {
			for _, p := range res1.Data {
				if p.ID == pageID {
					health.PageAuthorized = true
					break
				}
			}
		}
	}

	// 2. Check IG Linkage to Page
	if pageID != "" {
		u2 := fmt.Sprintf("%s/%s/%s?fields=instagram_business_account{id}&access_token=%s", c.baseURL, c.version, pageID, token)
		resp2, err := c.client.Get(u2)
		if err == nil {
			defer resp2.Body.Close()
			var res2 struct {
				InstagramBusinessAccount struct {
					ID string `json:"id"`
				} `json:"instagram_business_account"`
			}
			if json.NewDecoder(resp2.Body).Decode(&res2) == nil {
				if res2.InstagramBusinessAccount.ID == igID {
					health.InstagramLinked = true
				}
			}
		}
	}

	// 3. Check IG Authorization directly
	if igID != "" {
		u3 := fmt.Sprintf("%s/%s/%s?fields=id,username&access_token=%s", c.baseURL, c.version, igID, token)
		resp3, err := c.client.Get(u3)
		if err == nil {
			defer resp3.Body.Close()
			if resp3.StatusCode == 200 {
				health.InstagramAuthorized = true
			} else {
				// Capture the LOUD error for direct visibility in health check
				body, _ := io.ReadAll(resp3.Body)
				health.Error = string(body)
			}
		}
	}

	return health, nil
}

// FetchAccountInsights retrieves high-level follower and reach stats for an IG user.
func (c *MetaMarketingClient) FetchAccountInsights(igUserID string, startDate, endDate string) (*AccountInsights, error) {
	token, ok := c.getToken()
	if !ok {
		return nil, fmt.Errorf("Meta access token not configured")
	}

	// 1. Fetch Follower Count (Field - Plural v22.0 requirement)
	u1 := fmt.Sprintf("%s/%s/%s?fields=followers_count&access_token=%s", c.baseURL, c.version, igUserID, token)
	results := &AccountInsights{
		Breakdowns: make(map[string]int),
	}
	resp1, err := c.client.Get(u1)
	if err == nil {
		defer resp1.Body.Close()
		var userResult struct {
			FollowersCount int `json:"followers_count"`
		}
		if err := decodeResponse(resp1, &userResult); err == nil {
			results.FollowerCount = userResult.FollowersCount
		}
	}

	// Fetch Reach (Periodic)
	uReach := fmt.Sprintf("%s/%s/%s/insights?metric=reach&period=day&access_token=%s", c.baseURL, c.version, igUserID, token)
	respR, err := c.client.Get(uReach)
	if err == nil {
		defer respR.Body.Close()
		var rRes struct {
			Data []struct {
				Values []struct {
					Value int `json:"value"`
				} `json:"values"`
			} `json:"data"`
		}
		if err := decodeResponse(respR, &rRes); err == nil && len(rRes.Data) > 0 {
			for _, item := range rRes.Data {
				for _, v := range item.Values {
					results.Reach += v.Value
				}
			}
		}
	}

	// Fetch Views (Total Value - v22.0 requirement)
	uViews := fmt.Sprintf("%s/%s/%s/insights?metric=views&metric_type=total_value&access_token=%s", c.baseURL, c.version, igUserID, token)
	respV, err := c.client.Get(uViews)
	if err == nil {
		defer respV.Body.Close()
		var vRes struct {
			Data []struct {
				TotalValue struct {
					Value int `json:"value"`
				} `json:"total_value"`
			} `json:"data"`
		}
		if err := decodeResponse(respV, &vRes); err == nil && len(vRes.Data) > 0 {
			results.Views = vRes.Data[0].TotalValue.Value
		}
	}

	// 3. Fetch Demographic Breakdowns (Follower vs Non-follower)
	sTime, _ := time.Parse("2006-01-02", startDate)
	uTime, _ := time.Parse("2006-01-02", endDate)
	uDemo := fmt.Sprintf("%s/%s/%s/insights?metric=reached_audience_demographics&breakdown=follow_type&metric_type=total_value&since=%d&until=%d&access_token=%s",
		c.baseURL, c.version, igUserID, sTime.Unix(), uTime.Add(24*time.Hour).Unix(), token)
	respD, err := c.client.Get(uDemo)
	if err == nil {
		defer respD.Body.Close()
		var dRes struct {
			Data []struct {
				Name       string `json:"name"`
				ID         string `json:"id"`
				TotalValue struct {
					Breakdowns []struct {
						Dimension string `json:"dimension"`
						Results   []struct {
							DimensionValues []string `json:"dimension_values"`
							Value           int      `json:"value"`
						} `json:"results"`
					} `json:"breakdowns"`
				} `json:"total_value"`
			} `json:"data"`
		}
		if err := decodeResponse(respD, &dRes); err == nil && len(dRes.Data) > 0 {
			for _, item := range dRes.Data {
				for _, b := range item.TotalValue.Breakdowns {
					if b.Dimension == "follow_type" {
						for _, r := range b.Results {
							if len(r.DimensionValues) > 0 {
								key := strings.ToLower(r.DimensionValues[0])
								results.Breakdowns[key] = r.Value
							}
						}
					}
				}
			}
		}
	}

	return results, nil
}

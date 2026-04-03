package handler

import (
	"encoding/json"
	"log"
	"mi-tech/internal/marketing"
	"net/http"
	"strings"
)

type MarketingHandler struct {
	metaClient *marketing.MetaMarketingClient
}

func NewMarketingHandler(metaClient *marketing.MetaMarketingClient) *MarketingHandler {
	return &MarketingHandler{
		metaClient: metaClient,
	}
}

func (h *MarketingHandler) GetMetaOverview(w http.ResponseWriter, r *http.Request) {
	// 1. Get Configured Ad Account ID
	configID := h.metaClient.GetConfiguredAdAccountID()
	
	// Always fetch available accounts to ensure UI has account names
	accounts, err := h.metaClient.FetchAdAccounts()
	if err != nil {
		log.Printf("ERROR: FetchAdAccounts failed: %v", err)
	}

	// If no ID is configured but we have accounts, pick the first one
	if configID == "" && len(accounts) > 0 {
		configID = accounts[0].ID
	}

	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	var insights []marketing.Insight
	if configID != "" {
		log.Printf("DEBUG: Fetching Overview Insights (level: campaign) for Account: %s (%s to %s)", configID, startDate, endDate)
		var fetchErr error
		insights, fetchErr = h.metaClient.FetchInsights(configID, "campaign", startDate, endDate)
		if fetchErr != nil {
			log.Printf("ERROR: FetchInsights failed: %v", fetchErr)
			// Return a 401 if it's a token error
			if strings.Contains(fetchErr.Error(), "Error validating access token") || strings.Contains(fetchErr.Error(), "expired") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"message": "Meta Session Expired. Please update your API token in Settings.",
				})
				return
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"accounts":  accounts,
		"insights":  insights,
		"active_id": configID,
	})
}

func (h *MarketingHandler) GetMetaCampaigns(w http.ResponseWriter, r *http.Request) {
	accountID := r.URL.Query().Get("ad_account_id")
	if accountID == "" {
		http.Error(w, "ad_account_id is required", http.StatusBadRequest)
		return
	}

	campaigns, err := h.metaClient.FetchCampaigns(accountID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"campaigns": campaigns,
	})
}

func (h *MarketingHandler) GetMetaAdSets(w http.ResponseWriter, r *http.Request) {
	campaignID := r.URL.Query().Get("campaign_id")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	if campaignID == "" {
		http.Error(w, "campaign_id is required", http.StatusBadRequest)
		return
	}

	log.Printf("DEBUG: Incoming GetMetaAdSets for Campaign: %s", campaignID)
	adsets, err := h.metaClient.FetchAdSets(campaignID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Fetch insights for each adset (optional optimization: fetch in parallel)
	insights, _ := h.metaClient.FetchInsights(campaignID, "adset", startDate, endDate)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"adsets":   adsets,
		"insights": insights,
	})
}

func (h *MarketingHandler) GetMetaAds(w http.ResponseWriter, r *http.Request) {
	adsetID := r.URL.Query().Get("adset_id")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	if adsetID == "" {
		http.Error(w, "adset_id is required", http.StatusBadRequest)
		return
	}

	log.Printf("DEBUG: Incoming GetMetaAds for AdSet: %s", adsetID)
	ads, err := h.metaClient.FetchAds(adsetID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Fetch insights for each ad
	insights, _ := h.metaClient.FetchInsights(adsetID, "ad", startDate, endDate)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"ads":      ads,
		"insights": insights,
	})
}



package handler

import (
	"encoding/json"
	_ "mi-tech/internal/marketing"
	"mi-tech/internal/service"
	"net/http"
)

type SMMHandler struct {
	socialService service.SocialService
}

func NewSMMHandler(socialService service.SocialService) *SMMHandler {
	return &SMMHandler{socialService: socialService}
}

// GetOverview returns historical insights for a platform.
// @Summary Social Media Overview
// @Description Fetch historical engagement metrics and growth insights for FB, IG, or Threads.
// @Tags social
// @Security Bearer
// @Produce json
// @Param platform query string true "Platform (facebook, instagram, threads)"
// @Param start_date query string false "Start date"
// @Param end_date query string false "End date"
// @Success 200 {object} map[string]interface{}
// @Router /marketing/smm/overview [get]
func (h *SMMHandler) GetOverview(w http.ResponseWriter, r *http.Request) {
	platform := r.URL.Query().Get("platform")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	if platform == "" {
		http.Error(w, "platform is required", http.StatusBadRequest)
		return
	}

	overview, err := h.socialService.GetOverview(platform, startDate, endDate)
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"overview": overview,
	})
}

// CheckHealth performs a real-time audit of Meta asset alignment.
// @Summary SMM Health Check
// @Description Audit the visibility and linkage of configured Meta Page and Instagram IDs.
// @Tags social
// @Security Bearer
// @Produce json
// @Success 200 {object} marketing.AssetHealth
// @Router /marketing/smm/health [get]
func (h *SMMHandler) CheckHealth(w http.ResponseWriter, r *http.Request) {
	health, err := h.socialService.CheckAssetHealth()
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"health":  health,
	})
}

// PostContent handles cross-platform posting.
// @Summary Cross-platform Post
// @Description Publish content to FB, IG, or Threads.
// @Tags social
// @Security Bearer
// @Accept json
// @Produce json
// @Param body body map[string]string true "Post content"
// @Success 200 {object} map[string]interface{}
// @Router /marketing/smm/post [post]
func (h *SMMHandler) PostContent(w http.ResponseWriter, r *http.Request) {
	var body map[string]string
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	platform := body["platform"]
	if platform == "" {
		http.Error(w, "platform is required", http.StatusBadRequest)
		return
	}

	postID, err := h.socialService.PostContent(platform, body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"post_id": postID,
	})
}

// Sync triggers a manual data fetch from Meta.
// @Summary Sync Social Data
// @Description Manually trigger a deep sync of metrics and post history.
// @Tags social
// @Security Bearer
// @Produce json
// @Param platform query string true "Platform (facebook, instagram, threads)"
// @Success 200 {object} map[string]interface{}
// @Router /marketing/smm/sync [post]
func (h *SMMHandler) Sync(w http.ResponseWriter, r *http.Request) {
	platform := r.URL.Query().Get("platform")
	if platform == "" {
		http.Error(w, "platform is required", http.StatusBadRequest)
		return
	}

	// 1. Sync daily metrics
	if err := h.socialService.SyncPlatformMetrics(platform); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 2. Sync historical posts (async-like but sequential for simplicity here)
	if err := h.socialService.SyncHistoricalInsights(platform); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Sync completed successfully",
	})
}

// GetPostInsights returns granular insights for a specific post.
func (h *SMMHandler) GetPostInsights(w http.ResponseWriter, r *http.Request) {
	postID := r.URL.Query().Get("id")
	if postID == "" {
		http.Error(w, "post id is required", http.StatusBadRequest)
		return
	}
	mediaType := r.URL.Query().Get("media_type")

	insights, err := h.socialService.GetPostInsights(postID, mediaType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"insights": insights,
	})
}

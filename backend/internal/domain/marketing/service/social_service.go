package service

import (
	"fmt"
	"log"
	repository "mi-tech/internal/domain/marketing/repository"
	"time"
)

type SocialService interface {
	SyncPlatformMetrics(platform string) error
	GetOverview(platform string, startDate, endDate string) (map[string]interface{}, error)
	PostContent(platform string, content map[string]string) (string, error)
	SyncHistoricalInsights(platform string) error
	GetPostInsights(postID string, mediaType string) (*DetailedInsights, error)
	CheckAssetHealth() (*AssetHealth, error)
}

type socialService struct {
	repo       repository.SocialRepository
	metaClient *MetaMarketingClient
}

func NewSocialService(repo repository.SocialRepository, metaClient *MetaMarketingClient) SocialService {
	return &socialService{
		repo:       repo,
		metaClient: metaClient,
	}
}

func (s *socialService) SyncPlatformMetrics(platform string) error {
	// Persistence disabled: Meta API is the single source of truth.
	log.Printf("DEBUG: Sync requested for %s (No-DB mode)", platform)
	return nil
}

func (s *socialService) GetOverview(platform string, startDate, endDate string) (map[string]interface{}, error) {
	log.Printf("DEBUG: Fetching live overview for %s from %s to %s", platform, startDate, endDate)

	// Shared logic for all platforms
	start, _ := time.Parse("2006-01-02", startDate)
	end, _ := time.Parse("2006-01-02", endDate)
	const metaLayout = "2006-01-02T15:04:05-0700"

	type mediaResult struct {
		item       map[string]interface{}
		engagement int
		views      int
		reach      int
		ok         bool
	}

	switch platform {
	case "instagram":
		igID := s.metaClient.GetConfiguredIGID()
		if igID == "" {
			return nil, nil
		}

		// 1. Fetch Account-level insights (historical/engagement)
		accountStats, err := s.metaClient.FetchAccountInsights(igID, startDate, endDate)
		if err != nil {
			log.Printf("WARNING: FetchAccountInsights failed: %v", err)
			accountStats = &AccountInsights{} // Default empty to avoid panic
		}

		accountInsights, err := s.metaClient.FetchInstagramInsights(igID, startDate, endDate)
		if err != nil {
			log.Printf("WARNING: FetchInstagramInsights failed (Graceful Degradation): %v", err)
			accountInsights = map[string]int{"reach": 0, "total_interactions": 0}
		}

		// 2. Fetch Media and filter by date
		media, err := s.metaClient.FetchInstagramMedia(igID)
		if err != nil {
			log.Printf("WARNING: FetchInstagramMedia failed (Graceful Degradation): %v", err)
			media = []SocialPost{}
		}
		log.Printf("DEBUG: Meta returned %d raw media items", len(media))

		var filteredMedia []map[string]interface{}
		totalEngagement := 0
		totalViews := 0

		resultsChan := make(chan mediaResult, len(media))
		sem := make(chan struct{}, 10) // Concurrency limit of 10

		for _, m := range media {
			pubTime, err := time.Parse(metaLayout, m.Timestamp)
			if err != nil {
				pubTime, _ = time.Parse(time.RFC3339, m.Timestamp)
			}

			if (startDate == "" || pubTime.After(start) || pubTime.Equal(start)) &&
				(endDate == "" || pubTime.Before(end) || pubTime.Equal(end)) {

				go func(m SocialPost) {
					sem <- struct{}{}        // Acquire
					defer func() { <-sem }() // Release

					// Fetch insights for this specific media
					mInsights, err := s.metaClient.FetchMediaInsights(m.ID, m.MediaType)
					if err != nil {
						log.Printf("WARNING: FetchMediaInsights failed for %s (%s): %v", m.ID, m.MediaType, err)
						// Graceful Degradation
						resultsChan <- mediaResult{
							ok: true,
							item: map[string]interface{}{
								"id":            m.ID,
								"content":       m.Caption,
								"media_url":     m.MediaURL,
								"thumbnail_url": m.ThumbnailURL,
								"permalink":     m.Permalink,
								"published_at":  m.Timestamp,
								"media_type":    m.MediaType,
								"reach":         0,
								"views":         0,
								"engagement":    0,
								"restricted":    true,
							},
							engagement: 0,
							views:      0,
						}
						return
					}

					// Standardize Metrics (v22.0 Alignment)
					reach := mInsights["reach"]
					views := mInsights["views"]
					if m.MediaType != "VIDEO" {
						views = reach // Use reach as a proxy for image views
					}
					engagement := mInsights["engagement"]

					resultsChan <- mediaResult{
						ok: true,
						item: map[string]interface{}{
							"id":            m.ID,
							"content":       m.Caption,
							"media_url":     m.MediaURL,
							"thumbnail_url": m.ThumbnailURL,
							"permalink":     m.Permalink,
							"published_at":  m.Timestamp,
							"media_type":    m.MediaType,
							"reach":         reach,
							"views":         views,
							"engagement":    engagement,
						},
						engagement: engagement,
						views:      views,
					}
				}(m)
			} else {
				resultsChan <- mediaResult{ok: false}
			}
		}

		// Collect ALL results (one per item in the media slice)
		for i := 0; i < len(media); i++ {
			res := <-resultsChan
			if res.ok && res.item != nil {
				// Only add to result if it's actually in our date range
				pubTime, err := time.Parse(metaLayout, res.item["published_at"].(string))
				if err != nil {
					pubTime, _ = time.Parse(time.RFC3339, res.item["published_at"].(string))
				}

				if (startDate == "" || pubTime.After(start) || pubTime.Equal(start)) &&
					(endDate == "" || pubTime.Before(end) || pubTime.Equal(end)) {
					filteredMedia = append(filteredMedia, res.item)
					totalEngagement += res.engagement
					totalViews += res.views
				}
			}
		}

		log.Printf("DEBUG: Parallel sync complete. Total media in range: %d", len(filteredMedia))
		log.Printf("DEBUG: Total media filtered in range: %d", len(filteredMedia))

		return map[string]interface{}{
			"account": map[string]interface{}{
				"follower_count": accountStats.FollowerCount,
				"total_reach":    accountStats.Reach,
				"total_views":    accountStats.Views,
				"breakdowns":     accountStats.Breakdowns,
			},
			"platform":         "instagram",
			"total_reach":      accountInsights["reach"],
			"total_views":      accountInsights["views"],
			"total_engagement": totalEngagement,
			"posts":            filteredMedia,
			"success":          true,
		}, nil

	case "facebook":
		pageID := s.metaClient.GetConfiguredPageID()
		if pageID == "" {
			return nil, fmt.Errorf("Facebook Page ID not configured")
		}

		// 1. Fetch Page Overall Insights
		pageInsights, err := s.metaClient.FetchPageInsights(pageID, startDate, endDate)
		if err != nil {
			log.Printf("WARNING: FetchPageInsights failed (Graceful Degradation): %v", err)
			pageInsights = map[string]int{"page_impressions": 0, "page_post_engagements": 0}
		}

		// 2. Fetch Page Media (Posts)
		media, err := s.metaClient.FetchFacebookPageMedia(pageID)
		if err != nil {
			log.Printf("WARNING: FetchFacebookPageMedia failed (Graceful Degradation): %v", err)
			media = []SocialPost{}
		}

		// 3. Process Media Insights in Parallel
		var filteredMedia []map[string]interface{}
		totalEngagement := 0
		totalViews := 0
		totalReach := 0

		resultsChan := make(chan mediaResult, len(media))
		sem := make(chan struct{}, 10) // Concurrency limit

		for _, m := range media {
			pubTime, err := time.Parse(time.RFC3339, m.Timestamp)
			if err != nil {
				pubTime, _ = time.Parse("2006-01-02T15:04:05-0700", m.Timestamp) // FB alt layout
			}

			if (startDate == "" || pubTime.After(start) || pubTime.Equal(start)) &&
				(endDate == "" || pubTime.Before(end) || pubTime.Equal(end)) {

				go func(m SocialPost) {
					sem <- struct{}{}
					defer func() { <-sem }()

					mInsights, err := s.metaClient.FetchFacebookPostInsights(m.ID)
					if err != nil {
						log.Printf("WARNING: FetchFacebookPostInsights failed for %s: %v", m.ID, err)
						// Graceful Degradation: If insights are blocked, return post with 0 stats
						resultsChan <- mediaResult{
							ok: true,
							item: map[string]interface{}{
								"id":            m.ID,
								"content":       m.Message,
								"media_url":     m.MediaURL,
								"thumbnail_url": m.MediaURL,
								"permalink":     m.Permalink,
								"published_at":  m.Timestamp,
								"media_type":    m.MediaType,
								"reach":         0,
								"views":         0,
								"engagement":    0,
								"restricted":    true, // Flag for UI visibility
							},
							engagement: 0,
							views:      0,
							reach:      0,
						}
						return
					}

					reach := mInsights["post_impressions_unique"]
					engagement := mInsights["post_engaged_users"]
					views := reach // FB doesn't have a distinct 'views' metric for generic posts, using impressions

					resultsChan <- mediaResult{
						ok: true,
						item: map[string]interface{}{
							"id":            m.ID,
							"content":       m.Message,
							"media_url":     m.MediaURL,
							"thumbnail_url": m.MediaURL,
							"permalink":     m.Permalink,
							"published_at":  m.Timestamp,
							"media_type":    m.MediaType,
							"reach":         reach,
							"views":         views,
							"engagement":    engagement,
						},
						engagement: engagement,
						views:      views,
						reach:      reach,
					}
				}(m)
			} else {
				resultsChan <- mediaResult{ok: false}
			}
		}

		// Collect results
		processedCount := 0
		for _, m := range media {
			pubTime, err := time.Parse(time.RFC3339, m.Timestamp)
			if err != nil {
				pubTime, _ = time.Parse("2006-01-02T15:04:05-0700", m.Timestamp)
			}
			if (startDate == "" || pubTime.After(start) || pubTime.Equal(start)) &&
				(endDate == "" || pubTime.Before(end) || pubTime.Equal(end)) {
				processedCount++
			}
		}

		for i := 0; i < processedCount; i++ {
			res := <-resultsChan
			if res.ok && res.item != nil {
				filteredMedia = append(filteredMedia, res.item)
				totalEngagement += res.engagement
				totalViews += res.views
				totalReach += res.reach
			}
		}

		return map[string]interface{}{
			"account": map[string]interface{}{
				"follower_count": pageInsights["page_fans"],
				"total_reach":    pageInsights["page_impressions"],
				"total_views":    pageInsights["page_views_total"],
			},
			"platform":         "facebook",
			"total_reach":      pageInsights["page_impressions"],
			"total_views":      pageInsights["page_views_total"],
			"total_engagement": pageInsights["page_post_engagements"],
			"posts":            filteredMedia,
			"success":          true,
		}, nil
	}

	return nil, nil
}

func (s *socialService) GetPostInsights(postID string, mediaType string) (*DetailedInsights, error) {
	return s.metaClient.FetchDetailedMediaInsights(postID, mediaType)
}

func (s *socialService) SyncHistoricalInsights(platform string) error {
	// Persistence disabled
	log.Printf("DEBUG: Historical sync requested for %s (No-DB mode)", platform)
	return nil
}

func (s *socialService) PostContent(platform string, content map[string]string) (string, error) {
	switch platform {
	case "facebook":
		pageID := content["page_id"]
		message := content["message"]
		return s.metaClient.PostToFacebookPage(pageID, message)
	case "instagram":
		igID := content["ig_id"]
		imageURL := content["image_url"]
		caption := content["caption"]
		return s.metaClient.PostToInstagram(igID, imageURL, caption)
	case "threads":
		threadsID := content["threads_id"]
		text := content["text"]
		return s.metaClient.PostToThreads(threadsID, text)
	}
	return "", nil
}

func (s *socialService) CheckAssetHealth() (*AssetHealth, error) {
	return s.metaClient.CheckAssetAlignment()
}

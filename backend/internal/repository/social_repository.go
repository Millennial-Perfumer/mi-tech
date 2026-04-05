package repository

import (
	"mi-tech/internal/entity"
	"time"

	"gorm.io/gorm"
)

type socialRepository struct {
	db *gorm.DB
}

func NewSocialRepository(db *gorm.DB) SocialRepository {
	return &socialRepository{db: db}
}

func (r *socialRepository) UpsertAccount(account entity.SocialAccount) error {
	account.UpdatedAt = time.Now()
	return r.db.Save(&account).Error
}

func (r *socialRepository) GetAccount(platform string) (entity.SocialAccount, error) {
	var account entity.SocialAccount
	err := r.db.Where("platform = ? AND is_active = ?", platform, true).First(&account).Error
	return account, err
}

func (r *socialRepository) UpsertPost(post entity.SocialPost) error {
	return r.db.Where("post_id = ?", post.PostID).
		Assign(post).
		FirstOrCreate(&post).Error
}

func (r *socialRepository) ListPosts(platform string, limit int) ([]entity.SocialPost, error) {
	var posts []entity.SocialPost
	err := r.db.Where("platform = ?", platform).
		Order("published_at DESC").
		Limit(limit).
		Find(&posts).Error
	return posts, err
}

func (r *socialRepository) UpsertMetricSnapshot(metric entity.SocialMetricHistory) error {
	return r.db.Where("platform = ? AND post_id = ? AND metric_date = ?", metric.Platform, metric.PostID, metric.MetricDate).
		Assign(metric).
		FirstOrCreate(&metric).Error
}

func (r *socialRepository) GetHistoricalMetrics(platform string, postID string, days int) ([]entity.SocialMetricHistory, error) {
	var results []entity.SocialMetricHistory
	cutoff := time.Now().AddDate(0, 0, -days)
	
	query := r.db.Where("platform = ? AND metric_date >= ?", platform, cutoff)
	if postID != "" {
		query = query.Where("post_id = ?", postID)
	} else {
		query = query.Where("post_id IS NULL")
	}
	
	err := query.Order("metric_date ASC").Find(&results).Error
	return results, err
}

func (r *socialRepository) GetPlatformSummary(platform string, startDate, endDate string) (map[string]interface{}, error) {
	var result struct {
		TotalLikes       int `json:"total_likes"`
		TotalComments    int `json:"total_comments"`
		TotalShares      int `json:"total_shares"`
		TotalReach       int `json:"total_reach"`
		TotalImpressions int `json:"total_impressions"`
	}

	err := r.db.Model(&entity.SocialMetricHistory{}).
		Select("SUM(likes) as total_likes, SUM(comments) as total_comments, SUM(shares) as total_shares, SUM(reach) as total_reach, SUM(impressions) as total_impressions").
		Where("platform = ? AND metric_date BETWEEN ? AND ? AND post_id IS NULL", platform, startDate, endDate).
		Scan(&result).Error

	if err != nil {
		return nil, err
	}

	// Also get top posts by engagement
	var topPosts []struct {
		PostID     string `json:"post_id"`
		Engagement int    `json:"engagement"`
	}

	r.db.Model(&entity.SocialMetricHistory{}).
		Select("post_id, SUM(likes + comments + shares) as engagement").
		Where("platform = ? AND metric_date BETWEEN ? AND ? AND post_id IS NOT NULL", platform, startDate, endDate).
		Group("post_id").
		Order("engagement DESC").
		Limit(5).
		Scan(&topPosts)

	return map[string]interface{}{
		"totals":    result,
		"top_posts": topPosts,
	}, nil
}

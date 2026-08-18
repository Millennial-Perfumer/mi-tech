package repository

import (
	"time"

	"mi-tech/internal/domain/marketing/entity"

	"gorm.io/gorm"
)

// SocialRepository defines all data access for social media persistence and history.
type SocialRepository interface {
	UpsertAccount(account entity.SocialAccount) error
	GetAccount(platform string) (entity.SocialAccount, error)
	UpsertPost(post entity.SocialPost) error
	ListPosts(platform string, limit int) ([]entity.SocialPost, error)
	UpsertMetricSnapshot(metric entity.SocialMetricHistory) error
	GetHistoricalMetrics(platform string, postID string, days int) ([]entity.SocialMetricHistory, error)
	GetPlatformSummary(platform string, startDate, endDate string) (map[string]interface{}, error)
	CreateQueuePost(post entity.SocialQueuePost) (entity.SocialQueuePost, error)
	ListQueuePosts(limit int) ([]entity.SocialQueuePost, error)
	GetAppConfig(key string) (string, error)
}

type gormSocialRepository struct {
	db *gorm.DB
}

// NewSocialRepository creates a new GORM-backed SocialRepository.
func NewSocialRepository(db *gorm.DB) SocialRepository {
	if db != nil {
		_ = db.AutoMigrate(&entity.SocialQueuePost{})
	}
	return &gormSocialRepository{db: db}
}

func (r *gormSocialRepository) UpsertAccount(account entity.SocialAccount) error {
	account.UpdatedAt = time.Now()
	return r.db.Save(&account).Error
}

func (r *gormSocialRepository) GetAccount(platform string) (entity.SocialAccount, error) {
	var account entity.SocialAccount
	err := r.db.Where("platform = ? AND is_active = ?", platform, true).First(&account).Error
	return account, err
}

func (r *gormSocialRepository) UpsertPost(post entity.SocialPost) error {
	return r.db.Where("post_id = ?", post.PostID).
		Assign(post).
		FirstOrCreate(&post).Error
}

func (r *gormSocialRepository) ListPosts(platform string, limit int) ([]entity.SocialPost, error) {
	var posts []entity.SocialPost
	err := r.db.Where("platform = ?", platform).
		Order("published_at DESC").
		Limit(limit).
		Find(&posts).Error
	return posts, err
}

func (r *gormSocialRepository) UpsertMetricSnapshot(metric entity.SocialMetricHistory) error {
	return r.db.Where("platform = ? AND post_id = ? AND metric_date = ?", metric.Platform, metric.PostID, metric.MetricDate).
		Assign(metric).
		FirstOrCreate(&metric).Error
}

func (r *gormSocialRepository) GetHistoricalMetrics(platform string, postID string, days int) ([]entity.SocialMetricHistory, error) {
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

func (r *gormSocialRepository) GetPlatformSummary(platform string, startDate, endDate string) (map[string]interface{}, error) {
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

func (r *gormSocialRepository) CreateQueuePost(post entity.SocialQueuePost) (entity.SocialQueuePost, error) {
	post.CreatedAt = time.Now()
	post.UpdatedAt = time.Now()
	if r.db == nil {
		return post, nil
	}
	err := r.db.Create(&post).Error
	return post, err
}

func (r *gormSocialRepository) ListQueuePosts(limit int) ([]entity.SocialQueuePost, error) {
	var posts []entity.SocialQueuePost
	if r.db == nil {
		return posts, nil
	}
	if limit <= 0 {
		limit = 20
	}
	err := r.db.Order("created_at DESC").Limit(limit).Find(&posts).Error
	return posts, err
}

func (r *gormSocialRepository) GetAppConfig(key string) (string, error) {
	if r.db == nil {
		return "", nil
	}
	var cfg struct {
		Value string `gorm:"column:value"`
	}
	err := r.db.Table("app_configs").Select("value").Where("key = ?", key).First(&cfg).Error
	if err != nil {
		return "", err
	}
	return cfg.Value, nil
}


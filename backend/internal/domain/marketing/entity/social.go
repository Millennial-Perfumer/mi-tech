package entity

import (
	"time"
)

// SocialAccount represents a linked social media platform account (Page/User).
type SocialAccount struct {
	ID          int       `gorm:"primaryKey" json:"id"`
	Platform    string    `gorm:"not null" json:"platform"` // 'facebook', 'instagram', 'threads'
	PlatformID  string    `gorm:"unique;not null" json:"platform_id"`
	AccountName string    `json:"account_name"`
	AccessToken string    `json:"access_token,omitempty"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SocialPost archives organic media/posts from platforms.
type SocialPost struct {
	ID           int       `gorm:"primaryKey" json:"id"`
	Platform     string    `gorm:"not null" json:"platform"`
	PostID       string    `gorm:"unique;not null" json:"post_id"`
	Content      string    `json:"content"`
	MediaURL     string    `json:"media_url"`
	ThumbnailURL string    `json:"thumbnail_url"`
	Permalink    string    `json:"permalink"`
	PublishedAt  time.Time `json:"published_at"`
	CreatedAt    time.Time `json:"created_at"`
}

// SocialMetricHistory stores daily snapshots of post engagement.
type SocialMetricHistory struct {
	ID          int       `gorm:"primaryKey" json:"id"`
	PostID      *string   `json:"post_id"` // Normalized platform post ID, NULL for account level
	Platform    string    `gorm:"not null" json:"platform"`
	MetricDate  time.Time `gorm:"not null" json:"metric_date"`
	Likes       int       `json:"likes"`
	Shares      int       `json:"shares"`
	Comments    int       `json:"comments"`
	Engagement  int       `json:"engagement"`
	Reach       int       `json:"reach"`
	Impressions int       `json:"impressions"`
	Saves       int       `json:"saves"`
	CreatedAt   time.Time `json:"created_at"`
}

func (SocialMetricHistory) TableName() string {
	return "social_metrics_history"
}

func (SocialPost) TableName() string {
	return "social_post_history"
}

// SocialQueuePost tracks posts queued for Google Drive / n8n auto-publishing.
type SocialQueuePost struct {
	ID              int       `gorm:"primaryKey" json:"id"`
	FolderName      string    `gorm:"uniqueIndex;not null" json:"folder_name"`
	GDriveFolderID  string    `gorm:"column:gdrive_folder_id" json:"gdrive_folder_id"`
	PostType        string    `gorm:"not null" json:"post_type"` // 'SINGLE_PHOTO', 'CAROUSEL', 'VIDEO'
	Caption         string    `json:"caption"`
	Hashtags        string    `json:"hashtags"`
	MediaFilenames  string    `json:"media_filenames"`  // JSON array string
	TargetPlatforms string    `json:"target_platforms"` // JSON array string e.g. ["facebook","instagram","threads","x"]
	Status          string    `gorm:"default:'QUEUED'" json:"status"` // 'QUEUED', 'PUBLISHING', 'PUBLISHED', 'FAILED'
	ErrorMessage    string    `json:"error_message,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (SocialQueuePost) TableName() string {
	return "social_queue_posts"
}


package entity

import "time"

// PublishedReview represents a review successfully posted to Judge.me and stored in DB.
type PublishedReview struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ReviewID     string    `gorm:"type:varchar(100)" json:"review_id"`
	ProductID    string    `gorm:"type:varchar(100);not null;index" json:"product_id"`
	ProductTitle string    `gorm:"type:varchar(255);not null" json:"product_title"`
	ReviewerName string    `gorm:"type:varchar(150);not null" json:"reviewer_name"`
	Gender       string    `gorm:"type:varchar(20);default:'unspecified'" json:"gender"`
	Email        string    `gorm:"type:varchar(150);not null" json:"email"`
	Rating       int       `gorm:"not null" json:"rating"`
	Title        string    `gorm:"type:text;not null" json:"title"`
	Body         string    `gorm:"type:text;not null" json:"body"`
	ShopDomain   string    `gorm:"type:varchar(150);not null" json:"shop_domain"`
	Status       string    `gorm:"type:varchar(50);not null;default:'SUCCESS'" json:"status"`
	StatusCode   int       `gorm:"not null;default:200" json:"status_code"`
	PublishedAt  time.Time `gorm:"default:CURRENT_TIMESTAMP;index" json:"published_at"`
	CreatedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (PublishedReview) TableName() string {
	return "published_reviews"
}

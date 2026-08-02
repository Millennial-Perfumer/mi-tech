package repository

import (
	"time"

	"mi-tech/internal/domain/feedback/dto"
	"mi-tech/internal/domain/feedback/entity"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// FeedbackRepository defines the data access operations for the customer feedback system.
type FeedbackRepository interface {
	SaveCustomerFeedback(feedback entity.CustomerFeedback) error
	GetCustomerFeedback() ([]dto.FeedbackResponse, error)
	GetFeedbackByID(id int) (entity.CustomerFeedback, error)
	UpdateFeedbackAdminComment(id int, comment string) error
	MarkJudgeMePosted(id int) error
	MarkGoogleReviewRequested(id int) error
}

type gormFeedbackRepository struct {
	db *gorm.DB
}

// NewFeedbackRepository creates a new GORM-backed FeedbackRepository.
func NewFeedbackRepository(db *gorm.DB) FeedbackRepository {
	return &gormFeedbackRepository{db: db}
}

func (r *gormFeedbackRepository) SaveCustomerFeedback(feedback entity.CustomerFeedback) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "order_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"rating", "message", "updated_at"}),
	}).Create(&feedback).Error
}

func (r *gormFeedbackRepository) GetCustomerFeedback() ([]dto.FeedbackResponse, error) {
	var results []dto.FeedbackResponse
	err := r.db.Table("customer_feedback").
		Select("customer_feedback.id, customer_feedback.order_id, orders.order_number, orders.customer_name, orders.customer_phone, customer_feedback.rating, customer_feedback.message as comment, customer_feedback.admin_comment, customer_feedback.judgeme_posted, customer_feedback.judgeme_posted_at, customer_feedback.google_review_requested, customer_feedback.google_review_requested_at, customer_feedback.created_at").
		Joins("JOIN orders ON orders.id = customer_feedback.order_id").
		Order("customer_feedback.created_at DESC").
		Scan(&results).Error
	return results, err
}

func (r *gormFeedbackRepository) GetFeedbackByID(id int) (entity.CustomerFeedback, error) {
	var feedback entity.CustomerFeedback
	err := r.db.Where("id = ?", id).First(&feedback).Error
	return feedback, err
}

func (r *gormFeedbackRepository) UpdateFeedbackAdminComment(id int, comment string) error {
	return r.db.Model(&entity.CustomerFeedback{}).Where("id = ?", id).Update("admin_comment", comment).Error
}

func (r *gormFeedbackRepository) MarkJudgeMePosted(id int) error {
	now := time.Now()
	return r.db.Model(&entity.CustomerFeedback{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"judgeme_posted":    true,
			"judgeme_posted_at": now,
		}).Error
}

func (r *gormFeedbackRepository) MarkGoogleReviewRequested(id int) error {
	now := time.Now()
	return r.db.Model(&entity.CustomerFeedback{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"google_review_requested":    true,
			"google_review_requested_at": now,
		}).Error
}

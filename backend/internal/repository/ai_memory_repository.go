package repository

import (
	"mi-tech/internal/entity"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type gormAIMemoryRepository struct {
	db *gorm.DB
}

func NewAIMemoryRepository(db *gorm.DB) AIMemoryRepository {
	return &gormAIMemoryRepository{db: db}
}

func (r *gormAIMemoryRepository) Upsert(memory *entity.AIMemory) error {
	memory.UpdatedAt = time.Now()
	if memory.CreatedAt.IsZero() {
		memory.CreatedAt = memory.UpdatedAt
	}

	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"content", "category", "metadata", "updated_at"}),
	}).Create(memory).Error
}

func (r *gormAIMemoryRepository) List(category string) ([]entity.AIMemory, error) {
	var memories []entity.AIMemory
	query := r.db.Order("updated_at DESC")
	if category != "" {
		query = query.Where("category = ?", category)
	}
	err := query.Find(&memories).Error
	return memories, err
}

func (r *gormAIMemoryRepository) GetByKey(key string) (*entity.AIMemory, error) {
	var memory entity.AIMemory
	err := r.db.Where("key = ?", key).First(&memory).Error
	if err != nil {
		return nil, err
	}
	return &memory, nil
}

func (r *gormAIMemoryRepository) Delete(key string) error {
	return r.db.Where("key = ?", key).Delete(&entity.AIMemory{}).Error
}

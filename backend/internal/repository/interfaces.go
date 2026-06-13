package repository

import (
	"mi-tech/internal/entity"
)

// AIMemoryRepository defines data access for AI persistent memory.
type AIMemoryRepository interface {
	Upsert(memory *entity.AIMemory) error
	List(category string) ([]entity.AIMemory, error)
	GetByKey(key string) (*entity.AIMemory, error)
	Delete(key string) error
}

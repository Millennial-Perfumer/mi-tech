package repository

import (
	"mi-tech/internal/mcp/entity"

	"gorm.io/gorm"
)

// gormAuditLogRepository is the GORM implementation of AuditLogRepository.
type gormAuditLogRepository struct {
	db *gorm.DB
}

// NewAuditLogRepository creates a new AuditLogRepository.
func NewAuditLogRepository(db *gorm.DB) AuditLogRepository {
	return &gormAuditLogRepository{db: db}
}

func (r *gormAuditLogRepository) Create(log *entity.MCPAuditLog) error {
	return r.db.Create(log).Error
}

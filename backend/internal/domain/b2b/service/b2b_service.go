package service

import (
	"gorm.io/gorm"
	"mi-tech/internal/domain/b2b/repository"
	"mi-tech/internal/shared/config"
)

type B2BService struct {
	repo     repository.B2BRepository
	settings *config.SettingsProvider
	db       *gorm.DB // For raw database transitions
}

// lineGSTRate preserves the legacy 18% default while allowing an explicit 0%
// line item rate from the invoice builder.
func lineGSTRate(rate *float64) float64 {
	if rate == nil {
		return 18
	}
	return *rate
}

func NewB2BService(repo repository.B2BRepository, settings *config.SettingsProvider, db *gorm.DB) *B2BService {
	return &B2BService{
		repo:     repo,
		settings: settings,
		db:       db,
	}
}

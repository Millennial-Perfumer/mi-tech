package service

import (
	"mi-tech/internal/domain/production/entity"
	"mi-tech/internal/domain/production/repository"
)

type OilInventoryService struct {
	repo repository.OilInventoryRepository
}

func NewOilInventoryService(repo repository.OilInventoryRepository) *OilInventoryService {
	return &OilInventoryService{repo: repo}
}

func (s *OilInventoryService) ListOils(search string) ([]entity.OilInventory, error) {
	return s.repo.List(search)
}

func (s *OilInventoryService) ListOilsPage(search, sort string, page, limit int) ([]entity.OilInventory, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	return s.repo.ListPage(search, sort, page, limit)
}

func (s *OilInventoryService) GetOil(id int) (entity.OilInventory, error) {
	return s.repo.GetByID(id)
}

func (s *OilInventoryService) CreateOil(item *entity.OilInventory) error {
	return s.repo.Create(item)
}

func (s *OilInventoryService) UpdateOil(item *entity.OilInventory) error {
	return s.repo.Update(item)
}

func (s *OilInventoryService) DeleteOil(id int) error {
	return s.repo.Delete(id)
}

func (s *OilInventoryService) BulkDeleteOils(ids []int) error {
	return s.repo.BulkDelete(ids)
}

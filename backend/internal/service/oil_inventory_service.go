package service

import (
	"mi-tech/internal/entity"
	"mi-tech/internal/repository"
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

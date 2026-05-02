package service

import (
	"mi-tech/internal/entity"
	"mi-tech/internal/repository"
)

type SupplierService struct {
	repo repository.SupplierRepository
}

func NewSupplierService(repo repository.SupplierRepository) *SupplierService {
	return &SupplierService{repo: repo}
}

func (s *SupplierService) ListSuppliers(search string) ([]entity.Supplier, error) {
	return s.repo.List(search)
}

func (s *SupplierService) GetSupplier(id int) (entity.Supplier, error) {
	return s.repo.GetByID(id)
}

func (s *SupplierService) CreateSupplier(supplier *entity.Supplier) error {
	return s.repo.Create(supplier)
}

func (s *SupplierService) UpdateSupplier(supplier *entity.Supplier) error {
	return s.repo.Update(supplier)
}

func (s *SupplierService) DeleteSupplier(id int) error {
	return s.repo.Delete(id)
}

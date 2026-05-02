package repository

import (
	"mi-tech/internal/entity"

	"gorm.io/gorm"
)

type gormSupplierRepository struct {
	db *gorm.DB
}

func NewSupplierRepository(db *gorm.DB) SupplierRepository {
	return &gormSupplierRepository{db: db}
}

func (r *gormSupplierRepository) List(search string) ([]entity.Supplier, error) {
	var suppliers []entity.Supplier
	query := r.db
	if search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}
	err := query.Find(&suppliers).Error
	return suppliers, err
}

func (r *gormSupplierRepository) GetByID(id int) (entity.Supplier, error) {
	var supplier entity.Supplier
	err := r.db.First(&supplier, id).Error
	return supplier, err
}

func (r *gormSupplierRepository) Create(supplier *entity.Supplier) error {
	return r.db.Create(supplier).Error
}

func (r *gormSupplierRepository) Update(supplier *entity.Supplier) error {
	return r.db.Save(supplier).Error
}

func (r *gormSupplierRepository) Delete(id int) error {
	return r.db.Delete(&entity.Supplier{}, id).Error
}

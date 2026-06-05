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
	if err != nil {
		return nil, err
	}

	type SupplierOilCount struct {
		SupplierID int
		Count      int
	}
	var counts []SupplierOilCount
	err = r.db.Model(&entity.OilInventory{}).
		Select("supplier_id, count(*) as count").
		Where("supplier_id IS NOT NULL").
		Group("supplier_id").
		Scan(&counts).Error
	if err == nil {
		countMap := make(map[int]int)
		for _, c := range counts {
			countMap[c.SupplierID] = c.Count
		}
		for i := range suppliers {
			suppliers[i].OilsCount = countMap[suppliers[i].ID]
		}
	}
	return suppliers, nil
}

func (r *gormSupplierRepository) GetByID(id int) (entity.Supplier, error) {
	var supplier entity.Supplier
	err := r.db.First(&supplier, id).Error
	if err != nil {
		return supplier, err
	}
	var count int64
	r.db.Model(&entity.OilInventory{}).Where("supplier_id = ?", id).Count(&count)
	supplier.OilsCount = int(count)
	return supplier, nil
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

package repository

import (
	"mi-tech/internal/domain/production/entity"

	"gorm.io/gorm"
)

// PurchaseOrderRepository defines all data access for raw material purchases.
type PurchaseOrderRepository interface {
	List() ([]entity.PurchaseOrder, error)
	ListRecent(limit int) ([]entity.PurchaseOrder, error)
	ListRecentDays(days, page int) ([]entity.PurchaseOrder, int64, error)
	GetByID(id int) (*entity.PurchaseOrder, error)
	Create(po *entity.PurchaseOrder) error
	Update(po *entity.PurchaseOrder) error
	Delete(id int) error
}

type pgPurchaseOrderRepository struct {
	db *gorm.DB
}

func NewPurchaseOrderRepository(db *gorm.DB) PurchaseOrderRepository {
	return &pgPurchaseOrderRepository{db: db}
}

func (r *pgPurchaseOrderRepository) List() ([]entity.PurchaseOrder, error) {
	var pos []entity.PurchaseOrder
	err := r.db.Preload("OilInventory").Preload("Supplier").Order("purchase_date desc").Find(&pos).Error
	return pos, err
}

func (r *pgPurchaseOrderRepository) ListRecent(limit int) ([]entity.PurchaseOrder, error) {
	var pos []entity.PurchaseOrder
	err := r.db.Preload("OilInventory").Preload("Supplier").Order("purchase_date DESC").Limit(limit).Find(&pos).Error
	return pos, err
}

// ListRecentDays returns every purchase order from the most recent distinct purchase dates.
// A single bill date can contain multiple oil line items, so it must not be truncated by row count.
func (r *pgPurchaseOrderRepository) ListRecentDays(days, page int) ([]entity.PurchaseOrder, int64, error) {
	var pos []entity.PurchaseOrder
	var totalDates int64
	if err := r.db.Model(&entity.PurchaseOrder{}).Select("COUNT(DISTINCT DATE(purchase_date))").Scan(&totalDates).Error; err != nil {
		return nil, 0, err
	}

	recentDates := r.db.Model(&entity.PurchaseOrder{}).
		Select("DISTINCT DATE(purchase_date)").
		Order("DATE(purchase_date) DESC").
		Offset((page - 1) * days).
		Limit(days)

	err := r.db.
		Preload("OilInventory").
		Preload("Supplier").
		Where("DATE(purchase_date) IN (?)", recentDates).
		Order("purchase_date DESC, id DESC").
		Find(&pos).Error
	return pos, totalDates, err
}

func (r *pgPurchaseOrderRepository) GetByID(id int) (*entity.PurchaseOrder, error) {
	var po entity.PurchaseOrder
	err := r.db.First(&po, id).Error
	return &po, err
}

func (r *pgPurchaseOrderRepository) Create(po *entity.PurchaseOrder) error {
	return r.db.Create(po).Error
}

func (r *pgPurchaseOrderRepository) Update(po *entity.PurchaseOrder) error {
	return r.db.Save(po).Error
}

func (r *pgPurchaseOrderRepository) Delete(id int) error {
	return r.db.Delete(&entity.PurchaseOrder{}, id).Error
}

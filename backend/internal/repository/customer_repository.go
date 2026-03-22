package repository

import (
	"context"
	"mi-tech/internal/entity"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CustomerRepository struct {
	db *gorm.DB
}

func NewCustomerRepository(db *gorm.DB) *CustomerRepository {
	return &CustomerRepository{db: db}
}

// UpsertByPhone performs a phone-number based upsert.
// If the customer exists, it updates the fields.
func (r *CustomerRepository) UpsertByPhone(ctx context.Context, customer *entity.Customer) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "phone_number"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"first_name", "last_name", "email", "address1", "address2", 
			"city", "state", "country", "zip_code", "total_orders", 
			"total_spent", "updated_at",
		}),
	}).Create(customer).Error
}

// UpdateStats updates only the total_orders and total_spent for a customer.
func (r *CustomerRepository) UpdateStats(ctx context.Context, phoneNumber string, orderDelta int, spentDelta float64) error {
	return r.db.WithContext(ctx).Model(&entity.Customer{}).
		Where("phone_number = ?", phoneNumber).
		Updates(map[string]interface{}{
			"total_orders": gorm.Expr("total_orders + ?", orderDelta),
			"total_spent":  gorm.Expr("total_spent + ?", spentDelta),
			"updated_at":   gorm.Expr("CURRENT_TIMESTAMP"),
		}).Error
}

func (r *CustomerRepository) GetByPhone(ctx context.Context, phone string) (*entity.Customer, error) {
	var customer entity.Customer
	err := r.db.WithContext(ctx).Where("phone_number = ?", phone).First(&customer).Error
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

func (r *CustomerRepository) List(ctx context.Context, search, sortBy, sortOrder string, offset, limit int) ([]entity.Customer, int64, error) {
	var customers []entity.Customer
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Customer{})

	if search != "" {
		s := "%" + search + "%"
		query = query.Where("phone_number LIKE ? OR first_name ILIKE ? OR last_name ILIKE ? OR email ILIKE ?", s, s, s, s)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	if sortBy != "" {
		order := sortBy
		if sortOrder == "ASC" || sortOrder == "DESC" {
			order += " " + sortOrder
		}
		query = query.Order(order)
	} else {
		query = query.Order("updated_at DESC")
	}

	err = query.Offset(offset).Limit(limit).Find(&customers).Error
	if err != nil {
		return nil, 0, err
	}

	return customers, total, nil
}

// UpsertBatch performs a batch upsert of customers.
func (r *CustomerRepository) UpsertBatch(ctx context.Context, customers []entity.Customer) error {
	if len(customers) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "phone_number"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"first_name", "last_name", "email", "address1", "address2", 
			"city", "state", "country", "zip_code", "total_orders", 
			"total_spent", "updated_at",
		}),
	}).CreateInBatches(customers, 100).Error
}

func (r *CustomerRepository) DeleteAll(ctx context.Context) error {
	return r.db.WithContext(ctx).Exec("DELETE FROM customers").Error
}

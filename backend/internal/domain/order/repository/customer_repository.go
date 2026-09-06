package repository

import (
	"context"
	"errors"
	"mi-tech/internal/domain/order/entity"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CustomerRepository struct {
	db     *gorm.DB
	events *gormEventRepository
}

// customerListAllowedSortColumns defines the permitted columns for sorting to prevent SQL injection.
// Hoisted to package level to avoid redundant map allocations per request.
var customerListAllowedSortColumns = map[string]bool{
	"phone_number": true, "first_name": true, "last_name": true,
	"email": true, "city": true, "state": true, "total_orders": true,
	"total_spent": true, "updated_at": true, "created_at": true, "source_id": true,
}

func NewCustomerRepository(db *gorm.DB) *CustomerRepository {
	return &CustomerRepository{db: db, events: newEventRepository(db)}
}

// UpsertByPhone performs a phone-number based upsert using raw SQL to ensure
// exact matching with partial unique indexes and avoid column ambiguity.
func (r *CustomerRepository) UpsertByPhone(ctx context.Context, c *entity.Customer) error {
	return r.upsertByPhone(ctx, c, nil)
}

// UpsertByPhoneForOrder is used when customer data was refreshed from a
// particular order, allowing the customer event to retain that order context.
func (r *CustomerRepository) UpsertByPhoneForOrder(ctx context.Context, c *entity.Customer, orderID int64) error {
	return r.upsertByPhone(ctx, c, &orderID)
}

func (r *CustomerRepository) upsertByPhone(ctx context.Context, c *entity.Customer, orderID *int64) error {
	now := time.Now()
	query := `
		INSERT INTO customers (
			phone_number, first_name, last_name, email, address1, address2,
			city, state, country, zip_code, total_orders, total_spent,
			source_id, external_id, created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		)
		ON CONFLICT (phone_number) WHERE deleted_at IS NULL
		DO UPDATE SET
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			email = EXCLUDED.email,
			address1 = EXCLUDED.address1,
			address2 = EXCLUDED.address2,
			city = EXCLUDED.city,
			state = EXCLUDED.state,
			country = EXCLUDED.country,
			zip_code = EXCLUDED.zip_code,
			total_orders = EXCLUDED.total_orders,
			total_spent = EXCLUDED.total_spent,
			updated_at = EXCLUDED.updated_at
		RETURNING id`

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var before entity.Customer
		lookupErr := tx.WithContext(ctx).Where("phone_number = ?", c.PhoneNumber).First(&before).Error
		if lookupErr != nil && !errors.Is(lookupErr, gorm.ErrRecordNotFound) {
			return lookupErr
		}

		if err := tx.WithContext(ctx).Raw(query,
			c.PhoneNumber, c.FirstName, c.LastName, c.Email, c.Address1, c.Address2,
			c.City, c.State, c.Country, c.ZipCode, c.TotalOrders, c.TotalSpent,
			c.SourceID, c.ExternalID, now, now,
		).Scan(&c.ID).Error; err != nil {
			return err
		}

		var after entity.Customer
		if err := tx.WithContext(ctx).First(&after, c.ID).Error; err != nil {
			return err
		}
		if errors.Is(lookupErr, gorm.ErrRecordNotFound) {
			return r.events.recordCustomerChanges(tx, nil, &after, after.SourceID, "sync", orderID)
		}
		return r.events.recordCustomerChanges(tx, &before, &after, after.SourceID, "sync", orderID)
	})
}

// UpdateStats updates only the total_orders and total_spent for a customer.
func (r *CustomerRepository) UpdateStats(ctx context.Context, phoneNumber string, orderDelta int, spentDelta float64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var before entity.Customer
		if err := tx.WithContext(ctx).Where("phone_number = ?", phoneNumber).First(&before).Error; err != nil {
			return err
		}
		if err := tx.WithContext(ctx).Model(&entity.Customer{}).
			Where("phone_number = ?", phoneNumber).
			Updates(map[string]interface{}{
				"total_orders": gorm.Expr("total_orders + ?", orderDelta),
				"total_spent":  gorm.Expr("total_spent + ?", spentDelta),
				"updated_at":   gorm.Expr("CURRENT_TIMESTAMP"),
			}).Error; err != nil {
			return err
		}
		var after entity.Customer
		if err := tx.WithContext(ctx).Where("id = ?", before.ID).First(&after).Error; err != nil {
			return err
		}
		return r.events.recordCustomerChanges(tx, &before, &after, after.SourceID, "system", nil)
	})
}

func (r *CustomerRepository) Create(ctx context.Context, customer *entity.Customer) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Create(customer).Error; err != nil {
			return err
		}
		return r.events.recordCustomerChanges(tx, nil, customer, customer.SourceID, "manual", nil)
	})
}

func (r *CustomerRepository) Update(ctx context.Context, customer *entity.Customer) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var before entity.Customer
		if err := tx.WithContext(ctx).First(&before, customer.ID).Error; err != nil {
			return err
		}
		if err := tx.WithContext(ctx).Save(customer).Error; err != nil {
			return err
		}
		var after entity.Customer
		if err := tx.WithContext(ctx).First(&after, customer.ID).Error; err != nil {
			return err
		}
		return r.events.recordCustomerChanges(tx, &before, &after, after.SourceID, "manual", nil)
	})
}

func (r *CustomerRepository) GetByID(ctx context.Context, id int64) (*entity.Customer, error) {
	var customer entity.Customer
	err := r.db.WithContext(ctx).First(&customer, id).Error
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

func (r *CustomerRepository) GetByPhones(ctx context.Context, phones []string) ([]entity.Customer, error) {
	var customers []entity.Customer
	err := r.db.WithContext(ctx).Where("phone_number IN ?", phones).Find(&customers).Error
	return customers, err
}

func (r *CustomerRepository) GetByPhone(ctx context.Context, phone string) (*entity.Customer, error) {
	var customer entity.Customer
	err := r.db.WithContext(ctx).Where("phone_number = ?", phone).First(&customer).Error
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

func (r *CustomerRepository) GetByExternalID(ctx context.Context, externalID string) (*entity.Customer, error) {
	var customer entity.Customer
	err := r.db.WithContext(ctx).Where("external_id = ?", externalID).First(&customer).Error
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

func (r *CustomerRepository) List(ctx context.Context, search, sortBy, sortOrder string, sourceID string, minSpent, maxSpent float64, minOrders int, city, state string, firstName, lastName, email string, firstNameEmpty, lastNameEmpty, emailEmpty bool, offset, limit int) ([]entity.Customer, int64, error) {
	var customers []entity.Customer
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Customer{})

	if search != "" {
		s := "%" + search + "%"
		query = query.Where("phone_number LIKE ? OR first_name ILIKE ? OR last_name ILIKE ? OR email ILIKE ?", s, s, s, s)
	}

	if sourceID != "" {
		query = query.Where("source_id = ?", sourceID)
	}
	if minSpent > 0 {
		query = query.Where("total_spent >= ?", minSpent)
	}
	if maxSpent > 0 {
		query = query.Where("total_spent <= ?", maxSpent)
	}
	if minOrders > 0 {
		query = query.Where("total_orders >= ?", minOrders)
	}
	if city != "" {
		query = query.Where("city ILIKE ?", "%"+city+"%")
	}
	if state != "" {
		query = query.Where("state ILIKE ?", "%"+state+"%")
	}

	// Exact matches
	if firstName != "" {
		query = query.Where("first_name = ?", firstName)
	}
	if lastName != "" {
		query = query.Where("last_name = ?", lastName)
	}
	if email != "" {
		query = query.Where("email = ?", email)
	}

	// Empty checks
	if firstNameEmpty {
		query = query.Where("(first_name IS NULL OR first_name = '')")
	}
	if lastNameEmpty {
		query = query.Where("(last_name IS NULL OR last_name = '')")
	}
	if emailEmpty {
		query = query.Where("(email IS NULL OR email = '')")
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Security: Use allowlist for sortBy to prevent SQL injection.
	// Performance: Uses pre-allocated package-level map.
	if sortBy != "" && customerListAllowedSortColumns[sortBy] {
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
// Optimization: Replaces iterative raw SQL calls with a single GORM batch create.
// Expected Impact: Reduces database roundtrips from O(N) to O(1).
func (r *CustomerRepository) UpsertBatch(ctx context.Context, customers []entity.Customer) error {
	if len(customers) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		phones := make([]string, 0, len(customers))
		for _, customer := range customers {
			phones = append(phones, customer.PhoneNumber)
		}

		var existingCustomers []entity.Customer
		if err := tx.WithContext(ctx).Where("phone_number IN ?", phones).Find(&existingCustomers).Error; err != nil {
			return err
		}
		existingByPhone := make(map[string]entity.Customer, len(existingCustomers))
		for _, customer := range existingCustomers {
			existingByPhone[customer.PhoneNumber] = customer
		}

		// We use clause.OnConflict to handle the upsert logic.
		// The partial unique index 'idx_customers_phone_unique_active' is targeted via TargetWhere.
		if err := tx.WithContext(ctx).Clauses(clause.OnConflict{
			TargetWhere: clause.Where{
				Exprs: []clause.Expression{
					clause.Expr{SQL: "deleted_at IS NULL"},
				},
			},
			Columns: []clause.Column{{Name: "phone_number"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"first_name", "last_name", "email", "address1", "address2",
				"city", "state", "country", "zip_code", "total_orders", "total_spent",
				"updated_at",
			}),
		}).Create(&customers).Error; err != nil {
			return err
		}

		for i := range customers {
			var after entity.Customer
			if err := tx.WithContext(ctx).Where("phone_number = ?", customers[i].PhoneNumber).First(&after).Error; err != nil {
				return err
			}
			if before, found := existingByPhone[customers[i].PhoneNumber]; found {
				if err := r.events.recordCustomerChanges(tx, &before, &after, after.SourceID, "sync", nil); err != nil {
					return err
				}
			} else if err := r.events.recordCustomerChanges(tx, nil, &after, after.SourceID, "sync", nil); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *CustomerRepository) GetByIDs(ctx context.Context, ids []uint) ([]entity.Customer, error) {
	var customers []entity.Customer
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&customers).Error
	return customers, err
}

func (r *CustomerRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&entity.Customer{}, id).Error
}

func (r *CustomerRepository) BulkDelete(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Delete(&entity.Customer{}, ids).Error
}

func (r *CustomerRepository) DeleteAll(ctx context.Context) error {
	return r.db.WithContext(ctx).Exec("DELETE FROM customers").Error
}

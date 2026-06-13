package repository

import (
	"mi-tech/internal/domain/user/entity"

	"gorm.io/gorm"
)

// UserRepository defines the database operations for users.
type UserRepository interface {
	GetUsers() ([]entity.User, error)
	GetByUsername(username string) (entity.User, error)
	Create(user *entity.User) error
	Update(user *entity.User) error
	GetTemplateIDByName(name string) (int, error)
}

type gormUserRepository struct {
	db *gorm.DB
}

// NewUserRepository constructs a new GORM-backed UserRepository.
func NewUserRepository(db *gorm.DB) UserRepository {
	return &gormUserRepository{db: db}
}

func (r *gormUserRepository) GetUsers() ([]entity.User, error) {
	var users []entity.User
	if err := r.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *gormUserRepository) GetByUsername(username string) (entity.User, error) {
	var user entity.User
	err := r.db.Where("username = ?", username).First(&user).Error
	return user, err
}

func (r *gormUserRepository) Create(user *entity.User) error {
	return r.db.Create(user).Error
}

func (r *gormUserRepository) Update(user *entity.User) error {
	return r.db.Save(user).Error
}

func (r *gormUserRepository) GetTemplateIDByName(name string) (int, error) {
	var templateID int
	err := r.db.Raw("SELECT id FROM automation_templates WHERE template_name = ?", name).Scan(&templateID).Error
	return templateID, err
}

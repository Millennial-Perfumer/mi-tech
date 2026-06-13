package service

import (
	"errors"

	"mi-tech/internal/domain/user/entity"
	"mi-tech/internal/domain/user/repository"

	"golang.org/x/crypto/bcrypt"
)

// UserService manages user accounts.
type UserService struct {
	repo repository.UserRepository
}

// NewUserService constructs a new UserService.
func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetUsers() ([]entity.User, error) {
	return s.repo.GetUsers()
}

func (s *UserService) CreateUser(username, password, role string) error {
	if role != "admin" && role != "read" {
		return errors.New("invalid role")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := entity.User{
		Username:     username,
		PasswordHash: string(hashedPassword),
		Role:         role,
	}

	return s.repo.Create(&user)
}

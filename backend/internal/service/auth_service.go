package service

import (
	"errors"
	"fmt"
	"mi-tech/internal/config"
	"time"

	"mi-tech/internal/entity"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	db       *gorm.DB
	settings *config.SettingsProvider
}

func NewAuthService(db *gorm.DB, settings *config.SettingsProvider) *AuthService {
	return &AuthService{
		db:       db,
		settings: settings,
	}
}

// Login verifies credentials and returns a JWT token.
func (s *AuthService) Login(username, password string) (string, error) {
	var user entity.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("invalid username or password")
		}
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid username or password")
	}

	return s.GenerateToken(user)
}

// GenerateToken creates a new JWT token for a user.
func (s *AuthService) GenerateToken(user entity.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.settings.GetJWTSecret()))
}

// Register (for initial setup) hashes password and saves user.
func (s *AuthService) Register(username, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := entity.User{
		Username:     username,
		PasswordHash: string(hashedPassword),
	}

	return s.db.Create(&user).Error
}

// ValidateToken parses and validates a JWT token.
func (s *AuthService) ValidateToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.settings.GetJWTSecret()), nil
	})
}

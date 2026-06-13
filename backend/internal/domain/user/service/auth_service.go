package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"time"

	"mi-tech/internal/domain/shared/config"
	"mi-tech/internal/domain/user/entity"
	"mi-tech/internal/domain/user/repository"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

// Messenger defines the interface for sending WhatsApp notifications.
type Messenger interface {
	SendTemplateMessage(storeID string, templateID int, orderID int64, phoneNumber, templateName, languageCode string, components []interface{}) error
}

// AuthService coordinates session management and 2FA.
type AuthService struct {
	repo            repository.UserRepository
	settings        *config.SettingsProvider
	messagesService Messenger
}

// NewAuthService constructs a new AuthService.
func NewAuthService(repo repository.UserRepository, settings *config.SettingsProvider, messagesService Messenger) *AuthService {
	return &AuthService{
		repo:            repo,
		settings:        settings,
		messagesService: messagesService,
	}
}

// Login verifies credentials. If 2FA is enabled, it sends an OTP and returns a special error.
func (s *AuthService) Login(username, password string) (string, bool, error) {
	user, err := s.repo.GetByUsername(username)
	if err != nil {
		return "", false, errors.New("invalid username or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", false, errors.New("invalid username or password")
	}

	if user.TwoFactorEnabled {
		if user.PhoneNumber == "" {
			return "", false, errors.New("2FA enabled but no phone number configured. Contact admin.")
		}

		err := s.SendOTP(&user)
		if err != nil {
			return "", false, fmt.Errorf("failed to send verification code: %v", err)
		}
		return "", true, nil
	}

	token, err := s.GenerateToken(user)
	return token, false, err
}

func (s *AuthService) SendOTP(user *entity.User) error {
	otp := make([]byte, 6)
	if _, err := io.ReadAtLeast(rand.Reader, otp, 6); err != nil {
		return err
	}
	for i := 0; i < len(otp); i++ {
		otp[i] = (otp[i] % 10) + '0'
	}
	otpStr := string(otp)

	expiry := time.Now().Add(5 * time.Minute)
	user.OTPCode = otpStr
	user.OTPExpiry = &expiry

	if err := s.repo.Update(user); err != nil {
		return err
	}

	components := []interface{}{
		map[string]interface{}{
			"type": "body",
			"parameters": []interface{}{
				map[string]interface{}{
					"type": "text",
					"text": otpStr,
				},
			},
		},
		map[string]interface{}{
			"type":     "button",
			"sub_type": "url",
			"index":    "0",
			"parameters": []interface{}{
				map[string]interface{}{
					"type": "text",
					"text": otpStr,
				},
			},
		},
	}

	templateID, err := s.repo.GetTemplateIDByName("login_verification_template")
	if err != nil {
		return fmt.Errorf("failed to query OTP template ID: %w", err)
	}

	return s.messagesService.SendTemplateMessage(config.StoreIDShopify, templateID, 0, user.PhoneNumber, "login_verification_template", "en", components)
}

func (s *AuthService) VerifyOTP(username, otp string) (string, error) {
	user, err := s.repo.GetByUsername(username)
	if err != nil {
		return "", errors.New("invalid session")
	}

	if user.OTPCode == "" || user.OTPExpiry == nil || time.Now().After(*user.OTPExpiry) {
		return "", errors.New("verification code expired or not found")
	}

	if user.OTPCode != otp {
		return "", errors.New("invalid verification code")
	}

	user.OTPCode = ""
	user.OTPExpiry = nil
	if err := s.repo.Update(&user); err != nil {
		return "", err
	}

	return s.GenerateToken(user)
}

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

func (s *AuthService) Register(username, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := entity.User{
		Username:     username,
		PasswordHash: string(hashedPassword),
	}

	return s.repo.Create(&user)
}

func (s *AuthService) ValidateToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.settings.GetJWTSecret()), nil
	})
}

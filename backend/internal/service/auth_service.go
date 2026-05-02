package service

import (
	"errors"
	"fmt"
	"mi-tech/internal/config"
	"mi-tech/internal/entity"
	"time"

	"crypto/rand"
	"crypto/subtle"
	"io"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Messenger interface {
	SendTemplateMessage(storeID string, templateID int, orderID int64, phoneNumber, templateName, languageCode string, components []interface{}) error
}

type AuthService struct {
	db              *gorm.DB
	settings        *config.SettingsProvider
	messagesService Messenger
}

func NewAuthService(db *gorm.DB, settings *config.SettingsProvider, messagesService Messenger) *AuthService {
	return &AuthService{
		db:              db,
		settings:        settings,
		messagesService: messagesService,
	}
}

// Login verifies credentials. If 2FA is enabled, it sends an OTP and returns a special error.
func (s *AuthService) Login(username, password string) (string, bool, error) {
	var user entity.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", false, errors.New("invalid username or password")
		}
		return "", false, err
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
	// Generate 6-digit OTP
	otp := make([]byte, 6)
	if _, err := io.ReadAtLeast(rand.Reader, otp, 6); err != nil {
		return err
	}
	// Convert to digits 0-9
	for i := 0; i < len(otp); i++ {
		otp[i] = (otp[i] % 10) + '0'
	}
	otpStr := string(otp)

	expiry := time.Now().Add(5 * time.Minute)
	user.OTPCode = otpStr
	user.OTPExpiry = &expiry

	if err := s.db.Save(user).Error; err != nil {
		return err
	}

	// Prepare WhatsApp components (Position 1 is the code)
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

	// Find template ID for logging
	var templateID int
	s.db.Raw("SELECT id FROM automation_templates WHERE template_name = ?", "login_verification_template").Scan(&templateID)

	// Send via WhatsApp using the specific template
	// storeID config.StoreIDShopify, orderID 0
	return s.messagesService.SendTemplateMessage(config.StoreIDShopify, templateID, 0, user.PhoneNumber, "login_verification_template", "en", components)
}

func (s *AuthService) VerifyOTP(username, otp string) (string, error) {
	var user entity.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		return "", errors.New("invalid session")
	}

	if user.OTPCode == "" || user.OTPExpiry == nil || time.Now().After(*user.OTPExpiry) {
		return "", errors.New("verification code expired or not found")
	}

	if subtle.ConstantTimeCompare([]byte(user.OTPCode), []byte(otp)) != 1 {
		return "", errors.New("invalid verification code")
	}

	// Clear OTP after successful verification
	s.db.Model(&user).Updates(map[string]interface{}{
		"otp_code":   "",
		"otp_expiry": nil,
	})

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

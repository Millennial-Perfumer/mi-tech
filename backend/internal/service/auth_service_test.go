package service

import (
	"fmt"
	"mi-tech/internal/config"
	"mi-tech/internal/entity"
	"mi-tech/internal/testutil"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthService_Login(t *testing.T) {
	db, err := testutil.SetupTestDB()
	if err != nil {
		t.Skip("DB not available")
	}
	defer testutil.CleanupTestDB(db)

	service := NewAuthService(db, nil, nil)

	_, _, err = service.Login("admin@millennialperfumer.in", "admin123")

	if err != nil {
		// We expect an error about 2FA because SettingsProvider is nil/not found
		assert.Contains(t, err.Error(), "2FA enabled but no phone number configured")
	}
}

func TestAuthService_Register(t *testing.T) {
	db, err := testutil.SetupTestDB()
	if err != nil {
		t.Skip("DB not available")
	}
	defer testutil.CleanupTestDB(db)

	service := NewAuthService(db, nil, nil)
	username := fmt.Sprintf("user_%d@example.com", time.Now().UnixNano())
	err = service.Register(username, "password123")
	assert.NoError(t, err)

	var user entity.User
	err = db.Where("username = ?", username).First(&user).Error
	assert.NoError(t, err)
	assert.Equal(t, username, user.Username)

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("password123"))
	assert.NoError(t, err)
}

func TestAuthService_VerifyOTP(t *testing.T) {
	db, err := testutil.SetupTestDB()
	if err != nil {
		t.Skip("DB not available")
	}
	defer testutil.CleanupTestDB(db)

	settings := &config.SettingsProvider{} // Mocked settings
	service := NewAuthService(db, settings, nil)

	username := "otp_user@example.com"
	otp := "123456"
	expiry := time.Now().Add(5 * time.Minute)

	// Create a user with a valid OTP
	user := entity.User{
		Username:     username,
		PasswordHash: "hashed_password",
		OTPCode:      otp,
		OTPExpiry:    &expiry,
	}
	err = db.Create(&user).Error
	assert.NoError(t, err)

	// Test successful verification
	// We need to handle GenerateToken which requires JWT_SECRET.
	// We'll set it in the environment for the test.
	t.Setenv("JWT_SECRET", "test_secret")

	token, err := service.VerifyOTP(username, otp)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Verify OTP is cleared
	var updatedUser entity.User
	err = db.Where("username = ?", username).First(&updatedUser).Error
	assert.NoError(t, err)
	assert.Empty(t, updatedUser.OTPCode)
	assert.Nil(t, updatedUser.OTPExpiry)

	// Test failed verification (wrong OTP)
	user.OTPCode = "654321"
	user.OTPExpiry = &expiry
	db.Save(&user)

	_, err = service.VerifyOTP(username, "wrong_otp")
	assert.Error(t, err)
	assert.Equal(t, "invalid verification code", err.Error())

	// Test expired OTP
	pastExpiry := time.Now().Add(-5 * time.Minute)
	user.OTPCode = otp
	user.OTPExpiry = &pastExpiry
	db.Save(&user)

	_, err = service.VerifyOTP(username, otp)
	assert.Error(t, err)
	assert.Equal(t, "verification code expired or not found", err.Error())
}

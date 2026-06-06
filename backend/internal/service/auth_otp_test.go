package service

import (
	"mi-tech/internal/entity"
	"mi-tech/internal/testutil"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAuthService_VerifyOTP(t *testing.T) {
	db, err := testutil.SetupTestDB()
	if err != nil {
		t.Skip("DB not available")
	}
	defer testutil.CleanupTestDB(db)

	service := NewAuthService(db, nil, nil)
	username := "otp_user@example.com"
	otp := "123456"
	expiry := time.Now().Add(5 * time.Minute)

	user := entity.User{
		Username:  username,
		OTPCode:   otp,
		OTPExpiry: &expiry,
	}
	db.Create(&user)

	// Test correct OTP
	_, err = service.VerifyOTP(username, otp)
	assert.NoError(t, err)

	// Test incorrect OTP
	expiry2 := time.Now().Add(5 * time.Minute)
	user2 := entity.User{
		Username:  "wrong_otp_user@example.com",
		OTPCode:   otp,
		OTPExpiry: &expiry2,
	}
	db.Create(&user2)
	_, err = service.VerifyOTP("wrong_otp_user@example.com", "wrong")
	assert.Error(t, err)
	assert.Equal(t, "invalid verification code", err.Error())
}

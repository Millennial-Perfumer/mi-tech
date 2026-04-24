package service

import (
	"fmt"
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

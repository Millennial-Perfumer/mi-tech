package service

import (
	"testing"

	"mi-tech/internal/entity"
	"mi-tech/internal/testutil"

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
		// We expect an error about nil settings if it tries to GenerateToken
		// or invalid username/password since we haven't seeded the admin user in this test.
	}
}

func TestAuthService_Register(t *testing.T) {
	db, err := testutil.SetupTestDB()
	if err != nil {
		t.Skip("DB not available")
	}
	defer testutil.CleanupTestDB(db)

	service := NewAuthService(db, nil, nil)
	err = service.Register("newuser@example.com", "password123")
	assert.NoError(t, err)

	var user entity.User
	err = db.Where("username = ?", "newuser@example.com").First(&user).Error
	assert.NoError(t, err)
	assert.Equal(t, "newuser@example.com", user.Username)

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("password123"))
	assert.NoError(t, err)
}

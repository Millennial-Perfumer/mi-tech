package test

import (
	"fmt"
	"testing"
	"time"

	"mi-tech/internal/domain/user/repository"
	"mi-tech/internal/domain/user/service"
	"mi-tech/internal/shared/testutil"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthService_Login(t *testing.T) {
	db, err := testutil.SetupTestDB()
	if err != nil {
		t.Skip("DB not available")
	}
	defer testutil.CleanupTestDB(db)

	repo := repository.NewUserRepository(db)
	svc := service.NewAuthService(repo, nil, nil)

	username := fmt.Sprintf("test_login_%d@example.com", time.Now().UnixNano())
	err = svc.Register(username, "password123")
	assert.NoError(t, err)

	// Disable 2FA for test user to verify standard JWT token generation
	user, err := repo.GetByUsername(username)
	assert.NoError(t, err)
	user.TwoFactorEnabled = false
	err = repo.Update(&user)
	assert.NoError(t, err)

	token, requires2FA, err := svc.Login(username, "password123")
	assert.NoError(t, err)
	assert.False(t, requires2FA)
	assert.NotEmpty(t, token)
}

func TestAuthService_Register(t *testing.T) {
	db, err := testutil.SetupTestDB()
	if err != nil {
		t.Skip("DB not available")
	}
	defer testutil.CleanupTestDB(db)

	repo := repository.NewUserRepository(db)
	svc := service.NewAuthService(repo, nil, nil)
	username := fmt.Sprintf("user_%d@example.com", time.Now().UnixNano())
	err = svc.Register(username, "password123")
	assert.NoError(t, err)

	user, err := repo.GetByUsername(username)
	assert.NoError(t, err)
	assert.Equal(t, username, user.Username)

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("password123"))
	assert.NoError(t, err)
}

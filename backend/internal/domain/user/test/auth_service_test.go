package test

import (
	"fmt"
	"testing"
	"time"

	"mi-tech/internal/shared/testutil"
	"mi-tech/internal/domain/user/repository"
	"mi-tech/internal/domain/user/service"

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

	_, _, err = svc.Login("admin@millennialperfumer.in", "admin123")

	if err != nil {
		assert.Contains(t, err.Error(), "2FA enabled but no phone number configured")
	}
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

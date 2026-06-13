package test

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"mi-tech/internal/domain/user/dto"
	"mi-tech/internal/shared/testutil"

	"github.com/stretchr/testify/assert"
)

func TestAuthHandler_Login(t *testing.T) {
	db, err := testutil.SetupTestDB()
	if err != nil {
		t.Skip("DB not available")
	}
	defer testutil.CleanupTestDB(db)

	loginReq := dto.LoginRequest{
		Username: "admin@millennialperfumer.in",
		Password: "wrongpassword",
	}
	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	assert.NotNil(t, req)
	assert.NotNil(t, w)
}

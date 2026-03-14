package service

import (
	"testing"
	"mi-tech/internal/entity"
	"github.com/golang-jwt/jwt/v4"
)

// GenerateToken is pure logic
func TestGenerateToken(t *testing.T) {
	jwtSecret := "testsecret"
	svc := NewAuthService(nil, jwtSecret)

	user := entity.User{
		ID:       123,
		Username: "testuser",
	}

	token, err := svc.GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	if token == "" {
		t.Errorf("expected non-empty token")
	}

	// Validate the generated token
	parsedToken, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	if !parsedToken.Valid {
		t.Errorf("expected token to be valid")
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatalf("expected map claims")
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		t.Fatalf("expected user_id to be float64 in JSON map, got %T", claims["user_id"])
	}

	if uint(userIDFloat) != user.ID {
		t.Errorf("expected user_id %d, got %v", user.ID, userIDFloat)
	}

	if claims["username"] != user.Username {
		t.Errorf("expected username %s, got %v", user.Username, claims["username"])
	}
}

// TestValidateToken invalid token
func TestValidateToken_Invalid(t *testing.T) {
	jwtSecret := "testsecret"
	svc := NewAuthService(nil, jwtSecret)

	_, err := svc.ValidateToken("invalid.token.string")
	if err == nil {
		t.Errorf("expected error for invalid token, got nil")
	}
}

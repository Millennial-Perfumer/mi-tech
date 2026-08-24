package mcp

import (
	"errors"
	"strings"
	"testing"
	"time"

	"mi-tech/internal/mcp/entity"

	"github.com/stretchr/testify/require"
)

// fakeKeyRepo is an in-memory MachineKeyRepository for tests.
type fakeKeyRepo struct {
	keys []entity.MachineAPIKey
}

func (f *fakeKeyRepo) Create(key *entity.MachineAPIKey) error {
	key.ID = int64(len(f.keys) + 1)
	f.keys = append(f.keys, *key)
	return nil
}

func (f *fakeKeyRepo) FindByHash(hash string) (*entity.MachineAPIKey, error) {
	for i := range f.keys {
		if f.keys[i].KeyHash == hash {
			return &f.keys[i], nil
		}
	}
	return nil, errors.New("not found")
}

func (f *fakeKeyRepo) List() ([]entity.MachineAPIKey, error) {
	return f.keys, nil
}

func (f *fakeKeyRepo) Update(key *entity.MachineAPIKey) error {
	for i := range f.keys {
		if f.keys[i].ID == key.ID {
			f.keys[i] = *key
			return nil
		}
	}
	return errors.New("not found")
}

func (f *fakeKeyRepo) TouchLastUsed(id int64, at time.Time) error {
	for i := range f.keys {
		if f.keys[i].ID == id {
			f.keys[i].LastUsedAt = &at
			return nil
		}
	}
	return errors.New("not found")
}

func TestGenerateAndAuthenticate(t *testing.T) {
	svc := NewMachineKeyService(&fakeKeyRepo{})
	svc.now = func() time.Time { return time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC) }

	plaintext, key, err := svc.Generate(KeyOptions{
		Name:   "codex",
		Scopes: []string{ScopeOrders, ScopeMetrics},
	})
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(plaintext, KeyPrefix), "key should carry prefix")
	require.NotEqual(t, plaintext, key.KeyHash, "plaintext must never equal stored hash")

	authed, err := svc.Authenticate(plaintext)
	require.NoError(t, err)
	require.Equal(t, key.ID, authed.ID)
	require.True(t, authed.HasScope(ScopeOrders))
	require.False(t, authed.HasScope("orders:write"))
	require.NotNil(t, authed.LastUsedAt, "last_used_at should be touched")
}

func TestAuthenticateRejectsUnknownAndMalformed(t *testing.T) {
	svc := NewMachineKeyService(&fakeKeyRepo{})
	svc.now = func() time.Time { return time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC) }

	if _, err := svc.Authenticate("mtk_does-not-exist"); !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("expected ErrInvalidKey, got %v", err)
	}
	if _, err := svc.Authenticate("not-a-key"); !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("expected ErrInvalidKey for non-prefixed token, got %v", err)
	}
}

func TestAuthenticateRejectsExpired(t *testing.T) {
	now := time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC)
	svc := NewMachineKeyService(&fakeKeyRepo{})
	svc.now = func() time.Time { return now }

	expired := now.Add(-time.Hour)
	_, key, err := svc.Generate(KeyOptions{Name: "exp", Scopes: []string{ScopeOrders}, ExpiresAt: &expired})
	require.NoError(t, err)
	_ = key

	plaintext2, _, err := svc.Generate(KeyOptions{Name: "exp2", Scopes: []string{ScopeOrders}, ExpiresAt: &expired})
	require.NoError(t, err)
	if _, err := svc.Authenticate(plaintext2); !errors.Is(err, ErrKeyExpired) {
		t.Fatalf("expected ErrKeyExpired, got %v", err)
	}
}

func TestAuthenticateRejectsRevoked(t *testing.T) {
	svc := NewMachineKeyService(&fakeKeyRepo{})
	svc.now = func() time.Time { return time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC) }

	plaintext, key, err := svc.Generate(KeyOptions{Name: "rev", Scopes: []string{ScopeOrders}})
	require.NoError(t, err)

	require.NoError(t, svc.Revoke(key.ID))
	if _, err := svc.Authenticate(plaintext); !errors.Is(err, ErrKeyRevoked) {
		t.Fatalf("expected ErrKeyRevoked, got %v", err)
	}
}

func TestRotateChangesHashAndPreservesMetadata(t *testing.T) {
	svc := NewMachineKeyService(&fakeKeyRepo{})
	svc.now = func() time.Time { return time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC) }

	oldPlain, key, err := svc.Generate(KeyOptions{Name: "rot", Scopes: []string{ScopeGST}})
	require.NoError(t, err)
	oldHash := key.KeyHash

	newPlain, err := svc.Rotate(key.ID)
	require.NoError(t, err)
	require.NotEqual(t, oldPlain, newPlain)

	// Old plaintext must no longer authenticate; new one must.
	if _, err := svc.Authenticate(oldPlain); !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("expected old key to fail, got %v", err)
	}
	authed, err := svc.Authenticate(newPlain)
	require.NoError(t, err)
	require.Equal(t, key.ID, authed.ID)
	require.NotEqual(t, oldHash, authed.KeyHash)
}

func TestRateLimit(t *testing.T) {
	now := time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC)
	svc := NewMachineKeyService(&fakeKeyRepo{})
	svc.now = func() time.Time { return now }

	plaintext, _, err := svc.Generate(KeyOptions{Name: "rl", Scopes: []string{ScopeOrders}, RateLimitPerMin: 2})
	require.NoError(t, err)

	if _, err := svc.Authenticate(plaintext); err != nil {
		t.Fatalf("call 1 should pass, got %v", err)
	}
	if _, err := svc.Authenticate(plaintext); err != nil {
		t.Fatalf("call 2 should pass, got %v", err)
	}
	if _, err := svc.Authenticate(plaintext); !errors.Is(err, ErrRateLimited) {
		t.Fatalf("call 3 should be rate limited, got %v", err)
	}

	// New minute resets the window.
	svc.now = func() time.Time { return now.Add(2 * time.Minute) }
	if _, err := svc.Authenticate(plaintext); err != nil {
		t.Fatalf("new window should pass, got %v", err)
	}
}

func TestValidateScopes(t *testing.T) {
	svc := NewMachineKeyService(&fakeKeyRepo{})

	if err := svc.ValidateScopes([]string{ScopeOrders, ScopeSystem}); err != nil {
		t.Fatalf("valid scopes rejected: %v", err)
	}
	if err := svc.ValidateScopes([]string{"orders:write"}); !errors.Is(err, ErrInvalidScope) {
		t.Fatalf("expected ErrInvalidScope, got %v", err)
	}
}

func TestHashKeyIsStable(t *testing.T) {
	a := HashKey("mtk_test")
	b := HashKey("mtk_test")
	if a != b {
		t.Fatal("hash should be deterministic")
	}
	if HashKey("mtk_test") == HashKey("mtk_other") {
		t.Fatal("distinct keys should hash differently")
	}
}

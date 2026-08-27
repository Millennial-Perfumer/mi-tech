package mcp

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"mi-tech/internal/mcp/entity"
	"mi-tech/internal/mcp/repository"
)

const (
	// KeyPrefix marks machine API keys so they are distinguishable from user JWTs.
	KeyPrefix = "mtk_"
	// DefaultRateLimitPerMin is applied when no limit is specified.
	DefaultRateLimitPerMin = 60
)

var (
	// ErrInvalidKey is returned when a key is malformed or unknown.
	ErrInvalidKey = errors.New("invalid machine API key")
	// ErrKeyRevoked is returned when the key has been revoked.
	ErrKeyRevoked = errors.New("machine API key revoked")
	// ErrKeyExpired is returned when the key has expired.
	ErrKeyExpired = errors.New("machine API key expired")
	// ErrRateLimited is returned when the key exceeds its per-minute limit.
	ErrRateLimited = errors.New("machine API key rate limit exceeded")
	// ErrInvalidScope is returned when a requested scope is not known.
	ErrInvalidScope = errors.New("invalid scope")
)

// KeyOptions controls machine key generation.
type KeyOptions struct {
	Name            string
	Scopes          []string
	RateLimitPerMin int
	ExpiresAt       *time.Time
}

// MachineKeyService manages machine-to-machine API keys for the MCP server.
// Only SHA-256 hashes of keys are persisted; plaintext keys are returned once
// at creation and cannot be recovered later.
type MachineKeyService struct {
	repo repository.MachineKeyRepository
	now  func() time.Time

	mu   sync.Mutex
	rate map[int64]*rateWindow
}

type rateWindow struct {
	start time.Time
	used  int
}

// NewMachineKeyService creates a MachineKeyService backed by the given repository.
func NewMachineKeyService(repo repository.MachineKeyRepository) *MachineKeyService {
	return &MachineKeyService{
		repo: repo,
		now:  time.Now,
		rate: make(map[int64]*rateWindow),
	}
}

// Generate creates a new machine key and returns the plaintext key exactly once.
// Callers are responsible for showing it to the operator; it is not persisted.
func (s *MachineKeyService) Generate(opts KeyOptions) (plaintext string, key *entity.MachineAPIKey, err error) {
	if err := s.ValidateScopes(opts.Scopes); err != nil {
		return "", nil, err
	}

	plaintext, hash, err := newKeyMaterial()
	if err != nil {
		return "", nil, err
	}

	rateLimit := opts.RateLimitPerMin
	if rateLimit <= 0 {
		rateLimit = DefaultRateLimitPerMin
	}

	now := s.now()
	key = &entity.MachineAPIKey{
		Name:            opts.Name,
		KeyHash:         hash,
		Scopes:          opts.Scopes,
		RateLimitPerMin: rateLimit,
		ExpiresAt:       opts.ExpiresAt,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.repo.Create(key); err != nil {
		return "", nil, err
	}
	return plaintext, key, nil
}

// Authenticate validates a plaintext key and returns its record.
// It rejects malformed, unknown, revoked, expired, and rate-limited keys.
func (s *MachineKeyService) Authenticate(plaintext string) (*entity.MachineAPIKey, error) {
	if !strings.HasPrefix(plaintext, KeyPrefix) {
		return nil, ErrInvalidKey
	}
	hash := HashKey(plaintext)

	key, err := s.repo.FindByHash(hash)
	if err != nil {
		return nil, ErrInvalidKey
	}

	now := s.now()
	if key.RevokedAt != nil {
		return nil, ErrKeyRevoked
	}
	if key.ExpiresAt != nil && now.After(*key.ExpiresAt) {
		return nil, ErrKeyExpired
	}
	if limiter, ok := s.repo.(repository.DistributedRateLimiter); ok {
		allowed, err := limiter.AllowRateLimit(key.ID, key.RateLimitPerMin, now)
		if err != nil {
			return nil, fmt.Errorf("rate limit check failed: %w", err)
		}
		if !allowed {
			return nil, ErrRateLimited
		}
	} else if !s.allow(key.ID, key.RateLimitPerMin, now) {
		// In-memory repositories remain supported for unit tests. Production
		// repositories implement DistributedRateLimiter for cross-instance limits.
		return nil, ErrRateLimited
	}

	// Best-effort last-used update; failure is not fatal for the request.
	_ = s.repo.TouchLastUsed(key.ID, now)
	return key, nil
}

// ValidateScopes ensures every scope is present in the allowlisted catalog.
func (s *MachineKeyService) ValidateScopes(scopes []string) error {
	if len(scopes) == 0 {
		return fmt.Errorf("%w: at least one scope is required", ErrInvalidScope)
	}
	known := make(map[string]struct{}, len(DefaultCatalog.Scopes()))
	for _, sc := range DefaultCatalog.Scopes() {
		known[sc] = struct{}{}
	}
	for _, sc := range scopes {
		if _, ok := known[sc]; !ok {
			return fmt.Errorf("%w: %s", ErrInvalidScope, sc)
		}
	}
	return nil
}

// Revoke marks a key as revoked.
func (s *MachineKeyService) Revoke(id int64) error {
	key, err := s.getByID(id)
	if err != nil {
		return err
	}
	now := s.now()
	key.RevokedAt = &now
	key.UpdatedAt = now
	return s.repo.Update(key)
}

// Rotate replaces a key's secret material while preserving metadata.
// It returns the new plaintext key exactly once.
func (s *MachineKeyService) Rotate(id int64) (plaintext string, err error) {
	key, err := s.getByID(id)
	if err != nil {
		return "", err
	}
	if key.RevokedAt != nil {
		return "", ErrKeyRevoked
	}

	plaintext, hash, err := newKeyMaterial()
	if err != nil {
		return "", err
	}

	key.KeyHash = hash
	key.UpdatedAt = s.now()
	if err := s.repo.Update(key); err != nil {
		return "", err
	}
	return plaintext, nil
}

// List returns all machine keys without their hashes.
func (s *MachineKeyService) List() ([]entity.MachineAPIKey, error) {
	return s.repo.List()
}

// getByID loads a key record by id.
func (s *MachineKeyService) getByID(id int64) (*entity.MachineAPIKey, error) {
	// Repository lookups are keyed by hash; emulate an id lookup by listing.
	keys, err := s.repo.List()
	if err != nil {
		return nil, err
	}
	for i := range keys {
		if keys[i].ID == id {
			return &keys[i], nil
		}
	}
	return nil, ErrInvalidKey
}

// allow enforces a per-minute sliding window rate limit for a key.
func (s *MachineKeyService) allow(id int64, limit int, now time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	w, ok := s.rate[id]
	if !ok || now.Sub(w.start) >= time.Minute {
		s.rate[id] = &rateWindow{start: now, used: 1}
		return true
	}
	if w.used >= limit {
		return false
	}
	w.used++
	return true
}

// newKeyMaterial generates a random plaintext key and its SHA-256 hex hash.
func newKeyMaterial() (plaintext, hash string, err error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", err
	}
	raw := KeyPrefix + base64.RawURLEncoding.EncodeToString(buf)
	return raw, HashKey(raw), nil
}

// HashKey computes the stable SHA-256 hex hash used for key storage/lookup.
func HashKey(plaintext string) string {
	sum := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(sum[:])
}

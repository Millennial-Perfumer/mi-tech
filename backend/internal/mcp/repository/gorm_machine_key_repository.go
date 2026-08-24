package repository

import (
	"time"

	"mi-tech/internal/mcp/entity"

	"gorm.io/gorm"
)

// gormMachineKeyRepository is the GORM implementation of MachineKeyRepository.
type gormMachineKeyRepository struct {
	db *gorm.DB
}

// NewMachineKeyRepository creates a new MachineKeyRepository.
func NewMachineKeyRepository(db *gorm.DB) MachineKeyRepository {
	return &gormMachineKeyRepository{db: db}
}

func (r *gormMachineKeyRepository) Create(key *entity.MachineAPIKey) error {
	return r.db.Raw(`
		INSERT INTO machine_api_keys (name, key_hash, scopes, rate_limit_per_min, expires_at, created_at, updated_at)
		VALUES (?, ?, ?::text[], ?, ?, NOW(), NOW())
		RETURNING id`, key.Name, key.KeyHash, scopesLiteral(key.Scopes), key.RateLimitPerMin, key.ExpiresAt,
	).Scan(&key.ID).Error
}

func (r *gormMachineKeyRepository) FindByHash(hash string) (*entity.MachineAPIKey, error) {
	var key entity.MachineAPIKey
	if err := r.db.Where("key_hash = ?", hash).First(&key).Error; err != nil {
		return nil, err
	}
	return &key, nil
}

func (r *gormMachineKeyRepository) List() ([]entity.MachineAPIKey, error) {
	var keys []entity.MachineAPIKey
	if err := r.db.Order("created_at DESC").Find(&keys).Error; err != nil {
		return nil, err
	}
	return keys, nil
}

func (r *gormMachineKeyRepository) Update(key *entity.MachineAPIKey) error {
	return r.db.Exec(`
		UPDATE machine_api_keys
		SET scopes = ?::text[], rate_limit_per_min = ?, expires_at = ?, revoked_at = ?, updated_at = NOW()
		WHERE id = ?`,
		scopesLiteral(key.Scopes), key.RateLimitPerMin, key.ExpiresAt, key.RevokedAt, key.ID,
	).Error
}

func (r *gormMachineKeyRepository) TouchLastUsed(id int64, at time.Time) error {
	return r.db.Model(&entity.MachineAPIKey{}).Where("id = ?", id).
		Update("last_used_at", at).Error
}

func (r *gormMachineKeyRepository) AllowRateLimit(keyID int64, limit int, now time.Time) (bool, error) {
	var allowed bool
	err := r.db.Raw(`
		INSERT INTO machine_api_key_rate_windows (key_id, window_start, used)
		VALUES (?, date_trunc('minute', ?::timestamptz), 1)
		ON CONFLICT (key_id) DO UPDATE SET
			window_start = CASE
				WHEN machine_api_key_rate_windows.window_start < date_trunc('minute', EXCLUDED.window_start)
				THEN EXCLUDED.window_start
				ELSE machine_api_key_rate_windows.window_start
			END,
			used = CASE
				WHEN machine_api_key_rate_windows.window_start < date_trunc('minute', EXCLUDED.window_start)
					THEN 1
				WHEN machine_api_key_rate_windows.used < ?
					THEN machine_api_key_rate_windows.used + 1
				ELSE machine_api_key_rate_windows.used
			END
		RETURNING used <= ?`, keyID, now, limit, limit).Scan(&allowed).Error
	return allowed, err
}

// scopesLiteral renders a Go []string as a Postgres array literal.
func scopesLiteral(scopes []string) string {
	if len(scopes) == 0 {
		return "{}"
	}
	var b []byte
	b = append(b, '{')
	for i, s := range scopes {
		if i > 0 {
			b = append(b, ',')
		}
		b = append(b, '"')
		for j := 0; j < len(s); j++ {
			c := s[j]
			if c == '"' || c == '\\' {
				b = append(b, '\\')
			}
			b = append(b, c)
		}
		b = append(b, '"')
	}
	b = append(b, '}')
	return string(b)
}

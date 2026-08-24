package repository

import (
	"encoding/json"
	"time"

	"mi-tech/internal/mcp/entity"

	"gorm.io/gorm"
)

// gormMachineKeyRepository is the GORM implementation of MachineKeyRepository.
type gormMachineKeyRepository struct {
	db *gorm.DB
}

// machineKeyRow keeps the PostgreSQL array at the SQL boundary. Scanning a
// text[] directly into []string is not supported consistently by the pgx
// database/sql adapter used by GORM; array_to_json gives us a portable scalar
// value that can be decoded into the entity's public []string field.
type machineKeyRow struct {
	ID              int64      `gorm:"column:id"`
	Name            string     `gorm:"column:name"`
	KeyHash         string     `gorm:"column:key_hash"`
	ScopesJSON      string     `gorm:"column:scopes_json"`
	RateLimitPerMin int        `gorm:"column:rate_limit_per_min"`
	ExpiresAt       *time.Time `gorm:"column:expires_at"`
	RevokedAt       *time.Time `gorm:"column:revoked_at"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at"`
	LastUsedAt      *time.Time `gorm:"column:last_used_at"`
}

func (r machineKeyRow) entity() (entity.MachineAPIKey, error) {
	key := entity.MachineAPIKey{
		ID:              r.ID,
		Name:            r.Name,
		KeyHash:         r.KeyHash,
		RateLimitPerMin: r.RateLimitPerMin,
		ExpiresAt:       r.ExpiresAt,
		RevokedAt:       r.RevokedAt,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
		LastUsedAt:      r.LastUsedAt,
	}
	if r.ScopesJSON == "" || r.ScopesJSON == "null" {
		return key, nil
	}
	if err := json.Unmarshal([]byte(r.ScopesJSON), &key.Scopes); err != nil {
		return entity.MachineAPIKey{}, err
	}
	return key, nil
}

const machineKeySelect = `
	SELECT id, name, key_hash,
	       COALESCE(array_to_json(scopes)::text, '[]') AS scopes_json,
	       rate_limit_per_min, expires_at, revoked_at,
	       created_at, updated_at, last_used_at
	FROM machine_api_keys`

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
	var row machineKeyRow
	result := r.db.Raw(machineKeySelect+` WHERE key_hash = ?`, hash).Scan(&row)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	key, err := row.entity()
	if err != nil {
		return nil, err
	}
	return &key, nil
}

func (r *gormMachineKeyRepository) List() ([]entity.MachineAPIKey, error) {
	var rows []machineKeyRow
	if err := r.db.Raw(machineKeySelect + ` ORDER BY created_at DESC`).Scan(&rows).Error; err != nil {
		return nil, err
	}
	keys := make([]entity.MachineAPIKey, 0, len(rows))
	for _, row := range rows {
		key, err := row.entity()
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
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

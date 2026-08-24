package repository

import (
	"time"

	"mi-tech/internal/mcp/entity"
)

// MachineKeyRepository defines database operations for machine API keys.
type MachineKeyRepository interface {
	// Create inserts a new machine key.
	Create(key *entity.MachineAPIKey) error
	// FindByHash loads a key by its SHA-256 hash.
	FindByHash(hash string) (*entity.MachineAPIKey, error)
	// List returns all machine keys (hashes excluded from output handling).
	List() ([]entity.MachineAPIKey, error)
	// Update persists key metadata (scopes, revocation, expiry, rate limit).
	Update(key *entity.MachineAPIKey) error
	// TouchLastUsed updates the last_used_at timestamp.
	TouchLastUsed(id int64, at time.Time) error
}

// DistributedRateLimiter provides an atomic, shared rate-limit window. The
// production repository implements this using PostgreSQL so limits hold
// across multiple API instances.
type DistributedRateLimiter interface {
	AllowRateLimit(keyID int64, limit int, now time.Time) (bool, error)
}

// AuditLogRepository defines database operations for MCP audit logs.
type AuditLogRepository interface {
	// Create inserts an audit log entry.
	Create(log *entity.MCPAuditLog) error
}

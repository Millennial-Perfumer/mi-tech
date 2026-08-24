package entity

import (
	"time"
)

// MachineAPIKey represents a machine-to-machine API key for the MCP server.
// Only the SHA-256 hash of the key is stored; the plaintext is never persisted.
type MachineAPIKey struct {
	ID              int64      `gorm:"column:id;primaryKey" json:"id"`
	Name            string     `gorm:"column:name" json:"name"`
	KeyHash         string     `gorm:"column:key_hash" json:"-"`
	Scopes          []string   `gorm:"column:scopes;type:text[]" json:"scopes"`
	RateLimitPerMin int        `gorm:"column:rate_limit_per_min" json:"rate_limit_per_min"`
	ExpiresAt       *time.Time `gorm:"column:expires_at" json:"expires_at"`
	RevokedAt       *time.Time `gorm:"column:revoked_at" json:"revoked_at"`
	CreatedAt       time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at" json:"updated_at"`
	LastUsedAt      *time.Time `gorm:"column:last_used_at" json:"last_used_at"`
}

// TableName matches the DB table name for MachineAPIKey.
func (MachineAPIKey) TableName() string { return "machine_api_keys" }

// IsActive reports whether the key is valid for use right now.
func (k *MachineAPIKey) IsActive(now time.Time) bool {
	if k.RevokedAt != nil {
		return false
	}
	if k.ExpiresAt != nil && now.After(*k.ExpiresAt) {
		return false
	}
	return true
}

// HasScope reports whether the key grants the given scope.
func (k *MachineAPIKey) HasScope(scope string) bool {
	for _, s := range k.Scopes {
		if s == scope {
			return true
		}
	}
	return false
}

// MCPAuditLog records a single MCP invocation. It intentionally omits the API
// key and authorization header to avoid leaking secrets.
type MCPAuditLog struct {
	ID         int64     `gorm:"column:id;primaryKey" json:"id"`
	KeyID      int64     `gorm:"column:key_id" json:"key_id"`
	KeyName    string    `gorm:"column:key_name" json:"key_name"`
	Scopes     []string  `gorm:"column:scopes;type:text[]" json:"scopes"`
	Tool       string    `gorm:"column:tool" json:"tool"`
	Outcome    string    `gorm:"column:outcome" json:"outcome"`
	Status     int       `gorm:"column:status" json:"status"`
	RemoteIP   string    `gorm:"column:remote_ip" json:"remote_ip"`
	RequestID  string    `gorm:"column:request_id" json:"request_id"`
	DurationMs int64     `gorm:"column:duration_ms" json:"duration_ms"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"created_at"`
}

// TableName matches the DB table name for MCPAuditLog.
func (MCPAuditLog) TableName() string { return "mcp_audit_logs" }

-- 132_mcp_audit_logs.sql
-- Audit log for every MCP invocation. Never stores the API key or the
-- Authorization header; it records tool, outcome, and metadata only.

CREATE TABLE IF NOT EXISTS mcp_audit_logs (
    id          BIGSERIAL PRIMARY KEY,
    key_id      BIGINT,
    key_name    TEXT,
    scopes      TEXT[] NOT NULL DEFAULT '{}',
    tool        TEXT NOT NULL,
    outcome     TEXT NOT NULL,
    status      INTEGER,
    remote_ip   TEXT,
    request_id  TEXT,
    duration_ms BIGINT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_mcp_audit_logs_created_at ON mcp_audit_logs (created_at);
CREATE INDEX IF NOT EXISTS idx_mcp_audit_logs_key_id ON mcp_audit_logs (key_id);
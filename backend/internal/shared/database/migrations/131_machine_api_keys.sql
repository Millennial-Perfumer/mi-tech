-- 131_machine_api_keys.sql
-- Dedicated machine-to-machine API keys for the read-only MCP server.
-- Only the SHA-256 hash of each key is stored; the plaintext key is shown once at creation.

CREATE TABLE IF NOT EXISTS machine_api_keys (
    id                BIGSERIAL PRIMARY KEY,
    name              TEXT NOT NULL,
    key_hash          TEXT NOT NULL UNIQUE,
    scopes            TEXT[] NOT NULL DEFAULT '{}',
    rate_limit_per_min INTEGER NOT NULL DEFAULT 60,
    expires_at        TIMESTAMPTZ,
    revoked_at        TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at      TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_machine_api_keys_key_hash ON machine_api_keys (key_hash);
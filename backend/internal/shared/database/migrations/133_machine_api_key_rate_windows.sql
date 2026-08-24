-- Shared per-minute rate-limit windows for horizontally scaled API instances.
CREATE TABLE IF NOT EXISTS machine_api_key_rate_windows (
    key_id       BIGINT PRIMARY KEY REFERENCES machine_api_keys(id) ON DELETE CASCADE,
    window_start TIMESTAMPTZ NOT NULL,
    used         INTEGER NOT NULL
);

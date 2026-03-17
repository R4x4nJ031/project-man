-- +goose Up
CREATE TABLE IF NOT EXISTS audit_logs (
    id TEXT PRIMARY KEY,
    actor_user_id TEXT NOT NULL,
    tenant_id TEXT NOT NULL,
    action TEXT NOT NULL,
    resource_type TEXT NOT NULL,
    resource_id TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS audit_logs_tenant_created_at_idx
    ON audit_logs (tenant_id, created_at DESC);

CREATE INDEX IF NOT EXISTS audit_logs_actor_created_at_idx
    ON audit_logs (actor_user_id, created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS audit_logs_actor_created_at_idx;
DROP INDEX IF EXISTS audit_logs_tenant_created_at_idx;
DROP TABLE IF EXISTS audit_logs;

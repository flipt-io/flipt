-- 001_init.sql
-- Core tables for approval workflow

CREATE TABLE IF NOT EXISTS approval_requests (
    id             UUID PRIMARY KEY,
    source_env     TEXT NOT NULL,
    target_env     TEXT NOT NULL,
    source_branch  TEXT,
    target_branch  TEXT,
    change_payload JSONB NOT NULL,
    status         TEXT NOT NULL CHECK (status IN ('PENDING', 'APPROVED', 'REJECTED')),
    requested_by   TEXT NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS approval_logs (
    id          BIGSERIAL PRIMARY KEY,
    request_id  UUID NOT NULL REFERENCES approval_requests(id) ON DELETE CASCADE,
    action      TEXT NOT NULL,
    actor       TEXT NOT NULL,
    timestamp   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

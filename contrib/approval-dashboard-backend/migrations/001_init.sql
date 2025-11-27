-- 001_init.sql
-- Flipt benzeri Approval Workflow için temel tablo şeması

CREATE TABLE IF NOT EXISTS approval_requests (
    id UUID PRIMARY KEY,
    source_env     TEXT        NOT NULL,
    target_env     TEXT        NOT NULL,
    change_payload JSONB       NOT NULL,
    status         TEXT        NOT NULL DEFAULT 'PENDING', -- PENDING / APPROVED / REJECTED
    requested_by   TEXT        NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS approval_logs (
    id SERIAL PRIMARY KEY,
    request_id UUID NOT NULL REFERENCES approval_requests(id) ON DELETE CASCADE,
    action     TEXT NOT NULL,       -- CREATED / APPROVED / REJECTED
    actor      TEXT NOT NULL,       -- yusuf, system, vs.
    timestamp  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

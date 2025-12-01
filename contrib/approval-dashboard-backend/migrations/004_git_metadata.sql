-- 004_git_metadata.sql
-- Add Git metadata to approval_requests

ALTER TABLE approval_requests
    ADD COLUMN IF NOT EXISTS repo_url      TEXT,
    ADD COLUMN IF NOT EXISTS source_commit TEXT,
    ADD COLUMN IF NOT EXISTS target_commit TEXT,
    ADD COLUMN IF NOT EXISTS change_type   TEXT;


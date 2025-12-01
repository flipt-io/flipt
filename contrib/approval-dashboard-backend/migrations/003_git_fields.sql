-- 003_git_fields.sql
-- Add Git branch fields to approval_requests

ALTER TABLE approval_requests
    ADD COLUMN IF NOT EXISTS source_branch TEXT,
    ADD COLUMN IF NOT EXISTS target_branch TEXT;

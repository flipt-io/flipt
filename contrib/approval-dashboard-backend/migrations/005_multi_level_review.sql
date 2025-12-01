-- 005_multi_level_review.sql
-- Multi-level approval: add review_state to track reviewer step

ALTER TABLE approval_requests
    ADD COLUMN IF NOT EXISTS review_state TEXT NOT NULL DEFAULT 'NONE';

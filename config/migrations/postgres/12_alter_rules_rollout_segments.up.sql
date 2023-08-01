-- Rules
ALTER TABLE rules DROP CONSTRAINT IF EXISTS rules_namespace_key_segment_key_fkey CASCADE;

ALTER TABLE rules DROP COLUMN segment_key;

ALTER TABLE rules ADD COLUMN segment_operator INTEGER NOT NULL DEFAULT 0;

-- Rollouts
ALTER TABLE rollout_segments DROP CONSTRAINT IF EXISTS rollout_segments_namespace_key_fkey CASCADE;
ALTER TABLE rollout_segments DROP CONSTRAINT IF EXISTS rollout_segments_namespace_key_segment_key_fkey CASCADE;

ALTER TABLE rollout_segments DROP COLUMN segment_key;
ALTER TABLE rollout_segments DROP COLUMN namespace_key;

ALTER TABLE rollout_segments ADD COLUMN segment_operator INTEGER NOT NULL DEFAULT 0;
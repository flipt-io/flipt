-- Rules
ALTER TABLE IF EXISTS rules DROP CONSTRAINT fk_namespace_key_ref_segments;

ALTER TABLE IF EXISTS rules DROP COLUMN segment_key;

ALTER TABLE IF EXISTS rules ADD COLUMN segment_operator INTEGER NOT NULL DEFAULT 0;

-- Rollouts
ALTER TABLE IF EXISTS rollout_segments DROP CONSTRAINT fk_namespace_key_ref_segments;
ALTER TABLE IF EXISTS rollout_segments DROP CONSTRAINT fk_namespace_key_ref_namespaces;

ALTER TABLE IF EXISTS rollout_segments DROP COLUMN segment_key;
ALTER TABLE IF EXISTS rollout_segments DROP COLUMN namespace_key;

ALTER TABLE IF EXISTS rollout_segments ADD COLUMN segment_operator INTEGER NOT NULL DEFAULT 0;
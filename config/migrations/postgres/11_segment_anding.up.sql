-- Rules
CREATE TABLE IF NOT EXISTS rule_segments (
  rule_id VARCHAR(255) NOT NULL REFERENCES rules ON DELETE CASCADE,
  namespace_key VARCHAR(255) NOT NULL,
  segment_key VARCHAR(255) NOT NULL,
  FOREIGN KEY (namespace_key, segment_key) REFERENCES segments (namespace_key, key) ON DELETE CASCADE
);

INSERT INTO rule_segments (rule_id, namespace_key, segment_key) SELECT id AS rule_id, namespace_key, segment_key FROM rules;

ALTER TABLE rules DROP CONSTRAINT IF EXISTS rules_namespace_key_segment_key_fkey CASCADE;

ALTER TABLE rules DROP COLUMN segment_key;

ALTER TABLE rules ADD COLUMN segment_operator INTEGER NOT NULL DEFAULT 0;

-- Rollouts
CREATE TABLE IF NOT EXISTS rollout_segment_references (
  rollout_segment_id VARCHAR(255) NOT NULL REFERENCES rollout_segments ON DELETE CASCADE,
  namespace_key VARCHAR(255) NOT NULL,
  segment_key VARCHAR(255) NOT NULL,
  FOREIGN KEY (namespace_key, segment_key) REFERENCES segments (namespace_key, key) ON DELETE CASCADE
);

INSERT INTO rollout_segment_references (rollout_segment_id, namespace_key, segment_key) SELECT id AS rollout_segment_id, namespace_key, segment_key FROM rollout_segments;

ALTER TABLE rollout_segments DROP CONSTRAINT IF EXISTS rollout_segments_namespace_key_fkey CASCADE;
ALTER TABLE rollout_segments DROP CONSTRAINT IF EXISTS rollout_segments_namespace_key_segment_key_fkey CASCADE;

ALTER TABLE rollout_segments DROP COLUMN segment_key;
ALTER TABLE rollout_segments DROP COLUMN namespace_key;

ALTER TABLE rollout_segments ADD COLUMN segment_operator INTEGER NOT NULL DEFAULT 0;
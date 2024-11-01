BEGIN;
-- Rules
CREATE TABLE IF NOT EXISTS rule_segments (
  rule_id VARCHAR(255) NOT NULL REFERENCES rules ON DELETE CASCADE,
  namespace_key VARCHAR(255) NOT NULL,
  segment_key VARCHAR(255) NOT NULL,
  UNIQUE (rule_id, namespace_key, segment_key),
  CONSTRAINT fk_namespace_key_ref_segments FOREIGN KEY (namespace_key, segment_key) REFERENCES segments (namespace_key, key) ON DELETE CASCADE
);
COMMIT;

INSERT INTO rule_segments (rule_id, namespace_key, segment_key) SELECT id AS rule_id, namespace_key, segment_key FROM rules;

BEGIN;
-- Rollouts
CREATE TABLE IF NOT EXISTS rollout_segment_references (
  rollout_segment_id VARCHAR(255) NOT NULL REFERENCES rollout_segments ON DELETE CASCADE,
  namespace_key VARCHAR(255) NOT NULL,
  segment_key VARCHAR(255) NOT NULL,
  UNIQUE (rollout_segment_id, namespace_key, segment_key),
  CONSTRAINT fk_namespace_key_ref_segments FOREIGN KEY (namespace_key, segment_key) REFERENCES segments (namespace_key, key) ON DELETE CASCADE
);
COMMIT;

INSERT INTO rollout_segment_references (rollout_segment_id, namespace_key, segment_key) SELECT id AS rollout_segment_id, namespace_key, segment_key FROM rollout_segments;

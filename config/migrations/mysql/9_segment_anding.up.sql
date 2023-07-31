-- Rules
CREATE TABLE IF NOT EXISTS rule_segments (
  rule_id VARCHAR(255) NOT NULL,
  namespace_key VARCHAR(255) NOT NULL,
  segment_key VARCHAR(255) NOT NULL,
  FOREIGN KEY (rule_id) REFERENCES rules (id) ON DELETE CASCADE,
  FOREIGN KEY (namespace_key, segment_key) REFERENCES segments (namespace_key, `key`) ON DELETE CASCADE
);

INSERT INTO rule_segments (rule_id, namespace_key, segment_key) SELECT id AS rule_id, namespace_key, segment_key FROM rules;

ALTER TABLE rules DROP FOREIGN KEY `rules_ibfk_3`;

ALTER TABLE rules DROP COLUMN segment_key;

ALTER TABLE rules ADD COLUMN segment_operator INTEGER NOT NULL DEFAULT 0;

-- Rollouts
CREATE TABLE IF NOT EXISTS rollout_segment_references (
  rollout_segment_id VARCHAR(255) NOT NULL,
  namespace_key VARCHAR(255) NOT NULL,
  segment_key VARCHAR(255) NOT NULL,
  FOREIGN KEY (rollout_segment_id) REFERENCES rollout_segments (id) ON DELETE CASCADE,
  FOREIGN KEY (namespace_key, segment_key) REFERENCES segments (namespace_key, `key`) ON DELETE CASCADE
);

INSERT INTO rollout_segment_references (rollout_segment_id, namespace_key, segment_key) SELECT id AS rollout_segment_id, namespace_key, segment_key FROM rollout_segments;

ALTER TABLE rollout_segments DROP FOREIGN KEY `rollout_segments_ibfk_1`;
ALTER TABLE rollout_segments DROP FOREIGN KEY `rollout_segments_ibfk_3`;

ALTER TABLE rollout_segments DROP COLUMN segment_key;
ALTER TABLE rollout_segments DROP COLUMN namespace_key;

ALTER TABLE rollout_segments ADD COLUMN segment_operator INTEGER NOT NULL DEFAULT 0;
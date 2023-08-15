-- Rules
CREATE TABLE IF NOT EXISTS rule_segments (
  rule_id VARCHAR(255) NOT NULL,
  namespace_key VARCHAR(255) NOT NULL,
  segment_key VARCHAR(255) NOT NULL,
  CONSTRAINT rule_id_namespace_segment UNIQUE (rule_id, namespace_key, segment_key),
  FOREIGN KEY (rule_id) REFERENCES rules (id) ON DELETE CASCADE,
  FOREIGN KEY (namespace_key, segment_key) REFERENCES segments (namespace_key, `key`) ON DELETE CASCADE
);

INSERT INTO rule_segments (rule_id, namespace_key, segment_key) SELECT id AS rule_id, namespace_key, segment_key FROM rules;

-- Rollouts
CREATE TABLE IF NOT EXISTS rollout_segment_references (
  rollout_segment_id VARCHAR(255) NOT NULL,
  namespace_key VARCHAR(255) NOT NULL,
  segment_key VARCHAR(255) NOT NULL,
  CONSTRAINT rollout_segment_id_namespace_segment UNIQUE (rollout_segment_id, namespace_key, segment_key),
  FOREIGN KEY (rollout_segment_id) REFERENCES rollout_segments (id) ON DELETE CASCADE,
  FOREIGN KEY (namespace_key, segment_key) REFERENCES segments (namespace_key, `key`) ON DELETE CASCADE
);

INSERT INTO rollout_segment_references (rollout_segment_id, namespace_key, segment_key) SELECT id AS rollout_segment_id, namespace_key, segment_key FROM rollout_segments;
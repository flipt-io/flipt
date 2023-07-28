-- Rules
CREATE TABLE IF NOT EXISTS rule_segments (
  rule_id VARCHAR(255) NOT NULL REFERENCES rules ON DELETE CASCADE,
  namespace_key VARCHAR(255) NOT NULL,
  segment_key VARCHAR(255) NOT NULL,
  FOREIGN KEY (namespace_key, segment_key) REFERENCES segments (namespace_key, key) ON DELETE CASCADE
);

INSERT INTO rule_segments (rule_id, namespace_key, segment_key) SELECT id AS rule_id, namespace_key, segment_key FROM rules;

CREATE TABLE IF NOT EXISTS rules_temp (
  id VARCHAR(255) PRIMARY KEY UNIQUE NOT NULL,
  flag_key VARCHAR(255) NOT NULL,
  rank INTEGER DEFAULT 1 NOT NULL,
  segment_operator INTEGER DEFAULT 0 NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  namespace_key VARCHAR(255) NOT NULL DEFAULT 'default' REFERENCES namespaces ON DELETE CASCADE,
  FOREIGN KEY (namespace_key, flag_key) REFERENCES flags (namespace_key, key) ON DELETE CASCADE
);

INSERT INTO rules_temp (id, flag_key, rank, created_at, updated_at, namespace_key) SELECT id, flag_key, rank, created_at, updated_at, namespace_key FROM rules;

DROP TABLE rules;

ALTER TABLE rules_temp RENAME TO rules;

-- Rollouts
CREATE TABLE IF NOT EXISTS rollout_segment_references (
  rollout_segment_id VARCHAR(255) NOT NULL REFERENCES rollout_segments ON DELETE CASCADE,
  namespace_key VARCHAR(255) NOT NULL,
  segment_key VARCHAR(255) NOT NULL,
  FOREIGN KEY (namespace_key, segment_key) REFERENCES segments (namespace_key, key) ON DELETE CASCADE
);

INSERT INTO rollout_segment_references (rollout_segment_id, namespace_key, segment_key) SELECT id AS rollout_segment_id, namespace_key, segment_key FROM rollout_segments;

CREATE TABLE IF NOT EXISTS rollout_segments_temp (
  id VARCHAR(255) PRIMARY KEY UNIQUE NOT NULL,
  rollout_id VARCHAR(255) NOT NULL REFERENCES rollouts ON DELETE CASCADE,
  value BOOLEAN DEFAULT FALSE NOT NULL,
  segment_operator INTEGER DEFAULT 0 NOT NULL
);

INSERT INTO rollout_segments_temp (id, rollout_id, value) SELECT id, rollout_id, value FROM rollout_segments;

DROP TABLE rollout_segments;

ALTER TABLE rollout_segments_temp RENAME TO rollout_segments;

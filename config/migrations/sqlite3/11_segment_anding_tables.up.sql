-- Rules
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

-- Copy data from distributions table to temporary distributions table since distributions depends on rules
CREATE TABLE distributions_temp (
  id VARCHAR(255) PRIMARY KEY UNIQUE NOT NULL,
  rule_id VARCHAR(255) NOT NULL REFERENCES rules_temp ON DELETE CASCADE,
  variant_id VARCHAR(255) NOT NULL REFERENCES variants ON DELETE CASCADE,
  rollout float DEFAULT 0 NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

INSERT INTO distributions_temp (id, rule_id, variant_id, rollout, created_at, updated_at)
  SELECT id, rule_id, variant_id, rollout, created_at, updated_at
  FROM distributions;

CREATE TABLE IF NOT EXISTS rule_segments (
  rule_id VARCHAR(255) NOT NULL REFERENCES rules_temp ON DELETE CASCADE,
  namespace_key VARCHAR(255) NOT NULL,
  segment_key VARCHAR(255) NOT NULL,
  UNIQUE (rule_id, namespace_key, segment_key),
  FOREIGN KEY (namespace_key, segment_key) REFERENCES segments (namespace_key, key) ON DELETE CASCADE
);

INSERT INTO rule_segments (rule_id, namespace_key, segment_key) SELECT id AS rule_id, namespace_key, segment_key FROM rules;

-- Rollouts
CREATE TABLE IF NOT EXISTS rollout_segments_temp (
  id VARCHAR(255) PRIMARY KEY UNIQUE NOT NULL,
  rollout_id VARCHAR(255) NOT NULL REFERENCES rollouts ON DELETE CASCADE,
  value BOOLEAN DEFAULT FALSE NOT NULL,
  segment_operator INTEGER DEFAULT 0 NOT NULL
);

INSERT INTO rollout_segments_temp (id, rollout_id, value) SELECT id, rollout_id, value FROM rollout_segments;

CREATE TABLE IF NOT EXISTS rollout_segment_references (
  rollout_segment_id VARCHAR(255) NOT NULL REFERENCES rollout_segments_temp ON DELETE CASCADE,
  namespace_key VARCHAR(255) NOT NULL,
  segment_key VARCHAR(255) NOT NULL,
  UNIQUE (rollout_segment_id, namespace_key, segment_key),
  FOREIGN KEY (namespace_key, segment_key) REFERENCES segments (namespace_key, key) ON DELETE CASCADE
);

INSERT INTO rollout_segment_references (rollout_segment_id, namespace_key, segment_key) SELECT id AS rollout_segment_id, namespace_key, segment_key FROM rollout_segments;

-- Drop old rules table
DROP TABLE rules;

-- Rename temporary rules table to rules
ALTER TABLE rules_temp RENAME TO rules;

-- Drop old rollout_segments table
DROP TABLE rollout_segments;

-- Rename temporary rollout_segments table to rollout_segments
ALTER TABLE rollout_segments_temp RENAME TO rollout_segments;

-- Drop distributions table
DROP TABLE distributions;

-- Rename distributions
ALTER TABLE distributions_temp RENAME TO distributions;
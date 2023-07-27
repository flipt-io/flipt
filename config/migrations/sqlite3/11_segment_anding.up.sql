-- Create new rule_segments table
CREATE TABLE IF NOT EXISTS rule_segments (
  rule_id VARCHAR(255) NOT NULL REFERENCES rules ON DELETE CASCADE,
  namespace_key VARCHAR(255) NOT NULL,
  segment_key VARCHAR(255) NOT NULL,
  FOREIGN KEY (namespace_key, segment_key) REFERENCES segments (namespace_key, key) ON DELETE CASCADE
);

INSERT INTO rule_segments (rule_id, namespace_key, segment_key) SELECT id AS rule_id, namespace_key, segment_key FROM rules;

-- Create temporary rules table
CREATE TABLE IF NOT EXISTS rules_new (
  id VARCHAR(255) PRIMARY KEY UNIQUE NOT NULL,
  flag_key VARCHAR(255) NOT NULL,
  rank INTEGER DEFAULT 1 NOT NULL,
  segment_match_type INTEGER DEFAULT 0 NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  namespace_key VARCHAR(255) NOT NULL DEFAULT 'default' REFERENCES namespaces ON DELETE CASCADE,
  FOREIGN KEY (namespace_key, flag_key) REFERENCES flags (namespace_key, key) ON DELETE CASCADE
);

INSERT INTO rules_new (id, flag_key, rank, created_at, updated_at, namespace_key) SELECT id, flag_key, rank, created_at, updated_at, namespace_key FROM rules;

DROP TABLE rules;

ALTER TABLE rules_new RENAME TO rules;

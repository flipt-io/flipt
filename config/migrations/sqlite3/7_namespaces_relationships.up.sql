-- Create temporary flags table
CREATE TABLE IF NOT EXISTS flags_temp (
  key VARCHAR(255) NOT NULL,
  name VARCHAR(255) NOT NULL,
  description TEXT NOT NULL,
  enabled BOOLEAN DEFAULT FALSE NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  namespace_key VARCHAR(255) NOT NULL DEFAULT 'default' REFERENCES namespaces ON DELETE CASCADE,
  PRIMARY KEY (namespace_key, key)
);

-- Copy data from flags table to temporary flags table
INSERT INTO flags_temp (key, name, description, enabled, created_at, updated_at)
  SELECT key, name, description, enabled, created_at, updated_at
  FROM flags;

-- Create temporary variants table
CREATE TABLE IF NOT EXISTS variants_temp (
  id VARCHAR(255) PRIMARY KEY UNIQUE NOT NULL,
  flag_key VARCHAR(255) NOT NULL,
  key VARCHAR(255) NOT NULL,
  name VARCHAR(255) NOT NULL,
  description TEXT NOT NULL,
  attachment TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  namespace_key VARCHAR(255) NOT NULL DEFAULT 'default' REFERENCES namespaces ON DELETE CASCADE,
  UNIQUE (namespace_key, flag_key, key),
  FOREIGN KEY (namespace_key, flag_key) REFERENCES flags_temp (namespace_key, key) ON DELETE CASCADE
);

-- Copy data from variants table to temporary variants table
INSERT INTO variants_temp (id, flag_key, key, name, description, attachment, created_at, updated_at)
  SELECT id, flag_key, key, name, description, attachment, created_at, updated_at
  FROM variants;

-- Create temporary segments table
CREATE TABLE IF NOT EXISTS segments_temp (
  key VARCHAR(255) NOT NULL,
  name VARCHAR(255) NOT NULL,
  description TEXT NOT NULL,
  match_type INTEGER DEFAULT 0 NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  namespace_key VARCHAR(255) NOT NULL DEFAULT 'default' REFERENCES namespaces ON DELETE CASCADE,
  PRIMARY KEY (namespace_key, key)
);

-- Copy data from segments table to temporary segments table
INSERT INTO segments_temp (key, name, description, match_type, created_at, updated_at)
  SELECT key, name, description, match_type, created_at, updated_at
  FROM segments;

-- Create temporary constraints table
CREATE TABLE IF NOT EXISTS constraints_temp (
  id VARCHAR(255) PRIMARY KEY UNIQUE NOT NULL,
  segment_key VARCHAR(255) NOT NULL,
  type INTEGER DEFAULT 0 NOT NULL,
  property VARCHAR(255) NOT NULL,
  operator VARCHAR(255) NOT NULL,
  value TEXT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  namespace_key VARCHAR(255) NOT NULL DEFAULT 'default' REFERENCES namespaces ON DELETE CASCADE,
  FOREIGN KEY (namespace_key, segment_key) REFERENCES segments_temp (namespace_key, key) ON DELETE CASCADE
);

-- Copy data from constraints table to temporary constraints table
INSERT INTO constraints_temp (id, segment_key, type, property, operator, value, created_at, updated_at)
  SELECT id, segment_key, type, property, operator, value, created_at, updated_at
  FROM constraints;

-- Create temporary rules table
CREATE TABLE IF NOT EXISTS rules_temp (
  id VARCHAR(255) PRIMARY KEY UNIQUE NOT NULL,
  flag_key VARCHAR(255) NOT NULL,
  segment_key VARCHAR(255) NOT NULL,
  rank INTEGER DEFAULT 1 NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  namespace_key VARCHAR(255) NOT NULL DEFAULT 'default' REFERENCES namespaces ON DELETE CASCADE,
  FOREIGN KEY (namespace_key, flag_key) REFERENCES flags_temp (namespace_key, key) ON DELETE CASCADE,
  FOREIGN KEY (namespace_key, segment_key) REFERENCES segments_temp (namespace_key, key) ON DELETE CASCADE
);

-- Copy data from rules table to temporary rules table
INSERT INTO rules_temp (id, flag_key, segment_key, rank, created_at, updated_at)
  SELECT id, flag_key, segment_key, rank, created_at, updated_at
  FROM rules;

-- Copy data from distributions table to temporary distributions table
CREATE TABLE distributions_temp (
  id VARCHAR(255) PRIMARY KEY UNIQUE NOT NULL,
  rule_id VARCHAR(255) NOT NULL REFERENCES rules_temp ON DELETE CASCADE,
  variant_id VARCHAR(255) NOT NULL REFERENCES variants_temp ON DELETE CASCADE,
  rollout float DEFAULT 0 NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

INSERT INTO distributions_temp (id, rule_id, variant_id, rollout, created_at, updated_at)
  SELECT id, rule_id, variant_id, rollout, created_at, updated_at
  FROM distributions;

-- Drop old distributions table
DROP TABLE distributions;
-- Rename temporary distributions table to distributions
ALTER TABLE distributions_temp RENAME TO distributions;

-- Drop old rules table
DROP TABLE rules;
-- Rename temporary rules table to rules
ALTER TABLE rules_temp RENAME TO rules;

-- Drop old flags table
DROP TABLE flags;
-- Rename temporary flags table to flags
ALTER TABLE flags_temp RENAME TO flags;

-- Drop old variants table
DROP TABLE variants;
-- Rename temporary variants table to variants
ALTER TABLE variants_temp RENAME TO variants;

-- Drop old segments table
DROP TABLE segments;
-- Rename temporary segments table to segments
ALTER TABLE segments_temp RENAME TO segments;

-- Drop old constraints table
DROP TABLE constraints;
-- Rename temporary constraints table to constraints
ALTER TABLE constraints_temp RENAME TO constraints;

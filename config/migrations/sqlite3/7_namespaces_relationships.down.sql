PRAGMA foreign_keys=off;

CREATE TABLE IF NOT EXISTS rules_temp (
  id VARCHAR(255) PRIMARY KEY UNIQUE NOT NULL,
  flag_key VARCHAR(255) NOT NULL,
  segment_key VARCHAR(255) NOT NULL,
  rank INTEGER DEFAULT 1 NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

INSERT INTO rules_temp (id, flag_key, segment_key, rank, created_at, updated_at)
  SELECT id, flag_key, segment_key, rank, created_at, updated_at
  FROM rules;

DROP TABLE rules;
ALTER TABLE rules_temp RENAME TO rules;

CREATE TABLE IF NOT EXISTS flags_temp (
  key VARCHAR(255) NOT NULL,
  name VARCHAR(255) NOT NULL,
  description TEXT NOT NULL,
  enabled BOOLEAN DEFAULT FALSE NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  PRIMARY KEY (key)
);

INSERT INTO flags_temp (key, name, description, enabled, created_at, updated_at)
  SELECT key, name, description, enabled, created_at, updated_at
  FROM flags;

DROP TABLE flags;
ALTER TABLE flags_temp RENAME TO flags;

CREATE TABLE IF NOT EXISTS variants_temp (
  id VARCHAR(255) PRIMARY KEY UNIQUE NOT NULL,
  flag_key VARCHAR(255) NOT NULL,
  key VARCHAR(255) NOT NULL,
  name VARCHAR(255) NOT NULL,
  description TEXT NOT NULL,
  attachment TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  UNIQUE (flag_key, key),
  FOREIGN KEY (flag_key) REFERENCES flags (key) ON DELETE CASCADE
);

INSERT INTO variants_temp (id, flag_key, key, name, description, attachment, created_at, updated_at)
  SELECT id, flag_key, key, name, description, attachment, created_at, updated_at
  FROM variants;

DROP TABLE variants;
ALTER TABLE variants_temp RENAME TO variants;

CREATE TABLE IF NOT EXISTS segments_temp (
  key VARCHAR(255) NOT NULL,
  name VARCHAR(255) NOT NULL,
  description TEXT NOT NULL,
  match_type INTEGER DEFAULT 0 NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  PRIMARY KEY (key)
);

INSERT INTO segments_temp (key, name, description, match_type, created_at, updated_at)
  SELECT key, name, description, match_type, created_at, updated_at
  FROM segments;

DROP TABLE segments;
ALTER TABLE segments_temp RENAME TO segments;

CREATE TABLE IF NOT EXISTS constraints_temp (
  id VARCHAR(255) PRIMARY KEY UNIQUE NOT NULL,
  segment_key VARCHAR(255) NOT NULL,
  type INTEGER DEFAULT 0 NOT NULL,
  property VARCHAR(255) NOT NULL,
  operator VARCHAR(255) NOT NULL,
  value TEXT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  FOREIGN KEY (segment_key) REFERENCES segments (key) ON DELETE CASCADE
);

INSERT INTO constraints_temp (id, segment_key, type, property, operator, value, created_at, updated_at)
  SELECT id, segment_key, type, property, operator, value, created_at, updated_at
  FROM constraints;

DROP TABLE constraints;
ALTER TABLE constraints_temp RENAME TO constraints;

CREATE TABLE IF NOT EXISTS rules_temp (
  id VARCHAR(255) PRIMARY KEY UNIQUE NOT NULL,
  flag_key VARCHAR(255) NOT NULL,
  segment_key VARCHAR(255) NOT NULL,
  rank INTEGER DEFAULT 1 NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  FOREIGN KEY (flag_key) REFERENCES flags (key) ON DELETE CASCADE,
  FOREIGN KEY (segment_key) REFERENCES segments (key) ON DELETE CASCADE
);

INSERT INTO rules_temp (id, flag_key, segment_key, rank, created_at, updated_at)
  SELECT id, flag_key, segment_key, rank, created_at, updated_at
  FROM rules;

DROP TABLE rules;
ALTER TABLE rules_temp RENAME TO rules;

PRAGMA foreign_keys=on;

PRAGMA foreign_keys=off;

/* Create namespaces table */
CREATE TABLE IF NOT EXISTS namespaces (
  key VARCHAR(255) PRIMARY KEY UNIQUE NOT NULL,
  name VARCHAR(255) NOT NULL,
  description TEXT NOT NULL,
  protected BOOLEAN DEFAULT FALSE NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

/* Create default namespace */
INSERT INTO namespaces (key, name, description, protected) VALUES ('default', 'Default', 'Default namespace', true);

/* SQLite doesn't allow you to drop unique constraints with ALTER TABLE
   so we have to create a new table with the schema we want and copy the data over.
   https://www.sqlite.org/lang_altertable.html
*/

/* Create temp tables */

/* flags */
CREATE TABLE IF NOT EXISTS flags_temp (
  key VARCHAR(255) NOT NULL,
  name VARCHAR(255) NOT NULL,
  description TEXT NOT NULL,
  enabled BOOLEAN DEFAULT FALSE NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  namespace_key VARCHAR(255) NOT NULL DEFAULT 'default' REFERENCES namespaces ON DELETE CASCADE,
  PRIMARY KEY (namespace_key, key)); /* composite primary key unique by default */

INSERT INTO flags_temp (key, name, description, enabled, created_at, updated_at)
  SELECT key, name, description, enabled, created_at, updated_at
  FROM flags;

DROP TABLE flags;

ALTER TABLE flags_temp RENAME TO flags;

/* variants */
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
  FOREIGN KEY (namespace_key, flag_key) REFERENCES flags (namespace_key, key) ON DELETE CASCADE
);

INSERT INTO variants_temp (id, flag_key, key, name, description, attachment, created_at, updated_at)
  SELECT id, flag_key, key, name, description, attachment, created_at, updated_at
  FROM variants;

DROP TABLE variants;

ALTER TABLE variants_temp RENAME TO variants;

/* rules */
CREATE TABLE IF NOT EXISTS rules_temp (
  id VARCHAR(255) PRIMARY KEY UNIQUE NOT NULL,
  flag_key VARCHAR(255) NOT NULL,
  segment_key VARCHAR(255) NOT NULL REFERENCES segments ON DELETE CASCADE,
  rank INTEGER DEFAULT 1 NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  namespace_key VARCHAR(255) NOT NULL DEFAULT 'default' REFERENCES namespaces ON DELETE CASCADE,
  FOREIGN KEY (namespace_key, flag_key) REFERENCES flags (namespace_key, key) ON DELETE CASCADE
);

INSERT INTO rules_temp (id, flag_key, segment_key, rank, created_at, updated_at)
  SELECT id, flag_key, segment_key, rank, created_at, updated_at
  FROM rules;

DROP TABLE rules;

ALTER TABLE rules_temp RENAME TO rules;

PRAGMA foreign_keys=on;
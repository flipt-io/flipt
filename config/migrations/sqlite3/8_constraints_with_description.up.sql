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
  description TEXT,
  FOREIGN KEY (namespace_key, segment_key) REFERENCES segments (namespace_key, key) ON DELETE CASCADE
);

-- Copy data from constraints table to temporary constraints table
INSERT INTO constraints_temp (id, segment_key, type, property, operator, value, created_at, updated_at, namespace_key)
  SELECT id, segment_key, type, property, operator, value, created_at, updated_at, namespace_key
  FROM constraints;

 -- Drop old constraints table
DROP TABLE constraints;

-- Rename temporary constraints table to constraints
ALTER TABLE constraints_temp RENAME TO constraints;

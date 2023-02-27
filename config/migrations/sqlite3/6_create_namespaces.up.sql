PRAGMA foreign_keys=off;

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

/* Add namespace_key column to all tables */
ALTER TABLE flags ADD COLUMN namespace_key VARCHAR(255) NOT NULL DEFAULT 'default' REFERENCES namespaces ON DELETE CASCADE;
ALTER TABLE segments ADD COLUMN namespace_key VARCHAR(255) NOT NULL DEFAULT 'default' REFERENCES namespaces ON DELETE CASCADE;
ALTER TABLE variants ADD COLUMN namespace_key VARCHAR(255) NOT NULL DEFAULT 'default' REFERENCES namespaces ON DELETE CASCADE;
ALTER TABLE constraints ADD COLUMN namespace_key VARCHAR(255) NOT NULL DEFAULT 'default' REFERENCES namespaces ON DELETE CASCADE;

/* Add unique indexes per namespace */
CREATE UNIQUE INDEX IF NOT EXISTS flags_namespace_key_idx ON flags (namespace_key, key);
CREATE UNIQUE INDEX IF NOT EXISTS segments_namespace_key_idx ON segments (namespace_key, key);

PRAGMA foreign_keys=on;
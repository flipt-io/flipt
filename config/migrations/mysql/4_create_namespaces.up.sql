/* Create namespaces table */
CREATE TABLE IF NOT EXISTS namespaces (
  `key` VARCHAR(255) PRIMARY KEY UNIQUE NOT NULL,
  name VARCHAR(255) NOT NULL,
  description TEXT NOT NULL,
  protected BOOLEAN DEFAULT FALSE NOT NULL,
  created_at TIMESTAMP(6) DEFAULT CURRENT_TIMESTAMP(6) NOT NULL,
  updated_at TIMESTAMP(6) DEFAULT CURRENT_TIMESTAMP(6) NOT NULL
);

/* Create default namespace */
INSERT INTO namespaces (`key`, name, description, protected) VALUES ('default', 'Default', 'Default namespace', true);
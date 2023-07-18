CREATE TABLE IF NOT EXISTS rollouts (
  id VARCHAR(255) UNIQUE NOT NULL,
  namespace_key VARCHAR(255) NOT NULL,
  flag_key VARCHAR(255) NOT NULL,
  type INTEGER DEFAULT 0 NOT NULL,
  description TEXT NOT NULL,
  `rank` INTEGER DEFAULT 1 NOT NULL,
  created_at TIMESTAMP(6) DEFAULT CURRENT_TIMESTAMP(6) NOT NULL,
  updated_at TIMESTAMP(6) DEFAULT CURRENT_TIMESTAMP(6) NOT NULL,
  PRIMARY KEY (id),
  FOREIGN KEY (namespace_key) REFERENCES namespaces (`key`) ON DELETE CASCADE,
  FOREIGN KEY (namespace_key, flag_key) REFERENCES flags (namespace_key, `key`) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS rollout_thresholds (
  id VARCHAR(255) UNIQUE NOT NULL,
  namespace_key VARCHAR(255) NOT NULL,
  rollout_id VARCHAR(255) UNIQUE NOT NULL,
  percentage float DEFAULT 0 NOT NULL,
  value BOOLEAN DEFAULT FALSE NOT NULL,
  PRIMARY KEY (id),
  FOREIGN KEY (namespace_key) REFERENCES namespaces (`key`) ON DELETE CASCADE,
  FOREIGN KEY (rollout_id) REFERENCES rollouts (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS rollout_segments (
  id VARCHAR(255) UNIQUE NOT NULL,
  namespace_key VARCHAR(255) NOT NULL,
  rollout_id VARCHAR(255) NOT NULL,
  segment_key VARCHAR(255) NOT NULL,
  value BOOLEAN DEFAULT FALSE NOT NULL,
  PRIMARY KEY (id),
  FOREIGN KEY (namespace_key) REFERENCES namespaces (`key`) ON DELETE CASCADE,
  FOREIGN KEY (rollout_id) REFERENCES rollouts (id) ON DELETE CASCADE,
  FOREIGN KEY (namespace_key, segment_key) REFERENCES segments (namespace_key, `key`) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS operation_lock (
  operation VARCHAR(255) NOT NULL,
  version INTEGER DEFAULT 0 NOT NULL CHECK (version >= 0),
  last_acquired_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  acquired_until TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  PRIMARY KEY (operation)
);

CREATE INDEX IF NOT EXISTS idx_lock_timestamps ON operation_lock(last_acquired_at, acquired_until);

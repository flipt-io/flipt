-- Operation lock table for managing distributed operation locks
CREATE TABLE IF NOT EXISTS operation_lock (
  operation VARCHAR(255) UNIQUE NOT NULL,
  version INTEGER DEFAULT 0 NOT NULL,
  last_acquired_at TIMESTAMP(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  acquired_until TIMESTAMP(6) DEFAULT CURRENT_TIMESTAMP(6),
  PRIMARY KEY (`operation`),
  INDEX idx_lock_timestamps (last_acquired_at, acquired_until)
);

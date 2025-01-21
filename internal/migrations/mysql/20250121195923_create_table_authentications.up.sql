CREATE TABLE IF NOT EXISTS authentications (
  id VARCHAR(255) UNIQUE NOT NULL,
  hashed_client_token VARCHAR(255) UNIQUE NOT NULL,
  method INTEGER DEFAULT 0 NOT NULL CHECK (method >= 0),
  metadata TEXT,
  expires_at TIMESTAMP(6),
  created_at TIMESTAMP(6) DEFAULT CURRENT_TIMESTAMP(6) NOT NULL,
  updated_at TIMESTAMP(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6) NOT NULL,
  PRIMARY KEY (`id`),
  INDEX idx_authentications_expires_at (expires_at)
);

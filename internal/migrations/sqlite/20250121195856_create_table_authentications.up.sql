CREATE TABLE IF NOT EXISTS authentications (
  id VARCHAR(255) NOT NULL,
  hashed_client_token VARCHAR(255) UNIQUE NOT NULL,
  method INTEGER DEFAULT 0 NOT NULL CHECK (method >= 0),
  metadata TEXT,
  expires_at TIMESTAMP,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_authentications_expires_at ON authentications(expires_at);
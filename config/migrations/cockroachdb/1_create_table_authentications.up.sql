CREATE TABLE IF NOT EXISTS authentications (
  id VARCHAR(255) PRIMARY KEY UNIQUE NOT NULL,
  hashed_client_token VARCHAR(255) UNIQUE NOT NULL,
  method VARCHAR(255) NOT NULL,
  metadata TEXT,
  expires_at TIMESTAMP,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE UNIQUE INDEX hashed_client_token_authentications_index ON authentications (hashed_client_token);

CREATE TABLE IF NOT EXISTS authentications (
  id VARCHAR(255) UNIQUE NOT NULL,
  hashed_client_token VARCHAR(255) UNIQUE NOT NULL,
  method INTEGER DEFAULT 0 NOT NULL,
  metadata TEXT,
  expires_at TIMESTAMP(6),
  created_at TIMESTAMP(6) DEFAULT CURRENT_TIMESTAMP(6) NOT NULL,
  updated_at TIMESTAMP(6) DEFAULT CURRENT_TIMESTAMP(6) NOT NULL,
  PRIMARY KEY (`id`)
);

CREATE UNIQUE INDEX hashed_client_token_authentications_index ON authentications (hashed_client_token);

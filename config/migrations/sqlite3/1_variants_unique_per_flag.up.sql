/* SQLite doesn't allow you to drop unique constraints with ALTER TABLE
   so we have to create a new table with the schema we want and copy the data over.
   https://www.sqlite.org/lang_altertable.html
*/

PRAGMA foreign_keys=off;

CREATE TABLE variants_temp
(
  id VARCHAR(255) PRIMARY KEY UNIQUE NOT NULL,
  flag_key VARCHAR(255) NOT NULL REFERENCES flags ON DELETE CASCADE,
  key VARCHAR(255) NOT NULL,
  name VARCHAR(255) NOT NULL,
  description TEXT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  UNIQUE (flag_key, key)
);

INSERT INTO variants_temp (id, flag_key, key, name, description, created_at, updated_at)
  SELECT id, flag_key, key, name, description, created_at, updated_at
  FROM variants;

DROP TABLE variants;

ALTER TABLE variants_temp RENAME TO variants;

PRAGMA foreign_keys=on;

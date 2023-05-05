PRAGMA foreign_keys = 0;
-- Add description column to constraints
ALTER TABLE constraints ADD COLUMN description TEXT NOT NULL DEFAULT 'desc';
PRAGMA foriegn_keys = 1;
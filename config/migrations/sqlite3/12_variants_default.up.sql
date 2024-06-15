ALTER TABLE variants ADD COLUMN `default` BOOLEAN DEFAULT FALSE NOT NULL;
CREATE UNIQUE INDEX variants_default_unique_index ON variants (namespace_key, flag_key) WHERE `default` = TRUE;
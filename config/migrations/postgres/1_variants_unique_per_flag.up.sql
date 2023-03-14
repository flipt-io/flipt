ALTER TABLE variants DROP CONSTRAINT IF EXISTS variants_key_key;
ALTER TABLE variants ADD UNIQUE(flag_key, key);

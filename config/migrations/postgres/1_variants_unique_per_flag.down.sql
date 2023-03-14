ALTER TABLE variants DROP CONSTRAINT IF EXISTS variants_flag_key_key_key;
ALTER TABLE variants ADD UNIQUE(key);

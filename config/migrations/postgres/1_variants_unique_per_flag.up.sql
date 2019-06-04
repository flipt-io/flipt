ALTER TABLE variants DROP CONSTRAINT variants_key_key;
ALTER TABLE variants ADD UNIQUE(flag_key, key);

-- Add column namespace_key with a default value
ALTER TABLE flags ADD COLUMN namespace_key VARCHAR(255) NOT NULL DEFAULT 'default';

-- Add foreign key constraint on namespace_key column referencing key column of namespaces table
ALTER TABLE flags ADD FOREIGN KEY (namespace_key) REFERENCES namespaces(key) ON DELETE CASCADE;

-- Drop primary key constraint and add a new composite primary key on namespace_key and key columns
ALTER TABLE flags ALTER PRIMARY KEY USING COLUMNS (namespace_key, key);
DROP INDEX IF EXISTS flags_key_key CASCADE;

ALTER TABLE variants ADD COLUMN namespace_key VARCHAR(255) NOT NULL DEFAULT 'default';
DROP INDEX IF EXISTS "variants_flag_key_key" CASCADE;
ALTER TABLE variants ADD CONSTRAINT "variants_namespace_flag_key" UNIQUE (namespace_key, flag_key, key);
ALTER TABLE variants ADD FOREIGN KEY (namespace_key) REFERENCES namespaces(key) ON DELETE CASCADE;
ALTER TABLE variants ADD FOREIGN KEY (namespace_key, flag_key) REFERENCES flags(namespace_key, key) ON DELETE CASCADE;
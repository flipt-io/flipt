-- Add column namespace_key with a default value
ALTER TABLE rules ADD COLUMN namespace_key VARCHAR(255) NOT NULL DEFAULT 'default';

-- Add foreign key constraint on namespace_key column referencing key column of namespaces table
ALTER TABLE rules ADD FOREIGN KEY (namespace_key) REFERENCES namespaces(key) ON DELETE CASCADE;
ALTER TABLE rules ADD FOREIGN KEY (namespace_key, flag_key) REFERENCES flags(namespace_key, key) ON DELETE CASCADE;
ALTER TABLE rules ADD FOREIGN KEY (namespace_key, segment_key) REFERENCES segments(namespace_key, key) ON DELETE CASCADE;
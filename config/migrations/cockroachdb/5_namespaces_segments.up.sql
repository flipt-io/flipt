-- Add column namespace_key with a default value
ALTER TABLE segments ADD COLUMN namespace_key VARCHAR(255) NOT NULL DEFAULT 'default';

-- Add foreign key constraint on namespace_key column referencing key column of namespaces table
ALTER TABLE segments ADD FOREIGN KEY (namespace_key) REFERENCES namespaces(key) ON DELETE CASCADE;

-- Drop primary key constraint and add a new composite primary key on namespace_key and key columns
ALTER TABLE segments DROP CONSTRAINT IF EXISTS segments_pkey CASCADE;
ALTER TABLE segments ADD CONSTRAINT segments_pkey PRIMARY KEY (namespace_key, key);

-- Add column namespace_key with a default value
ALTER TABLE constraints ADD COLUMN namespace_key VARCHAR(255) NOT NULL DEFAULT 'default';
-- Add foreign key constraint on namespace_key column referencing key column of namespaces table
ALTER TABLE constraints ADD FOREIGN KEY (namespace_key) REFERENCES namespaces(key) ON DELETE CASCADE;

-- Add foreign key constraint on namespace_key and segment_key columns referencing namespace_key and key columns of segments table
ALTER TABLE constraints ADD FOREIGN KEY (namespace_key) REFERENCES namespaces(key) ON DELETE CASCADE;
ALTER TABLE constraints ADD FOREIGN KEY (namespace_key, segment_key) REFERENCES segments(namespace_key, key) ON DELETE CASCADE;
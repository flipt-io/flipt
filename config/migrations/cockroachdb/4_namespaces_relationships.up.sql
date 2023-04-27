-- Flags
------------------

-- Add column namespace_key with a default value
ALTER TABLE flags ADD COLUMN namespace_key VARCHAR(255) NOT NULL DEFAULT 'default';

-- Add foreign key constraint on namespace_key column referencing key column of namespaces table
ALTER TABLE flags ADD FOREIGN KEY (namespace_key) REFERENCES namespaces(key) ON DELETE CASCADE;

-- Drop primary key constraint and add a new composite primary key on namespace_key and key columns
ALTER TABLE flags ALTER PRIMARY KEY USING COLUMNS (namespace_key, key);
DROP INDEX IF EXISTS flags_key_key CASCADE;

-- Variants
------------------

-- Add column namespace_key with a default value
ALTER TABLE variants ADD COLUMN namespace_key VARCHAR(255) NOT NULL DEFAULT 'default';

-- Drop previously created unique index
DROP INDEX IF EXISTS "variants_flag_key_key" CASCADE;

-- Add unique index on namespace_key, flag_key and key columns
ALTER TABLE variants ADD CONSTRAINT "variants_namespace_flag_key" UNIQUE (namespace_key, flag_key, key);

-- Add foreign key constraint on namespace_key column referencing key column of namespaces table
ALTER TABLE variants ADD FOREIGN KEY (namespace_key) REFERENCES namespaces(key) ON DELETE CASCADE;

-- Add foreign key constraint on namespace_key and flag_key columns referencing namespace_key and key columns of flags table
ALTER TABLE variants ADD FOREIGN KEY (namespace_key, flag_key) REFERENCES flags(namespace_key, key) ON DELETE CASCADE;

-- Segments
------------------

-- Add column namespace_key with a default value
ALTER TABLE segments ADD COLUMN namespace_key VARCHAR(255) NOT NULL DEFAULT 'default';

-- Add foreign key constraint on namespace_key column referencing key column of namespaces table
ALTER TABLE segments ADD FOREIGN KEY (namespace_key) REFERENCES namespaces(key) ON DELETE CASCADE;

-- Drop primary key constraint and add a new composite primary key on namespace_key and key columns
ALTER TABLE segments ALTER PRIMARY KEY USING COLUMNS (namespace_key, key);
DROP INDEX IF EXISTS segments_key_key CASCADE;

-- Constraints
------------------

-- Add column namespace_key with a default value
ALTER TABLE constraints ADD COLUMN namespace_key VARCHAR(255) NOT NULL DEFAULT 'default';

-- Add foreign key constraint on namespace_key column referencing key column of namespaces table
ALTER TABLE constraints ADD FOREIGN KEY (namespace_key) REFERENCES namespaces(key) ON DELETE CASCADE;

-- Add foreign key constraint on namespace_key and segment_key columns referencing namespace_key and key columns of segments table
ALTER TABLE constraints ADD FOREIGN KEY (namespace_key, segment_key) REFERENCES segments(namespace_key, key) ON DELETE CASCADE;

-- Rules
------------------

-- Add column namespace_key with a default value
ALTER TABLE rules ADD COLUMN namespace_key VARCHAR(255) NOT NULL DEFAULT 'default';

-- Add foreign key constraint on namespace_key column referencing key column of namespaces table
ALTER TABLE rules ADD FOREIGN KEY (namespace_key) REFERENCES namespaces(key) ON DELETE CASCADE;
ALTER TABLE rules ADD FOREIGN KEY (namespace_key, flag_key) REFERENCES flags(namespace_key, key) ON DELETE CASCADE;
ALTER TABLE rules ADD FOREIGN KEY (namespace_key, segment_key) REFERENCES segments(namespace_key, key) ON DELETE CASCADE;
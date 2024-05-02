-- Drop previously created foreign key
ALTER TABLE constraints DROP FOREIGN KEY `constraints_ibfk_1`;
ALTER TABLE rules DROP FOREIGN KEY `rules_ibfk_1`;
ALTER TABLE rules DROP FOREIGN KEY `rules_ibfk_2`;
ALTER TABLE variants DROP FOREIGN KEY `variants_ibfk_1`;

-- Flags

-- Add column namespace_key with a default value
ALTER TABLE flags ADD COLUMN namespace_key VARCHAR(255) NOT NULL DEFAULT 'default';

-- Drop previously created unique index
ALTER TABLE flags DROP INDEX `key`, ADD INDEX `key` (`key`) USING BTREE;

-- Drop primary key constraint and add a new composite primary key on namespace_key and key columns
ALTER TABLE flags DROP PRIMARY KEY, ADD PRIMARY KEY (`namespace_key`, `key`);

-- Add foreign key constraint on namespace_key column referencing key column of namespaces table
ALTER TABLE flags ADD FOREIGN KEY (namespace_key) REFERENCES namespaces(`key`) ON DELETE CASCADE;

-- Variants

-- Add column namespace_key with a default value
ALTER TABLE variants ADD COLUMN namespace_key VARCHAR(255) NOT NULL DEFAULT 'default';

-- Drop previously created foreign key

-- Drop previously created unique index and add a new unique index on namespace_key, flag_key and key columns
ALTER TABLE variants DROP INDEX `variants_flag_key_key`, ADD UNIQUE INDEX `variants_namespace_flag_key` (`namespace_key`, `flag_key`, `key`) USING BTREE;

-- Add foreign key constraint on namespace_key column referencing key column of namespaces table
ALTER TABLE variants ADD FOREIGN KEY (namespace_key) REFERENCES namespaces(`key`) ON DELETE CASCADE;

-- Add foreign key constraint on namespace_key and flag_key columns referencing namespace_key and key columns of flags table
ALTER TABLE variants ADD FOREIGN KEY (namespace_key, flag_key) REFERENCES flags(`namespace_key`, `key`) ON DELETE CASCADE;

-- Segments

-- Drop previously created unique index and add a new unique index on namespace_key and key columns
ALTER TABLE segments DROP INDEX `key`, ADD INDEX `key` (`key`) USING BTREE;

-- Add column namespace_key with a default value
ALTER TABLE segments ADD COLUMN namespace_key VARCHAR(255) NOT NULL DEFAULT 'default';

-- Drop primary key constraint and add a new composite primary key on namespace_key and key columns
ALTER TABLE segments DROP PRIMARY KEY, ADD PRIMARY KEY (`namespace_key`, `key`);

-- Add foreign key constraint on namespace_key column referencing key column of namespaces table
ALTER TABLE segments ADD FOREIGN KEY (namespace_key) REFERENCES namespaces(`key`) ON DELETE CASCADE;

-- Constraints

-- Add column namespace_key with a default value
ALTER TABLE constraints ADD COLUMN namespace_key VARCHAR(255) NOT NULL DEFAULT 'default';

-- Drop previously created index and add a new index on namespace_key and segment_key columns
ALTER TABLE constraints DROP INDEX `segment_key`, ADD INDEX `constraints_namespace_segment_key` (`namespace_key`, `segment_key`) USING BTREE;

-- Add foreign key constraint on namespace_key column referencing key column of namespaces table
ALTER TABLE constraints ADD FOREIGN KEY (namespace_key) REFERENCES namespaces(`key`) ON DELETE CASCADE;

-- Add foreign key constraint on namespace_key and segment_key columns referencing namespace_key and key columns of segments table
ALTER TABLE constraints ADD FOREIGN KEY (namespace_key, segment_key) REFERENCES segments(`namespace_key`, `key`) ON DELETE CASCADE;

-- Rules

-- Add column namespace_key with a default value
ALTER TABLE rules ADD COLUMN namespace_key VARCHAR(255) NOT NULL DEFAULT 'default';

-- Drop previously created index and add a new index on namespace_key, flag_key and segment_key columns
ALTER TABLE rules DROP INDEX `flag_key`, ADD INDEX `rules_namespace_flag_key` (`namespace_key`, `flag_key`) USING BTREE;
ALTER TABLE rules DROP INDEX `segment_key`, ADD INDEX `rules_namespace_segment_key` (`namespace_key`, `segment_key`) USING BTREE;

-- Add foreign key constraint on namespace_key column referencing key column of namespaces table
ALTER TABLE rules ADD FOREIGN KEY (namespace_key) REFERENCES namespaces(`key`) ON DELETE CASCADE;

-- Add foreign key constraint on namespace_key and flag_key columns referencing namespace_key and key columns of flags table
ALTER TABLE rules ADD FOREIGN KEY (namespace_key, flag_key) REFERENCES flags(`namespace_key`, `key`) ON DELETE CASCADE;

-- Add foreign key constraint on namespace_key and segment_key columns referencing namespace_key and key columns of segments table
ALTER TABLE rules ADD FOREIGN KEY (namespace_key, segment_key) REFERENCES segments(`namespace_key`, `key`) ON DELETE CASCADE;

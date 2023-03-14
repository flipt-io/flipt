-- Rules

-- Drop foreign key constraint on namespace_key column
ALTER TABLE rules DROP FOREIGN KEY `rules_ibfk_1`;

-- Drop foreign key constraint on namespace_key and flag_key columns
ALTER TABLE rules DROP FOREIGN KEY `rules_ibfk_2`;

-- Drop foreign key constraint on namespace_key and segment_key columns
ALTER TABLE rules DROP FOREIGN KEY `rules_ibfk_3`;

-- Drop previously created index and add a new index on flag_key and segment_key columns
ALTER TABLE rules DROP INDEX `rules_namespace_flag_key`, ADD INDEX `flag_key` (`flag_key`) USING BTREE;
ALTER TABLE rules DROP INDEX `rules_namespace_segment_key`, ADD INDEX `segment_key` (`segment_key`) USING BTREE;

-- Add foreign key constraint on flag_key column referencing key column of flags table
ALTER TABLE rules ADD FOREIGN KEY (flag_key) REFERENCES flags(`key`) ON DELETE CASCADE;

-- Add foreign key constraint on segment_key column referencing key column of segments table
ALTER TABLE rules ADD FOREIGN KEY (segment_key) REFERENCES segments(`key`) ON DELETE CASCADE;

-- Drop column namespace_key
ALTER TABLE rules DROP COLUMN namespace_key;

-- Variants

-- Drop foreign key constraint on namespace_key column
ALTER TABLE variants DROP FOREIGN KEY `variants_ibfk_1`;

-- Drop foreign key constraint on namespace_key and flag_key columns
ALTER TABLE variants DROP FOREIGN KEY `variants_ibfk_2`;

-- Add foreign key constraint on flag_key column referencing key column of flags table
ALTER TABLE variants ADD FOREIGN KEY (flag_key) REFERENCES flags(`key`) ON DELETE CASCADE;

-- Drop previously created unique index and add a new unique index on flag_key and key columns
ALTER TABLE variants DROP INDEX `variants_namespace_flag_key`, ADD UNIQUE INDEX `variants_flag_key_key` (`flag_key`, `key`) USING BTREE;

-- Drop column namespace_key
ALTER TABLE variants DROP COLUMN namespace_key;

-- Flags 

-- Drop foreign key constraint on namespace_key column
ALTER TABLE flags DROP FOREIGN KEY `flags_ibfk_1`;

-- Drop primary key constraint and add a new primary key on key column
ALTER TABLE flags DROP PRIMARY KEY, ADD PRIMARY KEY (`key`);

-- Drop column namespace_key
ALTER TABLE flags DROP COLUMN namespace_key;

-- Constraints

-- Drop foreign key constraint on namespace_key column
ALTER TABLE constraints DROP FOREIGN KEY `constraints_ibfk_1`;

-- Drop foreign key constraint on namespace_key and segment_key columns
ALTER TABLE constraints DROP FOREIGN KEY `constraints_ibfk_2`;

-- Add foreign key constraint on segment_key column referencing key column of segments table
ALTER TABLE constraints ADD FOREIGN KEY (segment_key) REFERENCES segments(`key`) ON DELETE CASCADE;

-- Drop index on namespace_key and segment_key columns and add a new index on segment_key column
ALTER TABLE constraints DROP INDEX `constraints_namespace_segment_key`, ADD INDEX `segment_key` (`segment_key`) USING BTREE;

-- Drop column namespace_key
ALTER TABLE constraints DROP COLUMN namespace_key;

-- Segments 

-- Drop foreign key constraint on namespace_key column
ALTER TABLE segments DROP FOREIGN KEY `segments_ibfk_1`;

-- Drop primary key constraint and add a new primary key on key column
ALTER TABLE segments DROP PRIMARY KEY, ADD PRIMARY KEY (`key`);

-- Drop column namespace_key
ALTER TABLE segments DROP COLUMN namespace_key;
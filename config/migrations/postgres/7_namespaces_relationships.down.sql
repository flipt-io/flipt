-- Flags 
----------------

-- Drop foreign key constraint on namespace_key column
ALTER TABLE flags DROP CONSTRAINT IF EXISTS flags_namespace_key_fkey CASCADE;

-- Drop primary key constraint and add a new primary key on key column
ALTER TABLE flags DROP CONSTRAINT IF EXISTS flags_pkey CASCADE;
ALTER TABLE flags ADD CONSTRAINT flags_pkey PRIMARY KEY (key);

-- Drop column namespace_key
ALTER TABLE flags DROP COLUMN IF EXISTS namespace_key;

-- Variants
----------------

-- Drop foreign key constraints on namespace_key and flag_key columns
ALTER TABLE variants DROP CONSTRAINT IF EXISTS variants_namespace_key_flag_key_fkey CASCADE;
ALTER TABLE variants DROP CONSTRAINT IF EXISTS variants_namespace_key_fkey CASCADE;

-- Add foreign key constraint on flag_key column referencing key column of flags table
ALTER TABLE variants ADD FOREIGN KEY (flag_key) REFERENCES flags(key) ON DELETE CASCADE;

-- Drop unique constraint created by previous migration and add a new unique constraint on flag_key and key columns
ALTER TABLE variants ADD CONSTRAINT variants_flag_key UNIQUE (flag_key, key);

-- Drop column namespace_key
ALTER TABLE variants DROP COLUMN IF EXISTS namespace_key;

-- Segments
----------------

-- Drop foreign key constraint on namespace_key column
ALTER TABLE segments DROP CONSTRAINT IF EXISTS segments_namespace_key_fkey CASCADE;

-- Drop primary key constraint and add a new primary key constraint on key column
ALTER TABLE segments DROP CONSTRAINT IF EXISTS segments_pkey CASCADE;
ALTER TABLE segments ADD PRIMARY KEY (key);

-- Drop column namespace_key
ALTER TABLE segments DROP COLUMN IF EXISTS namespace_key;

-- Constraints
----------------

-- Drop foreign key constraints on namespace_key and segment_key columns
ALTER TABLE constraints DROP CONSTRAINT IF EXISTS constraints_namespace_key_segment_key_fkey CASCADE;
ALTER TABLE constraints DROP CONSTRAINT IF EXISTS constraints_namespace_key_fkey CASCADE;

-- Add foreign key constraint on segment_key column referencing key column of segments table
ALTER TABLE constraints ADD FOREIGN KEY (segment_key) REFERENCES segments(key) ON DELETE CASCADE;

-- Drop column namespace_key
ALTER TABLE constraints DROP COLUMN IF EXISTS namespace_key;

-- Rules
----------------

-- Drop foreign key constraints on namespace_key, flag_key and segment_key columns
ALTER TABLE rules DROP CONSTRAINT IF EXISTS rules_namespace_key_fkey CASCADE;
ALTER TABLE rules DROP CONSTRAINT IF EXISTS rules_namespace_key_flag_key_fkey CASCADE;
ALTER TABLE rules DROP CONSTRAINT IF EXISTS rules_namespace_key_segment_key_fkey CASCADE;

-- Add foreign key constraint on flag_key column referencing key column of flags table
ALTER TABLE rules ADD FOREIGN KEY (flag_key) REFERENCES flags(key) ON DELETE CASCADE;

-- Add foreign key constraint on segment_key column referencing key column of segments table
ALTER TABLE rules ADD FOREIGN KEY (segment_key) REFERENCES segments(key) ON DELETE CASCADE;

-- Drop column namespace_key
ALTER TABLE rules DROP COLUMN IF EXISTS namespace_key;

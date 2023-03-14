-- Add column namespace_key with a default value
ALTER TABLE flags ADD COLUMN namespace_key VARCHAR(255) NOT NULL DEFAULT 'default';

-- Add foreign key constraint on namespace_key column referencing key column of namespaces table
ALTER TABLE flags ADD FOREIGN KEY (namespace_key) REFERENCES namespaces(key) ON DELETE CASCADE;

-- Drop primary key constraint and add a new composite primary key on namespace_key and key columns
ALTER TABLE flags DROP CONSTRAINT IF EXISTS flags_pkey CASCADE;
ALTER TABLE flags ADD CONSTRAINT flags_pkey PRIMARY KEY (namespace_key, key);

ALTER TABLE variants ADD COLUMN namespace_key VARCHAR(255) NOT NULL DEFAULT 'default';
ALTER TABLE variants DROP CONSTRAINT IF EXISTS variants_flag_key_key_key CASCADE;
ALTER TABLE variants ADD CONSTRAINT variants_namespace_flag_key UNIQUE (namespace_key, flag_key, key);
ALTER TABLE variants ADD FOREIGN KEY (namespace_key) REFERENCES namespaces(key) ON DELETE CASCADE;
ALTER TABLE variants ADD FOREIGN KEY (namespace_key, flag_key) REFERENCES flags(namespace_key, key) ON DELETE CASCADE;

-- Add column namespace_key with a default value
ALTER TABLE segments ADD COLUMN namespace_key VARCHAR(255) NOT NULL DEFAULT 'default';

-- Add foreign key constraint on namespace_key column referencing key column of namespaces table
ALTER TABLE segments ADD FOREIGN KEY (namespace_key) REFERENCES namespaces(key) ON DELETE CASCADE;

-- Drop primary key constraint and add a new composite primary key on namespace_key and key columns
ALTER TABLE segments DROP CONSTRAINT IF EXISTS segments_pkey CASCADE;
ALTER TABLE segments ADD CONSTRAINT segments_pkey PRIMARY KEY (namespace_key, key);

-- Add column namespace_key with a default value
ALTER TABLE constraints ADD COLUMN namespace_key VARCHAR(255) NOT NULL DEFAULT 'default';

-- Add foreign key constraint on namespace_key and segment_key columns referencing namespace_key and key columns of segments table
ALTER TABLE constraints ADD FOREIGN KEY (namespace_key) REFERENCES namespaces(key) ON DELETE CASCADE;
ALTER TABLE constraints ADD FOREIGN KEY (namespace_key, segment_key) REFERENCES segments(namespace_key, key) ON DELETE CASCADE;

-- Add column namespace_key with a default value
ALTER TABLE rules ADD COLUMN namespace_key VARCHAR(255) NOT NULL DEFAULT 'default';

-- Add foreign key constraint on namespace_key column referencing key column of namespaces table
ALTER TABLE rules ADD FOREIGN KEY (namespace_key) REFERENCES namespaces(key) ON DELETE CASCADE;
ALTER TABLE rules ADD FOREIGN KEY (namespace_key, flag_key) REFERENCES flags(namespace_key, key) ON DELETE CASCADE;
ALTER TABLE rules ADD FOREIGN KEY (namespace_key, segment_key) REFERENCES segments(namespace_key, key) ON DELETE CASCADE;
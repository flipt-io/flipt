 -- drop previously created foreign keys
ALTER TABLE variants DROP FOREIGN KEY `variants_ibfk_1`;
ALTER TABLE variants DROP FOREIGN KEY `variants_ibfk_2`;

-- drop previously created unique index
ALTER TABLE variants DROP INDEX `variants_namespace_flag_key`, ADD UNIQUE INDEX `variants_flag_key_key` (`flag_key`, `key`) USING BTREE;
ALTER TABLE variants DROP COLUMN namespace_key;

-- drop foreign key to namespaces
ALTER TABLE flags DROP FOREIGN KEY `flags_ibfk_1`;
-- drop previously created index
ALTER TABLE flags DROP INDEX `key`;
-- drop namespaces, key primary key
ALTER TABLE flags DROP PRIMARY KEY, ADD PRIMARY KEY (`key`);
ALTER TABLE flags DROP COLUMN namespace_key;


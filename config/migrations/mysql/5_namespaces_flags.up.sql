ALTER TABLE flags ADD COLUMN namespace_key VARCHAR(255) NOT NULL DEFAULT 'default';
ALTER TABLE flags DROP INDEX `key`, ADD INDEX `key` (`key`) USING BTREE; /* drop previously created unique index */
ALTER TABLE flags ADD FOREIGN KEY (namespace_key) REFERENCES namespaces(`key`) ON DELETE CASCADE;
ALTER TABLE flags DROP PRIMARY KEY, ADD PRIMARY KEY (`namespace_key`, `key`);

ALTER TABLE variants ADD COLUMN namespace_key VARCHAR(255) NOT NULL DEFAULT 'default';
ALTER TABLE variants DROP FOREIGN KEY `variants_ibfk_1`; /* drop previously created foreign key */
ALTER TABLE variants DROP INDEX `variants_flag_key_key`, ADD UNIQUE INDEX `variants_namespace_flag_key` (`namespace_key`, `flag_key`, `key`) USING BTREE; /* drop previously created unique index */
ALTER TABLE variants ADD FOREIGN KEY (namespace_key) REFERENCES namespaces(`key`) ON DELETE CASCADE;
ALTER TABLE variants ADD FOREIGN KEY (namespace_key, flag_key) REFERENCES flags(`namespace_key`, `key`) ON DELETE CASCADE;

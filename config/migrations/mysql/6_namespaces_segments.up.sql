ALTER TABLE segments ADD COLUMN namespace_key VARCHAR(255) NOT NULL DEFAULT 'default';
-- drop previously created unique index
ALTER TABLE segments DROP INDEX `key`, ADD INDEX `key` (`key`) USING BTREE;
ALTER TABLE segments ADD FOREIGN KEY (namespace_key) REFERENCES namespaces(`key`) ON DELETE CASCADE;
ALTER TABLE segments DROP PRIMARY KEY, ADD PRIMARY KEY (`namespace_key`, `key`);

ALTER TABLE constraints ADD COLUMN namespace_key VARCHAR(255) NOT NULL DEFAULT 'default';
-- drop previously created foreign key
ALTER TABLE constraints DROP FOREIGN KEY `constraints_ibfk_1`;
-- drop previously created index
ALTER TABLE constraints DROP INDEX `segment_key`, ADD INDEX `constraints_namespace_segment_key` (`namespace_key`, `segment_key`) USING BTREE;
ALTER TABLE constraints ADD FOREIGN KEY (namespace_key) REFERENCES namespaces(`key`) ON DELETE CASCADE;
ALTER TABLE constraints ADD FOREIGN KEY (namespace_key, segment_key) REFERENCES segments(`namespace_key`, `key`) ON DELETE CASCADE;

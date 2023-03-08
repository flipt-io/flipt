ALTER TABLE rules ADD COLUMN namespace_key VARCHAR(255) NOT NULL DEFAULT 'default';
ALTER TABLE rules DROP FOREIGN KEY `rules_ibfk_1`; /* drop previously created foreign key */
ALTER TABLE rules DROP FOREIGN KEY `rules_ibfk_2`; /* drop previously created foreign key */
ALTER TABLE rules DROP INDEX `flag_key`, ADD INDEX `rules_namespace_flag_key` (`namespace_key`, `flag_key`) USING BTREE; /* drop previously created index */
ALTER TABLE rules DROP INDEX `segment_key`, ADD INDEX `rules_namespace_segment_key` (`namespace_key`, `segment_key`) USING BTREE; /* drop previously created index */
ALTER TABLE rules ADD FOREIGN KEY (namespace_key) REFERENCES namespaces(`key`) ON DELETE CASCADE;
ALTER TABLE rules ADD FOREIGN KEY (namespace_key, flag_key) REFERENCES flags(`namespace_key`, `key`) ON DELETE CASCADE;
ALTER TABLE rules ADD FOREIGN KEY (namespace_key, segment_key) REFERENCES segments(`namespace_key`, `key`) ON DELETE CASCADE;

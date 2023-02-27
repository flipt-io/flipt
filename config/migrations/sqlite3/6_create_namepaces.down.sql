DROP INDEX flags_namespace_key_idx;
DROP INDEX segments_namespace_key_idx;

ALTER TABLE flags DROP COLUMN namespace_key;
ALTER TABLE segments DROP COLUMN namespace_key;

DROP TABLE IF EXISTS namespaces;
-- rule_segments
CREATE TABLE rule_segments_temp (
  rule_id VARCHAR(255) NOT NULL REFERENCES rules ON DELETE CASCADE,
  namespace_key VARCHAR(255) NOT NULL,
  segment_key VARCHAR(255) NOT NULL,
  UNIQUE (rule_id, namespace_key, segment_key),
  FOREIGN KEY (namespace_key, segment_key) REFERENCES segments (namespace_key, key) ON DELETE RESTRICT 
);

INSERT INTO rule_segments_temp (rule_id, namespace_key, segment_key) SELECT rule_id, namespace_key, segment_key FROM rule_segments;
DROP TABLE rule_segments;
ALTER TABLE rule_segments_temp RENAME TO rule_segments;

-- rollout_segment_references
CREATE TABLE rollout_segment_references_temp (
  rollout_segment_id VARCHAR(255) NOT NULL REFERENCES rollout_segments ON DELETE CASCADE,
  namespace_key VARCHAR(255) NOT NULL,
  segment_key VARCHAR(255) NOT NULL,
  UNIQUE (rollout_segment_id, namespace_key, segment_key),
  FOREIGN KEY (namespace_key, segment_key) REFERENCES segments (namespace_key, key) ON DELETE RESTRICT 
);

INSERT INTO rollout_segment_references_temp (rollout_segment_id, namespace_key, segment_key) SELECT rollout_segment_id, namespace_key, segment_key FROM rollout_segment_references;
DROP TABLE rollout_segment_references;
ALTER TABLE rollout_segment_references_temp RENAME TO rollout_segment_references;

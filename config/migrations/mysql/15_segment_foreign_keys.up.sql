ALTER TABLE rule_segments DROP FOREIGN KEY `rule_segments_ibfk_2`;
ALTER TABLE rule_segments ADD CONSTRAINT `rule_segments_ibfk_2` FOREIGN KEY (namespace_key, segment_key) REFERENCES segments (namespace_key, `key`) ON DELETE RESTRICT;

ALTER TABLE rollout_segment_references DROP FOREIGN KEY `rollout_segment_references_ibfk_2`;  
ALTER TABLE rollout_segment_references
 ADD CONSTRAINT `rollout_segment_references_ibfk_2` FOREIGN KEY (namespace_key, segment_key) REFERENCES segments (namespace_key, `key`) ON DELETE RESTRICT; 

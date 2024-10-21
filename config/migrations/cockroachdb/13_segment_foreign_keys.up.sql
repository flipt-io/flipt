ALTER TABLE rule_segments DROP CONSTRAINT fk_namespace_key_ref_segments;
ALTER TABLE rule_segments ADD CONSTRAINT fk_namespace_key_ref_segments FOREIGN KEY (namespace_key, segment_key) REFERENCES segments (namespace_key, key) ON DELETE RESTRICT;

ALTER TABLE rollout_segment_references DROP CONSTRAINT fk_namespace_key_ref_segments;  
ALTER TABLE rollout_segment_references
 ADD CONSTRAINT fk_namespace_key_ref_segments FOREIGN KEY (namespace_key, segment_key) REFERENCES segments (namespace_key, key) ON DELETE RESTRICT; 

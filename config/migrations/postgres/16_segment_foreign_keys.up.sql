ALTER TABLE rule_segments DROP CONSTRAINT rule_segments_namespace_key_segment_key_fkey;
ALTER TABLE rule_segments ADD CONSTRAINT rule_segments_namespace_key_segment_key_fkey FOREIGN KEY (namespace_key, segment_key) REFERENCES segments (namespace_key, key) ON DELETE RESTRICT;

ALTER TABLE rollout_segment_references DROP CONSTRAINT rollout_segment_references_namespace_key_segment_key_fkey;  
ALTER TABLE rollout_segment_references
 ADD CONSTRAINT rollout_segment_references_namespace_key_segment_key_fkey FOREIGN KEY (namespace_key, segment_key) REFERENCES segments (namespace_key, key) ON DELETE RESTRICT; 

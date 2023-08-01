-- Rules
ALTER TABLE rules DROP FOREIGN KEY `rules_ibfk_3`;

ALTER TABLE rules DROP COLUMN segment_key;

ALTER TABLE rules ADD COLUMN segment_operator INTEGER NOT NULL DEFAULT 0;

-- Rollouts
ALTER TABLE rollout_segments DROP FOREIGN KEY `rollout_segments_ibfk_1`;
ALTER TABLE rollout_segments DROP FOREIGN KEY `rollout_segments_ibfk_3`;

ALTER TABLE rollout_segments DROP COLUMN segment_key;
ALTER TABLE rollout_segments DROP COLUMN namespace_key;

ALTER TABLE rollout_segments ADD COLUMN segment_operator INTEGER NOT NULL DEFAULT 0;
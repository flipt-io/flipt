CREATE TABLE IF NOT EXISTS flipt_counter_analytics (
    `time` DateTime, `name` String, `flag_key` String, `value` UInt32 
) Engine = MergeTree ORDER BY tuple();

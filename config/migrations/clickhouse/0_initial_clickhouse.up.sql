CREATE TABLE IF NOT EXISTS flipt_counter_analytics (
    `timestamp` DateTime('UTC'), `name` String, `namespace_key` String, `flag_key` String, `reason` LowCardinality(String), `match` Bool, `value` UInt32 
) Engine = MergeTree
ORDER BY timestamp
TTL timestamp + INTERVAL 1 WEEK;

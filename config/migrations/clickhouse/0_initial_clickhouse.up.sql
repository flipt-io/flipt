CREATE TABLE IF NOT EXISTS flipt_counter_analytics (
    `timestamp` DateTime('UTC'), `analytic_name` String, `namespace_key` String, `flag_key` String, `flag_type` Enum('VARIANT_FLAG_TYPE' = 1, 'BOOLEAN_FLAG_TYPE' = 2), `reason` Enum('UNKNOWN_EVALUATION_REASON' = 1, 'FLAG_DISABLED_EVALUATION_REASON' = 2, 'MATCH_EVALUATION_REASON' = 3, 'DEFAULT_EVALUATION_REASON' = 4), `match` Nullable(Bool), `evaluation_value` Nullable(String), `value` UInt32
) Engine = MergeTree
ORDER BY timestamp
TTL timestamp + INTERVAL 1 WEEK;

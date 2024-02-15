CREATE TABLE flipt_counter_aggregated_analytics
(
    `timestamp` DateTime('UTC'),
    `analytic_name` LowCardinality(String),
    `namespace_key` LowCardinality(String),
    `flag_key` LowCardinality(String),
    `reason` LowCardinality(String),
    `evaluation_value` LowCardinality(String),
    `value` UInt32
)
ENGINE = SummingMergeTree
ORDER BY (timestamp, analytic_name, namespace_key, flag_key, reason, evaluation_value)
TTL timestamp + INTERVAL 1 WEEK;
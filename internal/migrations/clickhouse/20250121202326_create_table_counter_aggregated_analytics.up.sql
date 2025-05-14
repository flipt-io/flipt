CREATE TABLE flipt_counter_aggregated_analytics_v2
(
    `timestamp` DateTime('UTC'),
    `analytic_name` LowCardinality(String),
    `environment_key` LowCardinality(String),
    `namespace_key` LowCardinality(String),
    `flag_key` LowCardinality(String),
    `reason` LowCardinality(String),
    `evaluation_value` LowCardinality(Nullable(String)),
    `value` UInt32
)
ENGINE = SummingMergeTree
ORDER BY (
    environment_key,
    namespace_key,
    flag_key,
    timestamp,
    analytic_name,
    reason,
    evaluation_value
)
TTL timestamp + INTERVAL 1 WEEK
SETTINGS allow_nullable_key = 1;
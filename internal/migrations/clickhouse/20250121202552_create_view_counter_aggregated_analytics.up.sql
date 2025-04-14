CREATE MATERIALIZED VIEW flipt_counter_aggregated_analytics_mv_v2
TO flipt_counter_aggregated_analytics_v2
AS
SELECT
    `timestamp`,
    analytic_name,
    environment_key,
    namespace_key,
    flag_key,
    reason,
    evaluation_value,
    sum(`value`) AS `value`
FROM flipt_counter_analytics_v2
GROUP BY
    timestamp,
    analytic_name,
    environment_key,
    namespace_key,
    flag_key,
    reason,
    evaluation_value;
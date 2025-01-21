CREATE MATERIALIZED VIEW flipt_counter_aggregated_analytics_mv
TO flipt_counter_aggregated_analytics
AS
SELECT
    `timestamp`,
    analytic_name,
    namespace_key,
    flag_key,
    reason,
    evaluation_value,
    sum(`value`) AS `value`
FROM flipt_counter_analytics
GROUP BY
    timestamp,
    analytic_name,
    namespace_key,
    flag_key,
    reason,
    evaluation_value;
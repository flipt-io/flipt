# OTLP Example

This example shows how you can run Flipt with an [OpenTelemetry Protocol](https://opentelemetry.io/docs/reference/specification/protocol/) exporter which recieves, aggregates, and in-turn exports traces to both Jaeger and Zipken backends.

## Requirements

To run this example application you'll need:

* [Docker](https://docs.docker.com/install/)
* [docker-compose](https://docs.docker.com/compose/install/)

## Running the Example

1. Run `docker-compose up` from this directory
1. Open the Flipt UI (default: [http://localhost:8080](http://localhost:8080))
1. Create some sample data: Flags/Segments/etc. Perform a few evaluations in the Console.

### Jaeger UI

!['Jaeger Example'](../../images/jaeger.jpg)

1. Open the Jaeger UI (default: [http://localhost:16686](http://localhost:16686))
1. Select 'flipt' from the Service dropdown
1. Click 'Find Traces'
1. You should see a list of traces to explore

### Zipkin UI

!['Zipkin Example'](../../images/zipkin.png)

1. Open the Zipkin UI (default: [http://localhost:9411](http://localhost:9411))
1. Select `serviceName=flipt` from the search box
1. Click 'Run Query'
1. You should see a list of traces to explore

### Datadog UI

!['Datadog Example'](../../images/datadog.png)

For exporting traces from [OpenTelemetry to Datadog](https://docs.datadoghq.com/opentelemetry/otel_collector_datadog_exporter) you have to configure the exporter in the `otel-collector-config.yaml`:

```yaml
exporters:
  datadog:
    api:
      site: datadoghq.com
      key: ${DD_API_KEY}
```

**Note:** The `DD_API_KEY` should be replaced with your actual api key from Datadog.

Furthermore, you also have to add `datadog` as an entry in `exporters` under `service.pipelines.traces.exporters`.

For example:

```yaml
service:
  extensions: [pprof, zpages, health_check]
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [logging, zipkin, jaeger, datadog]
```

1. Open the Datadog traces UI under the menu item on the left `APM` then `Traces`
1. You should see a list of traces to explore

### Clickhouse

#### Requirements

1. [Clickhouse CLI](https://clickhouse.com/docs/en/install)

#### Introduction

[Clickhouse](https://clickhouse.com/) has alpha support as a destination for traces collected by an open telemetry collector as stated in the [docs](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/exporter/clickhouseexporter/README.md). Because of this, the source code for the collector we have to use is [here](https://github.com/open-telemetry/opentelemetry-collector-contrib) as this collector can read `clickhouse` configuration details.

To accomodate a clean illustration for clickhouse due to the nuances, we separated out the example into the `clickhouse` directory.

The configuration for the OpenTelemetry collector should look familiar if you have followed the above examples, and this [yaml](./clickhouse/otel-collector-config.yaml) could serve as a base to fit your configuration needs.

#### Analytics

Once you run the `docker-compose`, you can start creating flags, making evaluations etc. like regular. As data gets collected on the clickhouse server you can connect to the server via the client you should have installed.

```bash
./clickhouse client clickhouse://localhost:9000 --user default
```

From this, you can glean very powerful analytical data from Flipt, such as average/quantile measurements from the flag evaluations.

A sample query could look something like:

```sql
SELECT SpanName, avg(Duration), quantile(0.9)(Duration) AS p90, quantile(0.95)(Duration) AS p95, quantile(0.99)(Duration) AS p99 FROM otel.otel_traces WHERE SpanName='flipt.evaluation.EvaluationService/Variant' OR SpanName='flipt.evaluation.EvaluationService/Boolean' GROUP BY SpanName;
```

!['Clickhouse Example'](../../images/clickhouse.png)

This query will return average, and quantile results (p90, p95, p99) for `Variant` and `Boolean` evaluations. To do more cool and insighful things via their SQL syntax, you can refer to the [Clickhouse SQL reference docs](https://clickhouse.com/docs/en/sql-reference).

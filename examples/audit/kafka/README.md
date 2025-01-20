# Audit Event Logging / Kafka Example

This example shows how you can run Flipt with audit event logging enabled using the `Kafka` audit sink.

This works by setting the environment variables:

```bash
 FLIPT_AUDIT_SINKS_KAFKA_ENABLED=true
 FLIPT_AUDIT_SINKS_KAFKA_TOPIC=flipt-audit-events
 FLIPT_AUDIT_SINKS_KAFKA_ENCODING=avro
 FLIPT_AUDIT_SINKS_KAFKA_INSECURE_SKIP_TLS=true
 FLIPT_AUDIT_SINKS_KAFKA_BOOTSTRAP_SERVERS=redpanda
 FLIPT_AUDIT_SINKS_KAFKA_SCHEMA_REGISTRY_URL=http://redpanda:8081
```

The auditable events currently are `created`, `updated`, and `deleted` operations on `flags`, `variants`, `segments`, `constraints`, `rules`, `distributions`, `namespaces`, and `tokens`. If you do any of these operations through the API, Flipt will emit an audit event log to the specified location.

## Requirements

To run this example application you'll need:

* [Docker](https://docs.docker.com/install/)
* [docker compose](https://docs.docker.com/compose/install/)

## Running the Example

1. Run `docker compose up` from this directory
1. Open the Flipt UI (default: [http://localhost:8080](http://localhost:8080))
1. Create some sample data: Flags/Segments/etc.
1. Open the Redpanda UI (default: [http://localhost:8888/topics/flipt-audit-events](http://localhost:8888/topics/flipt-audit-events))
1. You should see a topic of audit events. 

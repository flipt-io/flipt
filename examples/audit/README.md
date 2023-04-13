# Audit Event Logging Example

This example shows how you can run Flipt with Audit Event logging enabled to a file on disk.

This works by setting the two environment variables `FLIPT_AUDIT_SINKS_LOG_ENABLED` and `FLIPT_AUDIT_SINKS_LOG_FILE`:

```bash
FLIPT_AUDIT_SINKS_LOG_ENABLED=true
FLIPT_AUDIT_SINKS_LOG_FILE=/var/opt/flipt/audits.txt
```

## Requirements

To run this example application you'll need:

* [Docker](https://docs.docker.com/install/)
* [docker-compose](https://docs.docker.com/compose/install/)

## Running the Example

1. Run `docker-compose up` from this directory
1. Open the Flipt UI (default: [http://localhost:8080](http://localhost:8080))

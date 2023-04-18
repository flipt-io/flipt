# Audit Event Logging Example

**This feature is still under development, and this example is subject to change.**

This example shows how you can run Flipt with Audit Event logging enabled to a file on disk.

This works by setting the two environment variables `FLIPT_AUDIT_SINKS_LOG_ENABLED` and `FLIPT_AUDIT_SINKS_LOG_FILE`:

```bash
FLIPT_AUDIT_SINKS_LOG_ENABLED=true
FLIPT_AUDIT_SINKS_LOG_FILE=/var/log/audit.log
```

The auditable events currently are CRUD (except for read) operations on `flags`, `variants`, `segments`, `constraints`, `rules`, `distributions`, and `namespaces`. If you do any of these operations through the API, it should emit an audit event log to the specified location.

Since docker containers are ephemeral and data within the container is lost when the container exits, we mount a local file on the host to the audit event log location in the container as a volume. You would have to create the file [first](https://github.com/moby/moby/issues/21612#issuecomment-202984678) before starting the container:

```bash
mkdir -p /tmp/flipt && touch /tmp/flipt/audit.log
```

and `tail` the logs as you are making API requests to the Flipt server when the container is running.

```bash
tail -f /tmp/flipt/audit.log
```

## Requirements

To run this example application you'll need:

* [Docker](https://docs.docker.com/install/)
* [docker-compose](https://docs.docker.com/compose/install/)

## Running the Example

1. Run `docker-compose up` from this directory
1. Open the Flipt UI (default: [http://localhost:8080](http://localhost:8080))

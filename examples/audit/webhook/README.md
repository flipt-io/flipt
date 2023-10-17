# Audit Events to Webhook Examples

In these examples we spin up a simple HTTP server in Python to receive and log the received webhook events. You can see the code in [server.py](server.py).

These examples include:

- [Basics](#basics): Shows how to configure Flipt to send audit events to a webhook.
- [Templates](#templates): Shows how to configure Flipt to send audit events to a webhook using Flipts built-in templating.

## Requirements

To run these examples application you'll need:

- [Docker](https://docs.docker.com/install/)
- [docker-compose](https://docs.docker.com/compose/install/)

## Basics

This example shows how you can run Flipt with audit event logging enabled to POST to a webhook using the `webhook` audit sink. It also configures filtering of audit events to only send `created` and `updated` events for `flag` instead of the default of all events for all types.

This works by setting the three environment variables `FLIPT_AUDIT_SINKS_WEBHOOK_ENABLED`, `FLIPT_AUDIT_EVENT_SINKS_WEBHOOK_URL`, and `FLIPT_AUDIT_EVENTS`:

**Note**: Support for filtering and sending audit events to webhooks was added in [v1.27.0](https://github.com/flipt-io/flipt/releases/tag/v1.27.0) of Flipt.

```bash
FLIPT_AUDIT_SINKS_WEBHOOK_ENABLED=true
FLIPT_AUDIT_SINKS_WEBHOOK_URL=http://localhost:8081
FLIPT_AUDIT_EVENTS=flag:created flag:updated
```

### Running the Example

1. Run `docker-compose up` from this directory
1. Open the Flipt UI (default: [http://localhost:8080](http://localhost:8080))
1. Create some sample data: Flags/Segments/etc.
1. View the logs of the webhook server to see the audit events being sent to the webhook server. Note that only `created` and `updated` events for `flag` are sent as that is what we configured in the `FLIPT_AUDIT_EVENTS` environment variable.

```console
webhook-flipt-1    | 2023-09-13T17:41:53Z       DEBUG   create flag     {"server": "grpc", "request": "key:\"asdf\" name:\"asdf\" namespace_key:\"default\""}
webhook-flipt-1    | 2023-09-13T17:41:53Z       DEBUG   create flag     {"server": "grpc", "response": "key:\"asdf\" name:\"asdf\" created_at:{seconds:1694626913 nanos:835784629} updated_at:{seconds:1694626913 nanos:835784629} namespace_key:\"default\""}
webhook-flipt-1    | 2023-09-13T17:41:53Z       INFO    finished unary call with code OK        {"server": "grpc", "grpc.start_time": "2023-09-13T17:41:53Z", "system": "grpc", "span.kind": "server", "grpc.service": "flipt.Flipt", "grpc.method": "CreateFlag", "peer.address": "127.0.0.1:32914", "grpc.code": "OK", "grpc.time_ms": 1.906}
webhook-flipt-1    | 2023-09-13T17:41:53Z       DEBUG   performing batched sending of audit events      {"server": "grpc", "sink": "webhook", "batch size": 1}

webhook-webhook-1  | INFO:root:POST request,
webhook-webhook-1  | Path: /
webhook-webhook-1  | Headers:
webhook-webhook-1  | Host: webhook:8081
webhook-webhook-1  | User-Agent: Go-http-client/1.1
webhook-webhook-1  | Content-Length: 248
webhook-webhook-1  | Content-Type: application/json
webhook-webhook-1  | Accept-Encoding: gzip
webhook-webhook-1  | 
webhook-webhook-1  | 
webhook-webhook-1  | 
webhook-webhook-1  | Body:
webhook-webhook-1  | {"version":"0.1","type":"flag","action":"created","metadata":{"actor":{"authentication":"none","ip":"172.18.0.1"}},"payload":{"description":"","enabled":false,"key":"asdf","name":"asdf","namespace_key":"default"},"timestamp":"2023-09-13T17:41:53Z"}
webhook-webhook-1  | 
webhook-webhook-1  | 172.18.0.3 - - [13/Sep/2023 17:41:53] "POST / HTTP/1.1" 200 -
```

## Templates

This example shows how you can run Flipt with audit event logging enabled to POST to a webhook using the `webhook` audit sink and uses the [webhook templating functionality](https://www.flipt.io/docs/configuration/observability#webhook-templates).

Templating enables you to adapt the payload of the webhook request to formats that are expected by the webhook server such as those defined by Slack, PagerDuty, etc.

This example uses a similar configuration to the [Basics](#basics) example, however it leverages Flipts configuration file instead of environment variables to configure Flipt with the template configuration.

**Note**: Support for webhook templating was added in [v1.28.0](https://github.com/flipt-io/flipt/releases/tag/v1.28.0) of Flipt.

Here is the relevant configuration which configures the `webhook` audit sink to send audit events to the webhook server at `http://webhook:8081/` using the `POST` method and sets the `Content-Type` header to `application/json`. The body of the request is a JSON payload which uses the [Go template syntax](https://pkg.go.dev/text/template) to render the payload.

The template allows us to customize the payload sent to the webhook server. In this example we are transforming the `Type` and `Action` of the audit event into a new JSON field `event`. We are also forwarding on the `Timestamp` of the audit event into a JSON field `timestamp`.

```yaml
audit:
  events:
    - flag:created
    - flag:updated
  sinks:
    webhook:
      enabled: true
      templates:
        - url: http://webhook:8081/
          method: POST
          headers:
            Content-Type: application/json
          body: |
            {
              "event": "{{ .Type }} {{ .Action }}",
              "timestamp": "{{ .Timestamp }}"
            }
```

### Running the Example

1. Run `docker-compose -f docker-compose.template.yml up` from this directory
1. Open the Flipt UI (default: [http://localhost:8080](http://localhost:8080))
1. Create some sample data: Flags/Segments/etc.
1. View the logs of the webhook server to see the audit events being sent to the webhook server. Note that only `created` and `updated` events for `flag` are sent as that is what we configured in the `audit.events` configuration.
1. Note that the payload of the webhook request is different than the [Basics](#basics) example. The `event` field is a combination of the `Type` and `Action` of the audit event and the `timestamp` field is the `Timestamp` of the audit event.

```console
webhook-webhook-1  | INFO:root:POST request,
webhook-webhook-1  | Path: /
webhook-webhook-1  | Headers:
webhook-webhook-1  | Host: webhook:8081
webhook-webhook-1  | User-Agent: Go-http-client/1.1
webhook-webhook-1  | Content-Length: 69
webhook-webhook-1  | Content-Type: application/json
webhook-webhook-1  | Accept-Encoding: gzip
webhook-webhook-1  |
webhook-webhook-1  |
webhook-webhook-1  |
webhook-webhook-1  | Body:
webhook-webhook-1  | {
webhook-webhook-1  |   "event": "flag created",
webhook-webhook-1  |   "timestamp": "2023-10-02T16:10:49Z"
webhook-webhook-1  | }
webhook-webhook-1  |
```

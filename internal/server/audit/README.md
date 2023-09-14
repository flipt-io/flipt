# Audit Events

Audit Events are pieces of data that describe a particular thing that has happened in a system. At Flipt, we provide the functionality of processing and batching these audit events and an abstraction for sending these audit events to a sink.

If you have an idea of a sink that you would like to receive audit events on, there are certain steps you would need to take to contribute, which are detailed below.

## Filterable Audit Events

The ability to filter audit events was added in [v1.27.0](https://github.com/flipt-io/flipt/releases/tag/v1.27.0) of Flipt. The following audit events are currently filterable:

### Nouns

- `flag`
- `segment`
- `variant`
- `constraint`
- `rule`
- `distribution`
- `namespace`
- `rollout`
- `token`

### Verbs

- `created`
- `updated`
- `deleted`

Any combination of the above nouns and verbs can be used to filter audit events. For example, `flag:created` would filter audit events for only `created` events for `flags`.

You may also use the `*` wildcard to filter on all nouns or verbs. For example, `*:created` would filter audit events for only `created` events for all nouns.

Similarly, `flag:*` would filter audit events for all verbs for `flags`.

Finally, `*:*` would filter audit events for all nouns and verbs which is the default behavior.

## Contributing

The abstraction that we provide for implementation of receiving these audit events to a sink is [this](https://github.com/flipt-io/flipt/blob/d252d6c1fdaecd6506bf413add9a9979a68c0bd7/internal/server/audit/audit.go#L130-L134).

```go
type Sink interface {
	SendAudits([]Event) error
	Close() error
	fmt.Stringer
}
```

For contributions of new sinks, you can follow this pattern:

- Create a folder for your new sink under the `audit` package with a meaningful name of your sink
- Provide the implementation to how to send audit events to your sink via the `SendAudits`
- Provide the implementation of closing resources/connections to your sink via the `Close` method (this will be called asynchronously to the `SendAudits` method so account for that in your implementation)
- Provide the variables for configuration just like [here](https://github.com/flipt-io/flipt/blob/d252d6c1fdaecd6506bf413add9a9979a68c0bd7/internal/config/audit.go#L52) for connection details to your sink
- Add a conditional to see if your sink is enabled [here](https://github.com/flipt-io/flipt/blob/d252d6c1fdaecd6506bf413add9a9979a68c0bd7/internal/cmd/grpc.go#L261)
- Write respective tests

:rocket: you should be good to go!

Need help? Reach out to us on [GitHub](https://github.com/flipt-io/flipt), [Discord](https://www.flipt.io/discord), [Twitter](https://twitter.com/flipt_io), or [Mastodon](https://hachyderm.io/@flipt).

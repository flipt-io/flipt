# Deprecation Notices

This page is used to list deprecation notices for Flipt.

Deprecated configuration options will be removed after ~6 months from the time they were deprecated.

Deprecated API endpoints, fields and objects will be removed after ~1 year from the time they were deprecated.

## Active Deprecations

<!--

Template for new deprecations:

### property

> since [version](link to version)

Description.

=== Before

    ``` yaml
    foo: bar
    ```

=== After

    ``` yaml
    foo: bar
    ```

-->

### Jaeger Tracing Exporter

> since [v1.36.0](https://github.com/flipt-io/flipt/releases/tag/v1.36.0)

This module is no longer supported. OpenTelemetry dropped support for Jaeger exporter in July 2023. Jaeger officially accepts and recommends using OTLP.

### API ListFlagRequest, ListSegmentRequest, ListRuleRequest offset

> since [v1.13.0](https://github.com/flipt-io/flipt/releases/tag/v1.13.0)

`offset` has been deprecated in favor of `page_token`/`next_page_token` for `ListFlagRequest`, `ListSegmentRequest` and `ListRuleRequest`. See: [#936](https://github.com/flipt-io/flipt/issues/936).

## Expired Deprecation Notices

The following options were deprecated in the past and were already removed.

### tracing.jaeger.enabled

> deprecated in [v1.18.2](https://github.com/flipt-io/flipt/releases/tag/v1.18.2)
> removed in [v1.36.0](https://github.com/flipt-io/flipt/releases/tag/v1.36.0)

### ui.enabled

> deprecated in [v1.17.0](https://github.com/flipt-io/flipt/releases/tag/v1.17.0)
> removed in [v1.28.0](https://github.com/flipt-io/flipt/releases/tag/v1.28.0)

### db.migrations.path and db.migrations_path

> deprecated in [v1.14.0](https://github.com/flipt-io/flipt/releases/tag/v1.14.0)
> removed in [v1.28.0](https://github.com/flipt-io/flipt/releases/tag/v1.28.0)

### cache.memory.enabled

> deprecated in [v1.10.0](https://github.com/flipt-io/flipt/releases/tag/v1.10.0)
> removed in [v1.28.0](https://github.com/flipt-io/flipt/releases/tag/v1.28.0)

### cache.memory.expiration

> deprecated in [v1.10.0](https://github.com/flipt-io/flipt/releases/tag/v1.10.0)
> removed in [v1.28.0](https://github.com/flipt-io/flipt/releases/tag/v1.28.0)

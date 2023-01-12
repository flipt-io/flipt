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

### ui.enabled

> since [v1.17.0](https://github.com/flipt-io/flipt/releases/tag/v1.17.0)

An upcoming release will enable the UI always and this option will be removed.
There will be a new version of Flipt (headless) that will run Flipt without the UI and only include the API.

### db.migrations.path and db.migrations_path

> since [v1.14.0](https://github.com/flipt-io/flipt/releases/tag/v1.14.0)

These options are no longer considered during Flipt execution.
Database migrations are embedded directly within the Flipt binary.

### API ListFlagRequest, ListSegmentRequest, ListRuleRequest offset

> since [v1.13.0](https://github.com/flipt-io/flipt/releases/tag/v1.13.0)

`offset` has been deprecated in favor of `page_token`/`next_page_token` for `ListFlagRequest`, `ListSegmentRequest` and `ListRuleRequest`. See: [#936](https://github.com/flipt-io/flipt/issues/936).

### cache.memory.enabled

> since [v1.10.0](https://github.com/flipt-io/flipt/releases/tag/v1.10.0)

Enabling in-memory cache via `cache.memory` is deprecated in favor of setting the `cache.backend` to `memory` and `cache.enabled` to `true`.

=== Before

    ``` yaml
    cache:
      memory:
        enabled: true
    ```

=== After

    ``` yaml
    cache:
      enabled: true
      backend: memory
    ```

### cache.memory.expiration

> since [v1.10.0](https://github.com/flipt-io/flipt/releases/tag/v1.10.0)

Setting cache expiration via `cache.memory` is deprecated in favor of setting the `cache.backend` to `memory` and `cache.ttl` to the desired duration.

=== Before

    ``` yaml
    cache:
      memory:
        expiration: 1m
    ```

=== After

    ``` yaml
    cache:
      enabled: true
      backend: memory
      ttl: 1m
    ```

## Expired Deprecation Notices

The following options were deprecated in the past and were already removed.

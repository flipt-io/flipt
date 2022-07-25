# Deprecation Notices

This page is used to list deprecation notices for Flipt.

Deprecated configuration options will be removed after ~6 months from the time they were deprecated.

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

### cache.memory.enabled

> since [v1.10.0](https://github.com/markphelps/flipt/releases/tag/v1.10.0)

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

> since [v1.10.0](https://github.com/markphelps/flipt/releases/tag/v1.10.0)

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

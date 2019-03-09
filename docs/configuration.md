# Configuration

There are two ways to configure Flipt: using a configuration file or through environment variables.

## Configuration File

The default way that Flipt is configured is with the use of a configuration file `default.yml`.

This file is read when Flipt starts up and configures several important properties for the server.

You can edit any of these properties to your liking, and on restart Flipt will pick up the new changes.

!!! note
    These defaults are commented out in `default.yml` to give you an idea of what they are. To change them you'll first need to uncomment them.

These properties are as follows:

| Property | Description | Default |
|---|---|---|
| log.level | Level at which messages are logged (trace, debug, info, warn, error, fatal, panic) | info |
| cache.enabled | Enable caching | false |
| cache.size | Number of items cache can hold | 250 |
| server.host | The host address on which to serve the Flipt application | 0.0.0.0 |
| server.http_port | The port on which to serve the Flipt REST API and UI | 8080 |
| server.grpc_port | The port on which to serve the Flipt GRPC server | 9000 |
| db.name | The name given to the Flipt database (suffixed with .db) | flipt |
| db.path | Where to store the Flipt database | /var/opt/flipt |
| db.migrations.auto | If database migrations are run on Flipt startup | true |
| db.migrations.path | Where the Flipt database migration files are kept | /etc/flipt/config/migrations |

## Using Environment Variables

All options in the configuration file can be overridden using environment variables using the syntax:

```shell
FLIPT_<SectionName>_<KeyName>
```

!!! tip
    Using environment variables to override defaults is especially helpful when running with Docker as described in the [Installation](installation.md) documentation.

Everything should be upper case, `.` should be replaced by `_`. For example, given these configuration settings:

```yaml
server:
  grpc_port: 9000

db:
  name: flipt
  path: /var/opt/flipt
```

You can override them using:

```shell
export FLIPT_SERVER_GRPC_PORT=9001
export FLIPT_DB_NAME=my-db
export FLIPT_DB_PATH=/tmp/db
```

## Caching

In-memory caching is currently only available for flags. When enabled, in-memory caching has been shown to speed up the fetching of individual flags by 10x.

To enable caching set the following in your config:

```yaml
cache:
  enabled: true
```

Work is planned to add caching support to rule evaluation soon.

## Authentication

Flipt currently has no built in authentication, authorization or encryption as Flipt was designed to work inside your trusted architecture and not be exposed publicly.

If you do wish to expose the Flipt dashboard and REST API publicly using HTTP Basic Authentication, you can do so by using a reverse proxy. There is an [example](https://github.com/markphelps/flipt/tree/master/examples/auth) provided in the GitHub repository showing how this could work.

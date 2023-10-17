# LibSQL

[LibSQL](https://github.com/tursodatabase/libsql) was created as a fork by [Turso](https://turso.tech/) to fit some use cases that SQLite was originally designed for. It is fully compatible with the SQLite API, and has the added benefit of being ran behind an HTTP interface in a service called [sqld](https://github.com/tursodatabase/libsql/tree/main/libsql-server/sqld).

## Requirements

- [Docker](https://www.docker.com/)
- [docker-compose](https://docs.docker.com/compose/install/)

## Running the Example

1. Run `docker-compose up` from this directory
1. Open the `flipt-one` UI ([http://localhost:8080](http://localhost:8080)) or `flipt-two` UI ([http://localhost:8081](http://localhost:8081))

## Details

The `docker compose` will spin up two instances of Flipt (named both `flipt-one` and `flipt-two`). The names of these Flipt instances do not necessarily matter since writes will be directed to the primary anyway. The compose file also spins up two instances of `sqld` called `sqld-primary` and `sqld-replica`. As previously mentioned, all writes will be directed to the `sqld-primary` and data will be replicated to the `sqld-replica` per semantics of `sqld`.

<img src="./images/sqld-overview.png" alt="SQLD Overivew" width="500px" />

> The diagram above was taken from the [libsql](https://github.com/tursodatabase/libsql) repository itself, but gives a nice overview on how all the concepts mesh together.

## Data

Since we mount the directories `/tmp/data.db` and `/tmp/replica.db` as volumes to the `sqld` docker containers, you can explore the data on the host using the [sqlite3](https://www.sqlite.org/download.html) CLI. The data itself will live under `/tmp/data.db/dbs/default/data` for the `sqld-primary` instance and `/tmp/replica.db/dbs/default/data` for the `sqld-replica` instance.

# LibSQL

[LibSQL](https://github.com/tursodatabase/libsql) was created as a fork of SQLite by [Turso](https://turso.tech/) to fit some use cases that SQLite was not originally designed for. 

It's fully compatible with the SQLite API, and has the added benefit of being run behind an HTTP interface in a service called [sqld](https://github.com/tursodatabase/libsql/tree/main/libsql-server).

## Requirements

- [Docker](https://www.docker.com/)
- [docker-compose](https://docs.docker.com/compose/install/)

## Running the Example

1. Run `docker-compose up` from this directory
1. Open the `flipt-one` UI ([http://localhost:8080](http://localhost:8080)) or `flipt-two` UI ([http://localhost:8081](http://localhost:8081))
1. Create a new feature flag in either UI
1. Switch to the other UI and verify that the feature flag was replicated
1. Continue to modify data in either UI and verify that the data is replicated

## Details

`docker compose` will spin up two instances of Flipt (named both `flipt-one` and `flipt-two`). We also spin up two instances of `sqld` called `sqld-primary` and `sqld-replica`. All writes will be directed to the `sqld-primary` and data will be replicated to the `sqld-replica` per semantics of `sqld`.

<img src="./images/sqld-overview.png" alt="sqld Overview" width="500px" />

> The diagram above was taken from the [libsql](https://github.com/tursodatabase/libsql) repository itself, but gives a nice overview on how all the concepts mesh together.

## Data

Since we mount the directories `/tmp/data.db` and `/tmp/replica.db` as volumes to the `sqld` Docker containers, you can explore the data on the host using the [sqlite3](https://www.sqlite.org/download.html) CLI. 

The data will live under `/tmp/data.db/dbs/default/data` for the `sqld-primary` instance and `/tmp/replica.db/dbs/default/data` for the `sqld-replica` instance.

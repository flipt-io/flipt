Flipt at the Edge
------------

This lab seeks to explore running Flipt on the edge with novel underlying technologies. The use case here is for allowing a host service using a Flipt client to achieve even faster evaluations for flags that an otherwise traditional set up of Flipt in front of an RDBMS. Since feature flagging applications are generally more read than write heavy, it perfectly fits the use case of some of these previously mentioned novel technologies we will explore throughout this lab.

## LibSQL

[LibSQL](https://github.com/tursodatabase/libsql) was created as a fork by [Turso](https://turso.tech/) to fit some use cases that SQLite was originally designed for. It is fully compatible with the SQLite API, and has the added benefit of being ran behind an HTTP interface in a service called [sqld](https://github.com/tursodatabase/libsql/tree/main/libsql-server/sqld).

The project housed under the `libsql` directory runs Flipt against the `sqld` servers in two modes `primary`, and `replica`.

### Prerequisites

- [Docker](https://www.docker.com/)

### Instructions

From the `libsql` directory, run `docker-compose up`. This will spin up two instances of Flipt (named both `flipt-one` and `flipt-two`). The names of these Flipt instances do not necessarily matter since writes will be directed to the primary anyway. The `docker-compose.yml` file also spins up two instances of `sqld` called `sqld-primary` and `sqld-replica`. As previously mentioned, all writes will be directed to the `sqld-primary` and data will be replicated to the `sqld-replica` per semantics of `sqld`.

<img src="./images/sqld-overview.png" alt="SQLD Overivew" width="500px" />

> The diagram above was taken from the [libsql](https://github.com/tursodatabase/libsql) repository itself, but gives a nice overview on how all the concepts mesh together.

After all of the containers are successfully stood up, you can access the `flipt-one` instance via the url `http://127.0.0.1:8080`, and the `flipt-two` instance via the url `http://127.0.0.1:8081`, and start adding data as necessary.

### Data

Since we mount the directories `/tmp/data.db` and `/tmp/replica.db` as volumes to the `sqld` docker containers, you can explore the data on the host using the [sqlite3](https://www.sqlite.org/download.html) CLI. The data itself will live under `/tmp/data.db/dbs/default/data` for the `sqld-primary` instance and `/tmp/replica.db/dbs/default/data` for the `sqld-replica` instance.
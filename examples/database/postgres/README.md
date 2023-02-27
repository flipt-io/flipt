<p align="center">
    <img src="../../images/logos/postgresql.svg" alt="Postgres" width=250 height=250 />
</p>

# Postgres Example

This example shows how you can run Flipt with a Postgres database over the default SQLite.

This works by setting the environment variable `FLIPT_DB_URL` to point to the Postgres database running in a container:

```bash
FLIPT_DB_URL=postgres://postgres:password@postgres:5432/flipt?sslmode=disable
```

## Requirements

To run this example application you'll need:

* [Docker](https://docs.docker.com/install/)
* [docker-compose](https://docs.docker.com/compose/install/)

## Running the Example

1. Run `docker-compose up` from this directory
1. Open the Flipt UI (default: [http://localhost:8080](http://localhost:8080))

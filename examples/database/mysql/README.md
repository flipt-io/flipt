<p align="center">
    <img src="../../images/logos/mysql.svg" alt="MySQL" width=250 height=250 />
</p>

# MySQL Example

This example shows how you can run Flipt with a MySQL database over the default SQLite.

This works by setting the environment variable `FLIPT_DB_URL` to point to the MySQL database running in a container:

```bash
FLIPT_DB_URL=mysql://mysql:password@mysql:3306/flipt
```

## Requirements

To run this example application you'll need:

* [Docker](https://docs.docker.com/install/)
* [docker-compose](https://docs.docker.com/compose/install/)

## Running the Example

1. Run `docker-compose up` from this directory
1. Open the Flipt UI (default: [http://localhost:8080](http://localhost:8080))

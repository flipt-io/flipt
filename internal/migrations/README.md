# Database Migrations

This directory contains the database migrations for the v2 version of Flipt.

We use [golang-migrate](https://github.com/golang-migrate/migrate) to create and manage the migrations.

To create a new migration, run the following command:

```sh
migrate create -ext sql -dir ./migrations/{db} <migration_name>
```

Where `{db}` is the database type, e.g. `mysql`, `postgres`, `sqlite`, etc.

Example:

```sh
migrate create -ext sql -dir ./migrations/mysql create_table_X
```

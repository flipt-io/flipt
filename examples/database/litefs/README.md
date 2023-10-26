# LiteFS

LiteFS is a distributed file system that replicates SQLite databases to other nodes at the file system level. It works by using [FUSE](https://www.kernel.org/doc/html/next/filesystems/fuse.html) to detect writes to the file system and determines how and if those should be replicated.

This example will demonstrate how to run Flipt over LiteFS.

## Requirements

- [Docker](https://www.docker.com/)
- [docker-compose](https://docs.docker.com/compose/install/)

## Running the Example

1. Run `docker-compose up` from this directory
1. Open the Flipt UI (default: [http://localhost:8080](http://localhost:8080))

## Details

`docker compose` will spin up two instances of Flipt with embedded SQLite databases. On top of these instances is an nginx proxy that will forward "write" requests, anything but `GET`, to the primary. `GET` requests will be served from the instance's embedded SQLite database.

LiteFS talks about following this [caveat](https://fly.io/docs/litefs/proxy/#how-it-works) in order for it to work properly.

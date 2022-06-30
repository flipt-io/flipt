# Basic Example

This example shows how you could integrate Flipt into your application.

It uses the [Go GRPC Flipt client](https://github.com/flipt-io/flipt-grpc-go) to query for an existing flag and then shows different content whether or not that flag is enabled.

## Requirements

To run this example application you'll need:

* [Docker](https://docs.docker.com/install/)
* [docker-compose](https://docs.docker.com/compose/install/)

## Running the Example

1. Run `docker-compose up` from this directory
1. Open the Flipt UI (default: [http://localhost:8080](http://localhost:8080))
1. Open the example UI at [http://localhost:8000](http://localhost:8000)
1. Switch back to the Flipt UI
1. Disable / Enable the example flag `example` in the Flipt UI
1. Refresh the example UI, you should see the content change

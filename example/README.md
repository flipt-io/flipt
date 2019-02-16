# Example Application

This example shows how you could integrate Flipt into your application.

It uses the Go GRPC Flipt client to query for an existing flag and then show's different content whether or not that flag is enabled.

!["demo"](../docs/assets/images/demo.gif?raw=true)

## Requirements

To run this example application you'll need:

* Go 1.10+ installed
* The Flipt GRPC client on your `$GOPATH`.
* The Flipt server running. See [Installation](../docs/installation.md) documentation for how to install/run Flipt.

## Runtime Configuration

There are two options that can be passed to the demo application at startup.

| Option | Description | Default |
|---|---|---|
| --server | The address of the Flipt server backend | localhost:9000 |
| --flag | The flag key to query for the example | example |

## Running the Example

1. With the Flipt server running, open the Flipt UI (default: [http://localhost:8080](http://localhost:8080))
1. Create a flag with the key `example` or use any other key that you would like
1. Start the demo application by running `go run main.go` from this directory, optionally passing in any configuration flags.
1. Open the example UI at [http://localhost:8000](http://localhost:8000)
1. Disable / Enable your example flag `example` in the Flipt UI
1. Refresh the example UI, you should see the content change

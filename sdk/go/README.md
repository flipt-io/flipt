# Flipt Go SDK

[![Go Reference](https://pkg.go.dev/badge/go.flipt.io/flipt/sdk/go.svg)](https://pkg.go.dev/go.flipt.io/flipt/sdk/go)

The Flipt Go SDK supports developing applications in Go against Flipt.
It also supports the ability to access the resource management APIs and other systems, such as authentication and metadata.

The SDK supports both Flipts `gRPC` and `HTTP` RPC APIs.
A majority of this client is generated directly from Flipt's `protobuf` definitions.
The [Flipt SDK Generator](../../internal/cmd/protoc-gen-go-flipt-sdk/) can be found locally within this repository.

## Dependencies

- Go `>= v1.20`

## Get the SDK

```sh
go get go.flipt.io/flipt/sdk/go
```

## Construct and Authenticate the SDK

Constructing an SDK client is easy.

1. Pick your transport of choice. Both`grpc` or `http` are sub-packages with respective implementations.
2. Pass a constructed `Transport` implementation to `sdk.New(...)`.
3. Optionally pass in a `sdk.ClientTokenProvider` to authenticate your RPC calls.

```go
package main

import (
	sdk "go.flipt.io/flipt/sdk/go"
	sdkgrpc "go.flipt.io/flipt/sdk/grpc"
	grpc "google.golang.org/grpc"
)

func main() {
	token := sdk.StaticClientTokenProvider("a-flipt-client-token")

	conn := grpc.Dial("localhost:9090")
	transport := sdkgrpc.NewTransport(conn)

	client := sdk.New(transport, sdk.WithClientTokenProvider(token))
}
```

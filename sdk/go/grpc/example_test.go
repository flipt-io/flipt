package grpc

import (
	"context"

	sdk "go.flipt.io/flipt/sdk"
	grpc "google.golang.org/grpc"
)

func ExampleNewTransport() {
	cc, _ := grpc.Dial("localhost:9090")

	transport := NewTransport(cc)

	sdk.New(transport).Meta().GetInfo(context.Background())
}

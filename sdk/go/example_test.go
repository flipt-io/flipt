package sdk

import (
	context "context"

	flipt "go.flipt.io/flipt/rpc/flipt"
)

func ExampleNew() {
	// see the following subpackages for transport implementations:
	// - grpc
	var transport Transport

	client := New(transport)

	client.Flipt().GetFlag(context.Background(), &flipt.GetFlagRequest{Key: "my_flag"})
}

package sdk

import (
	context "context"
)

func ExampleNew() {
	// see the following subpackages for transport implementations:
	// - grpc
	var transport Transport

	client := New(transport)

	client.Flipt().GetFlag(context.Background(), "my_flag")
}

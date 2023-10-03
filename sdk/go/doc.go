/*
Package sdk is the official Flipt Go SDK.

The SDK exposes the various RPC for interfacing with a remote Flipt instance.
Both HTTP and gRPC protocols are supported via this unified Go API.

The main entrypoint within this package is [New], which takes an instance of [Transport] and returns an instance of [SDK].
[Transport] implementations can be found in the following sub-packages:

  - [go.flipt.io/flipt/sdk/go/grpc.NewTransport]
  - [go.flipt.io/flipt/sdk/go/http.NewTransport]

# GRPC Transport

The following is an example of creating and instance of the SDK using the gRPC transport.

	func main() {
		conn := grpc.Dial("localhost:9090")
		transport := grpc.NewTransport(conn)
		client := sdk.New(transport)
	}

# HTTP Transport

The following is an example of creating an instance of the SDK using the HTTP transport.

	func main() {
		transport := http.NewTransport("http://localhost:8080")
		client := sdk.New(transport)
	}

# Authenticating the SDK

The remote procedure calls mades by this SDK are authenticated via a [ClientTokenProvider] implementation.
This can be supplied to [New] via the [WithClientTokenProvider] option.

Currently, there only exists a single implementation [StaticClientTokenProvider]:

	func main() {
		provider := sdk.StaticClientTokenProvider("some-flipt-token")
		client := sdk.New(transport, sdk.WithClientTokenProvider(provider))
	}

# SDK Services

The Flipt [SDK] is split into four sections [Flipt], [Auth], [Meta], and [Evaluation]
Each of which provides access to different parts of the Flipt system.

# Flipt Service

The [Flipt] service is the core Flipt API service.
This service provides access to the Flipt resource CRUD APIs.

	client := sdk.New(transport).Flipt()

Flipt resources can be accessed and managed directly.

	flag, err := client.GetFlag(ctx, &flipt.GetFlagRequest{
		NamespaceKey: "my_namespace",
		Key: "my_flag_key",
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(flag.Name)
	fmt.Println(flag.Description)

# Evaluation Service

The [Evaluation] service provides access to the Flipt evaluation APIs.

	client := sdk.New(transport).Evaluation()

The [Evaluation] service provides three methods for evaluating a flag for a given entity:
Boolean, Variant, and Batch.

# Boolean

The Boolean method returns a response containing a boolean value indicating whether or not the flag is enabled for the given entity.
Learn more about the Boolean flag type: <https://www.flipt.io/docs/concepts#boolean-flags>

	resp, err := client.Boolean(ctx, &evaluation.EvaluationRequest{
		NamespaceKey: "my_namespace",
		FlagKey: "my_flag_key",
		EntityId: "my_entity_id",
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(resp.Enabled)

# Variant

The Variant method returns a response containing the variant key for the given entity.
Learn more about the Variant flag type: <https://www.flipt.io/docs/concepts#variant-flags>

	resp, err := client.Variant(ctx, &evaluation.EvaluationRequest{
		NamespaceKey: "my_namespace",
		FlagKey: "my_flag_key",
		EntityId: "my_entity_id",
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(resp.VariantKey)
	fmt.Println(resp.VariantAttachment) Optional

# Batch

The Batch method returns a response containing the evaluation results for a batch of requests. These requests can be for a mix of boolean and variant flags.

	resp, err := client.Batch(ctx, &evaluation.BatchRequest{
		RequestId: "my_request_id",
		Requests: []*evaluation.EvaluationRequest{
			{
				NamespaceKey: "my_namespace",
				FlagKey: "my_flag_key",
				EntityId: "my_entity_id",
			}
		},
	})
	if err != nil {
		panic(err)
	}

	for _, result := range resp.Responses {
		fmt.Println(result.Type)
	 }
*/
package sdk

// Package sdk is the official Flipt Go SDK.
// The SDK exposes the various RPC for interfacing with a remote Flipt instance.
// Both HTTP and gRPC protocols are supported via this unified Go API.
//
// The main entrypoint within this package is [New], which takes an instance of [Transport] and returns an instance of [SDK].
// [Transport] implementations can be found in the following sub-packages:
//
//   - [go.flipt.io/flipt/sdk/go/grpc.NewTransport]
//   - [go.flipt.io/flipt/sdk/go/http.NewTransport]
//
// # GRPC Transport
//
// The following is an example of creating and instance of the SDK using the gRPC transport.
//
//	func main() {
//	    conn := grpc.Dial("localhost:9090")
//	    transport := grpc.NewTransport(conn)
//	    sdk := sdk.New(transport)
//	}
//
// # HTTP Transport
//
// The following is an example of creating an instance of the SDK using the HTTP transport.
//
//	func main() {
//	    transport := http.NewTransport("http://localhost:8080")
//	    sdk := sdk.New(transport)
//	}

// # Authenticating the SDK
//
// The remote procedure calls mades by this SDK are authenticated via a [ClientTokenProvider] implementation.
// This can be supplied to [New] via the [WithClientTokenProvider] option.
//
// Currently, there only exists a single implementation [StaticClientTokenProvider]:
//
//	func main() {
//	    provider := sdk.StaticClientTokenProvider("some-flipt-token")
//	    sdk.New(transport, sdk.WithClientTokenProvider(provider))
//	}
//
// # SDK Services
//
// The Flipt [SDK] is split into three sections [Flipt], [Auth] and [Meta].
// Each of which provides access to different parts of the Flipt system.
//
// # Flipt Service
//
// The [Flipt] service is the core Flipt API service.
// This service provides access to evaluate flag configuration within your application.
// As well as exposing the Flipt resource CRUD APIs.
//
//	client := sdk.New(transport).Flipt()
//
// The following demonstrates how to evaluate the state of a flag.
//
//	result, err := client.Evaluate(ctx, &flipt.EvaluationRequest{
//	    RequestId: uuid.NewV4().String(),
//	    FlagKey: "my_flag_key",
//	    EntityId: userID,
//	    Context: map[string]string{
//	        "organization": orgName,
//	    },
//	})
//
// Additionally, Flipt resources can be accessed and managed directly.
//
//	flag, err := client.GetFlag(ctx, &flipt.GetFlagRequest{Key: "my_flag_key"})
//	if err != nil {
//	    panic(err)
//	}
//
//	fmt.Println(flag.Name)
//	fmt.Println(flag.Description)
package sdk

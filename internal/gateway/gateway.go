package gateway

import (
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.flipt.io/flipt/rpc/flipt"
	"google.golang.org/protobuf/encoding/protojson"
)

var (
	// JSONMarshaler is the base JSON marshaller for Flipts gRPC gateway handler stack.
	JSONMarshaler = flipt.NewV1toV2MarshallerAdapter()
	// JSONPrettyMarshaler is Flipt's JSON marshaler when the `pretty` directive is added
	// to the requested JSON mime-type.
	JSONPrettyMarshaler = &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			Indent:    "  ",
			Multiline: true, // Optional, implied by presence of "Indent".
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	}
)

// commonMuxOptions are options for gateway mux which are used for multiple instances.
// This is required to fix a backwards compatibility issue with the v2 marshaller where `null` map values
// cause an error because they are not allowed by the proto spec, but they were handled by the v1 marshaller.
//
// See: rpc/flipt/marshal.go
//
// See: https://github.com/flipt-io/flipt/issues/664
var commonMuxOptions = []runtime.ServeMuxOption{
	runtime.WithMarshalerOption(runtime.MIMEWildcard, JSONMarshaler),
	runtime.WithMarshalerOption("application/json+pretty", JSONPrettyMarshaler),
}

// NewGatewayServeMux builds a new gateway serve mux with common options.
func NewGatewayServeMux(opts ...runtime.ServeMuxOption) *runtime.ServeMux {
	return runtime.NewServeMux(append(commonMuxOptions, opts...)...)
}

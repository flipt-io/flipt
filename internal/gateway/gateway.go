package gateway

import (
	"sync"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
)

// commonMuxOptions are options for gateway mux which are used for multiple instances.
// This is required to fix a backwards compatibility issue with the v2 marshaller where `null` map values
// cause an error because they are not allowed by the proto spec, but they were handled by the v1 marshaller.
//
// See: rpc/flipt/marshal.go
//
// See: https://github.com/flipt-io/flipt/issues/664
var commonMuxOptions []runtime.ServeMuxOption
var once sync.Once

// NewGatewayServeMux builds a new gateway serve mux with common options.
func NewGatewayServeMux(logger *zap.Logger, opts ...runtime.ServeMuxOption) *runtime.ServeMux {
	once.Do(func() {
		commonMuxOptions = []runtime.ServeMuxOption{
			runtime.WithMarshalerOption(runtime.MIMEWildcard, flipt.NewV1toV2MarshallerAdapter(logger)),
			runtime.WithMarshalerOption("application/json+pretty", &runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{
					Indent:    "  ",
					Multiline: true, // Optional, implied by presence of "Indent".
				},
				UnmarshalOptions: protojson.UnmarshalOptions{
					DiscardUnknown: true,
				},
			}),
		}

	})

	return runtime.NewServeMux(append(opts, commonMuxOptions...)...)
}

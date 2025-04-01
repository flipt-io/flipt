package gateway

import (
	"net/http"
	"slices"
	"sync"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
)

// commonMuxOptions are options for gateway mux which are used for multiple instances.
// This is required to fix a backwards compatibility issue with the v2 marshaller where `null` map values
// cause an error because they are not allowed by the proto spec, but they were handled by the v1 marshaller.
//
// See: https://github.com/flipt-io/flipt/issues/664
var (
	commonMuxOptions []runtime.ServeMuxOption
	once             sync.Once
)

// NewGatewayServeMux builds a new gateway serve mux with common options.
func NewGatewayServeMux(logger *zap.Logger, opts ...runtime.ServeMuxOption) *runtime.ServeMux {
	once.Do(func() {
		commonMuxOptions = []runtime.ServeMuxOption{
			runtime.WithMarshalerOption(runtime.MIMEWildcard, NewV1toV2MarshallerAdapter(logger)),
			runtime.WithMarshalerOption("application/json+pretty", &runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{
					Indent:    "  ",
					Multiline: true, // Optional, implied by presence of "Indent".
				},
				UnmarshalOptions: protojson.UnmarshalOptions{
					DiscardUnknown: true,
				},
			}),
			// Make sure trace-related headers are propagated from otelhttp through
			// grpc-gateway.
			// Usually this would be taken care of by otelgrpc client interceptors,
			// but inprocgrpc does not provide client interceptors.
			runtime.WithIncomingHeaderMatcher(func(key string) (string, bool) {
				if h := http.CanonicalHeaderKey(key); slices.Contains([]string{"Traceparent", "Tracestate", "Baggage"}, h) {
					return h, true
				}
				return runtime.DefaultHeaderMatcher(key)
			}),
		}
	})

	return runtime.NewServeMux(append(opts, commonMuxOptions...)...)
}

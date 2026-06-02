package gateway

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"sync"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.flipt.io/flipt/internal/server/common"
	"go.flipt.io/flipt/rpc/v2/evaluation"
	"go.uber.org/zap"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
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

const eventStreamContentType = "text/event-stream"

// NewGatewayServeMux builds a new gateway serve mux with common options.
func NewGatewayServeMux(logger *zap.Logger, opts ...runtime.ServeMuxOption) *runtime.ServeMux {
	once.Do(func() {
		commonMuxOptions = []runtime.ServeMuxOption{
			runtime.WithMarshalerOption(runtime.MIMEWildcard, NewV1toV2MarshallerAdapter(logger)),
			runtime.WithMarshalerOption(eventStreamContentType, &eventSourceMarshaler{JSONPb: runtime.JSONPb{}}),
			runtime.WithStreamErrorHandler(func(_ context.Context, err error) *grpcstatus.Status {
				if errors.Is(err, context.Canceled) || grpcstatus.Code(err) == codes.Canceled {
					return grpcstatus.New(codes.Canceled, "client disconnected")
				}
				return grpcstatus.Convert(err)
			}),
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
				// Prevent clients from spoofing the internal OFREP stream marker.
				if isOFREPReservedHeader(key) {
					return "", false
				}
				return runtime.DefaultHeaderMatcher(key)
			}),
		}
	})

	return runtime.NewServeMux(append(opts, commonMuxOptions...)...)
}

type eventSourceMarshaler struct {
	runtime.JSONPb
}

func (m *eventSourceMarshaler) ContentType(_ any) string {
	return eventStreamContentType
}

func (m *eventSourceMarshaler) StreamContentType(_ any) string {
	return eventStreamContentType
}

func (m *eventSourceMarshaler) Marshal(v any) ([]byte, error) {
	switch val := v.(type) {
	case map[string]any:
		return m.marshalEventSourceMap(val)
	case map[string]proto.Message:
		converted := make(map[string]any, len(val))
		for k, v := range val {
			converted[k] = v
		}
		return m.marshalEventSourceMap(converted)
	case *status.Status:
		return m.marshalStatus(val)
	default:
		return nil, fmt.Errorf("unsupported event-stream payload %T", v)
	}
}

func (m *eventSourceMarshaler) Unmarshal(_ []byte, _ any) error {
	return errors.New("text/event-stream unmarshaler is unsupported")
}

func (m *eventSourceMarshaler) Delimiter() []byte {
	return []byte{}
}

func (m *eventSourceMarshaler) marshalEventSourceMap(val map[string]any) ([]byte, error) {
	// expected happy path
	if vv, ok := val["result"].(*evaluation.EvaluationNamespaceSnapshot); ok {
		b, err := m.JSONPb.Marshal(map[string]string{
			"type": "refetchEvaluation",
			"etag": vv.Digest,
		})
		if err != nil {
			return nil, err
		}
		return fmt.Appendf(nil, "id: %s\ndata: %s\n\n", vv.Digest[:min(len(vv.Digest), 9)], b), nil
	}

	if errStatus, ok := val["error"].(*status.Status); ok {
		return m.marshalStatus(errStatus)
	}

	return nil, fmt.Errorf("unsupported event-stream payload map: %T", val)
}

// isOFREPReservedHeader checks whether the given HTTP header key is reserved for
// internal use and must not be forwarded as gRPC metadata from client requests.
func isOFREPReservedHeader(key string) bool {
	canonicalKey := http.CanonicalHeaderKey(key)
	reservedCanonical := http.CanonicalHeaderKey(common.HeaderFliptOFREPStream)
	if canonicalKey == reservedCanonical {
		return true
	}
	// Also check the Grpc-Metadata-* prefixed form which DefaultHeaderMatcher
	// would forward after stripping the prefix.
	if trimmed, ok := strings.CutPrefix(key, "Grpc-Metadata-"); ok {
		if http.CanonicalHeaderKey(trimmed) == reservedCanonical {
			return true
		}
	}
	return false
}

func (m *eventSourceMarshaler) marshalStatus(errStatus *status.Status) ([]byte, error) {
	if codes.Code(errStatus.Code) == codes.Canceled {
		return nil, nil
	}
	b, err := m.JSONPb.Marshal(map[string]any{
		"type":    "error",
		"code":    errStatus.Code,
		"message": errStatus.Message,
	})
	if err != nil {
		return nil, err
	}
	return fmt.Appendf(nil, "data: %s\n\n", b), nil
}

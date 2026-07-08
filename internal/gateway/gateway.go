package gateway

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/grpc-ecosystem/grpc-gateway/v2/utilities"
	"go.flipt.io/flipt/internal/server/common"
	"go.flipt.io/flipt/rpc/flipt/auth"
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

const (
	// MIMEEventStream is the content type for server-sent events.
	MIMEEventStream = "text/event-stream"
	// MIMEFormURLEncoded is the content type for URL-encoded form data.
	MIMEFormURLEncoded = "application/x-www-form-urlencoded"
	// maxFormURLEncodedBodySize is the maximum size of a form-urlencoded request body.
	// 4 KiB is sufficient for OIDC back-channel logout requests and prevents memory exhaustion.
	maxFormURLEncodedBodySize = 4 * 1024
)

// NewGatewayServeMux builds a new gateway serve mux with common options.
func NewGatewayServeMux(logger *zap.Logger, opts ...runtime.ServeMuxOption) *runtime.ServeMux {
	once.Do(func() {
		commonMuxOptions = []runtime.ServeMuxOption{
			runtime.WithMarshalerOption(runtime.MIMEWildcard, NewV1toV2MarshallerAdapter(logger)),
			runtime.WithMarshalerOption(MIMEEventStream, &eventSourceMarshaler{JSONPb: runtime.JSONPb{}}),
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
	return MIMEEventStream
}

func (m *eventSourceMarshaler) StreamContentType(_ any) string {
	return MIMEEventStream
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

// formURLEncodedMarshaler is a custom marshaler for application/x-www-form-urlencoded requests.
// It reuses JSONBuiltin for marshalling responses but custom-decodes form-encoded bodies
// into protobuf messages (specifically RevokeOIDCRequest).
type formURLEncodedMarshaler struct {
	runtime.JSONBuiltin
}

// ContentType returns the content type of the response.
func (u formURLEncodedMarshaler) ContentType(v any) string {
	return u.JSONBuiltin.ContentType(v)
}

// Marshal encodes the response as JSON (delegates to JSONBuiltin).
func (u formURLEncodedMarshaler) Marshal(v any) ([]byte, error) {
	// can marshal the response in proto message format
	return u.JSONBuiltin.Marshal(v)
}

// NewDecoder returns a decoder that parses URL-encoded form data into a RevokeOIDCRequest.
func (u formURLEncodedMarshaler) NewDecoder(r io.Reader) runtime.Decoder {
	filter := &utilities.DoubleArray{}
	return runtime.DecoderFunc(func(p any) error {
		msg, ok := p.(*auth.RevokeOIDCRequest)
		if !ok {
			return fmt.Errorf("not proto message")
		}

		limitedReader := http.MaxBytesReader(nil, io.NopCloser(r), maxFormURLEncodedBodySize)
		formData, err := io.ReadAll(limitedReader)
		if err != nil {
			return err
		}

		values, err := url.ParseQuery(string(formData))
		if err != nil {
			return err
		}

		return runtime.PopulateQueryParameters(msg, values, filter)
	})
}

// NewFormURLEncodedMarshaler returns a ServeMuxOption that registers a custom marshaler
// for application/x-www-form-urlencoded content type, enabling form-encoded POST requests
// (such as OIDC back-channel logout from providers that send form-encoded bodies).
func NewFormURLEncodedMarshaler() runtime.ServeMuxOption {
	return runtime.WithMarshalerOption(MIMEFormURLEncoded, &formURLEncodedMarshaler{runtime.JSONBuiltin{}})
}

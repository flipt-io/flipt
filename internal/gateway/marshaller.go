package gateway

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"

	"github.com/golang/protobuf/jsonpb"
	grpc_gateway_v2 "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/runtime/protoiface"
)

var _ grpc_gateway_v2.Marshaler = &V1toV2MarshallerAdapter{}

// V1toV2MarshallerAdapter is a V1 to V2 marshaller adapter to be able to use the v1 marshaller
// from grpc-gateway v2.
//
// This is required to fix a backwards compatibility issue with the v2 marshaller where `null` map values
// cause an error because they are not allowed by the proto spec, but they were handled by the v1 marshaller.
//
// See: https://github.com/flipt-io/flipt/issues/664
//
// TODO: remove this custom marshaller for Flipt API v2 as we want to use the default v2 marshaller directly.
type V1toV2MarshallerAdapter struct {
	*grpc_gateway_v2.JSONPb
	logger *zap.Logger
}

func NewV1toV2MarshallerAdapter(logger *zap.Logger) *V1toV2MarshallerAdapter {
	return &V1toV2MarshallerAdapter{&grpc_gateway_v2.JSONPb{
		MarshalOptions: protojson.MarshalOptions{EmitDefaultValues: true},
	}, logger}
}

func (m *V1toV2MarshallerAdapter) ContentType(v any) string {
	return "application/json"
}

func (m *V1toV2MarshallerAdapter) Marshal(v any) ([]byte, error) {
	return m.JSONPb.Marshal(v)
}

// decoderInterceptor intercepts and modifies the outbound error return value for
// inputs that fail to unmarshal against the protobuf.
type decoderInterceptor struct {
	grpc_gateway_v2.Decoder
	logger       *zap.Logger
	unmarshaller protojson.UnmarshalOptions
}

func (c *decoderInterceptor) Decode(v any) error {
	var err error
	if pt, ok := v.(protoiface.MessageV1); ok {
		unmarshaler := &jsonpb.Unmarshaler{AllowUnknownFields: !c.unmarshaller.DiscardUnknown}
		// Decode into bytes for marshalling
		var b json.RawMessage
		if err := c.Decoder.Decode(&b); err != nil {
			return err
		}
		err = unmarshaler.Unmarshal(bytes.NewBuffer(b), pt)
	} else {
		err = c.Decoder.Decode(v)
	}
	if err != nil {
		c.logger.Debug("JSON decoding failed for inputs", zap.Error(err))
		var uerr *json.UnmarshalTypeError
		if errors.As(err, &uerr) {
			return errors.New("invalid values for key(s) in json body")
		}

		return err
	}

	return nil
}

func (m *V1toV2MarshallerAdapter) NewDecoder(r io.Reader) grpc_gateway_v2.Decoder {
	return &decoderInterceptor{Decoder: json.NewDecoder(r), logger: m.logger, unmarshaller: m.JSONPb.UnmarshalOptions}
}

func (m *V1toV2MarshallerAdapter) NewEncoder(w io.Writer) grpc_gateway_v2.Encoder {
	return m.JSONPb.NewEncoder(w)
}

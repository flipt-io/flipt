package flipt

import (
	"encoding/json"
	"errors"
	"io"

	grpc_gateway_v1 "github.com/grpc-ecosystem/grpc-gateway/runtime"
	grpc_gateway_v2 "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
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
	*grpc_gateway_v1.JSONPb
	logger *zap.Logger
}

func NewV1toV2MarshallerAdapter(logger *zap.Logger) *V1toV2MarshallerAdapter {
	return &V1toV2MarshallerAdapter{&grpc_gateway_v1.JSONPb{OrigName: false, EmitDefaults: true}, logger}
}

func (m *V1toV2MarshallerAdapter) ContentType(_ interface{}) string {
	return m.JSONPb.ContentType()
}

func (m *V1toV2MarshallerAdapter) Marshal(v interface{}) ([]byte, error) {
	return m.JSONPb.Marshal(v)
}

// decoderInterceptor intercepts and modifies the outbound error return value for
// inputs that fail to unmarshal against the protobuf.
type decoderInterceptor struct {
	grpc_gateway_v1.Decoder
	logger *zap.Logger
}

func (c *decoderInterceptor) Decode(v interface{}) error {
	err := c.Decoder.Decode(v)
	if err != nil {
		c.logger.Debug("JSON decoding failed for inputs", zap.Error(err))

		if _, ok := err.(*json.UnmarshalTypeError); ok {
			return errors.New("invalid values for key(s) in json body")
		}

		return err
	}

	return nil
}

func (m *V1toV2MarshallerAdapter) NewDecoder(r io.Reader) grpc_gateway_v2.Decoder {
	return &decoderInterceptor{Decoder: m.JSONPb.NewDecoder(r), logger: m.logger}
}

func (m *V1toV2MarshallerAdapter) NewEncoder(w io.Writer) grpc_gateway_v2.Encoder {
	return m.JSONPb.NewEncoder(w)
}

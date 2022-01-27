package flipt

import (
	"io"

	grpc_gateway_v1 "github.com/grpc-ecosystem/grpc-gateway/runtime"
	grpc_gateway_v2 "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

var _ grpc_gateway_v2.Marshaler = &V1toV2MarshallerAdapter{}

// V1toV2MarshallerAdapter is a V1 to V2 marshaller adapter to be able to use the v1 marshaller
// from grpc-gateway v2.
//
// This is required to fix a backwards compatibility issue with the v2 marshaller where `null` map values
// cause an error because they are not allowed by the proto spec, but they were handled by the v1 marshaller.
//
// See: https://github.com/markphelps/flipt/issues/664
//
// TODO: remove this custom marshaller for Flipt API v2 as we want to use the default v2 marshaller directly.
type V1toV2MarshallerAdapter struct {
	*grpc_gateway_v1.JSONPb
}

func NewV1toV2MarshallerAdapter() *V1toV2MarshallerAdapter {
	return &V1toV2MarshallerAdapter{&grpc_gateway_v1.JSONPb{OrigName: false, EmitDefaults: true}}
}

func (m *V1toV2MarshallerAdapter) ContentType(_ interface{}) string {
	return m.JSONPb.ContentType()
}

func (m *V1toV2MarshallerAdapter) Marshal(v interface{}) ([]byte, error) {
	return m.JSONPb.Marshal(v)
}

func (m *V1toV2MarshallerAdapter) NewDecoder(r io.Reader) grpc_gateway_v2.Decoder {
	return m.JSONPb.NewDecoder(r)
}

func (m *V1toV2MarshallerAdapter) NewEncoder(w io.Writer) grpc_gateway_v2.Encoder {
	return m.JSONPb.NewEncoder(w)
}

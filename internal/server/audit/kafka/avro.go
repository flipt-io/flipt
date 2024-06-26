package kafka

import (
	"bytes"
	"errors"

	"github.com/hamba/avro"
	"go.flipt.io/flipt/internal/server/audit"
	rpcaudit "go.flipt.io/flipt/rpc/flipt/audit"
)

func toAvro() encodingFn {
	schema := avro.MustParse(rpcaudit.AvroSchema)
	return func(v any) ([]byte, error) {
		if event, ok := v.(audit.Event); ok {
			var e audit.Event
			var err error
			event.CopyInto(&e)
			e.Payload, err = event.PayloadToMap()
			if err != nil {
				return nil, err
			}
			buf := &bytes.Buffer{}
			encoder := avro.NewEncoderForSchema(schema, buf)
			err = encoder.Encode(e)
			return buf.Bytes(), err
		}
		return nil, errors.New("unsupported struct")
	}
}

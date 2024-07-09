package kafka

import (
	"bytes"
	"errors"

	"github.com/hamba/avro/v2"
	"github.com/twmb/franz-go/pkg/sr"
	"go.flipt.io/flipt/internal/server/audit"
	protoaudit "go.flipt.io/flipt/rpc/flipt/audit"
)

func newAvroEncoder() *avroEncoder {
	schema := avro.MustParse(protoaudit.AvroSchema)
	return &avroEncoder{schema: schema}
}

type avroEncoder struct {
	schema avro.Schema
}

func (a *avroEncoder) Schema() sr.Schema {
	return sr.Schema{
		Schema: protoaudit.AvroSchema,
		Type:   sr.TypeAvro,
	}
}

func (a *avroEncoder) Encode(v any) ([]byte, error) {
	if event, ok := v.(audit.Event); ok {
		var e audit.Event
		var err error
		event.CopyInto(&e)
		e.Payload, err = event.PayloadToMap()
		if err != nil {
			return nil, err
		}
		buf := &bytes.Buffer{}
		encoder := avro.NewEncoderForSchema(a.schema, buf)
		err = encoder.Encode(e)
		return buf.Bytes(), err
	}
	return nil, errors.New("unsupported struct")
}

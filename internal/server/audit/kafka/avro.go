package kafka

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"

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
		event.CopyInto(&e)
		payload, err := event.PayloadToMap()
		if err != nil {
			return nil, err
		}

		for k, v := range payload {
			t := reflect.TypeOf(v)
			switch t.Kind() {
			case reflect.Map, reflect.Slice, reflect.Array, reflect.Struct:
				data, err := json.Marshal(v)
				if err != nil {
					return nil, err
				}
				payload[k] = string(data)
			}
		}

		e.Payload = payload
		buf := &bytes.Buffer{}
		encoder := avro.NewEncoderForSchema(a.schema, buf)
		err = encoder.Encode(e)
		return buf.Bytes(), err
	}
	return nil, errors.New("unsupported struct")
}

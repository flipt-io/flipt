package kafka

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/hamba/avro"
	"go.flipt.io/flipt/internal/server/audit"
	rpcaudit "go.flipt.io/flipt/rpc/flipt/audit"
)

func toAvro(v any) ([]byte, error) {
	if e, ok := v.(audit.Event); ok {
		if e.Payload != nil {
			//FIXME: this modifies the origin payload
			payloadString, err := json.Marshal(e.Payload)
			if err != nil {
				return nil, err
			}
			var payload map[string]any
			err = json.Unmarshal(payloadString, &payload)
			if err != nil {
				return nil, err
			}
			e.Payload = payload
		}
		buf := &bytes.Buffer{}
		enc, err := avro.NewEncoder(rpcaudit.AvroSchema, buf)
		if err != nil {
			return nil, err
		}
		err = enc.Encode(e)
		return buf.Bytes(), err
	}
	return nil, errors.New("unsupported struct")
}

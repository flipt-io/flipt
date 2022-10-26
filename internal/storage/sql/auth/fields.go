package auth

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"go.flipt.io/flipt/rpc/flipt/auth"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type timestamp struct {
	*timestamppb.Timestamp
}

func (t *timestamp) Scan(value interface{}) error {
	if v, ok := value.(time.Time); ok {
		val := timestamppb.New(v)
		if err := val.CheckValid(); err != nil {
			return err
		}

		t.Timestamp = val
	}

	return nil
}

func (t *timestamp) Value() (driver.Value, error) {
	return t.Timestamp.AsTime(), t.Timestamp.CheckValid()
}

type method auth.Method

func (m *method) Scan(v interface{}) error {
	var methodStr string
	switch b := v.(type) {
	case []byte:
		methodStr = string(b)
	case string:
		methodStr = b
	default:
		return fmt.Errorf("unexpected method type: %T", v)
	}

	*m = method(auth.Method_value[methodStr])

	return nil
}

func (m *method) Value() (driver.Value, error) {
	return auth.Method_name[int32(*m)], nil
}

type jsonField[T any] struct {
	t T
}

func (d *jsonField[T]) Scan(v any) error {
	var bytes []byte
	switch b := v.(type) {
	case []byte:
		bytes = b
	case string:
		bytes = []byte(b)
	default:
		return fmt.Errorf("unexpected type for data: %T", v)
	}

	return json.Unmarshal(bytes, &d.t)
}

func (d *jsonField[T]) Value() (driver.Value, error) {
	return json.Marshal(d.t)
}

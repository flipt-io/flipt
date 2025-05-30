package sql

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type Timestamp struct {
	*timestamppb.Timestamp
}

func (t *Timestamp) Scan(value interface{}) error {
	if v, ok := value.(time.Time); ok {
		val := timestamppb.New(v)
		if err := val.CheckValid(); err != nil {
			return err
		}

		t.Timestamp = val
	}

	return nil
}

func (t *Timestamp) Value() (driver.Value, error) {
	return t.Timestamp.AsTime(), t.Timestamp.CheckValid()
}

type NullableTimestamp Timestamp

func (t *NullableTimestamp) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	if v, ok := value.(time.Time); ok {
		val := timestamppb.New(v)
		if err := val.CheckValid(); err != nil {
			return err
		}

		t.Timestamp = val
	}

	return nil
}

func (t *NullableTimestamp) Value() (driver.Value, error) {
	if t.Timestamp == nil {
		return nil, nil
	}

	return t.Timestamp.AsTime(), t.Timestamp.CheckValid()
}

type JSONField[T any] struct {
	T T
}

func (f *JSONField[T]) Scan(v any) error {
	if v == nil {
		return nil
	}

	var bytes []byte
	switch b := v.(type) {
	case []byte:
		bytes = b
	case string:
		bytes = []byte(b)
	default:
		return fmt.Errorf("unexpected type for data: %T", v)
	}

	return json.Unmarshal(bytes, &f.T)
}

func (f *JSONField[T]) Value() (driver.Value, error) {
	v, err := json.Marshal(f.T)
	// INFO: This could potentially be a breaking change or introduce bugs.
	// Postgres stores data differently when using prepared statements versus the simple protocol.
	// If a byte array is returned, the simple protocol stores it as hex-encoded data.
	// For example, `{"user":"alice","role":"admin"}` would be stored as `\x7b2275736572223a22616c696365222c22726f6c65223a2261646d696e227d`.
	// To avoid this inconsistency, a string is returned instead.
	return string(v), err
}

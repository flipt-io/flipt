package common

import (
	"database/sql/driver"
	"time"

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

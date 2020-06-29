package common

import (
	"database/sql/driver"
	"time"

	"github.com/golang/protobuf/ptypes"

	proto "github.com/golang/protobuf/ptypes/timestamp"
)

type timestamp struct {
	*proto.Timestamp
}

func (t *timestamp) Scan(value interface{}) error {
	if v, ok := value.(time.Time); ok {
		val, err := ptypes.TimestampProto(v)
		if err != nil {
			return err
		}

		t.Timestamp = val
	}

	return nil
}

func (t *timestamp) Value() (driver.Value, error) {
	return ptypes.Timestamp(t.Timestamp)
}

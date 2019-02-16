package storage

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
	v, ok := value.(time.Time)
	if ok {
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

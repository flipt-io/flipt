package sql

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestTimestamp_Scan(t *testing.T) {
	ts := Timestamp{}

	rt, _ := time.Parse(time.RFC3339, "1970-01-01T00:00:00Z")
	err := ts.Scan(rt)

	assert.NoError(t, err)
}

func TestTimestamp_Value(t *testing.T) {
	ts := Timestamp{
		Timestamp: &timestamppb.Timestamp{
			Seconds: 0,
		},
	}

	v, err := ts.Value()
	require.NoError(t, err)

	if tv, ok := v.(time.Time); ok {
		rt, _ := time.Parse(time.RFC3339, "1970-01-01T00:00:00Z")
		assert.Equal(t, rt, tv)
	}
}

func TestNullableTimestamp_Scan(t *testing.T) {
	nts := NullableTimestamp{}

	err := nts.Scan(nil)
	require.NoError(t, err)

	rt, _ := time.Parse(time.RFC3339, "1970-01-01T00:00:00Z")
	err = nts.Scan(rt)
	require.NoError(t, err)
}

func TestNullableTimestamp_Value(t *testing.T) {
	nts := NullableTimestamp{
		Timestamp: nil,
	}

	v, err := nts.Value()
	require.NoError(t, err)
	assert.Nil(t, v)

	nts = NullableTimestamp{
		Timestamp: &timestamppb.Timestamp{
			Seconds: 0,
		},
	}

	v, err = nts.Value()
	require.NoError(t, err)

	if tv, ok := v.(time.Time); ok {
		rt, _ := time.Parse(time.RFC3339, "1970-01-01T00:00:00Z")
		assert.Equal(t, rt, tv)
	}
}

func TestJSONField_Scan(t *testing.T) {
	jf := JSONField[map[string]string]{}

	err := jf.Scan(`{"hello": "world"}`)
	require.NoError(t, err)

	err = jf.Scan([]byte(`{"hello": "world"}`))
	assert.NoError(t, err)
}

func TestJSONField_Value(t *testing.T) {
	jf := JSONField[map[string]string]{
		T: map[string]string{
			"hello": "world",
		},
	}

	b, err := jf.Value()
	require.NoError(t, err)

	if b, ok := b.([]byte); ok {
		assert.Equal(t, `{"hello":"world"}`, string(b))
	}
}

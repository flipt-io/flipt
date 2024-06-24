package kafka

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/server/audit"
	"go.flipt.io/flipt/rpc/flipt"
)

func TestEncoding(t *testing.T) {
	tests := []struct {
		name string
		f    func(any) ([]byte, error)
	}{
		{"protobuf", toProtobuf},
		{"avro", toAvro},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := flipt.NewRequest(flipt.ResourceFlag, flipt.ActionCreate, flipt.WithSubject(flipt.SubjectRule))
			e := audit.NewEvent(
				r,
				&audit.Actor{
					Authentication: "token",
					IP:             "127.0.0.1",
				},
				&audit.Flag{
					Key:          "this-flag",
					Name:         "this-flag",
					Description:  "this description",
					Enabled:      false,
					NamespaceKey: "default",
				},
			)

			b, err := tt.f(*e)
			require.NoError(t, err)
			require.NotEmpty(t, b)
		})
	}
}

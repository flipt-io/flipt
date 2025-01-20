package kafka

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/server/audit"
	"go.flipt.io/flipt/rpc/flipt"
)

func TestEncoding(t *testing.T) {
	tests := []struct {
		name    string
		encoder encodingFn
	}{
		{"protobuf", newProtobufEncoder().Encode},
		{"avro", newAvroEncoder().Encode},
	}

	dataset := []struct {
		name    string
		payload any
	}{
		{
			"flag",
			audit.NewFlag(&flipt.Flag{
				Key:          "this-flag",
				Name:         "this-flag",
				Description:  "this description",
				Enabled:      false,
				NamespaceKey: "default",
			}),
		},
		{
			"rollout-threshold",
			audit.NewRollout(&flipt.Rollout{
				Description:  "this description",
				NamespaceKey: "default",
				Rule: &flipt.Rollout_Threshold{
					Threshold: &flipt.RolloutThreshold{},
				},
			}),
		},
		{
			"rollout-segment",
			audit.NewRollout(&flipt.Rollout{
				Description:  "this description",
				NamespaceKey: "default",
				Rule: &flipt.Rollout_Segment{
					Segment: &flipt.RolloutSegment{
						SegmentOperator: flipt.SegmentOperator_AND_SEGMENT_OPERATOR,
						SegmentKeys:     []string{"segment-key", "some"},
					},
				},
			}),
		},
		{
			"auth",
			map[string]string{
				"method": "github",
				"org":    "someone",
			},
		},
		{
			"nil",
			nil,
		},
		{
			"segment",
			audit.NewSegment(&flipt.Segment{
				NamespaceKey: "default",
				Key:          "segment-key",
				Constraints:  []*flipt.Constraint{{Id: "constraint-id"}},
			}),
		},
	}

	for _, tt := range tests {
		for _, ds := range dataset {
			t.Run(tt.name+"/"+ds.name, func(t *testing.T) {
				r := flipt.NewRequest(flipt.ResourceFlag, flipt.ActionCreate, flipt.WithSubject(flipt.SubjectRule))
				e := audit.NewEvent(
					r,
					&audit.Actor{
						Authentication: "token",
						IP:             "127.0.0.1",
					},
					ds.payload,
				)

				b, err := tt.encoder(*e)
				require.NoError(t, err)
				require.NotEmpty(t, b)
			})
		}
	}
}

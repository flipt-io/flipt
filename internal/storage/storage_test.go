package storage

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/rpc/flipt/core"
)

func TestNewEvaluationSegment_WrapsConstraintError(t *testing.T) {
	tests := []struct {
		name       string
		constraint EvaluationConstraint
		wantErr    string
	}{
		{
			name: "invalid JSON for string set",
			constraint: EvaluationConstraint{
				Property: "tenant",
				Operator: flipt.OpIsOneOf,
				Type:     core.ComparisonType_STRING_COMPARISON_TYPE,
				Value:    `["bar", "baz"`,
			},
			wantErr: `segment "mysegment": constraint "tenant": parsing string set`,
		},
		{
			name: "invalid datetime",
			constraint: EvaluationConstraint{
				Property: "created_at",
				Operator: flipt.OpGTE,
				Type:     core.ComparisonType_DATETIME_COMPARISON_TYPE,
				Value:    "not a date",
			},
			wantErr: `segment "mysegment": constraint "created_at": parsing datetime from "not a date"`,
		},
		{
			name: "invalid JSON for number set",
			constraint: EvaluationConstraint{
				Property: "age",
				Operator: flipt.OpIsOneOf,
				Type:     core.ComparisonType_NUMBER_COMPARISON_TYPE,
				Value:    "[1, 1 2 2]",
			},
			wantErr: `segment "mysegment": constraint "age": parsing number set`,
		},
		{
			name: "invalid number for scalar operator",
			constraint: EvaluationConstraint{
				Property: "age",
				Operator: flipt.OpGTE,
				Type:     core.ComparisonType_NUMBER_COMPARISON_TYPE,
				Value:    "not a number",
			},
			wantErr: `segment "mysegment": constraint "age": parsing number`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewEvaluationSegment("mysegment", core.MatchType_ALL_MATCH_TYPE, []EvaluationConstraint{tt.constraint})
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

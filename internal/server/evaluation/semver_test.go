package evaluation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	errs "go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/rpc/flipt/core"
)

func Test_matchesSemver(t *testing.T) {
	tests := []struct {
		name       string
		constraint storage.EvaluationConstraint
		value      string
		wantMatch  bool
		wantErr    bool
	}{
		{
			name: "present",
			constraint: storage.EvaluationConstraint{
				Property: "version",
				Operator: "present",
			},
			value:     "1.0.0",
			wantMatch: true,
		},
		{
			name: "negative present",
			constraint: storage.EvaluationConstraint{
				Property: "version",
				Operator: "present",
			},
			value:     "",
			wantMatch: false,
		},
		{
			name: "negative present whitespace",
			constraint: storage.EvaluationConstraint{
				Property: "version",
				Operator: "present",
			},
			value:     "   ",
			wantMatch: false,
		},
		{
			name: "not present",
			constraint: storage.EvaluationConstraint{
				Property: "version",
				Operator: "notpresent",
			},
			value:     "",
			wantMatch: true,
		},
		{
			name: "not present whitespace",
			constraint: storage.EvaluationConstraint{
				Property: "version",
				Operator: "notpresent",
			},
			value:     "   ",
			wantMatch: true,
		},
		{
			name: "negative not present",
			constraint: storage.EvaluationConstraint{
				Property: "version",
				Operator: "notpresent",
			},
			value:     "1.0.0",
			wantMatch: false,
		},
		{
			name: "not a semver constraint value",
			constraint: storage.EvaluationConstraint{
				Property: "version",
				Operator: "eq",
				Value:    "invalid-semver",
			},
			value:   "1.0.0",
			wantErr: true,
		},
		{
			name: "not a semver context value",
			constraint: storage.EvaluationConstraint{
				Property: "version",
				Operator: "eq",
				Value:    "1.0.0",
			},
			value:   "invalid-semver",
			wantErr: true,
		},
		{
			name: "empty context value",
			constraint: storage.EvaluationConstraint{
				Property: "version",
				Operator: "eq",
				Value:    "1.0.0",
			},
			value:     "",
			wantMatch: false,
		},
		{
			name: "unknown operator",
			constraint: storage.EvaluationConstraint{
				Property: "version",
				Operator: "unknown",
				Value:    "1.0.0",
			},
			value:     "1.0.0",
			wantMatch: false,
		},
		{
			name: "eq exact",
			constraint: storage.EvaluationConstraint{
				Property: "version",
				Operator: "eq",
				Value:    "1.2.3",
			},
			value:     "1.2.3",
			wantMatch: true,
		},
		{
			name: "eq with v prefix",
			constraint: storage.EvaluationConstraint{
				Property: "version",
				Operator: "eq",
				Value:    "v1.2.3",
			},
			value:     "1.2.3",
			wantMatch: true,
		},
		{
			name: "eq build metadata ignored",
			constraint: storage.EvaluationConstraint{
				Property: "version",
				Operator: "eq",
				Value:    "1.2.3+build1",
			},
			value:     "1.2.3+build2",
			wantMatch: true,
		},
		{
			name: "negative eq",
			constraint: storage.EvaluationConstraint{
				Property: "version",
				Operator: "eq",
				Value:    "1.2.3",
			},
			value:     "1.2.4",
			wantMatch: false,
		},
		{
			name: "neq",
			constraint: storage.EvaluationConstraint{
				Property: "version",
				Operator: "neq",
				Value:    "1.2.3",
			},
			value:     "1.2.4",
			wantMatch: true,
		},
		{
			name: "negative neq",
			constraint: storage.EvaluationConstraint{
				Property: "version",
				Operator: "neq",
				Value:    "1.2.3",
			},
			value:     "1.2.3",
			wantMatch: false,
		},
		{
			name: "lt patch",
			constraint: storage.EvaluationConstraint{
				Property: "version",
				Operator: "lt",
				Value:    "1.2.3",
			},
			value:     "1.2.2",
			wantMatch: true,
		},
		{
			name: "lt minor",
			constraint: storage.EvaluationConstraint{
				Property: "version",
				Operator: "lt",
				Value:    "1.3.0",
			},
			value:     "1.2.9",
			wantMatch: true,
		},
		{
			name: "lt major",
			constraint: storage.EvaluationConstraint{
				Property: "version",
				Operator: "lt",
				Value:    "2.0.0",
			},
			value:     "1.9.9",
			wantMatch: true,
		},
		{
			name: "lt prerelease to release",
			constraint: storage.EvaluationConstraint{
				Property: "version",
				Operator: "lt",
				Value:    "1.0.0",
			},
			value:     "1.0.0-alpha",
			wantMatch: true,
		},
		{
			name: "lt prerelease alpha to beta",
			constraint: storage.EvaluationConstraint{
				Property: "version",
				Operator: "lt",
				Value:    "1.0.0-beta",
			},
			value:     "1.0.0-alpha",
			wantMatch: true,
		},
		{
			name: "lt prerelease numeric identifier",
			constraint: storage.EvaluationConstraint{
				Property: "version",
				Operator: "lt",
				Value:    "1.0.0-alpha.2",
			},
			value:     "1.0.0-alpha.1",
			wantMatch: true,
		},
		{
			name: "negative lt equal",
			constraint: storage.EvaluationConstraint{
				Property: "version",
				Operator: "lt",
				Value:    "1.2.3",
			},
			value:     "1.2.3",
			wantMatch: false,
		},
		{
			name: "negative lt greater",
			constraint: storage.EvaluationConstraint{
				Property: "version",
				Operator: "lt",
				Value:    "1.2.3",
			},
			value:     "1.2.4",
			wantMatch: false,
		},
		{
			name: "lte less",
			constraint: storage.EvaluationConstraint{
				Property: "version",
				Operator: "lte",
				Value:    "1.2.3",
			},
			value:     "1.2.2",
			wantMatch: true,
		},
		{
			name: "lte equal",
			constraint: storage.EvaluationConstraint{
				Property: "version",
				Operator: "lte",
				Value:    "1.2.3",
			},
			value:     "1.2.3",
			wantMatch: true,
		},
		{
			name: "negative lte",
			constraint: storage.EvaluationConstraint{
				Property: "version",
				Operator: "lte",
				Value:    "1.2.3",
			},
			value:     "1.2.4",
			wantMatch: false,
		},
		{
			name: "gt patch",
			constraint: storage.EvaluationConstraint{
				Property: "version",
				Operator: "gt",
				Value:    "1.2.3",
			},
			value:     "1.2.4",
			wantMatch: true,
		},
		{
			name: "gt minor",
			constraint: storage.EvaluationConstraint{
				Property: "version",
				Operator: "gt",
				Value:    "1.2.9",
			},
			value:     "1.3.0",
			wantMatch: true,
		},
		{
			name: "gt major",
			constraint: storage.EvaluationConstraint{
				Property: "version",
				Operator: "gt",
				Value:    "1.9.9",
			},
			value:     "2.0.0",
			wantMatch: true,
		},
		{
			name: "gt release over prerelease",
			constraint: storage.EvaluationConstraint{
				Property: "version",
				Operator: "gt",
				Value:    "1.0.0-rc.1",
			},
			value:     "1.0.0",
			wantMatch: true,
		},
		{
			name: "negative gt equal",
			constraint: storage.EvaluationConstraint{
				Property: "version",
				Operator: "gt",
				Value:    "1.2.3",
			},
			value:     "1.2.3",
			wantMatch: false,
		},
		{
			name: "negative gt less",
			constraint: storage.EvaluationConstraint{
				Property: "version",
				Operator: "gt",
				Value:    "1.2.3",
			},
			value:     "1.2.2",
			wantMatch: false,
		},
		{
			name: "gte greater",
			constraint: storage.EvaluationConstraint{
				Property: "version",
				Operator: "gte",
				Value:    "1.2.3",
			},
			value:     "1.2.4",
			wantMatch: true,
		},
		{
			name: "gte equal",
			constraint: storage.EvaluationConstraint{
				Property: "version",
				Operator: "gte",
				Value:    "1.2.3",
			},
			value:     "1.2.3",
			wantMatch: true,
		},
		{
			name: "negative gte",
			constraint: storage.EvaluationConstraint{
				Property: "version",
				Operator: "gte",
				Value:    "1.2.3",
			},
			value:     "1.2.2",
			wantMatch: false,
		},
		{
			name: "isoneof match",
			constraint: storage.EvaluationConstraint{
				Property: "version",
				Operator: "isoneof",
				Value:    `["1.0.0", "1.2.3", "2.0.0"]`,
			},
			value:     "1.2.3",
			wantMatch: true,
		},
		{
			name: "isoneof negative match",
			constraint: storage.EvaluationConstraint{
				Property: "version",
				Operator: "isoneof",
				Value:    `["1.0.0", "2.0.0"]`,
			},
			value:     "1.2.3",
			wantMatch: false,
		},
		{
			name: "isoneof invalid json",
			constraint: storage.EvaluationConstraint{
				Property: "version",
				Operator: "isoneof",
				Value:    `["invalid json`,
			},
			value:   "1.2.3",
			wantErr: true,
		},
		{
			name: "isoneof invalid semver in array",
			constraint: storage.EvaluationConstraint{
				Property: "version",
				Operator: "isoneof",
				Value:    `["1.0.0", "not-a-semver"]`,
			},
			value:   "1.2.3",
			wantErr: true,
		},
		{
			name: "isnotoneof match",
			constraint: storage.EvaluationConstraint{
				Property: "version",
				Operator: "isnotoneof",
				Value:    `["1.0.0", "2.0.0"]`,
			},
			value:     "1.2.3",
			wantMatch: true,
		},
		{
			name: "isnotoneof negative match",
			constraint: storage.EvaluationConstraint{
				Property: "version",
				Operator: "isnotoneof",
				Value:    `["1.0.0", "1.2.3"]`,
			},
			value:     "1.2.3",
			wantMatch: false,
		},
		{
			name: "isnotoneof invalid json",
			constraint: storage.EvaluationConstraint{
				Property: "version",
				Operator: "isnotoneof",
				Value:    `["invalid json`,
			},
			value:   "1.2.3",
			wantErr: true,
		},
		{
			name: "isnotoneof invalid semver in array",
			constraint: storage.EvaluationConstraint{
				Property: "version",
				Operator: "isnotoneof",
				Value:    `["not-a-semver"]`,
			},
			value:   "1.2.3",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			match, err := matchesSemver(tt.constraint, tt.value)

			if tt.wantErr {
				require.Error(t, err)
				var ierr errs.ErrInvalid
				assert.ErrorAs(t, err, &ierr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantMatch, match)
		})
	}
}

func Test_matchesSemverWithPreparedSets(t *testing.T) {
	tests := []struct {
		name       string
		constraint storage.EvaluationConstraint
		value      string
		wantMatch  bool
		wantErr    bool
	}{
		{
			name: "isoneof match",
			constraint: storage.EvaluationConstraint{
				Type:     core.ComparisonType_SEMVER_COMPARISON_TYPE,
				Property: "version",
				Operator: flipt.OpIsOneOf,
				Value:    "[\"1.0.0\", \"2.0.0\"]",
			},
			value:     "1.0.0",
			wantMatch: true,
		},
		{
			name: "isoneof no match",
			constraint: storage.EvaluationConstraint{
				Type:     core.ComparisonType_SEMVER_COMPARISON_TYPE,
				Property: "version",
				Operator: flipt.OpIsOneOf,
				Value:    "[\"1.0.0\", \"2.0.0\"]",
			},
			value:     "3.0.0",
			wantMatch: false,
		},
		{
			name: "isnotoneof match",
			constraint: storage.EvaluationConstraint{
				Type:     core.ComparisonType_SEMVER_COMPARISON_TYPE,
				Property: "version",
				Operator: flipt.OpIsNotOneOf,
				Value:    "[\"1.0.0\", \"2.0.0\"]",
			},
			value:     "3.0.0",
			wantMatch: true,
		},
		{
			name: "isnotoneof no match",
			constraint: storage.EvaluationConstraint{
				Type:     core.ComparisonType_SEMVER_COMPARISON_TYPE,
				Property: "version",
				Operator: flipt.OpIsNotOneOf,
				Value:    "[\"1.0.0\", \"2.0.0\"]",
			},
			value:     "2.0.0",
			wantMatch: false,
		},
		{
			name: "eq with v prefix prepared",
			constraint: storage.EvaluationConstraint{
				Type:     core.ComparisonType_SEMVER_COMPARISON_TYPE,
				Property: "version",
				Operator: flipt.OpEQ,
				Value:    "v1.2.3",
			},
			value:     "1.2.3",
			wantMatch: true,
		},
		{
			name: "gte with prerelease prepared",
			constraint: storage.EvaluationConstraint{
				Type:     core.ComparisonType_SEMVER_COMPARISON_TYPE,
				Property: "version",
				Operator: flipt.OpGTE,
				Value:    "1.0.0-alpha",
			},
			value:     "1.0.0-beta",
			wantMatch: true,
		},
		{
			name: "lt with prerelease prepared",
			constraint: storage.EvaluationConstraint{
				Type:     core.ComparisonType_SEMVER_COMPARISON_TYPE,
				Property: "version",
				Operator: flipt.OpLT,
				Value:    "1.0.0",
			},
			value:     "1.0.0-rc.1",
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			constraint := tt.constraint
			require.NoError(t, constraint.PrepareForEvaluation())
			match, err := matchesSemver(constraint, tt.value)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantMatch, match)
		})
	}
}

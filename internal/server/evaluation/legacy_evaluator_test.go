package evaluation

import (
	"context"
	"errors"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	ferrors "go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/rpc/flipt/evaluation"
	"go.uber.org/zap/zaptest"
)

func Test_matchesString(t *testing.T) {
	tests := []struct {
		name       string
		constraint storage.EvaluationConstraint
		value      string
		wantMatch  bool
	}{
		{
			name: "eq",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "eq",
				Value:    "bar",
			},
			value:     "bar",
			wantMatch: true,
		},
		{
			name: "negative eq",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "eq",
				Value:    "bar",
			},
			value: "baz",
		},
		{
			name: "neq",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "neq",
				Value:    "bar",
			},
			value:     "baz",
			wantMatch: true,
		},
		{
			name: "negative neq",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "neq",
				Value:    "bar",
			},
			value: "bar",
		},
		{
			name: "empty",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "empty",
			},
			value:     " ",
			wantMatch: true,
		},
		{
			name: "negative empty",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "empty",
			},
			value: "bar",
		},
		{
			name: "not empty",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "notempty",
			},
			value:     "bar",
			wantMatch: true,
		},
		{
			name: "negative not empty",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "notempty",
			},
			value: "",
		},
		{
			name: "unknown operator",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "foo",
				Value:    "bar",
			},
			value: "bar",
		},
		{
			name: "prefix",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "prefix",
				Value:    "ba",
			},
			value:     "bar",
			wantMatch: true,
		},
		{
			name: "negative prefix",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "prefix",
				Value:    "bar",
			},
			value: "nope",
		},
		{
			name: "suffix",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "suffix",
				Value:    "ar",
			},
			value:     "bar",
			wantMatch: true,
		},
		{
			name: "negative suffix",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "suffix",
				Value:    "bar",
			},
			value: "nope",
		},
		{
			name: "is one of",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "isoneof",
				Value:    "[\"bar\", \"baz\"]",
			},
			value:     "baz",
			wantMatch: true,
		},
		{
			name: "negative is one of",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "isoneof",
				Value:    "[\"bar\", \"baz\"]",
			},
			value: "nope",
		},
		{
			name: "negative is one of (invalid json)",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "isoneof",
				Value:    "[\"bar\", \"baz\"",
			},
			value: "bar",
		},
		{
			name: "negative is one of (non-string values)",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "isoneof",
				Value:    "[\"bar\", 5]",
			},
			value: "bar",
		},
		{
			name: "is not one of",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "isnotoneof",
				Value:    "[\"bar\", \"baz\"]",
			},
			value: "baz",
		},
		{
			name: "negative is not one of",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "isnotoneof",
				Value:    "[\"bar\", \"baz\"]",
			},
			value:     "nope",
			wantMatch: true,
		},
	}
	for _, tt := range tests {
		var (
			constraint = tt.constraint
			value      = tt.value
			wantMatch  = tt.wantMatch
		)

		t.Run(tt.name, func(t *testing.T) {
			match := matchesString(constraint, value)
			assert.Equal(t, wantMatch, match)
		})
	}
}

func Test_matchesNumber(t *testing.T) {
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
				Property: "foo",
				Operator: "present",
			},
			value:     "1",
			wantMatch: true,
		},
		{
			name: "negative present",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "present",
			},
		},
		{
			name: "not present",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "notpresent",
			},
			wantMatch: true,
		},
		{
			name: "negative notpresent",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "notpresent",
			},
			value: "1",
		},
		{
			name: "NAN constraint value",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "eq",
				Value:    "bar",
			},
			value:   "5",
			wantErr: true,
		},
		{
			name: "NAN context value",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "eq",
				Value:    "5",
			},
			value:   "foo",
			wantErr: true,
		},
		{
			name: "eq",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "eq",
				Value:    "42.0",
			},
			value:     "42.0",
			wantMatch: true,
		},
		{
			name: "negative eq",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "eq",
				Value:    "42.0",
			},
			value: "50",
		},
		{
			name: "neq",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "neq",
				Value:    "42.0",
			},
			value:     "50",
			wantMatch: true,
		},
		{
			name: "negative neq",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "neq",
				Value:    "42.0",
			},
			value: "42.0",
		},
		{
			name: "lt",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "lt",
				Value:    "42.0",
			},
			value:     "8",
			wantMatch: true,
		},
		{
			name: "negative lt",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "lt",
				Value:    "42.0",
			},
			value: "50",
		},
		{
			name: "lte",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "lte",
				Value:    "42.0",
			},
			value:     "42.0",
			wantMatch: true,
		},
		{
			name: "negative lte",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "lte",
				Value:    "42.0",
			},
			value: "102.0",
		},
		{
			name: "gt",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "gt",
				Value:    "10.11",
			},
			value:     "10.12",
			wantMatch: true,
		},
		{
			name: "negative gt",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "gt",
				Value:    "10.11",
			},
			value: "1",
		},
		{
			name: "gte",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "gte",
				Value:    "10.11",
			},
			value:     "10.11",
			wantMatch: true,
		},
		{
			name: "negative gte",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "gte",
				Value:    "10.11",
			},
			value: "0.11",
		},
		{
			name: "empty value",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "eq",
				Value:    "0.11",
			},
		},
		{
			name: "unknown operator",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "foo",
				Value:    "0.11",
			},
			value: "0.11",
		},
		{
			name: "negative suffix empty value",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "suffix",
				Value:    "bar",
			},
		},
		{
			name: "is one of",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "isoneof",
				Value:    "[3, 3.14159, 4]",
			},
			value:     "3.14159",
			wantMatch: true,
		},
		{
			name: "negative is one of",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "isoneof",
				Value:    "[5, 3.14159, 4]",
			},
			value: "9",
		},
		{
			name: "negative is one of (non-number values)",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "isoneof",
				Value:    "[5, \"str\"]",
			},
			value:   "5",
			wantErr: true,
		},
		{
			name: "is not one of",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "isnotoneof",
				Value:    "[5, 3.14159, 4]",
			},
			value:     "3",
			wantMatch: true,
		},
		{
			name: "negative is not one of",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "isnotoneof",
				Value:    "[5, 3.14159, 4]",
			},
			value:     "3.14159",
			wantMatch: false,
		},
		{
			name: "negative is not one of (invalid json)",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "isnotoneof",
				Value:    "[5, 6",
			},
			value:   "5",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		var (
			constraint = tt.constraint
			value      = tt.value
			wantMatch  = tt.wantMatch
			wantErr    = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			match, err := matchesNumber(constraint, value)

			if wantErr {
				require.Error(t, err)
				var ierr ferrors.ErrInvalid
				assert.ErrorAs(t, err, &ierr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, wantMatch, match)
		})
	}
}

func Test_matchesBool(t *testing.T) {
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
				Property: "foo",
				Operator: "present",
			},
			value:     "true",
			wantMatch: true,
		},
		{
			name: "negative present",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "present",
			},
		},
		{
			name: "not present",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "notpresent",
			},
			wantMatch: true,
		},
		{
			name: "negative notpresent",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "notpresent",
			},
			value: "true",
		},
		{
			name: "not a bool",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "true",
			},
			value:   "foo",
			wantErr: true,
		},
		{
			name: "is true",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "true",
			},
			value:     "true",
			wantMatch: true,
		},
		{
			name: "negative is true",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "true",
			},
			value: "false",
		},
		{
			name: "is false",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "false",
			},
			value:     "false",
			wantMatch: true,
		},
		{
			name: "negative is false",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "false",
			},
			value: "true",
		},
		{
			name: "empty value",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "false",
			},
		},
		{
			name: "unknown operator",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "foo",
			},
			value: "true",
		},
	}

	for _, tt := range tests {
		var (
			constraint = tt.constraint
			value      = tt.value
			wantMatch  = tt.wantMatch
			wantErr    = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			match, err := matchesBool(constraint, value)

			if wantErr {
				require.Error(t, err)
				var ierr ferrors.ErrInvalid
				assert.ErrorAs(t, err, &ierr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, wantMatch, match)
		})
	}
}

func Test_matchesDateTime(t *testing.T) {
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
				Property: "foo",
				Operator: "present",
			},
			value:     "2006-01-02T15:04:05Z",
			wantMatch: true,
		},
		{
			name: "negative present",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "present",
			},
		},
		{
			name: "not present",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "notpresent",
			},
			wantMatch: true,
		},
		{
			name: "negative notpresent",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "notpresent",
			},
			value: "2006-01-02T15:04:05Z",
		},
		{
			name: "not a datetime constraint value",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "eq",
				Value:    "bar",
			},
			value:   "2006-01-02T15:04:05Z",
			wantErr: true,
		},
		{
			name: "not a datetime context value",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "eq",
				Value:    "2006-01-02T15:04:05Z",
			},
			value:   "foo",
			wantErr: true,
		},
		{
			name: "eq",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "eq",
				Value:    "2006-01-02T15:04:05Z",
			},
			value:     "2006-01-02T15:04:05Z",
			wantMatch: true,
		},
		{
			name: "eq date only",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "eq",
				Value:    "2006-01-02",
			},
			value:     "2006-01-02",
			wantMatch: true,
		},
		{
			name: "negative eq",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "eq",
				Value:    "2006-01-02T15:04:05Z",
			},
			value: "2007-01-02T15:04:05Z",
		},
		{
			name: "neq",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "neq",
				Value:    "2006-01-02T15:04:05Z",
			},
			value:     "2007-01-02T15:04:05Z",
			wantMatch: true,
		},
		{
			name: "negative neq",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "neq",
				Value:    "2006-01-02T15:04:05Z",
			},
			value: "2006-01-02T15:04:05Z",
		},
		{
			name: "negative neq date only",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "neq",
				Value:    "2006-01-02",
			},
			value: "2006-01-02",
		},
		{
			name: "lt",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "lt",
				Value:    "2006-01-02T15:04:05Z",
			},
			value:     "2005-01-02T15:04:05Z",
			wantMatch: true,
		},
		{
			name: "negative lt",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "lt",
				Value:    "2005-01-02T15:04:05Z",
			},
			value: "2006-01-02T15:04:05Z",
		},
		{
			name: "lte",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "lte",
				Value:    "2006-01-02T15:04:05Z",
			},
			value:     "2006-01-02T15:04:05Z",
			wantMatch: true,
		},
		{
			name: "negative lte",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "lte",
				Value:    "2006-01-02T15:04:05Z",
			},
			value: "2007-01-02T15:04:05Z",
		},
		{
			name: "gt",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "gt",
				Value:    "2006-01-02T15:04:05Z",
			},
			value:     "2007-01-02T15:04:05Z",
			wantMatch: true,
		},
		{
			name: "negative gt",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "gt",
				Value:    "2007-01-02T15:04:05Z",
			},
			value: "2006-01-02T15:04:05Z",
		},
		{
			name: "gte",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "gte",
				Value:    "2006-01-02T15:04:05Z",
			},
			value:     "2006-01-02T15:04:05Z",
			wantMatch: true,
		},
		{
			name: "negative gte",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "gte",
				Value:    "2006-01-02T15:04:05Z",
			},
			value: "2005-01-02T15:04:05Z",
		},
		{
			name: "empty value",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "eq",
				Value:    "2006-01-02T15:04:05Z",
			},
		},
		{
			name: "unknown operator",
			constraint: storage.EvaluationConstraint{
				Property: "foo",
				Operator: "foo",
				Value:    "2006-01-02T15:04:05Z",
			},
			value: "2006-01-02T15:04:05Z",
		},
	}

	for _, tt := range tests {
		var (
			constraint = tt.constraint
			value      = tt.value
			wantMatch  = tt.wantMatch
			wantErr    = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			match, err := matchesDateTime(constraint, value)

			if wantErr {
				require.Error(t, err)
				var ierr ferrors.ErrInvalid
				assert.ErrorAs(t, err, &ierr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, wantMatch, match)
		})
	}
}

var (
	enabledFlag = &flipt.Flag{
		Key:     "foo",
		Enabled: true,
	}
	enabledFlagWithDefaultVariant = &flipt.Flag{
		Key:     "foo",
		Enabled: true,
		DefaultVariant: &flipt.Variant{
			Key:        "bar",
			Attachment: `{ "bar": "baz" }`,
		},
	}
	disabledFlag = &flipt.Flag{
		Key:     "foo",
		Enabled: false,
	}
	disabledFlagWithDefaultVariant = &flipt.Flag{
		Key:     "foo",
		Enabled: false,
		DefaultVariant: &flipt.Variant{
			Key:        "bar",
			Attachment: `{ "bar": "baz" }`,
		},
	}
)

func TestEvaluator_FlagDisabled(t *testing.T) {
	var (
		store  = &evaluationStoreMock{}
		logger = zaptest.NewLogger(t)
		s      = NewEvaluator(logger, store)
	)

	resp, err := s.Evaluate(context.TODO(), disabledFlag, &evaluation.EvaluationRequest{
		EntityId: "1",
		FlagKey:  "foo",
		Context: map[string]string{
			"bar": "boz",
		},
	})

	require.NoError(t, err)
	assert.False(t, resp.Match)
	assert.Empty(t, resp.Value)
	assert.Empty(t, resp.Attachment)
	assert.Equal(t, flipt.EvaluationReason_FLAG_DISABLED_EVALUATION_REASON, resp.Reason)
}

func TestEvaluator_FlagDisabled_DefaultVariant(t *testing.T) {
	var (
		store  = &evaluationStoreMock{}
		logger = zaptest.NewLogger(t)
		s      = NewEvaluator(logger, store)
	)

	resp, err := s.Evaluate(context.TODO(), disabledFlagWithDefaultVariant, &evaluation.EvaluationRequest{
		EntityId: "1",
		FlagKey:  "foo",
		Context: map[string]string{
			"bar": "boz",
		},
	})

	require.NoError(t, err)
	assert.False(t, resp.Match)
	assert.Equal(t, "bar", resp.Value)
	assert.Equal(t, `{ "bar": "baz" }`, resp.Attachment)
	assert.Equal(t, flipt.EvaluationReason_FLAG_DISABLED_EVALUATION_REASON, resp.Reason)
}

func TestEvaluator_NonVariantFlag(t *testing.T) {
	var (
		store  = &evaluationStoreMock{}
		logger = zaptest.NewLogger(t)
		s      = NewEvaluator(logger, store)
	)

	resp, err := s.Evaluate(context.TODO(), &flipt.Flag{
		Key:     "foo",
		Enabled: true,
		Type:    flipt.FlagType_BOOLEAN_FLAG_TYPE,
	}, &evaluation.EvaluationRequest{
		EntityId: "1",
		FlagKey:  "foo",
		Context: map[string]string{
			"bar": "boz",
		},
	})

	require.EqualError(t, err, "flag type BOOLEAN_FLAG_TYPE invalid")
	assert.False(t, resp.Match)
	assert.Equal(t, flipt.EvaluationReason_ERROR_EVALUATION_REASON, resp.Reason)
}

func TestEvaluator_FlagNoRules(t *testing.T) {
	var (
		store  = &evaluationStoreMock{}
		logger = zaptest.NewLogger(t)
		s      = NewEvaluator(logger, store)
	)

	store.On("GetEvaluationRules", mock.Anything, storage.NewResource("", "foo")).Return([]*storage.EvaluationRule{}, nil)

	resp, err := s.Evaluate(context.TODO(), enabledFlag, &evaluation.EvaluationRequest{
		EntityId: "1",
		FlagKey:  "foo",
		Context: map[string]string{
			"bar": "boz",
		},
	})

	require.NoError(t, err)
	assert.False(t, resp.Match)
	assert.Equal(t, flipt.EvaluationReason_UNKNOWN_EVALUATION_REASON, resp.Reason)
}

func TestEvaluator_FlagNoRules_DefaultVariant(t *testing.T) {
	var (
		store  = &evaluationStoreMock{}
		logger = zaptest.NewLogger(t)
		s      = NewEvaluator(logger, store)
	)

	tests := []struct {
		key   string
		rules []*storage.EvaluationRule
	}{
		{key: "foo", rules: []*storage.EvaluationRule{}},
		{key: "bar", rules: []*storage.EvaluationRule{
			{
				ID:              "entry",
				NamespaceKey:    enabledFlagWithDefaultVariant.NamespaceKey,
				SegmentOperator: flipt.SegmentOperator_AND_SEGMENT_OPERATOR,
				Segments: map[string]*storage.EvaluationSegment{
					"entity": {
						SegmentKey: "entity",
						MatchType:  flipt.MatchType_ALL_MATCH_TYPE,
						Constraints: []storage.EvaluationConstraint{
							{
								ID:       "entity",
								Type:     flipt.ComparisonType_ENTITY_ID_COMPARISON_TYPE,
								Operator: "eq",
								Value:    "some",
							},
						},
					},
				},
			},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			store.On("GetEvaluationRules", mock.Anything, storage.NewResource("", tt.key)).Return(tt.rules, nil)

			resp, err := s.Evaluate(context.TODO(), enabledFlagWithDefaultVariant, &evaluation.EvaluationRequest{
				EntityId: "1",
				FlagKey:  tt.key,
				Context: map[string]string{
					"bar": "boz",
				},
			})

			require.NoError(t, err)
			assert.False(t, resp.Match)
			assert.Equal(t, "bar", resp.Value)
			assert.Equal(t, `{ "bar": "baz" }`, resp.Attachment)
			assert.Equal(t, flipt.EvaluationReason_DEFAULT_EVALUATION_REASON, resp.Reason)
		})
	}
}

func TestEvaluator_ErrorGettingRules(t *testing.T) {
	var (
		store  = &evaluationStoreMock{}
		logger = zaptest.NewLogger(t)
		s      = NewEvaluator(logger, store)
	)

	store.On("GetEvaluationRules", mock.Anything, storage.NewResource("", "foo")).Return([]*storage.EvaluationRule{}, errors.New("error getting rules!"))

	resp, err := s.Evaluate(context.TODO(), enabledFlag, &evaluation.EvaluationRequest{
		EntityId: "1",
		FlagKey:  "foo",
		Context: map[string]string{
			"bar": "boz",
		},
	})

	require.Error(t, err)
	assert.False(t, resp.Match)
	assert.Equal(t, flipt.EvaluationReason_ERROR_EVALUATION_REASON, resp.Reason)
}

func TestEvaluator_RulesOutOfOrder(t *testing.T) {
	var (
		store  = &evaluationStoreMock{}
		logger = zaptest.NewLogger(t)
		s      = NewEvaluator(logger, store)
	)

	store.On("GetEvaluationRules", mock.Anything, storage.NewResource("", "foo")).Return(
		[]*storage.EvaluationRule{
			{
				ID:      "1",
				FlagKey: "foo",
				Rank:    1,
				Segments: map[string]*storage.EvaluationSegment{
					"bar": {
						SegmentKey: "bar",
						MatchType:  flipt.MatchType_ALL_MATCH_TYPE,
						Constraints: []storage.EvaluationConstraint{
							{
								ID:       "2",
								Type:     flipt.ComparisonType_STRING_COMPARISON_TYPE,
								Property: "bar",
								Operator: flipt.OpEQ,
								Value:    "baz",
							},
						},
					},
				},
			},
			{
				ID:      "2",
				FlagKey: "foo",
				Rank:    0,
				Segments: map[string]*storage.EvaluationSegment{
					"bar": {
						SegmentKey: "bar",
						MatchType:  flipt.MatchType_ALL_MATCH_TYPE,
						Constraints: []storage.EvaluationConstraint{
							{
								ID:       "2",
								Type:     flipt.ComparisonType_STRING_COMPARISON_TYPE,
								Property: "bar",
								Operator: flipt.OpEQ,
								Value:    "baz",
							},
						},
					},
				},
			},
		}, nil)

	resp, err := s.Evaluate(context.TODO(), enabledFlag, &evaluation.EvaluationRequest{
		EntityId: "1",
		FlagKey:  "foo",
		Context: map[string]string{
			"bar": "boz",
		},
	})

	require.Error(t, err)
	require.EqualError(t, err, "rule rank: 0 detected out of order")
	assert.False(t, resp.Match)
	assert.Equal(t, flipt.EvaluationReason_ERROR_EVALUATION_REASON, resp.Reason)
}

func TestEvaluator_ErrorParsingNumber(t *testing.T) {
	var (
		store  = &evaluationStoreMock{}
		logger = zaptest.NewLogger(t)
		s      = NewEvaluator(logger, store)
	)

	store.On("GetEvaluationRules", mock.Anything, storage.NewResource("", "foo")).Return(
		[]*storage.EvaluationRule{
			{
				ID:      "1",
				FlagKey: "foo",
				Rank:    1,
				Segments: map[string]*storage.EvaluationSegment{
					"bar": {
						SegmentKey: "bar",
						MatchType:  flipt.MatchType_ALL_MATCH_TYPE,
						Constraints: []storage.EvaluationConstraint{
							{
								ID:       "2",
								Type:     flipt.ComparisonType_NUMBER_COMPARISON_TYPE,
								Property: "bar",
								Operator: flipt.OpEQ,
								Value:    "boz",
							},
						},
					},
				},
			},
		}, nil)

	resp, err := s.Evaluate(context.TODO(), enabledFlag, &evaluation.EvaluationRequest{
		EntityId: "1",
		FlagKey:  "foo",
		Context: map[string]string{
			"bar": "baz",
		},
	})

	require.NoError(t, err)
	assert.False(t, resp.Match)
	assert.Equal(t, flipt.EvaluationReason_UNKNOWN_EVALUATION_REASON, resp.Reason)
}
func TestEvaluator_ErrorParsingDateTime(t *testing.T) {
	var (
		store  = &evaluationStoreMock{}
		logger = zaptest.NewLogger(t)
		s      = NewEvaluator(logger, store)
	)

	store.On("GetEvaluationRules", mock.Anything, storage.NewResource("", "foo")).Return(
		[]*storage.EvaluationRule{
			{
				ID:      "1",
				FlagKey: "foo",
				Rank:    1,
				Segments: map[string]*storage.EvaluationSegment{
					"bar": {
						SegmentKey: "bar",
						MatchType:  flipt.MatchType_ALL_MATCH_TYPE,
						Constraints: []storage.EvaluationConstraint{
							{
								ID:       "2",
								Type:     flipt.ComparisonType_DATETIME_COMPARISON_TYPE,
								Property: "bar",
								Operator: flipt.OpEQ,
								Value:    "boz",
							},
						},
					},
				},
			},
		}, nil)

	resp, err := s.Evaluate(context.TODO(), enabledFlag, &evaluation.EvaluationRequest{
		EntityId: "1",
		FlagKey:  "foo",
		Context: map[string]string{
			"bar": "baz",
		},
	})

	require.NoError(t, err)
	assert.False(t, resp.Match)
	assert.Equal(t, flipt.EvaluationReason_UNKNOWN_EVALUATION_REASON, resp.Reason)
}

func TestEvaluator_ErrorGettingDistributions(t *testing.T) {
	var (
		store  = &evaluationStoreMock{}
		logger = zaptest.NewLogger(t)
		s      = NewEvaluator(logger, store)
	)

	store.On("GetEvaluationRules", mock.Anything, storage.NewResource("", "foo")).Return(
		[]*storage.EvaluationRule{
			{
				ID:      "1",
				FlagKey: "foo",
				Rank:    0,
				Segments: map[string]*storage.EvaluationSegment{
					"bar": {
						SegmentKey: "bar",
						MatchType:  flipt.MatchType_ALL_MATCH_TYPE,
						Constraints: []storage.EvaluationConstraint{
							{
								ID:       "2",
								Type:     flipt.ComparisonType_STRING_COMPARISON_TYPE,
								Property: "bar",
								Operator: flipt.OpEQ,
								Value:    "baz",
							},
						},
					},
				},
			},
		}, nil)

	store.On("GetEvaluationDistributions", mock.Anything, storage.NewID("1")).Return([]*storage.EvaluationDistribution{}, errors.New("error getting distributions!"))

	resp, err := s.Evaluate(context.TODO(), enabledFlag, &evaluation.EvaluationRequest{
		EntityId: "1",
		FlagKey:  "foo",
		Context: map[string]string{
			"bar": "baz",
		},
	})

	require.Error(t, err)
	assert.False(t, resp.Match)
	assert.Equal(t, flipt.EvaluationReason_ERROR_EVALUATION_REASON, resp.Reason)
}

// Match ALL constraints
func TestEvaluator_MatchAll_NoVariants_NoDistributions(t *testing.T) {
	var (
		store  = &evaluationStoreMock{}
		logger = zaptest.NewLogger(t)
		s      = NewEvaluator(logger, store)
	)

	store.On("GetEvaluationRules", mock.Anything, storage.NewResource("", "foo")).Return(
		[]*storage.EvaluationRule{
			{
				ID:      "1",
				FlagKey: "foo",
				Rank:    0,
				Segments: map[string]*storage.EvaluationSegment{
					"bar": {
						SegmentKey: "bar",
						MatchType:  flipt.MatchType_ALL_MATCH_TYPE,
						Constraints: []storage.EvaluationConstraint{
							{
								ID:       "2",
								Type:     flipt.ComparisonType_STRING_COMPARISON_TYPE,
								Property: "bar",
								Operator: flipt.OpEQ,
								Value:    "baz",
							},
						},
					},
				},
			},
		}, nil)

	store.On("GetEvaluationDistributions", mock.Anything, storage.NewID("1")).Return([]*storage.EvaluationDistribution{}, nil)

	tests := []struct {
		name      string
		req       *evaluation.EvaluationRequest
		wantMatch bool
	}{
		{
			name: "match string value",
			req: &evaluation.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "1",
				Context: map[string]string{
					"bar": "baz",
				},
			},
			wantMatch: true,
		},
		{
			name: "no match string value",
			req: &evaluation.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "1",
				Context: map[string]string{
					"bar": "boz",
				},
			},
		},
	}

	for _, tt := range tests {
		var (
			req       = tt.req
			wantMatch = tt.wantMatch
		)

		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.Evaluate(context.TODO(), enabledFlag, req)
			require.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, "foo", resp.FlagKey)
			assert.Equal(t, req.Context, resp.RequestContext)

			if !wantMatch {
				assert.False(t, resp.Match)
				assert.Empty(t, resp.SegmentKey)
				return
			}

			assert.True(t, resp.Match)
			assert.Equal(t, "bar", resp.SegmentKey)
			assert.Empty(t, resp.Value)
			assert.Equal(t, flipt.EvaluationReason_MATCH_EVALUATION_REASON, resp.Reason)
		})
	}
}

func TestEvaluator_MatchAll_NoDistributions_DefaultVariant(t *testing.T) {
	var (
		store  = &evaluationStoreMock{}
		logger = zaptest.NewLogger(t)
		s      = NewEvaluator(logger, store)
	)

	store.On("GetEvaluationRules", mock.Anything, storage.NewResource("", "foo")).Return(
		[]*storage.EvaluationRule{
			{
				ID:      "1",
				FlagKey: "foo",
				Rank:    0,
				Segments: map[string]*storage.EvaluationSegment{
					"bar": {
						SegmentKey: "bar",
						MatchType:  flipt.MatchType_ALL_MATCH_TYPE,
						Constraints: []storage.EvaluationConstraint{
							{
								ID:       "2",
								Type:     flipt.ComparisonType_STRING_COMPARISON_TYPE,
								Property: "bar",
								Operator: flipt.OpEQ,
								Value:    "baz",
							},
						},
					},
				},
			},
		}, nil)

	store.On("GetEvaluationDistributions", mock.Anything, storage.NewID("1")).Return([]*storage.EvaluationDistribution{}, nil)

	resp, err := s.Evaluate(context.TODO(), enabledFlagWithDefaultVariant, &evaluation.EvaluationRequest{
		FlagKey:  "foo",
		EntityId: "1",
		Context: map[string]string{
			"bar": "baz",
		},
	})

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "foo", resp.FlagKey)
	assert.True(t, resp.Match)
	assert.Equal(t, "bar", resp.SegmentKey)
	assert.Equal(t, "bar", resp.Value)
	assert.Equal(t, `{ "bar": "baz" }`, resp.Attachment)
	assert.Equal(t, flipt.EvaluationReason_MATCH_EVALUATION_REASON, resp.Reason)
}

func TestEvaluator_MatchAll_MultipleSegments(t *testing.T) {
	var (
		store  = &evaluationStoreMock{}
		logger = zaptest.NewLogger(t)
		s      = NewEvaluator(logger, store)
	)

	store.On("GetEvaluationRules", mock.Anything, storage.NewResource("", "foo")).Return(
		[]*storage.EvaluationRule{
			{
				ID:              "1",
				FlagKey:         "foo",
				Rank:            0,
				SegmentOperator: flipt.SegmentOperator_AND_SEGMENT_OPERATOR,
				Segments: map[string]*storage.EvaluationSegment{
					"bar": {
						SegmentKey: "bar",
						MatchType:  flipt.MatchType_ALL_MATCH_TYPE,
						Constraints: []storage.EvaluationConstraint{
							{
								ID:       "2",
								Type:     flipt.ComparisonType_STRING_COMPARISON_TYPE,
								Property: "bar",
								Operator: flipt.OpEQ,
								Value:    "baz",
							},
						},
					},
					"foo": {
						SegmentKey: "foo",
						MatchType:  flipt.MatchType_ALL_MATCH_TYPE,
						Constraints: []storage.EvaluationConstraint{
							{
								ID:       "2",
								Type:     flipt.ComparisonType_STRING_COMPARISON_TYPE,
								Property: "company",
								Operator: flipt.OpEQ,
								Value:    "flipt",
							},
						},
					},
				},
			},
		}, nil)

	store.On("GetEvaluationDistributions", mock.Anything, storage.NewID("1")).Return([]*storage.EvaluationDistribution{}, nil)

	tests := []struct {
		name      string
		req       *evaluation.EvaluationRequest
		wantMatch bool
	}{
		{
			name: "match string value",
			req: &evaluation.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "1",
				Context: map[string]string{
					"bar":     "baz",
					"company": "flipt",
				},
			},
			wantMatch: true,
		},
		{
			name: "no match string value",
			req: &evaluation.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "1",
				Context: map[string]string{
					"bar": "boz",
				},
			},
		},
	}

	for _, tt := range tests {
		var (
			req       = tt.req
			wantMatch = tt.wantMatch
		)

		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.Evaluate(context.TODO(), enabledFlag, req)
			require.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, "foo", resp.FlagKey)
			assert.Equal(t, req.Context, resp.RequestContext)

			if !wantMatch {
				assert.False(t, resp.Match)
				assert.Empty(t, resp.SegmentKey, "segment key should be empty")
				return
			}

			assert.True(t, resp.Match)

			assert.Len(t, resp.SegmentKeys, 2)
			assert.Contains(t, resp.SegmentKeys, "bar")
			assert.Contains(t, resp.SegmentKeys, "foo")
			assert.Empty(t, resp.Value)
			assert.Equal(t, flipt.EvaluationReason_MATCH_EVALUATION_REASON, resp.Reason)
		})
	}
}

func TestEvaluator_DistributionNotMatched(t *testing.T) {
	var (
		store  = &evaluationStoreMock{}
		logger = zaptest.NewLogger(t)
		s      = NewEvaluator(logger, store)
	)

	store.On("GetEvaluationRules", mock.Anything, storage.NewResource("", "foo")).Return(
		[]*storage.EvaluationRule{
			{
				ID:      "1",
				FlagKey: "foo",
				Rank:    0,
				Segments: map[string]*storage.EvaluationSegment{
					"bar": {
						SegmentKey: "bar",
						MatchType:  flipt.MatchType_ALL_MATCH_TYPE,
						Constraints: []storage.EvaluationConstraint{
							// constraint: bar (string) == baz
							{
								ID:       "2",
								Type:     flipt.ComparisonType_STRING_COMPARISON_TYPE,
								Property: "bar",
								Operator: flipt.OpEQ,
								Value:    "baz",
							},
							// constraint: admin (bool) == true
							{
								ID:       "3",
								Type:     flipt.ComparisonType_BOOLEAN_COMPARISON_TYPE,
								Property: "admin",
								Operator: flipt.OpTrue,
							},
						},
					},
				},
			},
		}, nil)

	store.On("GetEvaluationDistributions", mock.Anything, storage.NewID("1")).Return(
		[]*storage.EvaluationDistribution{
			{
				ID:                "4",
				RuleID:            "1",
				VariantID:         "5",
				Rollout:           10,
				VariantKey:        "boz",
				VariantAttachment: `{"key":"value"}`,
			},
		}, nil)

	resp, err := s.Evaluate(context.TODO(), enabledFlag, &evaluation.EvaluationRequest{
		FlagKey:  "foo",
		EntityId: "123",
		Context: map[string]string{
			"bar":   "baz",
			"admin": "true",
		},
	})

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "foo", resp.FlagKey)

	assert.False(t, resp.Match, "distribution not matched")
}

func TestEvaluator_DistributionNotMatched_DefaultVariant(t *testing.T) {
	var (
		store  = &evaluationStoreMock{}
		logger = zaptest.NewLogger(t)
		s      = NewEvaluator(logger, store)
	)

	store.On("GetEvaluationRules", mock.Anything, storage.NewResource("", "foo")).Return(
		[]*storage.EvaluationRule{
			{
				ID:      "1",
				FlagKey: "foo",
				Rank:    0,
				Segments: map[string]*storage.EvaluationSegment{
					"bar": {
						SegmentKey: "bar",
						MatchType:  flipt.MatchType_ALL_MATCH_TYPE,
						Constraints: []storage.EvaluationConstraint{
							// constraint: bar (string) == baz
							{
								ID:       "2",
								Type:     flipt.ComparisonType_STRING_COMPARISON_TYPE,
								Property: "bar",
								Operator: flipt.OpEQ,
								Value:    "baz",
							},
							// constraint: admin (bool) == true
							{
								ID:       "3",
								Type:     flipt.ComparisonType_BOOLEAN_COMPARISON_TYPE,
								Property: "admin",
								Operator: flipt.OpTrue,
							},
						},
					},
				},
			},
		}, nil)

	store.On("GetEvaluationDistributions", mock.Anything, storage.NewID("1")).Return(
		[]*storage.EvaluationDistribution{
			{
				ID:                "4",
				RuleID:            "1",
				VariantID:         "5",
				Rollout:           10,
				VariantKey:        "boz",
				VariantAttachment: `{"key":"value"}`,
			},
		}, nil)

	resp, err := s.Evaluate(context.TODO(), enabledFlagWithDefaultVariant, &evaluation.EvaluationRequest{
		FlagKey:  "foo",
		EntityId: "123",
		Context: map[string]string{
			"bar":   "baz",
			"admin": "true",
		},
	})

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "foo", resp.FlagKey)

	assert.False(t, resp.Match, "distribution not matched")
	assert.Equal(t, "bar", resp.Value)
	assert.Equal(t, `{ "bar": "baz" }`, resp.Attachment)
	assert.Equal(t, flipt.EvaluationReason_DEFAULT_EVALUATION_REASON, resp.Reason)
}

func TestEvaluator_MatchAll_SingleVariantDistribution(t *testing.T) {
	var (
		store  = &evaluationStoreMock{}
		logger = zaptest.NewLogger(t)
		s      = NewEvaluator(logger, store)
	)

	store.On("GetEvaluationRules", mock.Anything, storage.NewResource("", "foo")).Return(
		[]*storage.EvaluationRule{
			{
				ID:      "1",
				FlagKey: "foo",
				Rank:    0,
				Segments: map[string]*storage.EvaluationSegment{
					"bar": {
						SegmentKey: "bar",
						MatchType:  flipt.MatchType_ALL_MATCH_TYPE,
						Constraints: []storage.EvaluationConstraint{
							// constraint: bar (string) == baz
							{
								ID:       "2",
								Type:     flipt.ComparisonType_STRING_COMPARISON_TYPE,
								Property: "bar",
								Operator: flipt.OpEQ,
								Value:    "baz",
							},
							// constraint: admin (bool) == true
							{
								ID:       "3",
								Type:     flipt.ComparisonType_BOOLEAN_COMPARISON_TYPE,
								Property: "admin",
								Operator: flipt.OpTrue,
							},
						},
					},
				},
			},
		}, nil)

	store.On("GetEvaluationDistributions", mock.Anything, storage.NewID("1")).Return(
		[]*storage.EvaluationDistribution{
			{
				ID:                "4",
				RuleID:            "1",
				VariantID:         "5",
				Rollout:           100,
				VariantKey:        "boz",
				VariantAttachment: `{"key":"value"}`,
			},
		}, nil)

	tests := []struct {
		name      string
		req       *evaluation.EvaluationRequest
		wantMatch bool
	}{
		{
			name: "matches all",
			req: &evaluation.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "1",
				Context: map[string]string{
					"bar":   "baz",
					"admin": "true",
				},
			},
			wantMatch: true,
		},
		{
			name: "no match all",
			req: &evaluation.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "1",
				Context: map[string]string{
					"bar":   "boz",
					"admin": "true",
				},
			},
		},
		{
			name: "no match just bool value",
			req: &evaluation.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "1",
				Context: map[string]string{
					"admin": "true",
				},
			},
		},
		{
			name: "no match just string value",
			req: &evaluation.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "1",
				Context: map[string]string{
					"bar": "baz",
				},
			},
		},
	}

	for _, tt := range tests {
		var (
			req       = tt.req
			wantMatch = tt.wantMatch
		)

		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.Evaluate(context.TODO(), enabledFlag, req)
			require.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, "foo", resp.FlagKey)
			assert.Equal(t, req.Context, resp.RequestContext)

			if !wantMatch {
				assert.False(t, resp.Match)
				assert.Empty(t, resp.SegmentKey)
				return
			}

			assert.True(t, resp.Match)
			assert.Equal(t, "bar", resp.SegmentKey)
			assert.Equal(t, "boz", resp.Value)
			assert.Equal(t, `{"key":"value"}`, resp.Attachment)
			assert.Equal(t, flipt.EvaluationReason_MATCH_EVALUATION_REASON, resp.Reason)
		})
	}
}

func TestEvaluator_MatchAll_RolloutDistribution(t *testing.T) {
	var (
		store  = &evaluationStoreMock{}
		logger = zaptest.NewLogger(t)
		s      = NewEvaluator(logger, store)
	)

	store.On("GetEvaluationRules", mock.Anything, storage.NewResource("", "foo")).Return(
		[]*storage.EvaluationRule{
			{
				ID:      "1",
				FlagKey: "foo",
				Rank:    0,
				Segments: map[string]*storage.EvaluationSegment{
					"bar": {
						SegmentKey: "bar",
						MatchType:  flipt.MatchType_ALL_MATCH_TYPE,
						Constraints: []storage.EvaluationConstraint{
							// constraint: bar (string) == baz
							{
								ID:       "2",
								Type:     flipt.ComparisonType_STRING_COMPARISON_TYPE,
								Property: "bar",
								Operator: flipt.OpEQ,
								Value:    "baz",
							},
						},
					},
				},
			},
		}, nil)

	store.On("GetEvaluationDistributions", mock.Anything, storage.NewID("1")).Return(
		[]*storage.EvaluationDistribution{
			{
				ID:         "4",
				RuleID:     "1",
				VariantID:  "5",
				Rollout:    50,
				VariantKey: "boz",
			},
			{
				ID:         "6",
				RuleID:     "1",
				VariantID:  "7",
				Rollout:    50,
				VariantKey: "booz",
			},
		}, nil)

	tests := []struct {
		name              string
		req               *evaluation.EvaluationRequest
		matchesVariantKey string
		wantMatch         bool
	}{
		{
			name: "match string value - variant 1",
			req: &evaluation.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "1",
				Context: map[string]string{
					"bar": "baz",
				},
			},
			matchesVariantKey: "boz",
			wantMatch:         true,
		},
		{
			name: "match string value - variant 2",
			req: &evaluation.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "2",
				Context: map[string]string{
					"bar": "baz",
				},
			},
			matchesVariantKey: "booz",
			wantMatch:         true,
		},
		{
			name: "no match string value",
			req: &evaluation.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "1",
				Context: map[string]string{
					"bar": "boz",
				},
			},
		},
	}

	for _, tt := range tests {
		var (
			req               = tt.req
			matchesVariantKey = tt.matchesVariantKey
			wantMatch         = tt.wantMatch
		)

		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.Evaluate(context.TODO(), enabledFlag, req)
			require.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, "foo", resp.FlagKey)
			assert.Equal(t, req.Context, resp.RequestContext)

			if !wantMatch {
				assert.False(t, resp.Match)
				assert.Empty(t, resp.SegmentKey)
				return
			}

			assert.True(t, resp.Match)
			assert.Equal(t, "bar", resp.SegmentKey)
			assert.Equal(t, matchesVariantKey, resp.Value)
			assert.Equal(t, flipt.EvaluationReason_MATCH_EVALUATION_REASON, resp.Reason)
		})
	}
}

func TestEvaluator_MatchAll_RolloutDistribution_MultiRule(t *testing.T) {
	var (
		store  = &evaluationStoreMock{}
		logger = zaptest.NewLogger(t)
		s      = NewEvaluator(logger, store)
	)

	store.On("GetEvaluationRules", mock.Anything, storage.NewResource("", "foo")).Return(
		[]*storage.EvaluationRule{
			{
				ID:      "1",
				FlagKey: "foo",
				Rank:    0,
				Segments: map[string]*storage.EvaluationSegment{
					"subscribers": {
						SegmentKey: "subscribers",
						MatchType:  flipt.MatchType_ALL_MATCH_TYPE,
						Constraints: []storage.EvaluationConstraint{
							// constraint: premium_user (bool) == true
							{
								ID:       "2",
								Type:     flipt.ComparisonType_BOOLEAN_COMPARISON_TYPE,
								Property: "premium_user",
								Operator: flipt.OpTrue,
							},
						},
					},
				},
			},
			{
				ID:      "2",
				FlagKey: "foo",
				Rank:    1,
				Segments: map[string]*storage.EvaluationSegment{
					"all_users": {
						SegmentKey: "all_users",
						MatchType:  flipt.MatchType_ALL_MATCH_TYPE,
					},
				},
			},
		}, nil)

	store.On("GetEvaluationDistributions", mock.Anything, storage.NewID("1")).Return(
		[]*storage.EvaluationDistribution{
			{
				ID:         "4",
				RuleID:     "1",
				VariantID:  "5",
				Rollout:    50,
				VariantKey: "released",
			},
			{
				ID:         "6",
				RuleID:     "1",
				VariantID:  "7",
				Rollout:    50,
				VariantKey: "unreleased",
			},
		}, nil)

	resp, err := s.Evaluate(context.TODO(), enabledFlag, &evaluation.EvaluationRequest{
		FlagKey:  "foo",
		EntityId: uuid.Must(uuid.NewV4()).String(),
		Context: map[string]string{
			"premium_user": "true",
		},
	})

	require.NoError(t, err)

	assert.NotNil(t, resp)
	assert.True(t, resp.Match)
	assert.Equal(t, "subscribers", resp.SegmentKey)
	assert.Equal(t, "foo", resp.FlagKey)
	assert.NotEmpty(t, resp.Value)
	assert.Equal(t, flipt.EvaluationReason_MATCH_EVALUATION_REASON, resp.Reason)
}

func TestEvaluator_MatchAll_NoConstraints(t *testing.T) {
	var (
		store  = &evaluationStoreMock{}
		logger = zaptest.NewLogger(t)
		s      = NewEvaluator(logger, store)
	)

	store.On("GetEvaluationRules", mock.Anything, storage.NewResource("", "foo")).Return(
		[]*storage.EvaluationRule{
			{
				ID:      "1",
				FlagKey: "foo",
				Rank:    0,
				Segments: map[string]*storage.EvaluationSegment{
					"bar": {
						SegmentKey: "bar",
						MatchType:  flipt.MatchType_ALL_MATCH_TYPE,
					},
				},
			},
		}, nil)

	store.On("GetEvaluationDistributions", mock.Anything, storage.NewID("1")).Return(
		[]*storage.EvaluationDistribution{
			{
				ID:         "4",
				RuleID:     "1",
				VariantID:  "5",
				Rollout:    50,
				VariantKey: "boz",
			},
			{
				ID:         "6",
				RuleID:     "1",
				VariantID:  "7",
				Rollout:    50,
				VariantKey: "moz",
			},
		}, nil)

	tests := []struct {
		name              string
		req               *evaluation.EvaluationRequest
		matchesVariantKey string
		wantMatch         bool
	}{
		{
			name: "match no value - variant 1",
			req: &evaluation.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "10",
				Context:  map[string]string{},
			},
			matchesVariantKey: "boz",
			wantMatch:         true,
		},
		{
			name: "match no value - variant 2",
			req: &evaluation.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "01",
				Context:  map[string]string{},
			},
			matchesVariantKey: "moz",
			wantMatch:         true,
		},
		{
			name: "match string value - variant 2",
			req: &evaluation.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "01",
				Context: map[string]string{
					"bar": "boz",
				},
			},
			matchesVariantKey: "moz",
			wantMatch:         true,
		},
	}

	for _, tt := range tests {
		var (
			req               = tt.req
			matchesVariantKey = tt.matchesVariantKey
			wantMatch         = tt.wantMatch
		)

		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.Evaluate(context.TODO(), enabledFlag, req)
			require.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, "foo", resp.FlagKey)
			assert.Equal(t, req.Context, resp.RequestContext)

			if !wantMatch {
				assert.False(t, resp.Match)
				assert.Empty(t, resp.SegmentKey)
				return
			}

			assert.True(t, resp.Match)
			assert.Equal(t, "bar", resp.SegmentKey)
			assert.Equal(t, matchesVariantKey, resp.Value)
			assert.Equal(t, flipt.EvaluationReason_MATCH_EVALUATION_REASON, resp.Reason)
		})
	}
}

// Match ANY constraints

func TestEvaluator_MatchAny_NoVariants_NoDistributions(t *testing.T) {
	var (
		store  = &evaluationStoreMock{}
		logger = zaptest.NewLogger(t)
		s      = NewEvaluator(logger, store)
	)

	store.On("GetEvaluationRules", mock.Anything, storage.NewResource("", "foo")).Return(
		[]*storage.EvaluationRule{
			{
				ID:      "1",
				FlagKey: "foo",
				Rank:    0,
				Segments: map[string]*storage.EvaluationSegment{
					"bar": {
						SegmentKey: "bar",
						MatchType:  flipt.MatchType_ANY_MATCH_TYPE,
						Constraints: []storage.EvaluationConstraint{
							{
								ID:       "2",
								Type:     flipt.ComparisonType_STRING_COMPARISON_TYPE,
								Property: "bar",
								Operator: flipt.OpEQ,
								Value:    "baz",
							},
						},
					},
				},
			},
		}, nil)

	store.On("GetEvaluationDistributions", mock.Anything, storage.NewID("1")).Return([]*storage.EvaluationDistribution{}, nil)

	tests := []struct {
		name      string
		req       *evaluation.EvaluationRequest
		wantMatch bool
	}{
		{
			name: "match string value",
			req: &evaluation.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "1",
				Context: map[string]string{
					"bar": "baz",
				},
			},
			wantMatch: true,
		},
		{
			name: "no match string value",
			req: &evaluation.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "1",
				Context: map[string]string{
					"bar": "boz",
				},
			},
		},
	}

	for _, tt := range tests {
		var (
			req       = tt.req
			wantMatch = tt.wantMatch
		)

		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.Evaluate(context.TODO(), enabledFlag, req)
			require.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, "foo", resp.FlagKey)
			assert.Equal(t, req.Context, resp.RequestContext)

			if !wantMatch {
				assert.False(t, resp.Match)
				assert.Empty(t, resp.SegmentKey)
				return
			}

			assert.True(t, resp.Match)
			assert.Equal(t, "bar", resp.SegmentKey)
			assert.Empty(t, resp.Value)
			assert.Equal(t, flipt.EvaluationReason_MATCH_EVALUATION_REASON, resp.Reason)
		})
	}
}

func TestEvaluator_MatchAny_NoDistributions_DefaultVariant(t *testing.T) {
	var (
		store  = &evaluationStoreMock{}
		logger = zaptest.NewLogger(t)
		s      = NewEvaluator(logger, store)
	)

	store.On("GetEvaluationRules", mock.Anything, storage.NewResource("", "foo")).Return(
		[]*storage.EvaluationRule{
			{
				ID:      "1",
				FlagKey: "foo",
				Rank:    0,
				Segments: map[string]*storage.EvaluationSegment{
					"bar": {
						SegmentKey: "bar",
						MatchType:  flipt.MatchType_ANY_MATCH_TYPE,
						Constraints: []storage.EvaluationConstraint{
							{
								ID:       "2",
								Type:     flipt.ComparisonType_STRING_COMPARISON_TYPE,
								Property: "bar",
								Operator: flipt.OpEQ,
								Value:    "baz",
							},
						},
					},
				},
			},
		}, nil)

	store.On("GetEvaluationDistributions", mock.Anything, storage.NewID("1")).Return([]*storage.EvaluationDistribution{}, nil)

	resp, err := s.Evaluate(context.TODO(), enabledFlagWithDefaultVariant, &evaluation.EvaluationRequest{
		FlagKey:  "foo",
		EntityId: "1",
		Context: map[string]string{
			"bar": "baz",
		},
	})

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "foo", resp.FlagKey)
	assert.True(t, resp.Match)
	assert.Equal(t, "bar", resp.SegmentKey)
	assert.Equal(t, "bar", resp.Value)
	assert.Equal(t, `{ "bar": "baz" }`, resp.Attachment)
	assert.Equal(t, flipt.EvaluationReason_MATCH_EVALUATION_REASON, resp.Reason)
}

func TestEvaluator_MatchAny_SingleVariantDistribution(t *testing.T) {
	var (
		store  = &evaluationStoreMock{}
		logger = zaptest.NewLogger(t)
		s      = NewEvaluator(logger, store)
	)

	store.On("GetEvaluationRules", mock.Anything, storage.NewResource("", "foo")).Return(
		[]*storage.EvaluationRule{
			{
				ID:      "1",
				FlagKey: "foo",
				Rank:    0,
				Segments: map[string]*storage.EvaluationSegment{
					"bar": {
						SegmentKey: "bar",
						MatchType:  flipt.MatchType_ANY_MATCH_TYPE,
						Constraints: []storage.EvaluationConstraint{
							// constraint: bar (string) == baz
							{
								ID:       "2",
								Type:     flipt.ComparisonType_STRING_COMPARISON_TYPE,
								Property: "bar",
								Operator: flipt.OpEQ,
								Value:    "baz",
							},
							// constraint: admin (bool) == true
							{
								ID:       "3",
								Type:     flipt.ComparisonType_BOOLEAN_COMPARISON_TYPE,
								Property: "admin",
								Operator: flipt.OpTrue,
							},
						},
					},
				},
			},
		}, nil)

	store.On("GetEvaluationDistributions", mock.Anything, storage.NewID("1")).Return(
		[]*storage.EvaluationDistribution{
			{
				ID:         "4",
				RuleID:     "1",
				VariantID:  "5",
				Rollout:    100,
				VariantKey: "boz",
			},
		}, nil)

	tests := []struct {
		name      string
		req       *evaluation.EvaluationRequest
		wantMatch bool
	}{
		{
			name: "matches all",
			req: &evaluation.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "1",
				Context: map[string]string{
					"bar":   "baz",
					"admin": "true",
				},
			},
			wantMatch: true,
		},
		{
			name: "matches one",
			req: &evaluation.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "1",
				Context: map[string]string{
					"bar":   "boz",
					"admin": "true",
				},
			},
			wantMatch: true,
		},
		{
			name: "matches just bool value",
			req: &evaluation.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "1",
				Context: map[string]string{
					"admin": "true",
				},
			},
			wantMatch: true,
		},
		{
			name: "matches just string value",
			req: &evaluation.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "1",
				Context: map[string]string{
					"bar": "baz",
				},
			},
			wantMatch: true,
		},
		{
			name: "no matches wrong bool value",
			req: &evaluation.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "1",
				Context: map[string]string{
					"admin": "false",
				},
			},
		},
		{
			name: "no matches wrong string value",
			req: &evaluation.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "1",
				Context: map[string]string{
					"bar": "boz",
				},
			},
		},
		{
			name: "no match none",
			req: &evaluation.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "1",
				Context: map[string]string{
					"boss": "boz",
					"user": "true",
				},
			},
		},
	}

	for _, tt := range tests {
		var (
			req       = tt.req
			wantMatch = tt.wantMatch
		)

		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.Evaluate(context.TODO(), enabledFlag, req)
			require.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, "foo", resp.FlagKey)
			assert.Equal(t, req.Context, resp.RequestContext)

			if !wantMatch {
				assert.False(t, resp.Match)
				assert.Empty(t, resp.SegmentKey)
				return
			}

			assert.True(t, resp.Match)
			assert.Equal(t, "bar", resp.SegmentKey)
			assert.Equal(t, "boz", resp.Value)
			assert.Equal(t, flipt.EvaluationReason_MATCH_EVALUATION_REASON, resp.Reason)
		})
	}
}

func TestEvaluator_MatchAny_RolloutDistribution(t *testing.T) {
	var (
		store  = &evaluationStoreMock{}
		logger = zaptest.NewLogger(t)
		s      = NewEvaluator(logger, store)
	)

	store.On("GetEvaluationRules", mock.Anything, storage.NewResource("", "foo")).Return(
		[]*storage.EvaluationRule{
			{
				ID:      "1",
				FlagKey: "foo",
				Rank:    0,
				Segments: map[string]*storage.EvaluationSegment{
					"bar": {
						SegmentKey: "bar",
						MatchType:  flipt.MatchType_ANY_MATCH_TYPE,
						Constraints: []storage.EvaluationConstraint{
							// constraint: bar (string) == baz
							{
								ID:       "2",
								Type:     flipt.ComparisonType_STRING_COMPARISON_TYPE,
								Property: "bar",
								Operator: flipt.OpEQ,
								Value:    "baz",
							},
						},
					},
				},
			},
		}, nil)

	store.On("GetEvaluationDistributions", mock.Anything, storage.NewID("1")).Return(
		[]*storage.EvaluationDistribution{
			{
				ID:         "4",
				RuleID:     "1",
				VariantID:  "5",
				Rollout:    50,
				VariantKey: "boz",
			},
			{
				ID:         "6",
				RuleID:     "1",
				VariantID:  "7",
				Rollout:    50,
				VariantKey: "booz",
			},
		}, nil)

	tests := []struct {
		name              string
		req               *evaluation.EvaluationRequest
		matchesVariantKey string
		wantMatch         bool
	}{
		{
			name: "match string value - variant 1",
			req: &evaluation.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "1",
				Context: map[string]string{
					"bar": "baz",
				},
			},
			matchesVariantKey: "boz",
			wantMatch:         true,
		},
		{
			name: "match string value - variant 2",
			req: &evaluation.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "2",
				Context: map[string]string{
					"bar": "baz",
				},
			},
			matchesVariantKey: "booz",
			wantMatch:         true,
		},
		{
			name: "no match string value",
			req: &evaluation.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "1",
				Context: map[string]string{
					"bar": "boz",
				},
			},
		},
	}

	for _, tt := range tests {
		var (
			req               = tt.req
			matchesVariantKey = tt.matchesVariantKey
			wantMatch         = tt.wantMatch
		)

		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.Evaluate(context.TODO(), enabledFlag, req)
			require.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, "foo", resp.FlagKey)
			assert.Equal(t, req.Context, resp.RequestContext)

			if !wantMatch {
				assert.False(t, resp.Match)
				assert.Empty(t, resp.SegmentKey)
				return
			}

			assert.True(t, resp.Match)
			assert.Equal(t, "bar", resp.SegmentKey)
			assert.Equal(t, matchesVariantKey, resp.Value)
			assert.Equal(t, flipt.EvaluationReason_MATCH_EVALUATION_REASON, resp.Reason)
		})
	}
}

func TestEvaluator_MatchAny_RolloutDistribution_MultiRule(t *testing.T) {
	var (
		store  = &evaluationStoreMock{}
		logger = zaptest.NewLogger(t)
		s      = NewEvaluator(logger, store)
	)

	store.On("GetEvaluationRules", mock.Anything, storage.NewResource("", "foo")).Return(
		[]*storage.EvaluationRule{
			{
				ID:      "1",
				FlagKey: "foo",
				Rank:    0,
				Segments: map[string]*storage.EvaluationSegment{
					"subscribers": {
						SegmentKey: "subscribers",
						MatchType:  flipt.MatchType_ANY_MATCH_TYPE,
						Constraints: []storage.EvaluationConstraint{
							// constraint: premium_user (bool) == true
							{
								ID:       "2",
								Type:     flipt.ComparisonType_BOOLEAN_COMPARISON_TYPE,
								Property: "premium_user",
								Operator: flipt.OpTrue,
							},
						},
					},
				},
			},
			{
				ID:      "2",
				FlagKey: "foo",
				Rank:    1,
				Segments: map[string]*storage.EvaluationSegment{
					"all_users": {
						SegmentKey: "all_users",
						MatchType:  flipt.MatchType_ALL_MATCH_TYPE,
						Constraints: []storage.EvaluationConstraint{
							// constraint: premium_user (bool) == true
							{
								ID:       "2",
								Type:     flipt.ComparisonType_BOOLEAN_COMPARISON_TYPE,
								Property: "premium_user",
								Operator: flipt.OpTrue,
							},
						},
					},
				},
			},
		}, nil)

	store.On("GetEvaluationDistributions", mock.Anything, storage.NewID("1")).Return(
		[]*storage.EvaluationDistribution{
			{
				ID:         "4",
				RuleID:     "1",
				VariantID:  "5",
				Rollout:    50,
				VariantKey: "released",
			},
			{
				ID:         "6",
				RuleID:     "1",
				VariantID:  "7",
				Rollout:    50,
				VariantKey: "unreleased",
			},
		}, nil)

	resp, err := s.Evaluate(context.TODO(), enabledFlag, &evaluation.EvaluationRequest{
		FlagKey:  "foo",
		EntityId: uuid.Must(uuid.NewV4()).String(),
		Context: map[string]string{
			"premium_user": "true",
		},
	})

	require.NoError(t, err)

	assert.NotNil(t, resp)
	assert.True(t, resp.Match)
	assert.Equal(t, "subscribers", resp.SegmentKey)
	assert.Equal(t, "foo", resp.FlagKey)
	assert.NotEmpty(t, resp.Value)
	assert.Equal(t, flipt.EvaluationReason_MATCH_EVALUATION_REASON, resp.Reason)
}

func TestEvaluator_MatchEntityId(t *testing.T) {
	var (
		store  = &evaluationStoreMock{}
		logger = zaptest.NewLogger(t)
		s      = NewEvaluator(logger, store)
	)

	store.On("GetEvaluationRules", mock.Anything, storage.NewResource("", "foo")).Return(
		[]*storage.EvaluationRule{
			{
				ID:      "1",
				FlagKey: "foo",
				Rank:    0,
				Segments: map[string]*storage.EvaluationSegment{
					"subscribers": {
						SegmentKey: "subscribers",
						MatchType:  flipt.MatchType_ANY_MATCH_TYPE,
						Constraints: []storage.EvaluationConstraint{
							{
								ID:       "2",
								Type:     flipt.ComparisonType_ENTITY_ID_COMPARISON_TYPE,
								Property: "entity",
								Operator: flipt.OpEQ,
								Value:    "user@flipt.io",
							},
						},
					},
				},
			},
		}, nil)

	store.On("GetEvaluationDistributions", mock.Anything, storage.NewID("1")).Return(
		[]*storage.EvaluationDistribution{
			{
				ID:         "4",
				RuleID:     "1",
				VariantID:  "5",
				Rollout:    50,
				VariantKey: "released",
			},
			{
				ID:         "6",
				RuleID:     "1",
				VariantID:  "7",
				Rollout:    50,
				VariantKey: "unreleased",
			},
		}, nil)

	resp, err := s.Evaluate(context.TODO(), enabledFlag, &evaluation.EvaluationRequest{
		FlagKey:  "foo",
		EntityId: "user@flipt.io",
		Context:  map[string]string{},
	})

	require.NoError(t, err)

	assert.NotNil(t, resp)
	assert.True(t, resp.Match)
	assert.Equal(t, "subscribers", resp.SegmentKey)
	assert.Equal(t, "foo", resp.FlagKey)
	assert.NotEmpty(t, resp.Value)
	assert.Equal(t, flipt.EvaluationReason_MATCH_EVALUATION_REASON, resp.Reason)
}

func TestEvaluator_MatchAny_NoConstraints(t *testing.T) {
	var (
		store  = &evaluationStoreMock{}
		logger = zaptest.NewLogger(t)
		s      = NewEvaluator(logger, store)
	)

	store.On("GetEvaluationRules", mock.Anything, storage.NewResource("", "foo")).Return(
		[]*storage.EvaluationRule{
			{
				ID:      "1",
				FlagKey: "foo",
				Rank:    0,
				Segments: map[string]*storage.EvaluationSegment{
					"bar": {
						SegmentKey: "bar",
						MatchType:  flipt.MatchType_ANY_MATCH_TYPE,
					},
				},
			},
		}, nil)

	store.On("GetEvaluationDistributions", mock.Anything, storage.NewID("1")).Return(
		[]*storage.EvaluationDistribution{
			{
				ID:         "4",
				RuleID:     "1",
				VariantID:  "5",
				Rollout:    50,
				VariantKey: "boz",
			},
			{
				ID:         "6",
				RuleID:     "1",
				VariantID:  "7",
				Rollout:    50,
				VariantKey: "moz",
			},
		}, nil)

	tests := []struct {
		name              string
		req               *evaluation.EvaluationRequest
		matchesVariantKey string
		wantMatch         bool
	}{
		{
			name: "match no value - variant 1",
			req: &evaluation.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "10",
				Context:  map[string]string{},
			},
			matchesVariantKey: "boz",
			wantMatch:         true,
		},
		{
			name: "match no value - variant 2",
			req: &evaluation.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "01",
				Context:  map[string]string{},
			},
			matchesVariantKey: "moz",
			wantMatch:         true,
		},
		{
			name: "match string value - variant 2",
			req: &evaluation.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "01",
				Context: map[string]string{
					"bar": "boz",
				},
			},
			matchesVariantKey: "moz",
			wantMatch:         true,
		},
	}

	for _, tt := range tests {
		var (
			req               = tt.req
			matchesVariantKey = tt.matchesVariantKey
			wantMatch         = tt.wantMatch
		)

		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.Evaluate(context.TODO(), enabledFlag, req)
			require.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, "foo", resp.FlagKey)
			assert.Equal(t, req.Context, resp.RequestContext)

			if !wantMatch {
				assert.False(t, resp.Match)
				assert.Empty(t, resp.SegmentKey)
				return
			}

			assert.True(t, resp.Match)
			assert.Equal(t, "bar", resp.SegmentKey)
			assert.Equal(t, matchesVariantKey, resp.Value)
			assert.Equal(t, flipt.EvaluationReason_MATCH_EVALUATION_REASON, resp.Reason)
		})
	}
}

// Since we skip rollout buckets that have 0% distribution, ensure that things still work
// when a 0% distribution is the first available one.
func TestEvaluator_FirstRolloutRuleIsZero(t *testing.T) {
	var (
		store  = &evaluationStoreMock{}
		logger = zaptest.NewLogger(t)
		s      = NewEvaluator(logger, store)
	)

	store.On("GetEvaluationRules", mock.Anything, storage.NewResource("", "foo")).Return(
		[]*storage.EvaluationRule{
			{
				ID:      "1",
				FlagKey: "foo",
				Rank:    0,
				Segments: map[string]*storage.EvaluationSegment{
					"bar": {
						SegmentKey: "bar",
						MatchType:  flipt.MatchType_ANY_MATCH_TYPE,
						Constraints: []storage.EvaluationConstraint{
							// constraint: bar (string) == baz
							{
								ID:       "2",
								Type:     flipt.ComparisonType_STRING_COMPARISON_TYPE,
								Property: "bar",
								Operator: flipt.OpEQ,
								Value:    "baz",
							},
						},
					},
				},
			},
		}, nil)

	store.On("GetEvaluationDistributions", mock.Anything, storage.NewID("1")).Return(
		[]*storage.EvaluationDistribution{
			{
				ID:         "4",
				RuleID:     "1",
				VariantID:  "5",
				Rollout:    0,
				VariantKey: "boz",
			},
			{
				ID:         "6",
				RuleID:     "1",
				VariantID:  "7",
				Rollout:    100,
				VariantKey: "booz",
			},
		}, nil)

	tests := []struct {
		name              string
		req               *evaluation.EvaluationRequest
		matchesVariantKey string
		wantMatch         bool
	}{
		{
			name: "match string value - variant 1",
			req: &evaluation.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "1",
				Context: map[string]string{
					"bar": "baz",
				},
			},
			matchesVariantKey: "booz",
			wantMatch:         true,
		},
	}

	for _, tt := range tests {
		var (
			req               = tt.req
			matchesVariantKey = tt.matchesVariantKey
			wantMatch         = tt.wantMatch
		)

		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.Evaluate(context.TODO(), enabledFlag, req)
			require.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, "foo", resp.FlagKey)
			assert.Equal(t, req.Context, resp.RequestContext)

			if !wantMatch {
				assert.False(t, resp.Match)
				assert.Empty(t, resp.SegmentKey)
				return
			}

			assert.True(t, resp.Match)
			assert.Equal(t, "bar", resp.SegmentKey)
			assert.Equal(t, matchesVariantKey, resp.Value)
			assert.Equal(t, flipt.EvaluationReason_MATCH_EVALUATION_REASON, resp.Reason)
		})
	}
}

// Ensure things work properly when many rollout distributions have a 0% value.
func TestEvaluator_MultipleZeroRolloutDistributions(t *testing.T) {
	var (
		store  = &evaluationStoreMock{}
		logger = zaptest.NewLogger(t)
		s      = NewEvaluator(logger, store)
	)

	store.On("GetEvaluationRules", mock.Anything, storage.NewResource("", "foo")).Return(
		[]*storage.EvaluationRule{
			{
				ID:      "1",
				FlagKey: "foo",
				Rank:    0,
				Segments: map[string]*storage.EvaluationSegment{
					"bar": {
						SegmentKey: "bar",
						MatchType:  flipt.MatchType_ALL_MATCH_TYPE,
						Constraints: []storage.EvaluationConstraint{
							// constraint: bar (string) == baz
							{
								ID:       "2",
								Type:     flipt.ComparisonType_STRING_COMPARISON_TYPE,
								Property: "bar",
								Operator: flipt.OpEQ,
								Value:    "baz",
							},
						},
					},
				},
			},
		}, nil)

	store.On("GetEvaluationDistributions", mock.Anything, storage.NewID("1")).Return(
		[]*storage.EvaluationDistribution{
			{
				ID:         "1",
				RuleID:     "1",
				VariantID:  "1",
				VariantKey: "1",
				Rollout:    0,
			},
			{
				ID:         "2",
				RuleID:     "1",
				VariantID:  "2",
				VariantKey: "2",
				Rollout:    0,
			},
			{
				ID:         "3",
				RuleID:     "1",
				VariantID:  "3",
				VariantKey: "3",
				Rollout:    50,
			},
			{
				ID:         "4",
				RuleID:     "4",
				VariantID:  "4",
				VariantKey: "4",
				Rollout:    0,
			},
			{
				ID:         "5",
				RuleID:     "1",
				VariantID:  "5",
				VariantKey: "5",
				Rollout:    0,
			},
			{
				ID:         "6",
				RuleID:     "1",
				VariantID:  "6",
				VariantKey: "6",
				Rollout:    50,
			},
		}, nil)

	tests := []struct {
		name              string
		req               *evaluation.EvaluationRequest
		matchesVariantKey string
		wantMatch         bool
	}{
		{
			name: "match string value - variant 1",
			req: &evaluation.EvaluationRequest{
				FlagKey:  "foo",
				EntityId: "1",
				Context: map[string]string{
					"bar": "baz",
				},
			},
			matchesVariantKey: "3",
			wantMatch:         true,
		},
	}

	for _, tt := range tests {
		var (
			req               = tt.req
			matchesVariantKey = tt.matchesVariantKey
			wantMatch         = tt.wantMatch
		)

		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.Evaluate(context.TODO(), enabledFlag, req)
			require.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, "foo", resp.FlagKey)
			assert.Equal(t, req.Context, resp.RequestContext)

			if !wantMatch {
				assert.False(t, resp.Match)
				assert.Empty(t, resp.SegmentKey)
				return
			}

			assert.True(t, resp.Match)
			assert.Equal(t, "bar", resp.SegmentKey)
			assert.Equal(t, matchesVariantKey, resp.Value)
			assert.Equal(t, flipt.EvaluationReason_MATCH_EVALUATION_REASON, resp.Reason)
		})
	}
}

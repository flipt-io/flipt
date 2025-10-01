package tracing_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.flipt.io/flipt/internal/server/tracing"
	"go.flipt.io/flipt/rpc/flipt/evaluation"
)

func TestReasonTransformation(t *testing.T) {
	for _, r := range evaluation.EvaluationReason_value {
		reason := evaluation.EvaluationReason(r)
		value := tracing.ReasonToValue(reason)
		got := tracing.ReasonFromValue(value)
		assert.Equal(t, reason, got)
	}
}

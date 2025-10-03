package tracing

import (
	"go.flipt.io/flipt/rpc/flipt/evaluation"
	"go.opentelemetry.io/otel/attribute"
)

const Event = "flipt.flag_evaluation"

var (
	AttributeFlag            = attribute.Key("flipt.key")
	AttributeFlagType        = attribute.Key("flipt.type")
	AttributeEnvironment     = attribute.Key("flipt.environment")
	AttributeNamespace       = attribute.Key("flipt.namespace")
	AttributeSegments        = attribute.Key("flipt.result.segments")
	AttributeReason          = attribute.Key("flipt.result.reason")
	AttributeVariant         = attribute.Key("flipt.result.variant")
	AttributeMatch           = attribute.Key("flipt.result.match")
	AttributeEntityID        = attribute.Key("flipt.context.entity_id")
	AttributeRequestID       = attribute.Key("flipt.context.request_id")
	AttributeFlagTypeVariant = AttributeFlagType.String("variant")
	AttributeFlagTypeBoolean = AttributeFlagType.String("boolean")
)

var (
	evaluationReasonToValues = map[evaluation.EvaluationReason]string{
		evaluation.EvaluationReason_UNKNOWN_EVALUATION_REASON:       "unknown",
		evaluation.EvaluationReason_FLAG_DISABLED_EVALUATION_REASON: "disabled",
		evaluation.EvaluationReason_MATCH_EVALUATION_REASON:         "match",
		evaluation.EvaluationReason_DEFAULT_EVALUATION_REASON:       "default",
	}
	evaluationReasonToNames = map[string]evaluation.EvaluationReason{
		"unknown":  evaluation.EvaluationReason_UNKNOWN_EVALUATION_REASON,
		"disabled": evaluation.EvaluationReason_FLAG_DISABLED_EVALUATION_REASON,
		"match":    evaluation.EvaluationReason_MATCH_EVALUATION_REASON,
		"default":  evaluation.EvaluationReason_DEFAULT_EVALUATION_REASON,
	}
)

func ReasonToValue(reason evaluation.EvaluationReason) string {
	return evaluationReasonToValues[reason]
}

func ReasonFromValue(value string) evaluation.EvaluationReason {
	return evaluationReasonToNames[value]
}

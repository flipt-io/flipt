package tracing

import (
	"go.flipt.io/flipt/rpc/v2/evaluation"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
)

const Event = "flipt_flag_evaluated"

// TODO: merge with metrics?
var (
	AttributeMatch       = attribute.Key("flipt_match")
	AttributeFlag        = attribute.Key("flipt_flag")
	AttributeFlagType    = attribute.Key("flipt_flag_type")
	AttributeEnvironment = attribute.Key("flipt_environment")
	AttributeNamespace   = attribute.Key("flipt_namespace")
	AttributeFlagEnabled = attribute.Key("flipt_flag_enabled")
	AttributeSegments    = attribute.Key("flipt_segments")
	AttributeReason      = attribute.Key("flipt_reason")
	AttributeValue       = attribute.Key("flipt_value")
	AttributeEntityID    = attribute.Key("flipt_entity_id")
	AttributeRequestID   = attribute.Key("flipt_request_id")
)

// Specific attributes for Semantic Conventions for Feature Flags in Spans
// https://opentelemetry.io/docs/specs/semconv/feature-flags/feature-flags-spans/
var (
	AttributeFlagKey         = semconv.FeatureFlagKey
	AttributeProviderName    = semconv.FeatureFlagProviderName("Flipt")
	AttributeFlagVariant     = semconv.FeatureFlagResultVariant
	AttributeFlagTypeVariant = AttributeFlagType.String(evaluation.EvaluationFlagType_VARIANT_FLAG_TYPE.String())
	AttributeFlagTypeBoolean = AttributeFlagType.String(evaluation.EvaluationFlagType_BOOLEAN_FLAG_TYPE.String())
)

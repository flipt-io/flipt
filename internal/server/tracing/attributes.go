package tracing

import (
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// TODO: merge with metrics?
var (
	AttributeMatch       = attribute.Key("flipt_match")
	AttributeFlag        = attribute.Key("flipt_flag")
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
	AttributeFlagKey      = semconv.FeatureFlagKey
	AttributeProviderName = semconv.FeatureFlagProviderName("Flipt")
	AttributeFlagVariant  = semconv.FeatureFlagVariant
)

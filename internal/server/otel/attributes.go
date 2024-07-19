package otel

import (
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

var (
	AttributeMatch       = attribute.Key("flipt.match")
	AttributeFlag        = attribute.Key("flipt.flag")
	AttributeNamespace   = attribute.Key("flipt.namespace")
	AttributeFlagEnabled = attribute.Key("flipt.flag_enabled")
	AttributeSegment     = attribute.Key("flipt.segment")
	AttributeSegments    = attribute.Key("flipt.segments")
	AttributeReason      = attribute.Key("flipt.reason")
	AttributeValue       = attribute.Key("flipt.value")
	AttributeEntityID    = attribute.Key("flipt.entity_id")
	AttributeRequestID   = attribute.Key("flipt.request_id")
)

// Specific attributes for Semantic Conventions for Feature Flags in Spans
// https://opentelemetry.io/docs/specs/semconv/feature-flags/feature-flags-spans/
var (
	AttributeFlagKey      = semconv.FeatureFlagKey
	AttributeProviderName = semconv.FeatureFlagProviderName("Flipt")
	AttributeFlagVariant  = semconv.FeatureFlagVariant
)

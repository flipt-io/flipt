package otel

import "go.opentelemetry.io/otel/attribute"

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

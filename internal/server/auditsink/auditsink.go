package auditsink

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
)

// AuditType represents what resource is being acted on.
type AuditType string

// AuditAction represents the action being taken on the resource.
type AuditAction string

const (
	ConstraintType   AuditType = "constraint"
	DistributionType AuditType = "distribution"
	FlagType         AuditType = "flag"
	RuleType         AuditType = "rule"
	SegmentType      AuditType = "segment"
	VariantType      AuditType = "variant"

	CreateAction AuditAction = "created"
	DeleteAction AuditAction = "deleted"
	UpdateAction AuditAction = "updated"
)

// AuditEvent holds information that represents an audit internally.
type AuditEvent struct {
	Version  string         `json:"version"`
	Metadata MetadataConfig `json:"metadata"`
	Payload  interface{}    `json:"payload"`
}

func (ae AuditEvent) DecodeToAttributes() []attribute.KeyValue {
	akv := make([]attribute.KeyValue, 0)

	if ae.Version != "" {
		akv = append(akv, attribute.KeyValue{
			Key:   "flipt.version",
			Value: attribute.StringValue(ae.Version),
		})
	}

	if ae.Metadata.Action != "" {
		akv = append(akv, attribute.KeyValue{
			Key:   "flipt.metadata.action",
			Value: attribute.StringValue(string(ae.Metadata.Action)),
		})
	}

	if ae.Metadata.Type != "" {
		akv = append(akv, attribute.KeyValue{
			Key:   "flipt.metadata.type",
			Value: attribute.StringValue(string(ae.Metadata.Type)),
		})
	}

	if ae.Payload != nil {
		b, _ := json.Marshal(ae.Payload)
		akv = append(akv, attribute.KeyValue{
			Key:   "flipt.payload",
			Value: attribute.StringValue(string(b)),
		})
	}

	return akv
}

func deserializeHelper(s string) string {
	ss := strings.Split(s, ".")
	return ss[len(ss)-1]
}

func (ae *AuditEvent) Valid() bool {
	return ae.Version != "" && ae.Metadata.Action != "" && ae.Metadata.Type != "" && ae.Payload != nil
}

func decodeToAuditEvent(kvs []attribute.KeyValue) (*AuditEvent, error) {
	ae := new(AuditEvent)
	for _, kv := range kvs {
		terminalKey := deserializeHelper(string(kv.Key))

		if terminalKey == "version" {
			ae.Version = kv.Value.AsString()
		}

		if terminalKey == "action" {
			ae.Metadata.Action = AuditAction(kv.Value.AsString())
		}

		if terminalKey == "type" {
			ae.Metadata.Type = AuditType(kv.Value.AsString())
		}

		if terminalKey == "payload" {
			var payload interface{}
			if err := json.Unmarshal([]byte(kv.Value.AsString()), &payload); err != nil {
				return nil, err
			}
			ae.Payload = payload
		}
	}

	if !ae.Valid() {
		return nil, fmt.Errorf("audit event not valid")
	}

	return ae, nil
}

// MetadataConfig holds information of what metadata an event will contain.
type MetadataConfig struct {
	Type   AuditType   `json:"type"`
	Action AuditAction `json:"action"`
}

// AuditSink is the abstraction for various audit sink configurations
// that Flipt will support.
type AuditSink interface {
	SendAudits([]AuditEvent) error
	fmt.Stringer
}

type SinkSpanExporter struct {
	sinks  []AuditSink
	logger *zap.Logger
}

func NewSinkSpanExporter(logger *zap.Logger, sinks []AuditSink) *SinkSpanExporter {
	return &SinkSpanExporter{
		sinks:  sinks,
		logger: logger,
	}
}

// ExportSpans completes one part of the implementation of a SpanExporter. Decodes span events to audit events.
func (s *SinkSpanExporter) ExportSpans(ctx context.Context, spans []trace.ReadOnlySpan) error {
	aes := make([]AuditEvent, 0)

	for _, span := range spans {
		events := span.Events()
		for _, e := range events {
			ae, err := decodeToAuditEvent(e.Attributes)
			if err != nil {
				s.logger.Debug("audit event not decodable")
				continue
			}
			aes = append(aes, *ae)
		}
	}

	return s.SendAudits(aes)
}

func (s *SinkSpanExporter) Shutdown(ctx context.Context) error {
	return nil
}

// SendAudits wraps the methods of sending audits to various sinks.
func (s *SinkSpanExporter) SendAudits(aes []AuditEvent) error {
	for _, sink := range s.sinks {
		err := sink.SendAudits(aes)
		if err != nil {
			s.logger.Debug("failed to send audits to sink", zap.Stringer("sink", sink))
		}
	}

	return nil
}

// NewAuditEvent is the constructor for an audit event.
func NewAuditEvent(auditType AuditType, auditAction AuditAction, payload interface{}, auditEventVersion string) *AuditEvent {
	return &AuditEvent{
		Version: auditEventVersion,
		Metadata: MetadataConfig{
			Type:   auditType,
			Action: auditAction,
		},
		Payload: payload,
	}
}

package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
)

const eventVersion = "v0.1"

// Type represents what resource is being acted on.
type Type string

// Action represents the action being taken on the resource.
type Action string

const (
	Constraint   Type = "constraint"
	Distribution Type = "distribution"
	Flag         Type = "flag"
	Namespace    Type = "namespace"
	Rule         Type = "rule"
	Segment      Type = "segment"
	Variant      Type = "variant"

	Create Action = "created"
	Delete Action = "deleted"
	Update Action = "updated"
)

// Event holds information that represents an audit internally.
type Event struct {
	Version  string         `json:"version"`
	Metadata MetadataConfig `json:"metadata"`
	Payload  interface{}    `json:"payload"`
}

// DecodeToAttributes provides a helper method for an Event that will return
// a value compatible to a SpanEvent.
func (e Event) DecodeToAttributes() []attribute.KeyValue {
	akv := make([]attribute.KeyValue, 0)

	if e.Version != "" {
		akv = append(akv, attribute.KeyValue{
			Key:   "flipt.event.version",
			Value: attribute.StringValue(e.Version),
		})
	}

	if e.Metadata.Action != "" {
		akv = append(akv, attribute.KeyValue{
			Key:   "flipt.event.metadata.action",
			Value: attribute.StringValue(string(e.Metadata.Action)),
		})
	}

	if e.Metadata.Type != "" {
		akv = append(akv, attribute.KeyValue{
			Key:   "flipt.event.metadata.type",
			Value: attribute.StringValue(string(e.Metadata.Type)),
		})
	}

	if e.Payload != nil {
		b, err := json.Marshal(e.Payload)
		if err == nil {
			akv = append(akv, attribute.KeyValue{
				Key:   "flipt.event.payload",
				Value: attribute.StringValue(string(b)),
			})
		}
	}

	return akv
}

func deserializeHelper(s string) string {
	ss := strings.Split(s, ".")
	return ss[len(ss)-1]
}

func (e *Event) Valid() bool {
	return e.Version != "" && e.Metadata.Action != "" && e.Metadata.Type != "" && e.Payload != nil
}

// decodeToEvent provides helper logic for turning to value of SpanEvents to
// an Event.
func decodeToEvent(kvs []attribute.KeyValue) (*Event, error) {
	e := new(Event)
	for _, kv := range kvs {
		terminalKey := deserializeHelper(string(kv.Key))

		switch terminalKey {
		case "version":
			e.Version = kv.Value.AsString()
		case "action":
			e.Metadata.Action = Action(kv.Value.AsString())
		case "type":
			e.Metadata.Type = Type(kv.Value.AsString())
		case "payload":
			var payload interface{}
			if err := json.Unmarshal([]byte(kv.Value.AsString()), &payload); err != nil {
				return nil, err
			}
			e.Payload = payload
		}
	}

	if !e.Valid() {
		return nil, fmt.Errorf("audit event not valid")
	}

	return e, nil
}

// MetadataConfig holds information of what metadata an event will contain.
type MetadataConfig struct {
	Type   Type   `json:"type"`
	Action Action `json:"action"`
}

// Sink is the abstraction for various audit sink configurations
// that Flipt will support.
type Sink interface {
	SendAudits([]Event) error
	Close() error
	fmt.Stringer
}

// SinkSpanExporter sends audit logs to configured sinks through intercepting span events.
type SinkSpanExporter struct {
	sinks  []Sink
	logger *zap.Logger
}

// EventExporter provides an API for exporting spans as Event(s).
type EventExporter interface {
	ExportSpans(ctx context.Context, spans []trace.ReadOnlySpan) error
	Shutdown(ctx context.Context) error
	SendAudits(es []Event) error
}

// NewSinkSpanExporter is the constructor for a SinkSpanExporter.
func NewSinkSpanExporter(logger *zap.Logger, sinks []Sink) EventExporter {
	return &SinkSpanExporter{
		sinks:  sinks,
		logger: logger,
	}
}

// ExportSpans completes one part of the implementation of a SpanExporter. Decodes span events to audit events.
func (s *SinkSpanExporter) ExportSpans(ctx context.Context, spans []trace.ReadOnlySpan) error {
	es := make([]Event, 0)

	for _, span := range spans {
		events := span.Events()
		for _, e := range events {
			e, err := decodeToEvent(e.Attributes)
			if err != nil {
				s.logger.Debug("audit event not decodable", zap.Any("audit event", e), zap.Error(err))
				continue
			}
			es = append(es, *e)
		}
	}

	return s.SendAudits(es)
}

// Shutdown will close all the registered sinks.
func (s *SinkSpanExporter) Shutdown(ctx context.Context) error {
	var result error

	for _, sink := range s.sinks {
		err := sink.Close()
		if err != nil {
			result = multierror.Append(result, err)
		}
	}

	return result
}

// SendAudits wraps the methods of sending audits to various sinks.
func (s *SinkSpanExporter) SendAudits(es []Event) error {
	for _, sink := range s.sinks {
		err := sink.SendAudits(es)
		if err != nil {
			s.logger.Debug("failed to send audits to sink", zap.Stringer("sink", sink))
		}
	}

	return nil
}

// NewEvent is the constructor for an audit event.
func NewEvent(auditType Type, action Action, payload interface{}) *Event {
	return &Event{
		Version: eventVersion,
		Metadata: MetadataConfig{
			Type:   auditType,
			Action: action,
		},
		Payload: payload,
	}
}

package audit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

const (
	eventVersion          = "0.1"
	eventVersionKey       = "flipt.event.version"
	eventActionKey        = "flipt.event.action"
	eventTypeKey          = "flipt.event.type"
	eventMetadataActorKey = "flipt.event.metadata.actor"
	eventPayloadKey       = "flipt.event.payload"
	eventTimestampKey     = "flipt.event.timestamp"
)

// Type represents what resource is being acted on.
type Type string

// Action represents the action being taken on the resource.
type Action string

const (
	ConstraintType   Type = "constraint"
	DistributionType Type = "distribution"
	FlagType         Type = "flag"
	NamespaceType    Type = "namespace"
	RolloutType      Type = "rollout"
	RuleType         Type = "rule"
	SegmentType      Type = "segment"
	TokenType        Type = "token"
	VariantType      Type = "variant"

	Create Action = "created"
	Delete Action = "deleted"
	Update Action = "updated"
)

// Event holds information that represents an audit internally.
type Event struct {
	Version string `json:"version"`
	Type    Type   `json:"type"`
	Action  Action `json:"action"`

	Metadata Metadata `json:"metadata"`

	Payload interface{} `json:"payload"`

	Timestamp string `json:"timestamp"`
}

// GRPCMethodToAction returns the Action from the gRPC method.
func GRPCMethodToAction(method string) Action {
	if strings.Contains(method, "Create") {
		return Create
	} else if strings.Contains(method, "Update") {
		return Update
	}

	return ""
}

// DecodeToAttributes provides a helper method for an Event that will return
// a value compatible to a SpanEvent.
func (e Event) DecodeToAttributes() []attribute.KeyValue {
	akv := make([]attribute.KeyValue, 0)

	if e.Version != "" {
		akv = append(akv, attribute.KeyValue{
			Key:   eventVersionKey,
			Value: attribute.StringValue(e.Version),
		})
	}

	if e.Action != "" {
		akv = append(akv, attribute.KeyValue{
			Key:   eventActionKey,
			Value: attribute.StringValue(string(e.Action)),
		})
	}

	if e.Type != "" {
		akv = append(akv, attribute.KeyValue{
			Key:   eventTypeKey,
			Value: attribute.StringValue(string(e.Type)),
		})
	}

	if e.Timestamp != "" {
		akv = append(akv, attribute.KeyValue{
			Key:   eventTimestampKey,
			Value: attribute.StringValue(e.Timestamp),
		})
	}

	b, err := json.Marshal(e.Metadata.Actor)
	if err == nil {
		akv = append(akv, attribute.KeyValue{
			Key:   eventMetadataActorKey,
			Value: attribute.StringValue(string(b)),
		})
	}

	if e.Payload != nil {
		b, err := json.Marshal(e.Payload)
		if err == nil {
			akv = append(akv, attribute.KeyValue{
				Key:   eventPayloadKey,
				Value: attribute.StringValue(string(b)),
			})
		}
	}

	return akv
}

func (e *Event) AddToSpan(ctx context.Context) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent("event", trace.WithAttributes(e.DecodeToAttributes()...))
}

func (e *Event) Valid() bool {
	return e.Version != "" && e.Action != "" && e.Type != "" && e.Timestamp != "" && e.Payload != nil
}

var errEventNotValid = errors.New("audit event not valid")

// decodeToEvent provides helper logic for turning to value of SpanEvents to
// an Event.
func decodeToEvent(kvs []attribute.KeyValue) (*Event, error) {
	e := new(Event)
	for _, kv := range kvs {
		switch string(kv.Key) {
		case eventVersionKey:
			e.Version = kv.Value.AsString()
		case eventActionKey:
			e.Action = Action(kv.Value.AsString())
		case eventTypeKey:
			e.Type = Type(kv.Value.AsString())
		case eventTimestampKey:
			e.Timestamp = kv.Value.AsString()
		case eventMetadataActorKey:
			var actor map[string]string
			if err := json.Unmarshal([]byte(kv.Value.AsString()), &actor); err != nil {
				return nil, err
			}
			e.Metadata.Actor = actor
		case eventPayloadKey:
			var payload interface{}
			if err := json.Unmarshal([]byte(kv.Value.AsString()), &payload); err != nil {
				return nil, err
			}
			e.Payload = payload
		}
	}

	if !e.Valid() {
		return nil, errEventNotValid
	}

	return e, nil
}

// Metadata holds information of what metadata an event will contain.
type Metadata struct {
	Actor map[string]string `json:"actor,omitempty"`
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
	ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error
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
func (s *SinkSpanExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	es := make([]Event, 0)

	for _, span := range spans {
		events := span.Events()
		for _, e := range events {
			e, err := decodeToEvent(e.Attributes)
			if err != nil {
				if !errors.Is(err, errEventNotValid) {
					s.logger.Error("audit event not decodable", zap.Error(err))
				}
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
	if len(es) < 1 {
		return nil
	}

	for _, sink := range s.sinks {
		s.logger.Debug("performing batched sending of audit events", zap.Stringer("sink", sink), zap.Int("batch size", len(es)))
		err := sink.SendAudits(es)
		if err != nil {
			s.logger.Debug("failed to send audits to sink", zap.Stringer("sink", sink))
		}
	}

	return nil
}

// NewEvent is the constructor for an audit event.
func NewEvent(eventType Type, action Action, actor map[string]string, payload interface{}) *Event {
	return &Event{
		Version: eventVersion,
		Action:  action,
		Type:    eventType,
		Metadata: Metadata{
			Actor: actor,
		},
		Payload:   payload,
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

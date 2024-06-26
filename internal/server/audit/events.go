package audit

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"go.flipt.io/flipt/rpc/flipt"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap/zapcore"
)

const (
	eventVersion          = "0.2"
	eventVersionKey       = "flipt.event.version"
	eventActionKey        = "flipt.event.action"
	eventTypeKey          = "flipt.event.type"
	eventMetadataActorKey = "flipt.event.metadata.actor"
	eventPayloadKey       = "flipt.event.payload"
	eventTimestampKey     = "flipt.event.timestamp"
	eventStatusKey        = "flipt.event.status"
)

// Event holds information that represents an action that was attempted in the system.
type Event struct {
	Version string `json:"version" avro:"version"`

	Type string `json:"type" avro:"type"`

	Action string `json:"action" avro:"action"`

	Metadata Metadata `json:"metadata" avro:"metadata"`

	Payload interface{} `json:"payload" avro:"payload"`

	Timestamp string `json:"timestamp" avro:"timestamp"`

	Status string `json:"status" avro:"status"`
}

// NewEvent is the constructor for an event.
func NewEvent(r flipt.Request, actor *Actor, payload interface{}) *Event {
	var action string

	switch r.Action {
	case flipt.ActionCreate:
		action = "created"
	case flipt.ActionUpdate:
		action = "updated"
	case flipt.ActionDelete:
		action = "deleted"
	default:
		action = "read"
	}

	typ := string(r.Resource)
	if r.Subject != "" {
		typ = string(r.Subject)
	}

	return &Event{
		Version: eventVersion,
		Action:  action,
		Type:    typ,
		Status:  string(r.Status),
		Metadata: Metadata{
			Actor: actor,
		},
		Payload:   payload,
		Timestamp: time.Now().Format(time.RFC3339),
	}
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
			Value: attribute.StringValue(e.Action),
		})
	}

	if e.Type != "" {
		akv = append(akv, attribute.KeyValue{
			Key:   eventTypeKey,
			Value: attribute.StringValue(e.Type),
		})
	}

	if e.Timestamp != "" {
		akv = append(akv, attribute.KeyValue{
			Key:   eventTimestampKey,
			Value: attribute.StringValue(e.Timestamp),
		})
	}

	if e.Status != "" {
		akv = append(akv, attribute.KeyValue{
			Key:   eventStatusKey,
			Value: attribute.StringValue(e.Status),
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
	return e.Version != "" && e.Action != "" && e.Type != "" && e.Timestamp != ""
}

func (e Event) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("version", e.Version)
	enc.AddString("type", e.Type)
	enc.AddString("action", e.Action)
	enc.AddString("status", e.Status)
	if err := enc.AddReflected("metadata", e.Metadata); err != nil {
		return err
	}
	if err := enc.AddReflected("payload", e.Payload); err != nil {
		return err
	}
	enc.AddString("timestamp", e.Timestamp)
	return nil
}

func (e *Event) CopyInto(out *Event) {
	*out = *e
	out.Metadata = e.Metadata
	out.Payload = e.Payload
}

func (e *Event) PayloadToMap() (map[string]any, error) {
	if e.Payload == nil {
		return map[string]any{}, nil
	}
	payloadString, err := json.Marshal(e.Payload)
	if err != nil {
		return nil, err
	}
	var payload map[string]any
	err = json.Unmarshal(payloadString, &payload)
	if err != nil {
		return nil, err
	}
	return payload, nil
}

var errEventNotValid = errors.New("event not valid")

// decodeToEvent provides helper logic for turning to value of SpanEvents to
// an Event.
func decodeToEvent(kvs []attribute.KeyValue) (*Event, error) {
	e := new(Event)
	for _, kv := range kvs {
		switch string(kv.Key) {
		case eventVersionKey:
			e.Version = kv.Value.AsString()
		case eventActionKey:
			e.Action = kv.Value.AsString()
		case eventTypeKey:
			e.Type = kv.Value.AsString()
		case eventTimestampKey:
			e.Timestamp = kv.Value.AsString()
		case eventStatusKey:
			e.Status = kv.Value.AsString()
		case eventMetadataActorKey:
			var actor Actor
			if err := json.Unmarshal([]byte(kv.Value.AsString()), &actor); err != nil {
				return nil, err
			}
			e.Metadata.Actor = &actor
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

type Actor struct {
	Authentication string `json:"authentication,omitempty" avro:"authentication"`
	IP             string `json:"ip,omitempty" avro:"ip"`
	Email          string `json:"email,omitempty" avro:"email"`
	Name           string `json:"name,omitempty" avro:"name"`
	Picture        string `json:"picture,omitempty" avro:"picture"`
}

// Metadata holds information of what metadata an event will contain.
type Metadata struct {
	Actor *Actor `json:"actor,omitempty" avro:"actor"`
}

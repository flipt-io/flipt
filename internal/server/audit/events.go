package audit

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"go.flipt.io/flipt/rpc/flipt"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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

// Event holds information that represents an action that has occurred in the system.
type Event struct {
	Version string `json:"version"`

	Type flipt.Subject `json:"type"`

	Action flipt.Action `json:"action"`

	Metadata Metadata `json:"metadata"`

	Payload interface{} `json:"payload"`

	Timestamp string `json:"timestamp"`
}

// NewEvent is the constructor for an event.
func NewEvent(r flipt.Request, actor *Actor, payload interface{}) *Event {
	return &Event{
		Version: eventVersion,
		Action:  r.Action,
		Type:    r.Subject,
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
			e.Action = flipt.Action(kv.Value.AsString())
		case eventTypeKey:
			e.Type = flipt.Subject(kv.Value.AsString())
		case eventTimestampKey:
			e.Timestamp = kv.Value.AsString()
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
	Authentication string `json:"authentication,omitempty"`
	IP             string `json:"ip,omitempty"`
	Email          string `json:"email,omitempty"`
	Name           string `json:"name,omitempty"`
	Picture        string `json:"picture,omitempty"`
}

// Metadata holds information of what metadata an event will contain.
type Metadata struct {
	Actor *Actor `json:"actor,omitempty"`
}

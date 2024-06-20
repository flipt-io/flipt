package kafka

import (
	"encoding/json"
	"errors"
	"time"

	"go.flipt.io/flipt/internal/server/audit"
	protoaudit "go.flipt.io/flipt/rpc/flipt/audit"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func toProtobuf(v any) ([]byte, error) {
	if e, ok := v.(audit.Event); ok {
		r := &protoaudit.Event{
			Version: e.Version,
			Action:  e.Action,
			Type:    e.Type,
			Status:  &e.Status,
			Metadata: &protoaudit.Metadata{
				Actor: &protoaudit.Actor{},
			},
		}
		if a := e.Metadata.Actor; a != nil {
			r.Metadata.Actor.Authentication = a.Authentication
			r.Metadata.Actor.Email = a.Email
			r.Metadata.Actor.Ip = a.IP
			r.Metadata.Actor.Name = a.Name
			r.Metadata.Actor.Picture = a.Picture
		}
		t, err := time.Parse(time.RFC3339, e.Timestamp)
		if err != nil {
			t = time.Now()
		}
		r.Timestamp = timestamppb.New(t)

		if e.Payload != nil {
			// FIXME: payload should be added here
			payload, err := json.Marshal(e.Payload)
			if err != nil {
				return nil, err
			}
			r.Payload = proto.String(string(payload))
		}

		b, err := proto.Marshal(r)
		return b, err
	}
	return nil, errors.New("unsupported struct")
}

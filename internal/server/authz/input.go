package authz

import (
	"go.flipt.io/flipt/rpc/flipt"
	authrpc "go.flipt.io/flipt/rpc/flipt/auth"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Input builds the policy evaluation input for a request and authentication.
//
// It uses only plain Go values (maps, strings and numbers) so policy engines
// can consume it without converting arbitrary Go structs. In particular this
// avoids a JSON round-trip per evaluation for the request and authentication
// values, which otherwise happens on every call.
func Input(request flipt.Request, auth *authrpc.Authentication) map[string]any {
	return map[string]any{
		"request":        RequestInput(request),
		"authentication": AuthenticationInput(auth),
	}
}

// RequestInput converts a request to plain Go values. The keys match the
// JSON field names of the request: unset environment and namespace are
// omitted.
func RequestInput(request flipt.Request) map[string]any {
	input := map[string]any{
		"scope":  string(request.Scope),
		"action": string(request.Action),
	}

	if request.Environment != nil {
		input["environment"] = *request.Environment
	}

	if request.Namespace != nil {
		input["namespace"] = *request.Namespace
	}

	return input
}

// AuthenticationInput converts an authentication to plain Go values. The keys
// match the JSON field names of the authentication: the metadata map is kept
// as is, while unset timestamps are omitted.
func AuthenticationInput(auth *authrpc.Authentication) map[string]any {
	input := map[string]any{
		"id":       auth.GetId(),
		"method":   int(auth.GetMethod()),
		"metadata": auth.GetMetadata(),
	}

	if expiresAt := auth.GetExpiresAt(); expiresAt != nil {
		input["expires_at"] = timestampInput(expiresAt)
	}

	if createdAt := auth.GetCreatedAt(); createdAt != nil {
		input["created_at"] = timestampInput(createdAt)
	}

	if updatedAt := auth.GetUpdatedAt(); updatedAt != nil {
		input["updated_at"] = timestampInput(updatedAt)
	}

	return input
}

// timestampInput converts a timestamp to plain Go values, matching its JSON
// representation: unset fields are omitted.
func timestampInput(ts *timestamppb.Timestamp) map[string]any {
	input := map[string]any{}

	if ts.GetSeconds() != 0 {
		input["seconds"] = float64(ts.GetSeconds())
	}

	if ts.GetNanos() != 0 {
		input["nanos"] = float64(ts.GetNanos())
	}

	return input
}

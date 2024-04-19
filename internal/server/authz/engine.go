package authz

import (
	"context"
	"encoding/json"
	"os"

	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/storage/inmem"
)

type Engine struct {
	query rego.PreparedEvalQuery
}

var defaultPolicy = `package authz.v1
import data

default allow = false

allow {
    input.role != ""
    input.action != ""
    input.subject != ""

    permissions := get_permissions(input.role)
    allowed(permissions, input.action, input.subject)
}

get_permissions(role) = result {
    some idx
    data.roles[idx].name = role
    result = data.roles[idx].rules
}

allowed(permissions, action, subject) {
    permissions[action]  # First, ensure the action key exists
    subject_in_list(permissions[action], subject)  # Check if the subject is in the list
}

allowed(permissions, action, subject) {
    permissions[action]["*"]  # Checks if all subjects are allowed for the action
}

# Handles cases where "*" is provided for all actions
allowed(permissions, action, subject) {
    permissions["*"] != null  # Check if wildcard for all actions exists
    subject_in_list(permissions["*"], subject)  # Check if subject is universally allowed or specific to an action
}

# Helper to handle array membership or wildcard
subject_in_list(list, subject) {
    list[_] = subject  # Subject is explicitly listed in the permissions
}

subject_in_list(list, subject) {
    list[_] = "*"  # Wildcard entry that permits all subjects
}
`

func NewEngine(ctx context.Context) (*Engine, error) {
	data := map[string]interface{}{}

	file, err := os.ReadFile("./policies/default.json")
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(file, &data); err != nil {
		return nil, err
	}

	store := inmem.NewFromObject(data)

	r := rego.New(
		rego.Query("data.authz.v1.allow"),
		rego.Module("policy.rego", defaultPolicy),
		rego.Store(store),
	)

	query, err := r.PrepareForEval(ctx)
	if err != nil {
		return nil, err
	}

	return &Engine{query: query}, nil
}

func (e *Engine) IsAllowed(ctx context.Context, input map[string]interface{}) (bool, error) {
	results, err := e.query.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return false, err
	}

	if len(results) == 0 {
		return false, nil
	}

	return results[0].Expressions[0].Value.(bool), nil
}

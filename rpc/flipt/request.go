package flipt

const (
	DefaultEnvironment = "default"
	DefaultNamespace   = "default"
)

type Requester interface {
	Request() []Request
}

// ScopeType represents the level at which the action is being performed
type ScopeType string

const (
	ScopeEnvironment ScopeType = "environment" // Operations within an environment
	ScopeNamespace   ScopeType = "namespace"   // Operations within a namespace
)

// Resource represents what resource or parent resource is being acted on.
type Resource string

// Action represents the action being taken on the resource.
type Action string

const (

	// Actions that can be performed
	ActionCreate Action = "create"
	ActionRead   Action = "read"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
)

// Request represents an authorization request
type Request struct {
	// Scope represents the level at which the action is being performed
	Scope ScopeType `json:"scope,omitempty"`
	// Environment is the environment in which the action is being performed
	Environment *string `json:"environment,omitempty"`
	// Namespace is the namespace in which the action is being performed
	Namespace *string `json:"namespace,omitempty"`
	// Action is the action being performed
	Action Action `json:"action,omitempty"`
}

// NewRequest creates a new Request with the given parameters
func NewRequest(scope ScopeType, a Action, opts ...func(*Request)) Request {
	req := Request{
		Scope:       scope,
		Environment: ptr(DefaultEnvironment),
		Action:      a,
	}

	if scope == ScopeNamespace {
		req.Namespace = ptr(DefaultNamespace)
	}

	for _, opt := range opts {
		opt(&req)
	}

	return req
}

// Option functions for Request

func WithEnvironment(env string) func(*Request) {
	return func(r *Request) {
		if env != "" {
			r.Environment = &env
		}
	}
}

func WithNamespace(ns string) func(*Request) {
	return func(r *Request) {
		if ns != "" {
			r.Namespace = &ns
		}
	}
}

func WithNoEnvironment() func(*Request) {
	return func(r *Request) {
		r.Environment = nil
	}
}

func WithNoNamespace() func(*Request) {
	return func(r *Request) {
		r.Namespace = nil
	}
}

func ptr[T any](v T) *T {
	return &v
}

func (req *ListFlagRequest) Request() []Request {
	return []Request{
		// Reading any resource in a namespace
		NewRequest(ScopeNamespace, ActionRead,
			WithEnvironment(req.EnvironmentKey),
			WithNamespace(req.NamespaceKey)),
	}
}

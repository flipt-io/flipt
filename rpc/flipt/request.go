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
	ScopeGlobal    ScopeType = "global"    // Global operations (full access)
	ScopeNamespace ScopeType = "namespace" // Namespace management
	ScopeResource  ScopeType = "resource"  // Operations within a namespace
)

// Resource represents what resource or parent resource is being acted on.
type Resource string

// Action represents the action being taken on the resource.
type Action string

// Status represents the status of the request.
type Status string

const (
	ResourceUnknown Resource = "-"

	// Core resources that can be managed
	ResourceEnvironment Resource = "environment" // Environment listing/reading
	ResourceNamespace   Resource = "namespace"   // Namespace management
	ResourceAny         Resource = "*"           // Any resource within a namespace

	// Actions that can be performed
	ActionCreate Action = "create"
	ActionRead   Action = "read"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"

	StatusSuccess Status = "success"
	StatusDenied  Status = "denied"
)

// Request represents an authorization request
type Request struct {
	// Scope represents the level at which the action is being performed
	Scope ScopeType `json:"scope,omitempty"`
	// Environment is the environment in which the action is being performed
	Environment *string `json:"environment,omitempty"`
	// Namespace is the namespace in which the action is being performed
	Namespace *string `json:"namespace,omitempty"`
	// Resource is the resource being acted upon
	Resource Resource `json:"resource,omitempty"`
	// Action is the action being performed
	Action Action `json:"action,omitempty"`
	// Status is the result of the authorization check
	Status *Status `json:"status,omitempty"`
}

// NewRequest creates a new Request with the given parameters
func NewRequest(scope ScopeType, r Resource, a Action, opts ...func(*Request)) Request {
	req := Request{
		Scope:    scope,
		Resource: r,
		Action:   a,
	}

	// Only set default environment and namespace for resource-level operations
	if scope == ScopeResource {
		req.Environment = ptr(DefaultEnvironment)
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

func WithStatus(s Status) func(*Request) {
	return func(r *Request) {
		r.Status = &s
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

// Example request implementation
func (req *ListFlagRequest) Request() []Request {
	return []Request{
		// Reading any resource in a namespace
		NewRequest(ScopeResource, ResourceAny, ActionRead,
			WithEnvironment(req.EnvironmentKey),
			WithNamespace(req.NamespaceKey)),
	}
}

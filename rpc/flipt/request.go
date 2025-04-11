package flipt

const (
	DefaultEnvironment = "default"
	DefaultNamespace   = "default"
)

type Requester interface {
	Request() []Request
}

// Resource represents what resource or parent resource is being acted on.
type Resource string

// Subject returns the subject of the request.
type Subject string

// Action represents the action being taken on the resource.
type Action string

// Status represents the status of the request.
type Status string

const (
	ResourceUnknown Resource = "-"

	ResourceNamespace   Resource = "namespace"
	ResourceEnvironment Resource = "environment"

	// TODO: remove these and subjects
	ResourceFlag           Resource = "flag"
	ResourceSegment        Resource = "segment"
	ResourceAuthentication Resource = "authentication"

	ActionCreate Action = "create"
	ActionDelete Action = "delete"
	ActionUpdate Action = "update"
	ActionRead   Action = "read"

	StatusSuccess Status = "success"
	StatusDenied  Status = "denied"
)

type Request struct {
	Namespace string   `json:"namespace,omitempty"`
	Resource  Resource `json:"resource,omitempty"`
	Action    Action   `json:"action,omitempty"`
	Status    Status   `json:"status,omitempty"`
}

func WithNoNamespace() func(*Request) {
	return func(r *Request) {
		r.Namespace = ""
	}
}

func WithNamespace(ns string) func(*Request) {
	return func(r *Request) {
		if ns != "" {
			r.Namespace = ns
		}
	}
}

func WithStatus(s Status) func(*Request) {
	return func(r *Request) {
		r.Status = s
	}
}

func NewRequest(r Resource, a Action, opts ...func(*Request)) Request {
	req := Request{
		Resource:  r,
		Action:    a,
		Status:    StatusSuccess,
		Namespace: DefaultNamespace,
	}

	for _, opt := range opts {
		opt(&req)
	}

	return req
}

func (req *ListFlagRequest) Request() []Request {
	return []Request{NewRequest(ResourceFlag, ActionRead, WithNamespace(req.NamespaceKey))}
}

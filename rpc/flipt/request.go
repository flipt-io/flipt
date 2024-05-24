package flipt

type Requester interface {
	Request() Request
}

// Resource represents what resource or parent resource is being acted on.
type Resource string

// Subject returns the subject of the request.
type Subject string

// Action represents the action being taken on the resource.
type Action string

const (
	ResourceNamespace Resource = "namespace"
	ResourceFlag      Resource = "flag"
	ResourceSegment   Resource = "segment"

	SubjectConstraint   Subject = "constraint"
	SubjectDistribution Subject = "distribution"
	SubjectFlag         Subject = "flag"
	SubjectNamespace    Subject = "namespace"
	SubjectRollout      Subject = "rollout"
	SubjectRule         Subject = "rule"
	SubjectSegment      Subject = "segment"
	SubjectToken        Subject = "token"
	SubjectVariant      Subject = "variant"

	ActionCreate Action = "create"
	ActionDelete Action = "delete"
	ActionUpdate Action = "update"
	ActionRead   Action = "read"
)

type Request struct {
	Namespaced
	Resource Resource `json:"resource"`
	Subject  Subject  `json:"subject"`
	Action   Action   `json:"action"`
}

func NewResourceScopedRequest(r Resource, s Subject, a Action) Request {
	return Request{
		Resource: r,
		Subject:  s,
		Action:   a,
	}
}

func NewRequest(s Subject, a Action) Request {
	return Request{
		Subject: s,
		Action:  a,
	}
}

func newFlagScopedRequest(s Subject, a Action) Request {
	return NewResourceScopedRequest(ResourceFlag, s, a)
}

func newSegmentScopedRequest(s Subject, a Action) Request {
	return NewResourceScopedRequest(ResourceSegment, s, a)
}

// Namespaces
func (req *GetNamespaceRequest) Request() Request {
	return NewRequest(SubjectNamespace, ActionRead)
}

func (req *ListNamespaceRequest) Request() Request {
	return NewRequest(SubjectNamespace, ActionRead)
}

func (req *CreateNamespaceRequest) Request() Request {
	return NewRequest(SubjectNamespace, ActionCreate)
}

func (req *UpdateNamespaceRequest) Request() Request {
	return NewRequest(SubjectNamespace, ActionUpdate)
}

func (req *DeleteNamespaceRequest) Request() Request {
	return NewRequest(SubjectNamespace, ActionDelete)
}

// Flags
func (req *GetFlagRequest) Request() Request {
	return NewRequest(SubjectFlag, ActionRead)
}

func (req *ListFlagRequest) Request() Request {
	return NewRequest(SubjectFlag, ActionRead)
}

func (req *CreateFlagRequest) Request() Request {
	return NewRequest(SubjectFlag, ActionCreate)
}

func (req *UpdateFlagRequest) Request() Request {
	return NewRequest(SubjectFlag, ActionUpdate)
}

func (req *DeleteFlagRequest) Request() Request {
	return NewRequest(SubjectFlag, ActionDelete)
}

// Variants
func (req *CreateVariantRequest) Request() Request {
	return newFlagScopedRequest(SubjectVariant, ActionCreate)
}

func (req *UpdateVariantRequest) Request() Request {
	return newFlagScopedRequest(SubjectVariant, ActionUpdate)
}

func (req *DeleteVariantRequest) Request() Request {
	return newFlagScopedRequest(SubjectVariant, ActionDelete)
}

// Rules
func (req *ListRuleRequest) Request() Request {
	return newFlagScopedRequest(SubjectRule, ActionRead)
}

func (req *GetRuleRequest) Request() Request {
	return newFlagScopedRequest(SubjectRule, ActionRead)
}

func (req *CreateRuleRequest) Request() Request {
	return newFlagScopedRequest(SubjectRule, ActionCreate)
}

func (req *UpdateRuleRequest) Request() Request {
	return newFlagScopedRequest(SubjectRule, ActionUpdate)
}

func (req *OrderRulesRequest) Request() Request {
	return newFlagScopedRequest(SubjectRule, ActionUpdate)
}

func (req *DeleteRuleRequest) Request() Request {
	return newFlagScopedRequest(SubjectRule, ActionDelete)
}

// Rollouts
func (req *ListRolloutRequest) Request() Request {
	return newFlagScopedRequest(SubjectRollout, ActionRead)
}

func (req *GetRolloutRequest) Request() Request {
	return newFlagScopedRequest(SubjectRollout, ActionRead)
}

func (req *CreateRolloutRequest) Request() Request {
	return newFlagScopedRequest(SubjectRollout, ActionCreate)
}

func (req *UpdateRolloutRequest) Request() Request {
	return newFlagScopedRequest(SubjectRollout, ActionUpdate)
}

func (req *OrderRolloutsRequest) Request() Request {
	return newFlagScopedRequest(SubjectRollout, ActionUpdate)
}

func (req *DeleteRolloutRequest) Request() Request {
	return newFlagScopedRequest(SubjectRollout, ActionDelete)
}

// Segments
func (req *GetSegmentRequest) Request() Request {
	return NewRequest(SubjectSegment, ActionRead)
}

func (req *ListSegmentRequest) Request() Request {
	return NewRequest(SubjectSegment, ActionRead)
}

func (req *CreateSegmentRequest) Request() Request {
	return NewRequest(SubjectSegment, ActionCreate)
}

func (req *UpdateSegmentRequest) Request() Request {
	return NewRequest(SubjectSegment, ActionUpdate)
}

func (req *DeleteSegmentRequest) Request() Request {
	return NewRequest(SubjectSegment, ActionDelete)
}

// Constraints
func (req *CreateConstraintRequest) Request() Request {
	return newSegmentScopedRequest(SubjectConstraint, ActionCreate)
}

func (req *UpdateConstraintRequest) Request() Request {
	return newSegmentScopedRequest(SubjectConstraint, ActionUpdate)
}

func (req *DeleteConstraintRequest) Request() Request {
	return newSegmentScopedRequest(SubjectConstraint, ActionDelete)
}

// Distributions
func (req *CreateDistributionRequest) Request() Request {
	return newSegmentScopedRequest(SubjectDistribution, ActionCreate)
}

func (req *UpdateDistributionRequest) Request() Request {
	return newSegmentScopedRequest(SubjectDistribution, ActionUpdate)
}

func (req *DeleteDistributionRequest) Request() Request {
	return newSegmentScopedRequest(SubjectDistribution, ActionDelete)
}

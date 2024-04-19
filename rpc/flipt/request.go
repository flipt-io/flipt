package flipt

type Requester interface {
	Request() Request
}

// Subject represents what resource is being acted on.
type Subject string

// Action represents the action being taken on the resource.
type Action string

const (
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
	Subject Subject `json:"subject"`
	Action  Action  `json:"action"`
}

func NewRequest(t Subject, a Action) Request {
	return Request{
		Subject: t,
		Action:  a,
	}
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
	return NewRequest(SubjectVariant, ActionCreate)
}

func (req *UpdateVariantRequest) Request() Request {
	return NewRequest(SubjectVariant, ActionUpdate)
}

func (req *DeleteVariantRequest) Request() Request {
	return NewRequest(SubjectVariant, ActionDelete)
}

// Rules
func (req *ListRuleRequest) Request() Request {
	return NewRequest(SubjectRule, ActionRead)
}

func (req *GetRuleRequest) Request() Request {
	return NewRequest(SubjectRule, ActionRead)
}

func (req *CreateRuleRequest) Request() Request {
	return NewRequest(SubjectRule, ActionCreate)
}

func (req *UpdateRuleRequest) Request() Request {
	return NewRequest(SubjectRule, ActionUpdate)
}

func (req *OrderRulesRequest) Request() Request {
	return NewRequest(SubjectRule, ActionUpdate)
}

func (req *DeleteRuleRequest) Request() Request {
	return NewRequest(SubjectRule, ActionDelete)
}

// Rollouts
func (req *ListRolloutRequest) Request() Request {
	return NewRequest(SubjectRollout, ActionRead)
}

func (req *GetRolloutRequest) Request() Request {
	return NewRequest(SubjectRollout, ActionRead)
}

func (req *CreateRolloutRequest) Request() Request {
	return NewRequest(SubjectRollout, ActionCreate)
}

func (req *UpdateRolloutRequest) Request() Request {
	return NewRequest(SubjectRollout, ActionUpdate)
}

func (req *OrderRolloutsRequest) Request() Request {
	return NewRequest(SubjectRollout, ActionUpdate)
}

func (req *DeleteRolloutRequest) Request() Request {
	return NewRequest(SubjectRollout, ActionDelete)
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
	return NewRequest(SubjectConstraint, ActionCreate)
}

func (req *UpdateConstraintRequest) Request() Request {
	return NewRequest(SubjectConstraint, ActionUpdate)
}

func (req *DeleteConstraintRequest) Request() Request {
	return NewRequest(SubjectConstraint, ActionDelete)
}

// Distributions
func (req *CreateDistributionRequest) Request() Request {
	return NewRequest(SubjectDistribution, ActionCreate)
}

func (req *UpdateDistributionRequest) Request() Request {
	return NewRequest(SubjectDistribution, ActionUpdate)
}

func (req *DeleteDistributionRequest) Request() Request {
	return NewRequest(SubjectDistribution, ActionDelete)
}

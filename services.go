package flipt

import "context"

type RuleService interface {
	Rule(ctx context.Context, r *GetRuleRequest) (*Rule, error)
	Rules(ctx context.Context, r *ListRuleRequest) ([]*Rule, error)
	CreateRule(ctx context.Context, r *CreateRuleRequest) (*Rule, error)
	UpdateRule(ctx context.Context, r *UpdateRuleRequest) (*Rule, error)
	DeleteRule(ctx context.Context, r *DeleteRuleRequest) error
	OrderRules(ctx context.Context, r *OrderRulesRequest) error
	CreateDistribution(ctx context.Context, r *CreateDistributionRequest) (*Distribution, error)
	UpdateDistribution(ctx context.Context, r *UpdateDistributionRequest) (*Distribution, error)
	DeleteDistribution(ctx context.Context, r *DeleteDistributionRequest) error
	Evaluate(ctx context.Context, r *EvaluationRequest) (*EvaluationResponse, error)
}

type FlagService interface {
	Flag(ctx context.Context, r *GetFlagRequest) (*Flag, error)
	Flags(ctx context.Context, r *ListFlagRequest) ([]*Flag, error)
	CreateFlag(ctx context.Context, r *CreateFlagRequest) (*Flag, error)
	UpdateFlag(ctx context.Context, r *UpdateFlagRequest) (*Flag, error)
	DeleteFlag(ctx context.Context, r *DeleteFlagRequest) error
	CreateVariant(ctx context.Context, r *CreateVariantRequest) (*Variant, error)
	UpdateVariant(ctx context.Context, r *UpdateVariantRequest) (*Variant, error)
	DeleteVariant(ctx context.Context, r *DeleteVariantRequest) error
}

type SegmentService interface {
	Segment(ctx context.Context, r *GetSegmentRequest) (*Segment, error)
	Segments(ctx context.Context, r *ListSegmentRequest) ([]*Segment, error)
	CreateSegment(ctx context.Context, r *CreateSegmentRequest) (*Segment, error)
	UpdateSegment(ctx context.Context, r *UpdateSegmentRequest) (*Segment, error)
	DeleteSegment(ctx context.Context, r *DeleteSegmentRequest) error
	CreateConstraint(ctx context.Context, r *CreateConstraintRequest) (*Constraint, error)
	UpdateConstraint(ctx context.Context, r *UpdateConstraintRequest) (*Constraint, error)
	DeleteConstraint(ctx context.Context, r *DeleteConstraintRequest) error
}

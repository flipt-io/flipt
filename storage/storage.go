package storage

import (
	"context"
	"fmt"

	flipt "github.com/markphelps/flipt/rpc/flipt"
)

// EvaluationRule represents a rule and constraints required for evaluating if a
// given flagKey matches a segment
type EvaluationRule struct {
	ID               string
	FlagKey          string
	SegmentKey       string
	SegmentMatchType flipt.MatchType
	Rank             int32
	Constraints      []EvaluationConstraint
}

// EvaluationConstraint represents a segment constraint that is used for evaluation
type EvaluationConstraint struct {
	ID       string
	Type     flipt.ComparisonType
	Property string
	Operator string
	Value    string
}

// EvaluationDistribution represents a rule distribution along with its variant for evaluation
type EvaluationDistribution struct {
	ID                string
	RuleID            string
	VariantID         string
	Rollout           float32
	VariantKey        string
	VariantAttachment string
}

type QueryParams struct {
	Limit  uint64
	Offset uint64
}

type QueryOption func(p *QueryParams)

func WithLimit(limit uint64) QueryOption {
	return func(p *QueryParams) {
		p.Limit = limit
	}
}

func WithOffset(offset uint64) QueryOption {
	return func(p *QueryParams) {
		p.Offset = offset
	}
}

type Store interface {
	FlagStore
	RuleStore
	SegmentStore
	EvaluationStore
	fmt.Stringer
}

// EvaluationStore returns data necessary for evaluation
type EvaluationStore interface {
	// GetEvaluationRules returns rules applicable to flagKey provided
	// Note: Rules MUST be returned in order by Rank
	GetEvaluationRules(ctx context.Context, flagKey string) ([]*EvaluationRule, error)
	GetEvaluationDistributions(ctx context.Context, ruleID string) ([]*EvaluationDistribution, error)
}

// FlagStore stores and retrieves flags and variants
type FlagStore interface {
	GetFlag(ctx context.Context, key string) (*flipt.Flag, error)
	ListFlags(ctx context.Context, opts ...QueryOption) ([]*flipt.Flag, error)
	CreateFlag(ctx context.Context, r *flipt.CreateFlagRequest) (*flipt.Flag, error)
	UpdateFlag(ctx context.Context, r *flipt.UpdateFlagRequest) (*flipt.Flag, error)
	DeleteFlag(ctx context.Context, r *flipt.DeleteFlagRequest) error
	CreateVariant(ctx context.Context, r *flipt.CreateVariantRequest) (*flipt.Variant, error)
	UpdateVariant(ctx context.Context, r *flipt.UpdateVariantRequest) (*flipt.Variant, error)
	DeleteVariant(ctx context.Context, r *flipt.DeleteVariantRequest) error
}

// SegmentStore stores and retrieves segments and constraints
type SegmentStore interface {
	GetSegment(ctx context.Context, key string) (*flipt.Segment, error)
	ListSegments(ctx context.Context, opts ...QueryOption) ([]*flipt.Segment, error)
	CreateSegment(ctx context.Context, r *flipt.CreateSegmentRequest) (*flipt.Segment, error)
	UpdateSegment(ctx context.Context, r *flipt.UpdateSegmentRequest) (*flipt.Segment, error)
	DeleteSegment(ctx context.Context, r *flipt.DeleteSegmentRequest) error
	CreateConstraint(ctx context.Context, r *flipt.CreateConstraintRequest) (*flipt.Constraint, error)
	UpdateConstraint(ctx context.Context, r *flipt.UpdateConstraintRequest) (*flipt.Constraint, error)
	DeleteConstraint(ctx context.Context, r *flipt.DeleteConstraintRequest) error
}

// RuleStore stores and retrieves rules and distributions
type RuleStore interface {
	GetRule(ctx context.Context, id string) (*flipt.Rule, error)
	ListRules(ctx context.Context, flagKey string, opts ...QueryOption) ([]*flipt.Rule, error)
	CreateRule(ctx context.Context, r *flipt.CreateRuleRequest) (*flipt.Rule, error)
	UpdateRule(ctx context.Context, r *flipt.UpdateRuleRequest) (*flipt.Rule, error)
	DeleteRule(ctx context.Context, r *flipt.DeleteRuleRequest) error
	OrderRules(ctx context.Context, r *flipt.OrderRulesRequest) error
	CreateDistribution(ctx context.Context, r *flipt.CreateDistributionRequest) (*flipt.Distribution, error)
	UpdateDistribution(ctx context.Context, r *flipt.UpdateDistributionRequest) (*flipt.Distribution, error)
	DeleteDistribution(ctx context.Context, r *flipt.DeleteDistributionRequest) error
}

package storage

import (
	"context"
	"fmt"

	"go.flipt.io/flipt/rpc/flipt"
)

const (
	// DefaultListLimit is the default limit applied to any list operation page size when one is not provided.
	DefaultListLimit uint64 = 25

	// MaxListLimit is the upper limit applied to any list operation page size.
	MaxListLimit uint64 = 100
)

// EvaluationRule represents a rule and constraints required for evaluating if a
// given flagKey matches a segment
type EvaluationRule struct {
	ID              string                        `json:"id,omitempty"`
	NamespaceKey    string                        `json:"namespace_key,omitempty"`
	FlagKey         string                        `json:"flag_key,omitempty"`
	Segments        map[string]*EvaluationSegment `json:"segments,omitempty"`
	Rank            int32                         `json:"rank,omitempty"`
	SegmentOperator flipt.SegmentOperator         `json:"segmentOperator,omitempty"`
}

type EvaluationSegment struct {
	SegmentKey  string                 `json:"segment_key,omitempty"`
	MatchType   flipt.MatchType        `json:"match_type,omitempty"`
	Constraints []EvaluationConstraint `json:"constraints,omitempty"`
}

// EvaluationRollout represents a rollout in the form that helps with evaluation.
type EvaluationRollout struct {
	NamespaceKey string            `json:"namespace_key,omitempty"`
	RolloutType  flipt.RolloutType `json:"rollout_type,omitempty"`
	Rank         int32             `json:"rank,omitempty"`
	Threshold    *RolloutThreshold `json:"threshold,omitempty"`
	Segment      *RolloutSegment   `json:"segment,omitempty"`
}

// RolloutThreshold represents Percentage(s) for use in evaluation.
type RolloutThreshold struct {
	Percentage float32 `json:"percentage,omitempty"`
	Value      bool    `json:"value,omitempty"`
}

// RolloutSegment represents Segment(s) for use in evaluation.
type RolloutSegment struct {
	Value           bool                          `json:"value,omitempty"`
	SegmentOperator flipt.SegmentOperator         `json:"segment_operator,omitempty"`
	Segments        map[string]*EvaluationSegment `json:"segments,omitempty"`
}

// EvaluationConstraint represents a segment constraint that is used for evaluation
type EvaluationConstraint struct {
	ID       string               `json:"id,omitempty"`
	Type     flipt.ComparisonType `json:"comparison_type,omitempty"`
	Property string               `json:"property,omitempty"`
	Operator string               `json:"operator,omitempty"`
	Value    string               `json:"value,omitempty"`
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
	Limit     uint64
	Offset    uint64 // deprecated
	PageToken string
	Order     Order // not exposed to the user yet
}

// Normalize adjusts query parameters within the enforced boundaries.
// For example, limit is adjusted to be in the range (0, max].
// Given the limit is not supplied (0) it is set to the default limit.
func (q *QueryParams) Normalize() {
	if q.Limit == 0 {
		q.Limit = DefaultListLimit
	}

	if q.Limit > MaxListLimit {
		q.Limit = MaxListLimit
	}
}

type QueryOption func(p *QueryParams)

func NewQueryParams(opts ...QueryOption) (params QueryParams) {
	for _, opt := range opts {
		opt(&params)
	}

	// NOTE(georgemac): I wanted to normalize under all circumstances
	// However, for legacy reasons the core flag state APIs expect
	// the default limit to be == 0. Normalize sets it to the default
	// constant which is > 0.
	// If we ever break this contract then we can normalize here.
	// params.Normalize()

	return params
}

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

func WithPageToken(pageToken string) QueryOption {
	return func(p *QueryParams) {
		p.PageToken = pageToken
	}
}

type Order uint8

const (
	OrderAsc Order = iota
	OrderDesc
)

func (o Order) String() string {
	switch o {
	case OrderAsc:
		return "ASC"
	case OrderDesc:
		return "DESC"
	}
	return ""
}

func WithOrder(order Order) QueryOption {
	return func(p *QueryParams) {
		p.Order = order
	}
}

// ReadOnlyStore is a storage implementation which only supports
// reading the various types of state configuring within Flipt
type ReadOnlyStore interface {
	ReadOnlyNamespaceStore
	ReadOnlyFlagStore
	ReadOnlySegmentStore
	ReadOnlyRuleStore
	ReadOnlyRolloutStore
	EvaluationStore
	fmt.Stringer
}

// Store supports reading and writing all the resources within Flipt
type Store interface {
	NamespaceStore
	FlagStore
	SegmentStore
	RuleStore
	RolloutStore
	EvaluationStore
	fmt.Stringer
}

type ResultSet[T any] struct {
	Results       []T    `json:"results"`
	NextPageToken string `json:"next_page_token"`
}

const DefaultNamespace = "default"

// EvaluationStore returns data necessary for evaluation
type EvaluationStore interface {
	// GetEvaluationRules returns rules applicable to namespaceKey + flagKey provided
	// Note: Rules MUST be returned in order by rank
	GetEvaluationRules(ctx context.Context, namespaceKey, flagKey string) ([]*EvaluationRule, error)
	GetEvaluationDistributions(ctx context.Context, ruleID string) ([]*EvaluationDistribution, error)
	// GetEvaluationRollouts returns rules applicable to namespaceKey + flagKey provided
	// Note: Rollouts MUST be returned in order by rank
	GetEvaluationRollouts(ctx context.Context, namespaceKey, flagKey string) ([]*EvaluationRollout, error)
}

// ReadOnlyNamespaceStore support retrieval of namespaces only
type ReadOnlyNamespaceStore interface {
	GetNamespace(ctx context.Context, key string) (*flipt.Namespace, error)
	ListNamespaces(ctx context.Context, opts ...QueryOption) (ResultSet[*flipt.Namespace], error)
	CountNamespaces(ctx context.Context) (uint64, error)
}

// NamespaceStore stores and retrieves namespaces
type NamespaceStore interface {
	ReadOnlyNamespaceStore
	CreateNamespace(ctx context.Context, r *flipt.CreateNamespaceRequest) (*flipt.Namespace, error)
	UpdateNamespace(ctx context.Context, r *flipt.UpdateNamespaceRequest) (*flipt.Namespace, error)
	DeleteNamespace(ctx context.Context, r *flipt.DeleteNamespaceRequest) error
}

// ReadOnlyFlagStore supports retrieval of flags
type ReadOnlyFlagStore interface {
	GetFlag(ctx context.Context, namespaceKey, key string) (*flipt.Flag, error)
	ListFlags(ctx context.Context, namespaceKey string, opts ...QueryOption) (ResultSet[*flipt.Flag], error)
	CountFlags(ctx context.Context, namespaceKey string) (uint64, error)
}

// FlagStore stores and retrieves flags and variants
type FlagStore interface {
	ReadOnlyFlagStore
	CreateFlag(ctx context.Context, r *flipt.CreateFlagRequest) (*flipt.Flag, error)
	UpdateFlag(ctx context.Context, r *flipt.UpdateFlagRequest) (*flipt.Flag, error)
	DeleteFlag(ctx context.Context, r *flipt.DeleteFlagRequest) error
	CreateVariant(ctx context.Context, r *flipt.CreateVariantRequest) (*flipt.Variant, error)
	UpdateVariant(ctx context.Context, r *flipt.UpdateVariantRequest) (*flipt.Variant, error)
	DeleteVariant(ctx context.Context, r *flipt.DeleteVariantRequest) error
}

// ReadOnlySegmentStore supports retrieval of segments and constraints
type ReadOnlySegmentStore interface {
	GetSegment(ctx context.Context, namespaceKey, key string) (*flipt.Segment, error)
	ListSegments(ctx context.Context, namespaceKey string, opts ...QueryOption) (ResultSet[*flipt.Segment], error)
	CountSegments(ctx context.Context, namespaceKey string) (uint64, error)
}

// SegmentStore stores and retrieves segments and constraints
type SegmentStore interface {
	ReadOnlySegmentStore
	CreateSegment(ctx context.Context, r *flipt.CreateSegmentRequest) (*flipt.Segment, error)
	UpdateSegment(ctx context.Context, r *flipt.UpdateSegmentRequest) (*flipt.Segment, error)
	DeleteSegment(ctx context.Context, r *flipt.DeleteSegmentRequest) error
	CreateConstraint(ctx context.Context, r *flipt.CreateConstraintRequest) (*flipt.Constraint, error)
	UpdateConstraint(ctx context.Context, r *flipt.UpdateConstraintRequest) (*flipt.Constraint, error)
	DeleteConstraint(ctx context.Context, r *flipt.DeleteConstraintRequest) error
}

// ReadOnlyRuleStore supports retrieval of rules and distributions
type ReadOnlyRuleStore interface {
	GetRule(ctx context.Context, namespaceKey, id string) (*flipt.Rule, error)
	ListRules(ctx context.Context, namespaceKey, flagKey string, opts ...QueryOption) (ResultSet[*flipt.Rule], error)
	CountRules(ctx context.Context, namespaceKey, flagKey string) (uint64, error)
}

// RuleStore stores and retrieves rules and distributions
type RuleStore interface {
	ReadOnlyRuleStore
	CreateRule(ctx context.Context, r *flipt.CreateRuleRequest) (*flipt.Rule, error)
	UpdateRule(ctx context.Context, r *flipt.UpdateRuleRequest) (*flipt.Rule, error)
	DeleteRule(ctx context.Context, r *flipt.DeleteRuleRequest) error
	OrderRules(ctx context.Context, r *flipt.OrderRulesRequest) error
	CreateDistribution(ctx context.Context, r *flipt.CreateDistributionRequest) (*flipt.Distribution, error)
	UpdateDistribution(ctx context.Context, r *flipt.UpdateDistributionRequest) (*flipt.Distribution, error)
	DeleteDistribution(ctx context.Context, r *flipt.DeleteDistributionRequest) error
}

// ReadOnlyRolloutStore supports retrieval of rollouts
type ReadOnlyRolloutStore interface {
	GetRollout(ctx context.Context, namespaceKey, id string) (*flipt.Rollout, error)
	ListRollouts(ctx context.Context, namespaceKey, flagKey string, opts ...QueryOption) (ResultSet[*flipt.Rollout], error)
	CountRollouts(ctx context.Context, namespaceKey, flagKey string) (uint64, error)
}

// RolloutStore supports storing and retrieving rollouts
type RolloutStore interface {
	ReadOnlyRolloutStore
	CreateRollout(ctx context.Context, r *flipt.CreateRolloutRequest) (*flipt.Rollout, error)
	UpdateRollout(ctx context.Context, r *flipt.UpdateRolloutRequest) (*flipt.Rollout, error)
	DeleteRollout(ctx context.Context, r *flipt.DeleteRolloutRequest) error
	OrderRollouts(ctx context.Context, r *flipt.OrderRolloutsRequest) error
}

// ListRequest is a generic container for the parameters required to perform a list operation.
// It contains a generic type T intended for a list predicate.
// It also contains a QueryParams object containing pagination constraints.
type ListRequest[P any] struct {
	Predicate   P
	QueryParams QueryParams
}

// ListOption is a function which can configure a ListRequest.
type ListOption[T any] func(*ListRequest[T])

// ListWithQueryParamOptions takes a set of functional options for QueryParam and returns a ListOption
// which applies them in order on the provided ListRequest.
func ListWithQueryParamOptions[T any](opts ...QueryOption) ListOption[T] {
	return func(r *ListRequest[T]) {
		for _, opt := range opts {
			opt(&r.QueryParams)
		}
	}
}

// NewListRequest constructs a new ListRequest using the provided ListOption.
func NewListRequest[T any](opts ...ListOption[T]) *ListRequest[T] {
	req := &ListRequest[T]{}

	for _, opt := range opts {
		opt(req)
	}

	return req
}

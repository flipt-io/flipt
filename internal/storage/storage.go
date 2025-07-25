package storage

import (
	"context"
	"fmt"
	"path"

	"go.flipt.io/flipt/internal/containers"
	"go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/rpc/flipt/core"
	"go.flipt.io/flipt/rpc/v2/evaluation"
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
	SegmentOperator core.SegmentOperator          `json:"segmentOperator,omitempty"`
}

type EvaluationSegment struct {
	SegmentKey  string                 `json:"segment_key,omitempty"`
	MatchType   core.MatchType         `json:"match_type,omitempty"`
	Constraints []EvaluationConstraint `json:"constraints,omitempty"`
}

// EvaluationRollout represents a rollout in the form that helps with evaluation.
type EvaluationRollout struct {
	NamespaceKey string            `json:"namespace_key,omitempty"`
	RolloutType  core.RolloutType  `json:"rollout_type,omitempty"`
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
	SegmentOperator core.SegmentOperator          `json:"segment_operator,omitempty"`
	Segments        map[string]*EvaluationSegment `json:"segments,omitempty"`
}

// EvaluationConstraint represents a segment constraint that is used for evaluation
type EvaluationConstraint struct {
	ID       string              `json:"id,omitempty"`
	Type     core.ComparisonType `json:"comparison_type,omitempty"`
	Property string              `json:"property,omitempty"`
	Operator string              `json:"operator,omitempty"`
	Value    string              `json:"value,omitempty"`
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
	ReadOnlyFlagStore
	EvaluationStore
	EvaluationSnapshotStore
	fmt.Stringer
}

// Store supports reading and writing all the resources within Flipt
type Store interface {
	FlagStore
	EvaluationStore
	fmt.Stringer
}

type ResultSet[T any] struct {
	Results       []T    `json:"results"`
	NextPageToken string `json:"next_page_token"`
}

// EvaluationStore returns data necessary for evaluation
type EvaluationStore interface {
	// GetEvaluationRules returns rules applicable to flagKey provided
	// Note: Rules MUST be returned in order by Rank
	GetEvaluationRules(ctx context.Context, flag ResourceRequest) ([]*EvaluationRule, error)
	GetEvaluationDistributions(ctx context.Context, flag ResourceRequest, rule IDRequest) ([]*EvaluationDistribution, error)
	// GetEvaluationRollouts returns rollouts applicable to namespaceKey + flagKey provided
	// Note: Rollouts MUST be returned in order by rank
	GetEvaluationRollouts(ctx context.Context, flag ResourceRequest) ([]*EvaluationRollout, error)
}

// EvaluationSnapshotStore is a store from which entire namespace snapshots can be retrieved.
// These snapshots are intended to be sent to client-side evaluators.
type EvaluationSnapshotStore interface {
	EvaluationNamespaceSnapshot(_ context.Context, ns string) (*evaluation.EvaluationNamespaceSnapshot, error)
}

// ReadOnlyFlagStore supports retrieval of flags
type ReadOnlyFlagStore interface {
	GetFlag(ctx context.Context, req ResourceRequest) (*core.Flag, error)
	ListFlags(ctx context.Context, req *ListRequest[NamespaceRequest]) (ResultSet[*core.Flag], error)
	CountFlags(ctx context.Context, ns NamespaceRequest) (uint64, error)
}

// FlagStore stores and retrieves flags and variants
type FlagStore interface {
	ReadOnlyFlagStore
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

// PageParameter is any type which exposes limit and page token
// accessors used to identify pages.
type PageParameter interface {
	GetLimit() int32
	GetPageToken() string
}

// OffsetPageParameter is a type which exposes an additional offset
// accessor for legacy paging implementations (page token supersedes).
type OffsetPageParameter interface {
	PageParameter
	GetOffset() int32
}

// ListWithParameters constructs a new ListRequest using the page parameters
// exposed by the provided PageParameter implementation.
func ListWithParameters[T any](t T, p PageParameter) *ListRequest[T] {
	opts := []QueryOption{
		WithLimit(uint64(p.GetLimit())),
		WithPageToken(p.GetPageToken()),
	}

	if po, ok := p.(OffsetPageParameter); ok {
		opts = append(opts, WithOffset(uint64(po.GetOffset())))
	}

	return ListWithOptions(t, ListWithQueryParamOptions[T](opts...))
}

// ListWithOptions constructs a new ListRequest using the provided ListOption.
func ListWithOptions[T any](t T, opts ...ListOption[T]) *ListRequest[T] {
	req := &ListRequest[T]{
		Predicate: t,
	}

	for _, opt := range opts {
		opt(req)
	}

	return req
}

// Reference is a string which can refer to either a concrete
// revision or it can be an indirect named reference.
type Reference string

// ReferenceRequest is used to identify a request predicated solely by a revision reference.
// This is primarily used for namespaces as it is the highest level domain model.
type ReferenceRequest struct {
	Reference
}

// WithReference sets the provided reference identifier on a *ReferenceRequest.
func WithReference(ref string) containers.Option[ReferenceRequest] {
	return func(rp *ReferenceRequest) {
		if ref != "" {
			rp.Reference = Reference(ref)
		}
	}
}

// ListWithReference is a ListOption constrained to ReferenceRequest types.
// It sets the reference on the resulting list request.
func ListWithReference(ref string) ListOption[ReferenceRequest] {
	return func(rp *ListRequest[ReferenceRequest]) {
		if ref != "" {
			rp.Predicate.Reference = Reference(ref)
		}
	}
}

// NamespaceRequest is used to identify a request predicated by both a revision and a namespace.
// This is used to identify namespaces and list resources directly beneath them (e.g. flags and segments).
type NamespaceRequest struct {
	ReferenceRequest
	key string
}

// NewNamespace constructs a *NamespaceRequest from the provided key string.
// Optionally, the target storage revision reference can also be supplied.
func NewNamespace(key string, opts ...containers.Option[ReferenceRequest]) NamespaceRequest {
	p := NamespaceRequest{key: key}

	containers.ApplyAll(&p.ReferenceRequest, opts...)

	return p
}

// String returns the resolved target namespace key string.
// If the underlying key is the empty string, the default namespace ("default")
// is returned instead.
func (n NamespaceRequest) String() string {
	return n.Namespace()
}

// Namespace returns the resolved target namespace key string.
// If the underlying key is the empty string, the default namespace ("default")
// is returned instead.
func (n NamespaceRequest) Namespace() string {
	if n.key == "" {
		return flipt.DefaultNamespace
	}

	return n.key
}

// ResourceRequest is used to identify a request predicated by revision, namespace and a key.
// This is used for core resources (e.g. flag and segment) as well as to list sub-resources (e.g. rules and constraints).
type ResourceRequest struct {
	NamespaceRequest
	Key string
}

// NewResource constructs and configures and new *ResourceRequest from the provided namespace and resource keys.
// Optionally, the target storage revision reference can also be supplied.
func NewResource(ns, key string, opts ...containers.Option[ReferenceRequest]) ResourceRequest {
	p := ResourceRequest{
		NamespaceRequest: NamespaceRequest{
			key: ns,
		},
		Key: key,
	}

	containers.ApplyAll(&p.ReferenceRequest, opts...)

	return p
}

// String returns a representation of the combined resource namespace and key separated by a '/'.
func (p ResourceRequest) String() string {
	return path.Join(p.Namespace(), p.Key)
}

// IDRequest is used to identify any sub-resources which have a unique random identifier.
// This is used for sub-resources with no key identifiers (e.g. rules and rollouts).
type IDRequest struct {
	ReferenceRequest
	ID string
}

// NewID constructs and configures a new *IDRequest with the provided ID string.
// Optionally, the target storage revision reference can also be supplied.
func NewID(id string, opts ...containers.Option[ReferenceRequest]) IDRequest {
	p := IDRequest{ID: id}
	containers.ApplyAll(&p.ReferenceRequest, opts...)
	return p
}

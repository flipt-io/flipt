package sqlite

import (
	"context"
	"database/sql"

	"errors"

	sq "github.com/Masterminds/squirrel"
	errs "github.com/markphelps/flipt/errors"
	flipt "github.com/markphelps/flipt/rpc"
	"github.com/markphelps/flipt/storage"
	"github.com/markphelps/flipt/storage/db"
	"github.com/mattn/go-sqlite3"
)

func NewProvider(builder sq.StatementBuilderType, d *sql.DB) *Provider {
	var (
		flagStore       = db.NewFlagStore(builder)
		segmentStore    = db.NewSegmentStore(builder)
		ruleStore       = db.NewRuleStore(builder, d)
		evaluationStore = db.NewEvaluationStore(builder)
	)

	return &Provider{
		flagStore:       &FlagStore{flagStore},
		segmentStore:    &SegmentStore{segmentStore},
		ruleStore:       &RuleStore{ruleStore},
		evaluationStore: &EvaluationStore{evaluationStore},
	}
}

type Provider struct {
	flagStore       *FlagStore
	segmentStore    *SegmentStore
	ruleStore       *RuleStore
	evaluationStore *EvaluationStore
}

func (p *Provider) FlagStore() storage.FlagStore {
	return p.flagStore
}

func (p *Provider) SegmentStore() storage.SegmentStore {
	return p.segmentStore
}

func (p *Provider) RuleStore() storage.RuleStore {
	return p.ruleStore
}

func (p *Provider) EvaluationStore() storage.EvaluationStore {
	return p.evaluationStore
}

var _ storage.FlagStore = &FlagStore{}

type FlagStore struct {
	*db.FlagStore
}

func (f *FlagStore) CreateFlag(ctx context.Context, r *flipt.CreateFlagRequest) (*flipt.Flag, error) {
	flag, err := f.FlagStore.CreateFlag(ctx, r)

	if err != nil {
		var serr sqlite3.Error

		if errors.As(err, &serr) && serr.Code == sqlite3.ErrConstraint {
			return nil, errs.ErrInvalidf("flag %q is not unique", r.Key)
		}

		return nil, err
	}

	return flag, nil
}

func (f *FlagStore) CreateVariant(ctx context.Context, r *flipt.CreateVariantRequest) (*flipt.Variant, error) {
	variant, err := f.FlagStore.CreateVariant(ctx, r)

	if err != nil {
		var serr sqlite3.Error

		if errors.As(err, &serr) {
			switch serr.ExtendedCode {
			case sqlite3.ErrConstraintForeignKey:
				return nil, errs.ErrNotFoundf("flag %q", r.FlagKey)
			case sqlite3.ErrConstraintUnique:
				return nil, errs.ErrInvalidf("variant %q is not unique", r.Key)
			}
		}

		return nil, err
	}

	return variant, nil
}

func (f *FlagStore) UpdateVariant(ctx context.Context, r *flipt.UpdateVariantRequest) (*flipt.Variant, error) {
	variant, err := f.FlagStore.UpdateVariant(ctx, r)

	if err != nil {
		var serr sqlite3.Error

		if errors.As(err, &serr) && serr.Code == sqlite3.ErrConstraint {
			return nil, errs.ErrInvalidf("variant %q is not unique", r.Key)
		}

		return nil, err
	}

	return variant, nil
}

var _ storage.SegmentStore = &SegmentStore{}

type SegmentStore struct {
	*db.SegmentStore
}

func (s *SegmentStore) CreateSegment(ctx context.Context, r *flipt.CreateSegmentRequest) (*flipt.Segment, error) {
	segment, err := s.SegmentStore.CreateSegment(ctx, r)

	if err != nil {
		var serr sqlite3.Error

		if errors.As(err, &serr) && serr.Code == sqlite3.ErrConstraint {
			return nil, errs.ErrInvalidf("segment %q is not unique", r.Key)
		}

		return nil, err
	}

	return segment, nil
}

func (s *SegmentStore) CreateConstraint(ctx context.Context, r *flipt.CreateConstraintRequest) (*flipt.Constraint, error) {
	constraint, err := s.SegmentStore.CreateConstraint(ctx, r)

	if err != nil {
		var serr sqlite3.Error

		if errors.As(err, &serr) && serr.Code == sqlite3.ErrConstraint {
			return nil, errs.ErrNotFoundf("segment %q", r.SegmentKey)
		}

		return nil, err
	}

	return constraint, nil
}

var _ storage.RuleStore = &RuleStore{}

type RuleStore struct {
	*db.RuleStore
}

func (s *RuleStore) CreateRule(ctx context.Context, r *flipt.CreateRuleRequest) (*flipt.Rule, error) {
	rule, err := s.RuleStore.CreateRule(ctx, r)

	if err != nil {
		var serr sqlite3.Error

		if errors.As(err, &serr) && serr.Code == sqlite3.ErrConstraint {
			return nil, errs.ErrNotFoundf("flag %q or segment %q", r.FlagKey, r.SegmentKey)
		}

		return nil, err
	}

	return rule, nil
}

func (s *RuleStore) CreateDistribution(ctx context.Context, r *flipt.CreateDistributionRequest) (*flipt.Distribution, error) {
	dist, err := s.RuleStore.CreateDistribution(ctx, r)

	if err != nil {
		var serr sqlite3.Error

		if errors.As(err, &serr) && serr.Code == sqlite3.ErrConstraint {
			return nil, errs.ErrNotFoundf("rule %q", r.RuleId)
		}

		return nil, err
	}

	return dist, nil
}

var _ storage.EvaluationStore = &EvaluationStore{}

type EvaluationStore struct {
	*db.EvaluationStore
}

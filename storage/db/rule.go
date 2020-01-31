package db

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/gofrs/uuid"
	proto "github.com/golang/protobuf/ptypes"
	"github.com/lib/pq"
	"github.com/markphelps/flipt/errors"
	flipt "github.com/markphelps/flipt/rpc"
	"github.com/markphelps/flipt/storage"
	sqlite3 "github.com/mattn/go-sqlite3"
)

var _ storage.RuleStore = &RuleStore{}

// RuleStore is a SQL RuleStore
type RuleStore struct {
	builder sq.StatementBuilderType
	db      *sql.DB
}

// NewRuleStore creates a RuleStore
func NewRuleStore(builder sq.StatementBuilderType, db *sql.DB) *RuleStore {
	return &RuleStore{
		builder: builder,
		db:      db,
	}
}

// GetRule gets a rule
func (s *RuleStore) GetRule(ctx context.Context, r *flipt.GetRuleRequest) (*flipt.Rule, error) {
	return s.rule(ctx, r.Id, r.FlagKey)
}

func (s *RuleStore) rule(ctx context.Context, id, flagKey string) (*flipt.Rule, error) {
	var (
		createdAt timestamp
		updatedAt timestamp

		rule = &flipt.Rule{}
	)

	if err := s.builder.Select("id, flag_key, segment_key, rank, created_at, updated_at").
		From("rules").
		Where(sq.And{sq.Eq{"id": id}, sq.Eq{"flag_key": flagKey}}).
		QueryRowContext(ctx).
		Scan(&rule.Id, &rule.FlagKey, &rule.SegmentKey, &rule.Rank, &createdAt, &updatedAt); err != nil {
		return nil, err
	}

	rule.CreatedAt = createdAt.Timestamp
	rule.UpdatedAt = updatedAt.Timestamp

	if err := s.distributions(ctx, rule); err != nil {
		return nil, err
	}

	return rule, nil
}

// ListRules lists all rules
func (s *RuleStore) ListRules(ctx context.Context, r *flipt.ListRuleRequest) ([]*flipt.Rule, error) {
	return s.rules(ctx, r)
}

func (s *RuleStore) rules(ctx context.Context, r *flipt.ListRuleRequest) ([]*flipt.Rule, error) {
	var (
		rules []*flipt.Rule

		query = s.builder.Select("id, flag_key, segment_key, rank, created_at, updated_at").
			From("rules").
			Where(sq.Eq{"flag_key": r.FlagKey}).
			OrderBy("rank ASC")
	)

	if r.Limit > 0 {
		query = query.Limit(uint64(r.Limit))
	}

	if r.Offset > 0 {
		query = query.Offset(uint64(r.Offset))
	}

	rows, err := query.QueryContext(ctx)
	if err != nil {
		return nil, err
	}

	defer func() {
		if cerr := rows.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	for rows.Next() {
		var (
			rule      flipt.Rule
			createdAt timestamp
			updatedAt timestamp
		)

		if err := rows.Scan(
			&rule.Id,
			&rule.FlagKey,
			&rule.SegmentKey,
			&rule.Rank,
			&createdAt,
			&updatedAt); err != nil {
			return nil, err
		}

		rule.CreatedAt = createdAt.Timestamp
		rule.UpdatedAt = updatedAt.Timestamp

		if err := s.distributions(ctx, &rule); err != nil {
			return nil, err
		}

		rules = append(rules, &rule)
	}

	return rules, rows.Err()
}

// CreateRule creates a rule
func (s *RuleStore) CreateRule(ctx context.Context, r *flipt.CreateRuleRequest) (*flipt.Rule, error) {
	var (
		now  = proto.TimestampNow()
		rule = &flipt.Rule{
			Id:         uuid.Must(uuid.NewV4()).String(),
			FlagKey:    r.FlagKey,
			SegmentKey: r.SegmentKey,
			Rank:       r.Rank,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
	)

	if _, err := s.builder.
		Insert("rules").
		Columns("id", "flag_key", "segment_key", "rank", "created_at", "updated_at").
		Values(rule.Id, rule.FlagKey, rule.SegmentKey, rule.Rank, &timestamp{rule.CreatedAt}, &timestamp{rule.UpdatedAt}).
		ExecContext(ctx); err != nil {
		switch ierr := err.(type) {
		case sqlite3.Error:
			if ierr.Code == sqlite3.ErrConstraint {
				return nil, errors.ErrNotFoundf("flag %q or segment %q", r.FlagKey, r.SegmentKey)
			}
		case *pq.Error:
			if ierr.Code.Name() == pgConstraintForeignKey {
				return nil, errors.ErrNotFoundf("flag %q or segment %q", r.FlagKey, r.SegmentKey)
			}
		}

		return nil, err
	}

	return rule, nil
}

// UpdateRule updates an existing rule
func (s *RuleStore) UpdateRule(ctx context.Context, r *flipt.UpdateRuleRequest) (*flipt.Rule, error) {
	query := s.builder.Update("rules").
		Set("segment_key", r.SegmentKey).
		Set("updated_at", &timestamp{proto.TimestampNow()}).
		Where(sq.And{sq.Eq{"id": r.Id}, sq.Eq{"flag_key": r.FlagKey}})

	res, err := query.ExecContext(ctx)
	if err != nil {
		return nil, err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	if count != 1 {
		return nil, errors.ErrNotFoundf("rule %q", r.Id)
	}

	return s.rule(ctx, r.Id, r.FlagKey)
}

// DeleteRule deletes a rule
func (s *RuleStore) DeleteRule(ctx context.Context, r *flipt.DeleteRuleRequest) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	// delete rule
	_, err = s.builder.Delete("rules").
		RunWith(tx).
		Where(sq.And{sq.Eq{"id": r.Id}, sq.Eq{"flag_key": r.FlagKey}}).
		ExecContext(ctx)

	// reorder existing rules after deletion
	rows, err := s.builder.Select("id").
		RunWith(tx).
		From("rules").
		Where(sq.Eq{"flag_key": r.FlagKey}).
		OrderBy("rank ASC").
		QueryContext(ctx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	defer func() {
		if cerr := rows.Close(); cerr != nil && err == nil {
			_ = tx.Rollback()
			err = cerr
		}
	}()

	var ruleIDs []string

	for rows.Next() {
		var ruleID string

		if err := rows.Scan(&ruleID); err != nil {
			_ = tx.Rollback()
			return err
		}

		ruleIDs = append(ruleIDs, ruleID)
	}

	if err := rows.Err(); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := s.orderRules(ctx, tx, r.FlagKey, ruleIDs); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

// OrderRules orders rules
func (s *RuleStore) OrderRules(ctx context.Context, r *flipt.OrderRulesRequest) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	if err := s.orderRules(ctx, tx, r.FlagKey, r.RuleIds); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (s *RuleStore) orderRules(ctx context.Context, runner sq.BaseRunner, flagKey string, ruleIDs []string) error {
	updatedAt := proto.TimestampNow()

	for i, id := range ruleIDs {
		_, err := s.builder.Update("rules").
			RunWith(runner).
			Set("rank", i+1).
			Set("updated_at", &timestamp{updatedAt}).
			Where(sq.And{sq.Eq{"id": id}, sq.Eq{"flag_key": flagKey}}).
			ExecContext(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

// CreateDistribution creates a distribution
func (s *RuleStore) CreateDistribution(ctx context.Context, r *flipt.CreateDistributionRequest) (*flipt.Distribution, error) {
	var (
		now = proto.TimestampNow()
		d   = &flipt.Distribution{
			Id:        uuid.Must(uuid.NewV4()).String(),
			RuleId:    r.RuleId,
			VariantId: r.VariantId,
			Rollout:   r.Rollout,
			CreatedAt: now,
			UpdatedAt: now,
		}
	)

	if _, err := s.builder.
		Insert("distributions").
		Columns("id", "rule_id", "variant_id", "rollout", "created_at", "updated_at").
		Values(d.Id, d.RuleId, d.VariantId, d.Rollout, &timestamp{d.CreatedAt}, &timestamp{d.UpdatedAt}).
		ExecContext(ctx); err != nil {
		switch ierr := err.(type) {
		case sqlite3.Error:
			if ierr.Code == sqlite3.ErrConstraint {
				return nil, errors.ErrNotFoundf("rule %q", r.RuleId)
			}
		case *pq.Error:
			if ierr.Code.Name() == pgConstraintForeignKey {
				return nil, errors.ErrNotFoundf("rule %q", r.RuleId)
			}
		}

		return nil, err
	}

	return d, nil
}

// UpdateDistribution updates an existing distribution
func (s *RuleStore) UpdateDistribution(ctx context.Context, r *flipt.UpdateDistributionRequest) (*flipt.Distribution, error) {
	query := s.builder.Update("distributions").
		Set("rollout", r.Rollout).
		Set("updated_at", &timestamp{proto.TimestampNow()}).
		Where(sq.And{sq.Eq{"id": r.Id}, sq.Eq{"rule_id": r.RuleId}, sq.Eq{"variant_id": r.VariantId}})

	res, err := query.ExecContext(ctx)
	if err != nil {
		return nil, err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	if count != 1 {
		return nil, errors.ErrNotFoundf("distribution %q", r.Id)
	}

	var (
		createdAt timestamp
		updatedAt timestamp

		distribution = &flipt.Distribution{}
	)

	if err := s.builder.Select("id, rule_id, variant_id, rollout, created_at, updated_at").
		From("distributions").
		Where(sq.And{sq.Eq{"id": r.Id}, sq.Eq{"rule_id": r.RuleId}, sq.Eq{"variant_id": r.VariantId}}).
		QueryRowContext(ctx).
		Scan(&distribution.Id, &distribution.RuleId, &distribution.VariantId, &distribution.Rollout, &createdAt, &updatedAt); err != nil {
		return nil, err
	}

	distribution.CreatedAt = createdAt.Timestamp
	distribution.UpdatedAt = updatedAt.Timestamp

	return distribution, nil
}

// DeleteDistribution deletes a distribution
func (s *RuleStore) DeleteDistribution(ctx context.Context, r *flipt.DeleteDistributionRequest) error {
	_, err := s.builder.Delete("distributions").
		Where(sq.And{sq.Eq{"id": r.Id}, sq.Eq{"rule_id": r.RuleId}, sq.Eq{"variant_id": r.VariantId}}).
		ExecContext(ctx)

	return err
}

func (s *RuleStore) distributions(ctx context.Context, rule *flipt.Rule) (err error) {
	query := s.builder.Select("id", "rule_id", "variant_id", "rollout", "created_at", "updated_at").
		From("distributions").
		Where(sq.Eq{"rule_id": rule.Id}).
		OrderBy("created_at ASC")

	rows, err := query.QueryContext(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if cerr := rows.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	for rows.Next() {
		var (
			distribution         flipt.Distribution
			createdAt, updatedAt timestamp
		)

		if err := rows.Scan(
			&distribution.Id,
			&distribution.RuleId,
			&distribution.VariantId,
			&distribution.Rollout,
			&createdAt,
			&updatedAt); err != nil {
			return err
		}

		distribution.CreatedAt = createdAt.Timestamp
		distribution.UpdatedAt = updatedAt.Timestamp

		rule.Distributions = append(rule.Distributions, &distribution)
	}

	return rows.Err()
}

package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"hash/crc32"
	"sort"
	"strconv"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/golang/protobuf/ptypes"
	proto "github.com/golang/protobuf/ptypes"
	flipt "github.com/markphelps/flipt/proto"
	sqlite3 "github.com/mattn/go-sqlite3"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
)

var _ RuleStore = &RuleStorage{}

// RuleStorage is a SQL RuleStore
type RuleStorage struct {
	logger  logrus.FieldLogger
	tx      sq.DBProxyBeginner
	builder sq.StatementBuilderType
}

// NewRuleStorage creates a RuleStorage
func NewRuleStorage(logger logrus.FieldLogger, tx sq.DBProxyBeginner, builder sq.StatementBuilderType) *RuleStorage {
	return &RuleStorage{
		logger:  logger.WithField("storage", "rule"),
		tx:      tx,
		builder: builder,
	}
}

// GetRule gets a rule
func (s *RuleStorage) GetRule(ctx context.Context, r *flipt.GetRuleRequest) (*flipt.Rule, error) {
	s.logger.WithField("request", r).Debug("get rule")
	rule, err := s.rule(ctx, r.Id, r.FlagKey)
	s.logger.WithField("response", rule).Debug("get rule")
	return rule, err
}

func (s *RuleStorage) rule(ctx context.Context, id, flagKey string) (*flipt.Rule, error) {
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
func (s *RuleStorage) ListRules(ctx context.Context, r *flipt.ListRuleRequest) ([]*flipt.Rule, error) {
	s.logger.WithField("request", r).Debug("list rules")
	rules, err := s.rules(ctx, r)
	s.logger.WithField("response", rules).Debug("list rules")
	return rules, err
}

func (s *RuleStorage) rules(ctx context.Context, r *flipt.ListRuleRequest) ([]*flipt.Rule, error) {
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
func (s *RuleStorage) CreateRule(ctx context.Context, r *flipt.CreateRuleRequest) (*flipt.Rule, error) {
	s.logger.WithField("request", r).Debug("create rule")

	var (
		now  = proto.TimestampNow()
		rule = &flipt.Rule{
			Id:         uuid.NewV4().String(),
			FlagKey:    r.FlagKey,
			SegmentKey: r.SegmentKey,
			Rank:       r.Rank,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
	)

	if _, err := s.builder.Insert("rules").
		Columns("id", "flag_key", "segment_key", "rank", "created_at", "updated_at").
		Values(rule.Id, rule.FlagKey, rule.SegmentKey, rule.Rank, &timestamp{rule.CreatedAt}, &timestamp{rule.UpdatedAt}).
		ExecContext(ctx); err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			if sqliteErr.ExtendedCode == sqlite3.ErrConstraintForeignKey {
				return nil, ErrNotFoundf("flag %q or segment %q", r.FlagKey, r.SegmentKey)
			}
		}
		return nil, err
	}

	s.logger.WithField("response", rule).Debug("create rule")
	return rule, nil
}

// UpdateRule updates an existing rule
func (s *RuleStorage) UpdateRule(ctx context.Context, r *flipt.UpdateRuleRequest) (*flipt.Rule, error) {
	s.logger.WithField("request", r).Debug("update rule")
	var (
		query = s.builder.Update("rules").
			Set("segment_key", r.SegmentKey).
			Set("updated_at", &timestamp{proto.TimestampNow()}).
			Where(sq.And{sq.Eq{"id": r.Id}, sq.Eq{"flag_key": r.FlagKey}})
	)

	res, err := query.ExecContext(ctx)
	if err != nil {
		return nil, err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	if count != 1 {
		return nil, ErrNotFoundf("rule %q", r.Id)
	}

	rule, err := s.rule(ctx, r.Id, r.FlagKey)
	s.logger.WithField("response", rule).Debug("update rule")
	return rule, err
}

// DeleteRule deletes a rule
func (s *RuleStorage) DeleteRule(ctx context.Context, r *flipt.DeleteRuleRequest) error {
	s.logger.WithField("request", r).Debug("delete rule")

	tx, err := s.tx.Begin()
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
func (s *RuleStorage) OrderRules(ctx context.Context, r *flipt.OrderRulesRequest) error {
	s.logger.WithField("request", r).Debug("order rules")

	tx, err := s.tx.Begin()
	if err != nil {
		return err
	}

	if err := s.orderRules(ctx, tx, r.FlagKey, r.RuleIds); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (s *RuleStorage) orderRules(ctx context.Context, tx *sql.Tx, flagKey string, ruleIDs []string) error {
	updatedAt := proto.TimestampNow()

	for i, id := range ruleIDs {
		_, err := s.builder.Update("rules").
			RunWith(tx).
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
func (s *RuleStorage) CreateDistribution(ctx context.Context, r *flipt.CreateDistributionRequest) (*flipt.Distribution, error) {
	s.logger.WithField("request", r).Debug("create distribution")

	var (
		now = proto.TimestampNow()
		d   = &flipt.Distribution{
			Id:        uuid.NewV4().String(),
			RuleId:    r.RuleId,
			VariantId: r.VariantId,
			Rollout:   r.Rollout,
			CreatedAt: now,
			UpdatedAt: now,
		}
	)

	if _, err := s.builder.Insert("distributions").
		Columns("id", "rule_id", "variant_id", "rollout", "created_at", "updated_at").
		Values(d.Id, d.RuleId, d.VariantId, d.Rollout, &timestamp{d.CreatedAt}, &timestamp{d.UpdatedAt}).
		ExecContext(ctx); err != nil {

		if sqliteErr, ok := err.(sqlite3.Error); ok {
			if sqliteErr.ExtendedCode == sqlite3.ErrConstraintForeignKey {
				return nil, ErrNotFoundf("rule %q", r.RuleId)
			}
		}
		return nil, err
	}

	s.logger.WithField("response", d).Debug("create distribution")
	return d, nil
}

// UpdateDistribution updates an existing distribution
func (s *RuleStorage) UpdateDistribution(ctx context.Context, r *flipt.UpdateDistributionRequest) (*flipt.Distribution, error) {
	s.logger.WithField("request", r).Debug("update distribution")

	var (
		query = s.builder.Update("distributions").
			Set("rollout", r.Rollout).
			Set("updated_at", &timestamp{proto.TimestampNow()}).
			Where(sq.And{sq.Eq{"id": r.Id}, sq.Eq{"rule_id": r.RuleId}, sq.Eq{"variant_id": r.VariantId}})
	)

	res, err := query.ExecContext(ctx)
	if err != nil {
		return nil, err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	if count != 1 {
		return nil, ErrNotFoundf("distribution %q", r.Id)
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

	s.logger.WithField("response", distribution).Debug("update distribution")
	return distribution, nil
}

// DeleteDistribution deletes a distribution
func (s *RuleStorage) DeleteDistribution(ctx context.Context, r *flipt.DeleteDistributionRequest) error {
	s.logger.WithField("request", r).Debug("delete distribution")

	_, err := s.builder.Delete("distributions").
		Where(sq.And{sq.Eq{"id": r.Id}, sq.Eq{"rule_id": r.RuleId}, sq.Eq{"variant_id": r.VariantId}}).
		ExecContext(ctx)

	return err
}

func (s *RuleStorage) distributions(ctx context.Context, rule *flipt.Rule) (err error) {
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

type constraint struct {
	Type     flipt.ComparisonType
	Property string
	Operator string
	Value    string
}

type rule struct {
	ID          string
	FlagKey     string
	SegmentKey  string
	Rank        int32
	Constraints []constraint
}

type distribution struct {
	RuleID     string
	VariantID  string
	Rollout    float32
	VariantKey string
}

// Evaluate evaluates a request for a given flag and entity
func (s *RuleStorage) Evaluate(ctx context.Context, r *flipt.EvaluationRequest) (*flipt.EvaluationResponse, error) {
	s.logger.WithField("request", r).Debug("evaluate")

	var (
		ts, _ = ptypes.TimestampProto(time.Now().UTC())
		resp  = &flipt.EvaluationResponse{
			RequestId:      r.RequestId,
			EntityId:       r.EntityId,
			RequestContext: r.Context,
			Timestamp:      ts,
			FlagKey:        r.FlagKey,
		}

		enabled bool

		err = s.builder.Select("enabled").
			From("flags").
			Where(sq.Eq{"key": r.FlagKey}).
			QueryRowContext(ctx).
			Scan(&enabled)
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return resp, ErrNotFoundf("flag %q", r.FlagKey)
		}
		return resp, err
	}

	if !enabled {
		return resp, ErrInvalidf("flag %q is disabled", r.FlagKey)
	}

	// get all rules for flag with their constraints
	rows, err := s.builder.Select("r.id, r.flag_key, r.segment_key, r.rank, c.type, c.property, c.operator, c.value").
		From("rules r").
		Join("constraints c ON (r.segment_key = c.segment_key)").
		Where(sq.Eq{"r.flag_key": r.FlagKey}).
		OrderBy("r.rank ASC").
		GroupBy("r.id").
		QueryContext(ctx)

	if err != nil {
		return resp, err
	}

	defer func() {
		if cerr := rows.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	var rules []rule

	for rows.Next() {
		var (
			existingRule   *rule
			tempRule       rule
			tempConstraint constraint
		)

		if err := rows.Scan(&tempRule.ID, &tempRule.FlagKey, &tempRule.SegmentKey, &tempRule.Rank, &tempConstraint.Type, &tempConstraint.Property, &tempConstraint.Operator, &tempConstraint.Value); err != nil {
			return resp, err
		}

		// current rule we know about
		if existingRule != nil && existingRule.ID == tempRule.ID {
			existingRule.Constraints = append(existingRule.Constraints, tempConstraint)
		} else {
			// haven't seen this rule before
			existingRule = &rule{
				ID:          tempRule.ID,
				FlagKey:     tempRule.FlagKey,
				SegmentKey:  tempRule.SegmentKey,
				Rank:        tempRule.Rank,
				Constraints: []constraint{tempConstraint},
			}
			rules = append(rules, *existingRule)
		}
	}

	if err := rows.Err(); err != nil {
		return resp, err
	}

	if len(rules) == 0 {
		return resp, nil
	}

	for _, rule := range rules {
		logger := s.logger.WithField("rule", rule)
		matchCount := 0

		for _, c := range rule.Constraints {
			if err := validate(c); err != nil {
				return resp, err
			}

			v := r.Context[c.Property]

			var (
				match bool
				err   error
			)

			switch c.Type {
			case flipt.ComparisonType_STRING_COMPARISON_TYPE:
				match, err = matchesString(c, v)
			case flipt.ComparisonType_NUMBER_COMPARISON_TYPE:
				match, err = matchesNumber(c, v)
			case flipt.ComparisonType_BOOLEAN_COMPARISON_TYPE:
				match, err = matchesBool(c, v)
			default:
				return resp, ErrInvalid("unknown constraint type")
			}

			if err != nil {
				return resp, err
			}

			logger = logger.WithFields(logrus.Fields{
				"constraint": c,
				"value":      v,
			})

			// constraint doesn't match, we can short circuit and move to the next rule
			// because we must match ALL constraints
			if !match {
				logger.Debug("does not match")
				break
			}

			// otherwise, increase the matchCount
			logger.Debug("matches")
			matchCount++
		}

		// all constraints did not match
		if matchCount != len(rule.Constraints) {
			continue
		}

		// otherwise, this is our matching rule, determine the flag variant to return
		// based on the distributions
		resp.Match = true
		resp.SegmentKey = rule.SegmentKey

		rows, err := s.builder.Select("d.rule_id", "d.variant_id", "d.rollout", "v.key").
			From("distributions d").
			Join("variants v ON (d.variant_id = v.id)").
			Where(sq.Eq{"d.rule_id": rule.ID}).
			QueryContext(ctx)

		if err != nil {
			return resp, err
		}

		defer func() {
			if cerr := rows.Close(); cerr != nil && err == nil {
				err = cerr
			}
		}()

		var (
			i             int
			distributions []distribution
			buckets       []int
		)

		for rows.Next() {
			var d distribution

			if err := rows.Scan(&d.RuleID, &d.VariantID, &d.Rollout, &d.VariantKey); err != nil {
				return resp, err
			}

			distributions = append(distributions, d)

			if i == 0 {
				buckets = append(buckets, int(d.Rollout*percentMultiplier))
			} else {
				buckets = append(buckets, buckets[i-1]+int(d.Rollout*percentMultiplier))
			}
			i++
		}

		if err := rows.Err(); err != nil {
			return resp, err
		}

		// no distributions for rule
		if len(distributions) == 0 {
			return resp, nil
		}

		var (
			bucket       = crc32Num(r.EntityId, r.FlagKey)
			index        = sort.SearchInts(buckets, int(bucket)+1)
			distribution = distributions[index]
		)

		resp.Value = distribution.VariantKey
		return resp, nil
	}

	return resp, nil
}

func crc32Num(entityID string, salt string) uint {
	return uint(crc32.ChecksumIEEE([]byte(salt+entityID))) % totalBucketNum
}

const (
	opEQ         = "eq"
	opNEQ        = "neq"
	opLT         = "lt"
	opLTE        = "lte"
	opGT         = "gt"
	opGTE        = "gte"
	opEmpty      = "empty"
	opNotEmpty   = "notempty"
	opTrue       = "true"
	opFalse      = "false"
	opPresent    = "present"
	opNotPresent = "notpresent"
)

var (
	validOperators = map[string]struct{}{
		opEQ:         {},
		opNEQ:        {},
		opLT:         {},
		opLTE:        {},
		opGT:         {},
		opGTE:        {},
		opEmpty:      {},
		opNotEmpty:   {},
		opTrue:       {},
		opFalse:      {},
		opPresent:    {},
		opNotPresent: {},
	}
	noValueOperators = map[string]struct{}{
		opEmpty:      {},
		opNotEmpty:   {},
		opPresent:    {},
		opNotPresent: {},
	}
	stringOperators = map[string]struct{}{
		opEQ:       {},
		opNEQ:      {},
		opEmpty:    {},
		opNotEmpty: {},
	}
	numberOperators = map[string]struct{}{
		opEQ:         {},
		opNEQ:        {},
		opLT:         {},
		opLTE:        {},
		opGT:         {},
		opGTE:        {},
		opPresent:    {},
		opNotPresent: {},
	}
	booleanOperators = map[string]struct{}{
		opTrue:       {},
		opFalse:      {},
		opPresent:    {},
		opNotPresent: {},
	}
)

const (
	// totalBucketNum represents how many buckets we can use to determine the consistent hashing
	// distribution and rollout
	totalBucketNum uint = 1000

	// percentMultiplier implies that the multiplier between percentage (100) and totalBucketNum
	percentMultiplier float32 = float32(totalBucketNum) / 100
)

func validate(c constraint) error {
	if c.Property == "" {
		return errors.New("empty property")
	}
	if c.Operator == "" {
		return errors.New("empty operator")
	}
	op := strings.ToLower(c.Operator)
	if _, ok := validOperators[op]; !ok {
		return fmt.Errorf("unsupported operator: %q", op)
	}

	return nil
}

func matchesString(c constraint, v string) (bool, error) {
	value := c.Value
	switch c.Operator {
	case opEQ:
		return value == v, nil
	case opNEQ:
		return value != v, nil
	case opEmpty:
		return len(strings.TrimSpace(v)) == 0, nil
	case opNotEmpty:
		return len(strings.TrimSpace(v)) != 0, nil
	}
	return false, nil
}

func matchesNumber(c constraint, v string) (bool, error) {
	switch c.Operator {
	case opNotPresent:
		return len(strings.TrimSpace(v)) == 0, nil
	case opPresent:
		return len(strings.TrimSpace(v)) != 0, nil
	}

	// can't parse an empty string
	if v == "" {
		return false, nil
	}

	n, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return false, fmt.Errorf("parsing number from %q", v)
	}

	value, err := strconv.ParseFloat(c.Value, 64)
	if err != nil {
		return false, fmt.Errorf("parsing number from %q", c.Value)
	}

	switch c.Operator {
	case opEQ:
		return value == n, nil
	case opNEQ:
		return value != n, nil
	case opLT:
		return n < value, nil
	case opLTE:
		return n <= value, nil
	case opGT:
		return n > value, nil
	case opGTE:
		return n >= value, nil
	}

	return false, nil
}

func matchesBool(c constraint, v string) (bool, error) {
	switch c.Operator {
	case opNotPresent:
		return len(strings.TrimSpace(v)) == 0, nil
	case opPresent:
		return len(strings.TrimSpace(v)) != 0, nil
	}

	// can't parse an empty string
	if v == "" {
		return false, nil
	}

	value, err := strconv.ParseBool(v)
	if err != nil {
		return false, fmt.Errorf("parsing boolean from %q", v)
	}

	switch c.Operator {
	case opTrue:
		return value, nil
	case opFalse:
		return !value, nil
	}
	return false, nil
}

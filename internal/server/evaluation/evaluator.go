package evaluation

import (
	"context"
	"errors"
	"hash/crc32"
	"sort"
	"strconv"
	"strings"
	"time"

	errs "go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/server/metrics"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

// Evaluator is an implementation of the MultiVariateEvaluator.
type Evaluator struct {
	logger *zap.Logger
	store  Storer
}

// NewEvaluator is the constructor for an Evaluator.
func NewEvaluator(logger *zap.Logger, store Storer) *Evaluator {
	return &Evaluator{
		logger: logger,
		store:  store,
	}
}

// MultiVariateEvaluator is an abstraction for evaluating a flag against a set of rules for multi-variate flags.
type MultiVariateEvaluator interface {
	Evaluate(ctx context.Context, r *flipt.EvaluationRequest) (*flipt.EvaluationResponse, error)
}

func (e *Evaluator) Evaluate(ctx context.Context, r *flipt.EvaluationRequest) (resp *flipt.EvaluationResponse, err error) {
	var (
		startTime     = time.Now().UTC()
		namespaceAttr = metrics.AttributeNamespace.String(r.NamespaceKey)
		flagAttr      = metrics.AttributeFlag.String(r.FlagKey)
	)

	metrics.EvaluationsTotal.Add(ctx, 1, metric.WithAttributeSet(attribute.NewSet(namespaceAttr, flagAttr)))

	defer func() {
		if err == nil {
			metrics.EvaluationResultsTotal.Add(ctx, 1,
				metric.WithAttributeSet(
					attribute.NewSet(
						namespaceAttr,
						flagAttr,
						metrics.AttributeMatch.Bool(resp.Match),
						metrics.AttributeSegment.String(resp.SegmentKey),
						metrics.AttributeReason.String(resp.Reason.String()),
						metrics.AttributeValue.String(resp.Value),
					),
				),
			)
		} else {
			metrics.EvaluationErrorsTotal.Add(ctx, 1, metric.WithAttributeSet(attribute.NewSet(namespaceAttr, flagAttr)))
		}

		metrics.EvaluationLatency.Record(
			ctx,
			float64(time.Since(startTime).Nanoseconds())/1e6,
			metric.WithAttributeSet(
				attribute.NewSet(
					namespaceAttr,
					flagAttr,
				),
			),
		)
	}()

	resp = &flipt.EvaluationResponse{
		RequestId:      r.RequestId,
		EntityId:       r.EntityId,
		RequestContext: r.Context,
		FlagKey:        r.FlagKey,
		NamespaceKey:   r.NamespaceKey,
	}

	flag, err := e.store.GetFlag(ctx, r.NamespaceKey, r.FlagKey)
	if err != nil {
		resp.Reason = flipt.EvaluationReason_ERROR_EVALUATION_REASON

		var errnf errs.ErrNotFound
		if errors.As(err, &errnf) {
			resp.Reason = flipt.EvaluationReason_FLAG_NOT_FOUND_EVALUATION_REASON
		}

		return resp, err
	}

	if !flag.Enabled {
		resp.Match = false
		resp.Reason = flipt.EvaluationReason_FLAG_DISABLED_EVALUATION_REASON
		return resp, nil
	}

	rules, err := e.store.GetEvaluationRules(ctx, r.NamespaceKey, r.FlagKey)
	if err != nil {
		resp.Reason = flipt.EvaluationReason_ERROR_EVALUATION_REASON
		return resp, err
	}

	if len(rules) == 0 {
		e.logger.Debug("no rules match")
		return resp, nil
	}

	var lastRank int32

	// rule loop
	for _, rule := range rules {
		if rule.Rank < lastRank {
			resp.Reason = flipt.EvaluationReason_ERROR_EVALUATION_REASON
			return resp, errs.ErrInvalidf("rule rank: %d detected out of order", rule.Rank)
		}

		lastRank = rule.Rank

		matched, err := doConstraintsMatch(e.logger, r.Context, rule.Constraints, rule.SegmentMatchType)
		if err != nil {
			resp.Reason = flipt.EvaluationReason_ERROR_EVALUATION_REASON
			return resp, err
		}

		if !matched {
			continue
		}

		// otherwise, this is our matching rule, determine the flag variant to return
		// based on the distributions
		resp.SegmentKey = rule.SegmentKey

		distributions, err := e.store.GetEvaluationDistributions(ctx, rule.ID)
		if err != nil {
			resp.Reason = flipt.EvaluationReason_ERROR_EVALUATION_REASON
			return resp, err
		}

		var (
			validDistributions []*storage.EvaluationDistribution
			buckets            []int
		)

		for _, d := range distributions {
			// don't include 0% rollouts
			if d.Rollout > 0 {
				validDistributions = append(validDistributions, d)

				if buckets == nil {
					bucket := int(d.Rollout * percentMultiplier)
					buckets = append(buckets, bucket)
				} else {
					bucket := buckets[len(buckets)-1] + int(d.Rollout*percentMultiplier)
					buckets = append(buckets, bucket)
				}
			}
		}

		// no distributions for rule
		if len(validDistributions) == 0 {
			e.logger.Info("no distributions for rule")
			resp.Match = true
			resp.Reason = flipt.EvaluationReason_MATCH_EVALUATION_REASON
			return resp, nil
		}

		var (
			bucket = crc32Num(r.EntityId, r.FlagKey)
			// sort.SearchInts searches for x in a sorted slice of ints and returns the index
			// as specified by Search. The return value is the index to insert x if x is
			// not present (it could be len(a)).
			index = sort.SearchInts(buckets, int(bucket)+1)
		)

		// if index is outside of our existing buckets then it does not match any distribution
		if index == len(validDistributions) {
			resp.Match = false
			e.logger.Debug("did not match any distributions")
			return resp, nil
		}

		d := validDistributions[index]
		e.logger.Debug("matched distribution", zap.Reflect("evaluation_distribution", d))

		resp.Match = true
		resp.Value = d.VariantKey
		resp.Attachment = d.VariantAttachment
		resp.Reason = flipt.EvaluationReason_MATCH_EVALUATION_REASON
		return resp, nil
	} // end rule loop

	return resp, nil
}

// doConstraintsMatch is a utility function that will return if all or any constraints have matched for a segment depending
// on the match type.
func doConstraintsMatch(logger *zap.Logger, evalCtx map[string]string, constraints []storage.EvaluationConstraint, segmentMatchType flipt.MatchType) (bool, error) {
	constraintMatches := 0

	for _, c := range constraints {
		v := evalCtx[c.Property]

		var (
			match bool
			err   error
		)

		switch c.Type {
		case flipt.ComparisonType_STRING_COMPARISON_TYPE:
			match = matchesString(c, v)
		case flipt.ComparisonType_NUMBER_COMPARISON_TYPE:
			match, err = matchesNumber(c, v)
		case flipt.ComparisonType_BOOLEAN_COMPARISON_TYPE:
			match, err = matchesBool(c, v)
		case flipt.ComparisonType_DATETIME_COMPARISON_TYPE:
			match, err = matchesDateTime(c, v)
		default:
			return false, errs.ErrInvalid("unknown constraint type")
		}

		if err != nil {
			return false, err
		}

		if match {
			// increase the matchCount
			constraintMatches++

			switch segmentMatchType {
			case flipt.MatchType_ANY_MATCH_TYPE:
				// can short circuit here since we had at least one match
				break
			default:
				// keep looping as we need to match all constraints
				continue
			}
		} else {
			// no match
			switch segmentMatchType {
			case flipt.MatchType_ALL_MATCH_TYPE:
				// we can short circuit because we must match all constraints
				break
			default:
				// keep looping to see if we match the next constraint
				continue
			}
		}
	}

	var matched = true

	switch segmentMatchType {
	case flipt.MatchType_ALL_MATCH_TYPE:
		if len(constraints) != constraintMatches {
			logger.Debug("did not match ALL constraints")
			matched = false
		}

	case flipt.MatchType_ANY_MATCH_TYPE:
		if len(constraints) > 0 && constraintMatches == 0 {
			logger.Debug("did not match ANY constraints")
			matched = false
		}
	default:
		logger.Error("unknown match type", zap.Int32("match_type", int32(segmentMatchType)))
		matched = false
	}

	return matched, nil
}

func crc32Num(entityID string, salt string) uint {
	return uint(crc32.ChecksumIEEE([]byte(salt+entityID))) % totalBucketNum
}

const (
	// totalBucketNum represents how many buckets we can use to determine the consistent hashing
	// distribution and rollout
	totalBucketNum uint = 1000

	// percentMultiplier implies that the multiplier between percentage (100) and totalBucketNum
	percentMultiplier float32 = float32(totalBucketNum) / 100
)

func matchesString(c storage.EvaluationConstraint, v string) bool {
	switch c.Operator {
	case flipt.OpEmpty:
		return len(strings.TrimSpace(v)) == 0
	case flipt.OpNotEmpty:
		return len(strings.TrimSpace(v)) != 0
	}

	if v == "" {
		return false
	}

	value := c.Value

	switch c.Operator {
	case flipt.OpEQ:
		return value == v
	case flipt.OpNEQ:
		return value != v
	case flipt.OpPrefix:
		return strings.HasPrefix(strings.TrimSpace(v), value)
	case flipt.OpSuffix:
		return strings.HasSuffix(strings.TrimSpace(v), value)
	}

	return false
}

func matchesNumber(c storage.EvaluationConstraint, v string) (bool, error) {
	switch c.Operator {
	case flipt.OpNotPresent:
		return len(strings.TrimSpace(v)) == 0, nil
	case flipt.OpPresent:
		return len(strings.TrimSpace(v)) != 0, nil
	}

	// can't parse an empty string
	if v == "" {
		return false, nil
	}

	n, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return false, errs.ErrInvalidf("parsing number from %q", v)
	}

	// TODO: we should consider parsing this at creation time since it doesn't change and it doesnt make sense to allow invalid constraint values
	value, err := strconv.ParseFloat(c.Value, 64)
	if err != nil {
		return false, errs.ErrInvalidf("parsing number from %q", c.Value)
	}

	switch c.Operator {
	case flipt.OpEQ:
		return value == n, nil
	case flipt.OpNEQ:
		return value != n, nil
	case flipt.OpLT:
		return n < value, nil
	case flipt.OpLTE:
		return n <= value, nil
	case flipt.OpGT:
		return n > value, nil
	case flipt.OpGTE:
		return n >= value, nil
	}

	return false, nil
}

func matchesBool(c storage.EvaluationConstraint, v string) (bool, error) {
	switch c.Operator {
	case flipt.OpNotPresent:
		return len(strings.TrimSpace(v)) == 0, nil
	case flipt.OpPresent:
		return len(strings.TrimSpace(v)) != 0, nil
	}

	// can't parse an empty string
	if v == "" {
		return false, nil
	}

	value, err := strconv.ParseBool(v)
	if err != nil {
		return false, errs.ErrInvalidf("parsing boolean from %q", v)
	}

	switch c.Operator {
	case flipt.OpTrue:
		return value, nil
	case flipt.OpFalse:
		return !value, nil
	}

	return false, nil
}

func matchesDateTime(c storage.EvaluationConstraint, v string) (bool, error) {
	switch c.Operator {
	case flipt.OpNotPresent:
		return len(strings.TrimSpace(v)) == 0, nil
	case flipt.OpPresent:
		return len(strings.TrimSpace(v)) != 0, nil
	}

	// can't parse an empty string
	if v == "" {
		return false, nil
	}

	d, err := tryParseDateTime(v)
	if err != nil {
		return false, err
	}

	value, err := tryParseDateTime(c.Value)
	if err != nil {
		return false, err
	}

	switch c.Operator {
	case flipt.OpEQ:
		return value.Equal(d), nil
	case flipt.OpNEQ:
		return !value.Equal(d), nil
	case flipt.OpLT:
		return d.Before(value), nil
	case flipt.OpLTE:
		return d.Before(value) || value.Equal(d), nil
	case flipt.OpGT:
		return d.After(value), nil
	case flipt.OpGTE:
		return d.After(value) || value.Equal(d), nil
	}

	return false, nil
}

func tryParseDateTime(v string) (time.Time, error) {
	if d, err := time.Parse(time.RFC3339, v); err == nil {
		return d.UTC(), nil
	}

	if d, err := time.Parse(time.DateOnly, v); err == nil {
		return d.UTC(), nil
	}

	return time.Time{}, errs.ErrInvalidf("parsing datetime from %q", v)
}

package evaluation

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	errs "go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/server/metrics"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/rpc/flipt/evaluation"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

// Evaluator is an evaluator for legacy flag evaluations.
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

func (e *Evaluator) Evaluate(ctx context.Context, flag *flipt.Flag, r *evaluation.EvaluationRequest) (resp *flipt.EvaluationResponse, err error) {
	// Flag should be of variant type by default. However, for maximum backwards compatibility
	// on this layer, we will check against the integer value of the flag type.
	if int(flag.Type) != 0 || flag.Type != flipt.FlagType_VARIANT_FLAG_TYPE {
		resp = &flipt.EvaluationResponse{}
		resp.Reason = flipt.EvaluationReason_ERROR_EVALUATION_REASON
		return resp, errs.ErrInvalidf("flag type %s invalid", flag.Type)
	}

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
						metrics.AttributeType.String("variant"),
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

	if flag.DefaultVariant != nil {
		resp.Reason = flipt.EvaluationReason_DEFAULT_EVALUATION_REASON
		resp.Value = flag.DefaultVariant.Key
		resp.Attachment = flag.DefaultVariant.Attachment
	}

	if !flag.Enabled {
		resp.Match = false
		resp.Reason = flipt.EvaluationReason_FLAG_DISABLED_EVALUATION_REASON
		return resp, nil
	}

	rules, err := e.store.GetEvaluationRules(ctx, storage.NewResource(r.NamespaceKey, r.FlagKey))
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

		segmentKeys := make([]string, 0, len(rule.Segments))
		segmentMatches := 0

		for k, v := range rule.Segments {
			matched, reason, err := e.matchConstraints(r.Context, v.Constraints, v.MatchType, r.EntityId)
			if err != nil {
				resp.Reason = flipt.EvaluationReason_ERROR_EVALUATION_REASON
				return resp, err
			}

			if matched {
				e.logger.Debug(reason)
				segmentKeys = append(segmentKeys, k)
				segmentMatches++
			}
		}

		switch rule.SegmentOperator {
		case flipt.SegmentOperator_OR_SEGMENT_OPERATOR:
			if segmentMatches < 1 {
				e.logger.Debug("did not match ANY segments")
				continue
			}
		case flipt.SegmentOperator_AND_SEGMENT_OPERATOR:
			if len(rule.Segments) != segmentMatches {
				e.logger.Debug("did not match ALL segments")
				continue
			}
		}

		// For legacy reasons of supporting SegmentKey.
		// The old EvaluationResponse will return both the SegmentKey, and SegmentKeys.
		// If there are multiple segmentKeys, the old EvaluationResponse will take the first value
		// in the segmentKeys slice.
		if len(segmentKeys) > 0 {
			resp.SegmentKey = segmentKeys[0]
			resp.SegmentKeys = segmentKeys
		}

		distributions, err := e.store.GetEvaluationDistributions(ctx, storage.NewID(rule.ID))
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
		// match is true here because it did match the segment/rule
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
			e.logger.Debug("did not match any distributions")
			resp.Match = false
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

// matchConstraints is a utility function that will return if all or any constraints have matched for a segment depending
// on the match type.
func (e *Evaluator) matchConstraints(evalCtx map[string]string, constraints []storage.EvaluationConstraint, segmentMatchType flipt.MatchType, entityId string) (bool, string, error) {
	constraintMatches := 0

	var reason string

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
		case flipt.ComparisonType_ENTITY_ID_COMPARISON_TYPE:
			match = matchesString(c, entityId)
		default:
			return false, reason, errs.ErrInvalid("unknown constraint type")
		}

		if err != nil {
			e.logger.Debug("error matching constraint", zap.String("property", c.Property), zap.Error(err))
			// don't return here because we want to continue to evaluate the other constraints
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
			reason = "did not match ALL constraints"
			matched = false
		}

	case flipt.MatchType_ANY_MATCH_TYPE:
		if len(constraints) > 0 && constraintMatches == 0 {
			reason = "did not match ANY constraints"
			matched = false
		}
	default:
		reason = fmt.Sprintf("unknown match type %d", segmentMatchType)
		matched = false
	}

	return matched, reason, nil
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
	case flipt.OpIsOneOf:
		values := []string{}
		if err := json.Unmarshal([]byte(value), &values); err != nil {
			return false
		}
		return slices.Contains(values, v)
	case flipt.OpIsNotOneOf:
		values := []string{}
		if err := json.Unmarshal([]byte(value), &values); err != nil {
			return false
		}
		return !slices.Contains(values, v)
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

	if c.Operator == flipt.OpIsOneOf {
		values := []float64{}
		if err := json.Unmarshal([]byte(c.Value), &values); err != nil {
			return false, errs.ErrInvalidf("Invalid value for constraint %q", c.Value)
		}
		return slices.Contains(values, n), nil
	} else if c.Operator == flipt.OpIsNotOneOf {
		values := []float64{}
		if err := json.Unmarshal([]byte(c.Value), &values); err != nil {
			return false, errs.ErrInvalidf("Invalid value for constraint %q", c.Value)
		}
		return !slices.Contains(values, n), nil
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

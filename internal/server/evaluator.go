package server

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
	flipt "go.flipt.io/flipt/rpc/flipt"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Evaluate evaluates a request for a given flag and entity
func (s *Server) Evaluate(ctx context.Context, r *flipt.EvaluationRequest) (*flipt.EvaluationResponse, error) {
	s.logger.Debug("evaluate", zap.Stringer("request", r))
	resp, err := s.evaluate(ctx, r)
	if err != nil {
		return resp, err
	}

	spanAttrs := []attribute.KeyValue{
		attribute.String("flipt.flag_key", r.FlagKey),
		attribute.String("flipt.entity_id", r.EntityId),
		attribute.String("flipt.request_id", r.RequestId),
	}

	if resp != nil {
		spanAttrs = append(spanAttrs,
			attribute.Bool("flipt.match", resp.Match),
			attribute.String("flipt.segment_key", resp.SegmentKey),
			attribute.String("flipt.value", resp.Value),
			attribute.String("flipt.reason", resp.Reason.String()),
			attribute.Float64("flipt.duration_ms", resp.RequestDurationMillis),
		)
	}

	// add otel attributes to span
	span := trace.SpanFromContext(ctx)
	span.AddEvent("flipt.evaluate", trace.WithAttributes(spanAttrs...))

	s.logger.Debug("evaluate", zap.Stringer("response", resp))
	return resp, nil
}

// BatchEvaluate evaluates a request for multiple flags and entities
func (s *Server) BatchEvaluate(ctx context.Context, r *flipt.BatchEvaluationRequest) (*flipt.BatchEvaluationResponse, error) {
	s.logger.Debug("batch-evaluate", zap.Stringer("request", r))
	resp, err := s.batchEvaluate(ctx, r)
	if err != nil {
		return nil, err
	}
	s.logger.Debug("batch-evaluate", zap.Stringer("response", resp))
	return resp, nil
}

func (s *Server) batchEvaluate(ctx context.Context, r *flipt.BatchEvaluationRequest) (*flipt.BatchEvaluationResponse, error) {
	res := flipt.BatchEvaluationResponse{
		Responses: make([]*flipt.EvaluationResponse, 0, len(r.GetRequests())),
	}

	// TODO: we should change this to a native batch query instead of looping through
	// each request individually
	for _, flag := range r.GetRequests() {
		// TODO: we also need to validate each request, we should likely do this in the validation middleware
		f, err := s.evaluate(ctx, flag)
		if err != nil {
			var errnf errs.ErrNotFound
			if r.GetExcludeNotFound() && errors.As(err, &errnf) {
				continue
			}
			return &res, err
		}
		f.RequestId = ""
		res.Responses = append(res.Responses, f)
	}

	return &res, nil
}

func (s *Server) evaluate(ctx context.Context, r *flipt.EvaluationRequest) (resp *flipt.EvaluationResponse, err error) {
	var (
		startTime = time.Now().UTC()
		flagAttr  = metrics.AttributeFlag.String(r.FlagKey)
	)
	metrics.EvaluationsTotal.Add(ctx, 1, flagAttr)
	defer func() {
		if err == nil {
			metrics.EvaluationResultsTotal.Add(ctx, 1,
				flagAttr,
				metrics.AttributeMatch.Bool(resp.Match),
				metrics.AttributeSegment.String(resp.SegmentKey),
				metrics.AttributeReason.String(resp.Reason.String()),
				metrics.AttributeValue.String(resp.Value),
			)
		} else {
			metrics.EvaluationErrorsTotal.Add(ctx, 1, flagAttr)
		}

		metrics.EvaluationLatency.Record(
			ctx,
			float64(time.Since(startTime).Nanoseconds())/1e6,
			flagAttr,
		)
	}()

	resp = &flipt.EvaluationResponse{
		RequestId:      r.RequestId,
		EntityId:       r.EntityId,
		RequestContext: r.Context,
		FlagKey:        r.FlagKey,
	}

	flag, err := s.store.GetFlag(ctx, r.FlagKey)
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

	rules, err := s.store.GetEvaluationRules(ctx, r.FlagKey)
	if err != nil {
		resp.Reason = flipt.EvaluationReason_ERROR_EVALUATION_REASON
		return resp, err
	}

	if len(rules) == 0 {
		s.logger.Debug("no rules match")
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

		constraintMatches := 0

		// constraint loop
		for _, c := range rule.Constraints {
			v := r.Context[c.Property]

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
			default:
				resp.Reason = flipt.EvaluationReason_ERROR_EVALUATION_REASON
				return resp, errs.ErrInvalid("unknown constraint type")
			}

			if err != nil {
				resp.Reason = flipt.EvaluationReason_ERROR_EVALUATION_REASON
				return resp, err
			}

			if match {
				s.logger.Debug("constraint matches", zap.Reflect("constraint", c))

				// increase the matchCount
				constraintMatches++

				switch rule.SegmentMatchType {
				case flipt.MatchType_ANY_MATCH_TYPE:
					// can short circuit here since we had at least one match
					break
				default:
					// keep looping as we need to match all constraints
					continue
				}
			} else {
				// no match
				s.logger.Debug("constraint does not match", zap.Reflect("constraint", c))

				switch rule.SegmentMatchType {
				case flipt.MatchType_ALL_MATCH_TYPE:
					// we can short circuit because we must match all constraints
					break
				default:
					// keep looping to see if we match the next constraint
					continue
				}
			}
		} // end constraint loop

		switch rule.SegmentMatchType {
		case flipt.MatchType_ALL_MATCH_TYPE:
			if len(rule.Constraints) != constraintMatches {
				// all constraints did not match, continue to next rule
				s.logger.Debug("did not match ALL constraints")
				continue
			}

		case flipt.MatchType_ANY_MATCH_TYPE:
			if len(rule.Constraints) > 0 && constraintMatches == 0 {
				// no constraints matched, continue to next rule
				s.logger.Debug("did not match ANY constraints")
				continue
			}
		default:
			s.logger.Error("unknown match type", zap.Int32("match_type", int32(rule.SegmentMatchType)))
			continue
		}

		// otherwise, this is our matching rule, determine the flag variant to return
		// based on the distributions
		resp.SegmentKey = rule.SegmentKey

		distributions, err := s.store.GetEvaluationDistributions(ctx, rule.ID)
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
			s.logger.Info("no distributions for rule")
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
			s.logger.Debug("did not match any distributions")
			return resp, nil
		}

		d := validDistributions[index]
		s.logger.Debug("matched distribution", zap.Reflect("evaluation_distribution", d))

		resp.Match = true
		resp.Value = d.VariantKey
		resp.Attachment = d.VariantAttachment
		resp.Reason = flipt.EvaluationReason_MATCH_EVALUATION_REASON
		return resp, nil
	} // end rule loop

	return resp, nil
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

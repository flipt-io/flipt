package evaluation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	errs "go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/server/metrics"
	fliptotel "go.flipt.io/flipt/internal/server/otel"
	"go.flipt.io/flipt/internal/storage"
	flipt "go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/rpc/flipt/core"
	rpcevaluation "go.flipt.io/flipt/rpc/flipt/evaluation"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Variant evaluates a request for a multi-variate flag and entity.
func (s *Server) Variant(ctx context.Context, r *rpcevaluation.EvaluationRequest) (*rpcevaluation.VariantEvaluationResponse, error) {
	store, err := s.getEvalStore(ctx)
	if err != nil {
		return nil, err
	}

	flag, err := store.GetFlag(ctx, storage.NewResource(r.NamespaceKey, r.FlagKey, storage.WithReference(r.Reference)))
	if err != nil {
		return nil, err
	}

	if flag.Type != core.FlagType_VARIANT_FLAG_TYPE {
		return nil, errs.ErrInvalidf("flag type %s invalid", flag.Type)
	}

	resp, err := s.variant(ctx, store, flag, r)
	if err != nil {
		return nil, err
	}

	spanAttrs := []attribute.KeyValue{
		// TODO: fliptotel.AttributeEnvironment.String(r.EnvironmentKey),
		fliptotel.AttributeNamespace.String(r.NamespaceKey),
		fliptotel.AttributeFlag.String(r.FlagKey),
		fliptotel.AttributeEntityID.String(r.EntityId),
		fliptotel.AttributeRequestID.String(r.RequestId),
		fliptotel.AttributeMatch.Bool(resp.Match),
		fliptotel.AttributeValue.String(resp.VariantKey),
		fliptotel.AttributeReason.String(resp.Reason.String()),
		fliptotel.AttributeSegments.StringSlice(resp.SegmentKeys),
		fliptotel.AttributeFlagKey(resp.FlagKey),
		fliptotel.AttributeProviderName,
		fliptotel.AttributeFlagVariant(resp.VariantKey),
	}

	// add otel attributes to span
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(spanAttrs...)

	return resp, nil
}

func (s *Server) variant(ctx context.Context, store storage.ReadOnlyStore, flag *core.Flag, r *rpcevaluation.EvaluationRequest) (*rpcevaluation.VariantEvaluationResponse, error) {
	var (
		resp = &rpcevaluation.VariantEvaluationResponse{
			FlagKey:   flag.Key,
			RequestId: r.RequestId,
		}
		err      error
		lastRank int32

		startTime = time.Now().UTC()
		// TODO: environmentAttr = metrics.AttributeEnvironment.String(r.EnvironmentKey)
		namespaceAttr = metrics.AttributeNamespace.String(r.NamespaceKey)
		flagAttr      = metrics.AttributeFlag.String(r.FlagKey)
	)

	metrics.EvaluationsTotal.Add(ctx, 1, metric.WithAttributeSet(attribute.NewSet(namespaceAttr, flagAttr)))

	defer func() {
		if err == nil {
			metrics.EvaluationResultsTotal.Add(ctx, 1,
				metric.WithAttributeSet(
					attribute.NewSet(
						// TODO: environmentAttr,
						namespaceAttr,
						flagAttr,
						metrics.AttributeMatch.Bool(resp.Match),
						metrics.AttributeSegments.StringSlice(resp.SegmentKeys),
						metrics.AttributeReason.String(resp.Reason.String()),
						metrics.AttributeValue.String(resp.VariantKey),
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

	if flag.DefaultVariant != nil {
		dv, ok := getVariant(flag.GetDefaultVariant(), flag.Variants...)
		if !ok {
			return nil, fmt.Errorf("default variant not found: %q", flag.GetDefaultVariant())
		}

		var attachment string
		if dv.Attachment != nil {
			a, err := dv.Attachment.MarshalJSON()
			if err != nil {
				return nil, err
			}

			attachment = string(a)
		}

		resp.Reason = rpcevaluation.EvaluationReason_DEFAULT_EVALUATION_REASON
		resp.VariantKey = dv.Key
		resp.VariantAttachment = attachment
	}

	if !flag.Enabled {
		resp.Match = false
		resp.Reason = rpcevaluation.EvaluationReason_FLAG_DISABLED_EVALUATION_REASON
		return resp, nil
	}

	rules, err := store.GetEvaluationRules(ctx, storage.NewResource(r.NamespaceKey, r.FlagKey))
	if err != nil {
		return nil, err
	}

	if len(rules) == 0 {
		s.logger.Debug("no rules match")
		return resp, nil
	}

	// rule loop
	for _, rule := range rules {
		if rule.Rank < lastRank {
			return resp, errs.ErrInvalidf("rule rank: %d detected out of order", rule.Rank)
		}

		lastRank = rule.Rank

		segmentKeys := make([]string, 0, len(rule.Segments))
		segmentMatches := 0

		for k, v := range rule.Segments {
			matched, reason, err := s.matchConstraints(r.Context, v.Constraints, v.MatchType, r.EntityId)
			if err != nil {
				return resp, err
			}

			if matched {
				s.logger.Debug(reason)
				segmentKeys = append(segmentKeys, k)
				segmentMatches++
			}
		}

		switch rule.SegmentOperator {
		case core.SegmentOperator_OR_SEGMENT_OPERATOR:
			if segmentMatches < 1 {
				s.logger.Debug("did not match ANY segments")
				continue
			}
		case core.SegmentOperator_AND_SEGMENT_OPERATOR:
			if len(rule.Segments) != segmentMatches {
				s.logger.Debug("did not match ALL segments")
				continue
			}
		}

		if len(segmentKeys) > 0 {
			resp.SegmentKeys = segmentKeys
		}

		distributions, err := store.GetEvaluationDistributions(ctx, storage.NewResource(r.NamespaceKey, r.FlagKey), storage.NewID(rule.ID))
		if err != nil {
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
			s.logger.Info("no distributions for rule")
			resp.Match = true
			resp.Reason = rpcevaluation.EvaluationReason_MATCH_EVALUATION_REASON
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
			s.logger.Debug("did not match any distributions")
			resp.Match = false
			return resp, nil
		}

		d := validDistributions[index]
		s.logger.Debug("matched distribution", zap.Reflect("evaluation_distribution", d))

		resp.Match = true
		resp.VariantKey = d.VariantKey
		resp.VariantAttachment = d.VariantAttachment
		resp.Reason = rpcevaluation.EvaluationReason_MATCH_EVALUATION_REASON
		return resp, nil
	} // end rule loop

	return resp, nil
}

// Boolean evaluates a request for a boolean flag and entity.
func (s *Server) Boolean(ctx context.Context, r *rpcevaluation.EvaluationRequest) (*rpcevaluation.BooleanEvaluationResponse, error) {
	store, err := s.getEvalStore(ctx)
	if err != nil {
		return nil, err
	}

	flag, err := store.GetFlag(ctx, storage.NewResource(r.NamespaceKey, r.FlagKey, storage.WithReference(r.Reference)))
	if err != nil {
		return nil, err
	}

	if flag.Type != core.FlagType_BOOLEAN_FLAG_TYPE {
		return nil, errs.ErrInvalidf("flag type %s invalid", flag.Type)
	}

	resp, err := s.boolean(ctx, store, flag, r)
	if err != nil {
		return nil, err
	}

	spanAttrs := []attribute.KeyValue{
		// TODO: fliptotel.AttributeEnvironment.String(r.EnvironmentKey),
		fliptotel.AttributeNamespace.String(r.NamespaceKey),
		fliptotel.AttributeFlag.String(r.FlagKey),
		fliptotel.AttributeEntityID.String(r.EntityId),
		fliptotel.AttributeRequestID.String(r.RequestId),
		fliptotel.AttributeValue.Bool(resp.Enabled),
		fliptotel.AttributeReason.String(resp.Reason.String()),
		fliptotel.AttributeFlagKey(r.FlagKey),
		fliptotel.AttributeProviderName,
		fliptotel.AttributeFlagVariant(strconv.FormatBool(resp.Enabled)),
	}

	// add otel attributes to span
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(spanAttrs...)

	return resp, nil
}

func (s *Server) boolean(ctx context.Context, store storage.ReadOnlyStore, flag *core.Flag, r *rpcevaluation.EvaluationRequest) (*rpcevaluation.BooleanEvaluationResponse, error) {
	rollouts, err := store.GetEvaluationRollouts(ctx, storage.NewResource(r.NamespaceKey, flag.Key, storage.WithReference(r.Reference)))
	if err != nil {
		return nil, err
	}

	var (
		resp = &rpcevaluation.BooleanEvaluationResponse{
			RequestId: r.RequestId,
		}
		lastRank int32
	)

	var (
		startTime = time.Now().UTC()
		// TODO: environmentAttr = metrics.AttributeEnvironment.String(r.EnvironmentKey)
		namespaceAttr = metrics.AttributeNamespace.String(r.NamespaceKey)
		flagAttr      = metrics.AttributeFlag.String(r.FlagKey)
	)

	metrics.EvaluationsTotal.Add(ctx, 1, metric.WithAttributeSet(attribute.NewSet(namespaceAttr, flagAttr)))

	defer func() {
		if err == nil {
			metrics.EvaluationResultsTotal.Add(ctx, 1,
				metric.WithAttributeSet(
					attribute.NewSet(
						// TODO: environmentAttr,
						namespaceAttr,
						flagAttr,
						metrics.AttributeValue.Bool(resp.Enabled),
						metrics.AttributeReason.String(resp.Reason.String()),
						metrics.AttributeType.String("boolean"),
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

	for _, rollout := range rollouts {
		if rollout.Rank < lastRank {
			return nil, fmt.Errorf("rollout rank: %d detected out of order", rollout.Rank)
		}

		lastRank = rollout.Rank

		if rollout.Threshold != nil {
			// consistent hashing based on the entity id and flag key.
			hash := crc32.ChecksumIEEE([]byte(r.EntityId + r.FlagKey))

			normalizedValue := float32(int(hash) % 100)

			// if this case does not hold, fall through to the next rollout.
			if normalizedValue < rollout.Threshold.Percentage {
				resp.Enabled = rollout.Threshold.Value
				resp.Reason = rpcevaluation.EvaluationReason_MATCH_EVALUATION_REASON
				resp.FlagKey = flag.Key
				s.logger.Debug("threshold based matched", zap.Int("rank", int(rollout.Rank)), zap.String("rollout_type", "threshold"))
				return resp, nil
			}
		} else if rollout.Segment != nil {

			var (
				segmentMatches = 0
				segmentKeys    = []string{}
			)

			for k, v := range rollout.Segment.Segments {
				segmentKeys = append(segmentKeys, k)
				matched, reason, err := s.matchConstraints(r.Context, v.Constraints, v.MatchType, r.EntityId)
				if err != nil {
					return nil, err
				}

				// if we don't match the segment, fall through to the next rollout.
				if matched {
					s.logger.Debug(reason)
					segmentMatches++
				}
			}

			switch rollout.Segment.SegmentOperator {
			case core.SegmentOperator_OR_SEGMENT_OPERATOR:
				if segmentMatches < 1 {
					s.logger.Debug("did not match ANY segments")
					continue
				}
			case core.SegmentOperator_AND_SEGMENT_OPERATOR:
				if len(rollout.Segment.Segments) != segmentMatches {
					s.logger.Debug("did not match ALL segments")
					continue
				}
			}

			resp.Enabled = rollout.Segment.Value
			resp.Reason = rpcevaluation.EvaluationReason_MATCH_EVALUATION_REASON
			resp.FlagKey = flag.Key

			s.logger.Debug("segment based matched", zap.Int("rank", int(rollout.Rank)), zap.Strings("segments", segmentKeys))
			return resp, nil
		}
	}

	// If we have exhausted all rollouts and we still don't have a match, return flag enabled value.
	resp.Reason = rpcevaluation.EvaluationReason_DEFAULT_EVALUATION_REASON
	resp.Enabled = flag.Enabled
	resp.FlagKey = flag.Key

	s.logger.Debug("default rollout matched", zap.Bool("enabled", flag.Enabled))
	return resp, nil
}

// Batch takes in a list of *evaluation.EvaluationRequest and returns their respective responses.
func (s *Server) Batch(ctx context.Context, b *rpcevaluation.BatchEvaluationRequest) (*rpcevaluation.BatchEvaluationResponse, error) {
	store, err := s.getEvalStore(ctx)
	if err != nil {
		return nil, err
	}

	resp := &rpcevaluation.BatchEvaluationResponse{
		Responses: make([]*rpcevaluation.EvaluationResponse, 0, len(b.Requests)),
	}

	for _, req := range b.GetRequests() {
		f, err := store.GetFlag(ctx, storage.NewResource(req.NamespaceKey, req.FlagKey, storage.WithReference(b.Reference)))
		if err != nil {
			var errnf errs.ErrNotFound
			if errors.As(err, &errnf) {
				eresp := &rpcevaluation.EvaluationResponse{
					Type: rpcevaluation.EvaluationResponseType_ERROR_EVALUATION_RESPONSE_TYPE,
					Response: &rpcevaluation.EvaluationResponse_ErrorResponse{
						ErrorResponse: &rpcevaluation.ErrorEvaluationResponse{
							FlagKey:      req.FlagKey,
							NamespaceKey: req.NamespaceKey,
							Reason:       rpcevaluation.ErrorEvaluationReason_NOT_FOUND_ERROR_EVALUATION_REASON,
						},
					},
				}

				resp.Responses = append(resp.Responses, eresp)
				continue
			}

			return nil, err
		}

		switch f.Type {
		case core.FlagType_BOOLEAN_FLAG_TYPE:
			res, err := s.boolean(ctx, store, f, req)
			if err != nil {
				return nil, err
			}

			eresp := &rpcevaluation.EvaluationResponse{
				Type: rpcevaluation.EvaluationResponseType_BOOLEAN_EVALUATION_RESPONSE_TYPE,
				Response: &rpcevaluation.EvaluationResponse_BooleanResponse{
					BooleanResponse: res,
				},
			}

			resp.Responses = append(resp.Responses, eresp)
		case core.FlagType_VARIANT_FLAG_TYPE:
			res, err := s.variant(ctx, store, f, req)
			if err != nil {
				return nil, err
			}
			eresp := &rpcevaluation.EvaluationResponse{
				Type: rpcevaluation.EvaluationResponseType_VARIANT_EVALUATION_RESPONSE_TYPE,
				Response: &rpcevaluation.EvaluationResponse_VariantResponse{
					VariantResponse: res,
				},
			}

			resp.Responses = append(resp.Responses, eresp)
		default:
			return nil, errs.ErrInvalidf("unknown flag type: %s", f.Type)
		}
	}

	return resp, nil
}

// matchConstraints is a utility function that will return if all or any constraints have matched for a segment depending
// on the match type.
func (s *Server) matchConstraints(evalCtx map[string]string, constraints []storage.EvaluationConstraint, segmentMatchType core.MatchType, entityId string) (bool, string, error) {
	constraintMatches := 0

	var reason string

	for _, c := range constraints {
		v := evalCtx[c.Property]

		var (
			match bool
			err   error
		)

		switch c.Type {
		case core.ComparisonType_STRING_COMPARISON_TYPE:
			match = matchesString(c, v)
		case core.ComparisonType_NUMBER_COMPARISON_TYPE:
			match, err = matchesNumber(c, v)
		case core.ComparisonType_BOOLEAN_COMPARISON_TYPE:
			match, err = matchesBool(c, v)
		case core.ComparisonType_DATETIME_COMPARISON_TYPE:
			match, err = matchesDateTime(c, v)
		case core.ComparisonType_ENTITY_ID_COMPARISON_TYPE:
			match = matchesString(c, entityId)
		default:
			return false, reason, errs.ErrInvalid("unknown constraint type")
		}

		if err != nil {
			s.logger.Debug("error matching constraint", zap.String("property", c.Property), zap.Error(err))
			// don't return here because we want to continue to evaluate the other constraints
		}

		if match {
			// increase the matchCount
			constraintMatches++

			switch segmentMatchType {
			case core.MatchType_ANY_MATCH_TYPE:
				// can short circuit here since we had at least one match
				break
			default:
				// keep looping as we need to match all constraints
				continue
			}
		} else {
			// no match
			switch segmentMatchType {
			case core.MatchType_ALL_MATCH_TYPE:
				// we can short circuit because we must match all constraints
				break
			default:
				// keep looping to see if we match the next constraint
				continue
			}
		}
	}

	matched := true

	switch segmentMatchType {
	case core.MatchType_ALL_MATCH_TYPE:
		if len(constraints) != constraintMatches {
			reason = "did not match ALL constraints"
			matched = false
		}

	case core.MatchType_ANY_MATCH_TYPE:
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
	case flipt.OpContains:
		return strings.Contains(v, value)
	case flipt.OpNotContains:
		return !strings.Contains(v, value)
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

func getVariant(key string, variants ...*core.Variant) (*core.Variant, bool) {
	for _, v := range variants {
		if v.Key == key {
			return v, true
		}
	}

	return nil, false
}

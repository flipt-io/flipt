package server

import (
	"context"
	"fmt"
	"hash/crc32"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/golang/protobuf/ptypes"
	"github.com/markphelps/flipt/errors"
	flipt "github.com/markphelps/flipt/rpc"
	"github.com/markphelps/flipt/storage"
)

// Evaluate evaluates a request for a given flag and entity
func (s *Server) Evaluate(ctx context.Context, req *flipt.EvaluationRequest) (*flipt.EvaluationResponse, error) {
	startTime := time.Now()

	// set request ID if not present
	if req.RequestId == "" {
		req.RequestId = uuid.Must(uuid.NewV4()).String()
	}

	resp, err := s.evaluate(ctx, req)
	if resp != nil {
		resp.RequestDurationMillis = float64(time.Since(startTime)) / float64(time.Millisecond)
	}

	if err != nil {
		return resp, err
	}

	return resp, nil
}

func (s *Server) evaluate(ctx context.Context, r *flipt.EvaluationRequest) (*flipt.EvaluationResponse, error) {
	s.logger.Debug("evaluate")

	var (
		ts, _ = ptypes.TimestampProto(time.Now().UTC())
		resp  = &flipt.EvaluationResponse{
			RequestId:      r.RequestId,
			EntityId:       r.EntityId,
			RequestContext: r.Context,
			Timestamp:      ts,
			FlagKey:        r.FlagKey,
		}
	)

	flag, err := s.FlagStore.GetFlag(ctx, &flipt.GetFlagRequest{Key: r.FlagKey})
	if err != nil {
		return resp, err
	}

	if !flag.Enabled {
		return resp, errors.ErrInvalidf("flag %q is disabled", r.FlagKey)
	}

	rules, err := s.EvaluationStore.GetEvaluationRules(ctx, r.FlagKey)
	if err != nil {
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
			return resp, errors.ErrInvalidf("rule rank: %d detected out of order", rule.Rank)
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
				return resp, errors.ErrInvalid("unknown constraint type")
			}

			if err != nil {
				return resp, err
			}

			if match {
				s.logger.Debugf("constraint: %+v matches", c)

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
				s.logger.Debugf("constraint: %+v does not match", c)

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
			s.logger.Errorf("unknown match type: %v", rule.SegmentMatchType)
			continue
		}

		// otherwise, this is our matching rule, determine the flag variant to return
		// based on the distributions
		resp.SegmentKey = rule.SegmentKey

		distributions, err := s.EvaluationStore.GetEvaluationDistributions(ctx, rule.ID)
		if err != nil {
			return resp, err
		}

		var (
			validDistributions []*storage.EvaluationDistribution
			buckets            []int
		)

		for i, d := range distributions {
			// don't include 0% rollouts
			if d.Rollout > 0 {
				validDistributions = append(validDistributions, d)

				if i == 0 {
					bucket := int(d.Rollout * percentMultiplier)
					buckets = append(buckets, bucket)
				} else {
					bucket := buckets[i-1] + int(d.Rollout*percentMultiplier)
					buckets = append(buckets, bucket)
				}
			}
		}

		// no distributions for rule
		if len(validDistributions) == 0 {
			s.logger.Info("no distributions for rule")
			resp.Match = true
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
		s.logger.Debugf("matched distribution: %+v", d)

		resp.Match = true
		resp.Value = d.VariantKey
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
		return false, fmt.Errorf("parsing number from %q", v)
	}

	value, err := strconv.ParseFloat(c.Value, 64)
	if err != nil {
		return false, fmt.Errorf("parsing number from %q", c.Value)
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
		return false, fmt.Errorf("parsing boolean from %q", v)
	}

	switch c.Operator {
	case flipt.OpTrue:
		return value, nil
	case flipt.OpFalse:
		return !value, nil
	}

	return false, nil
}

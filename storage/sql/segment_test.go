package sql

import (
	"context"
	"encoding/json"
	"time"

	flipt "go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/storage"
	"go.flipt.io/flipt/storage/sql/common"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *DBTestSuite) TestGetSegment() {
	t := s.T()

	segment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		MatchType:   flipt.MatchType_ALL_MATCH_TYPE,
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	got, err := s.store.GetSegment(context.TODO(), segment.Key)

	require.NoError(t, err)
	assert.NotNil(t, got)

	assert.Equal(t, segment.Key, got.Key)
	assert.Equal(t, segment.Name, got.Name)
	assert.Equal(t, segment.Description, got.Description)
	assert.NotZero(t, segment.CreatedAt)
	assert.NotZero(t, segment.UpdatedAt)
	assert.Equal(t, segment.MatchType, got.MatchType)
}

func (s *DBTestSuite) TestGetSegmentNotFound() {
	t := s.T()

	_, err := s.store.GetSegment(context.TODO(), "foo")
	assert.EqualError(t, err, "segment \"foo\" not found")
}

func (s *DBTestSuite) TestListSegments() {
	t := s.T()

	reqs := []*flipt.CreateSegmentRequest{
		{
			Key:         uuid.Must(uuid.NewV4()).String(),
			Name:        "foo",
			Description: "bar",
		},
		{
			Key:         uuid.Must(uuid.NewV4()).String(),
			Name:        "foo",
			Description: "bar",
		},
	}

	for _, req := range reqs {
		_, err := s.store.CreateSegment(context.TODO(), req)
		require.NoError(t, err)
	}

	res, err := s.store.ListSegments(context.TODO())
	require.NoError(t, err)
	got := res.Results
	assert.NotZero(t, len(got))
}

func (s *DBTestSuite) TestListSegmentsPagination_LimitOffset() {
	t := s.T()

	reqs := []*flipt.CreateSegmentRequest{
		{
			Key:         uuid.Must(uuid.NewV4()).String(),
			Name:        "foo",
			Description: "bar",
		},
		{
			Key:         uuid.Must(uuid.NewV4()).String(),
			Name:        "foo",
			Description: "bar",
		},
	}

	for _, req := range reqs {
		_, err := s.store.CreateSegment(context.TODO(), req)
		require.NoError(t, err)
	}

	res, err := s.store.ListSegments(context.TODO(), storage.WithLimit(1), storage.WithOffset(1))
	require.NoError(t, err)
	got := res.Results
	assert.Len(t, got, 1)
}

func (s *DBTestSuite) TestListSegmentsPagination_LimitWithNextPage() {
	t := s.T()

	reqs := []*flipt.CreateSegmentRequest{
		{
			Key:         uuid.Must(uuid.NewV4()).String(),
			Name:        "foo",
			Description: "bar",
		},
		{
			Key:         uuid.Must(uuid.NewV4()).String(),
			Name:        "foo",
			Description: "bar",
		},
	}

	for _, req := range reqs {
		if s.driver == MySQL {
			// required for MySQL since it only s.stores timestamps to the second and not millisecond granularity
			time.Sleep(time.Second)
		}
		_, err := s.store.CreateSegment(context.TODO(), req)
		require.NoError(t, err)
	}

	// TODO: the ordering (DESC) is required because the default ordering is ASC and we are not clearing the DB between tests
	opts := []storage.QueryOption{storage.WithOrder(storage.OrderDesc), storage.WithLimit(1)}

	res, err := s.store.ListSegments(context.TODO(), opts...)
	require.NoError(t, err)

	got := res.Results
	assert.Len(t, got, 1)
	assert.Equal(t, reqs[1].Key, got[0].Key)
	assert.NotEmpty(t, res.NextPageToken)

	pageToken := &common.PageToken{}
	err = json.Unmarshal([]byte(res.NextPageToken), pageToken)
	require.NoError(t, err)
	assert.Equal(t, reqs[0].Key, pageToken.Key)
	assert.NotZero(t, pageToken.Offset)

	opts = append(opts, storage.WithPageToken(res.NextPageToken))

	res, err = s.store.ListSegments(context.TODO(), opts...)
	require.NoError(t, err)

	got = res.Results
	assert.Len(t, got, 1)
	assert.Equal(t, reqs[0].Key, got[0].Key)
}

func (s *DBTestSuite) TestCreateSegment() {
	t := s.T()

	segment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		MatchType:   flipt.MatchType_ANY_MATCH_TYPE,
	})

	require.NoError(t, err)

	assert.Equal(t, t.Name(), segment.Key)
	assert.Equal(t, "foo", segment.Name)
	assert.Equal(t, "bar", segment.Description)
	assert.Equal(t, flipt.MatchType_ANY_MATCH_TYPE, segment.MatchType)
	assert.NotZero(t, segment.CreatedAt)
	assert.Equal(t, segment.CreatedAt.Seconds, segment.UpdatedAt.Seconds)
}

func (s *DBTestSuite) TestCreateSegment_DuplicateKey() {
	t := s.T()

	_, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)

	_, err = s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	assert.EqualError(t, err, "segment \"TestDBTestSuite/TestCreateSegment_DuplicateKey\" is not unique")
}

func (s *DBTestSuite) TestUpdateSegment() {
	t := s.T()

	segment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		MatchType:   flipt.MatchType_ALL_MATCH_TYPE,
	})

	require.NoError(t, err)

	assert.Equal(t, t.Name(), segment.Key)
	assert.Equal(t, "foo", segment.Name)
	assert.Equal(t, "bar", segment.Description)
	assert.Equal(t, flipt.MatchType_ALL_MATCH_TYPE, segment.MatchType)
	assert.NotZero(t, segment.CreatedAt)
	assert.Equal(t, segment.CreatedAt.Seconds, segment.UpdatedAt.Seconds)

	updated, err := s.store.UpdateSegment(context.TODO(), &flipt.UpdateSegmentRequest{
		Key:         segment.Key,
		Name:        segment.Name,
		Description: "foobar",
		MatchType:   flipt.MatchType_ANY_MATCH_TYPE,
	})

	require.NoError(t, err)

	assert.Equal(t, segment.Key, updated.Key)
	assert.Equal(t, segment.Name, updated.Name)
	assert.Equal(t, "foobar", updated.Description)
	assert.Equal(t, flipt.MatchType_ANY_MATCH_TYPE, updated.MatchType)
	assert.NotZero(t, updated.CreatedAt)
	assert.NotZero(t, updated.UpdatedAt)
}

func (s *DBTestSuite) TestUpdateSegment_NotFound() {
	t := s.T()

	_, err := s.store.UpdateSegment(context.TODO(), &flipt.UpdateSegmentRequest{
		Key:         "foo",
		Name:        "foo",
		Description: "bar",
	})

	assert.EqualError(t, err, "segment \"foo\" not found")
}

func (s *DBTestSuite) TestDeleteSegment() {
	t := s.T()

	segment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	err = s.store.DeleteSegment(context.TODO(), &flipt.DeleteSegmentRequest{Key: segment.Key})
	require.NoError(t, err)
}

func (s *DBTestSuite) TestDeleteSegment_ExistingRule() {
	t := s.T()
	// TODO
	t.SkipNow()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variant, err := s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	segment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	rule, err := s.store.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
		FlagKey:    flag.Key,
		SegmentKey: segment.Key,
		Rank:       1,
	})

	require.NoError(t, err)
	assert.NotNil(t, rule)

	// try to delete segment with attached rule
	err = s.store.DeleteSegment(context.TODO(), &flipt.DeleteSegmentRequest{
		Key: segment.Key,
	})

	assert.EqualError(t, err, "atleast one rule exists that matches this segment")

	// delete the rule, then try to delete the segment again
	err = s.store.DeleteRule(context.TODO(), &flipt.DeleteRuleRequest{
		Id:      rule.Id,
		FlagKey: flag.Key,
	})

	require.NoError(t, err)

	err = s.store.DeleteSegment(context.TODO(), &flipt.DeleteSegmentRequest{
		Key: segment.Key,
	})

	require.NoError(t, err)
}

func (s *DBTestSuite) TestDeleteSegment_NotFound() {
	t := s.T()

	err := s.store.DeleteSegment(context.TODO(), &flipt.DeleteSegmentRequest{Key: "foo"})
	require.NoError(t, err)
}

func (s *DBTestSuite) TestCreateConstraint() {
	t := s.T()

	segment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	constraint, err := s.store.CreateConstraint(context.TODO(), &flipt.CreateConstraintRequest{
		SegmentKey: segment.Key,
		Type:       flipt.ComparisonType_STRING_COMPARISON_TYPE,
		Property:   "foo",
		Operator:   "EQ",
		Value:      "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, constraint)

	assert.NotZero(t, constraint.Id)
	assert.Equal(t, segment.Key, constraint.SegmentKey)
	assert.Equal(t, flipt.ComparisonType_STRING_COMPARISON_TYPE, constraint.Type)
	assert.Equal(t, "foo", constraint.Property)
	assert.Equal(t, flipt.OpEQ, constraint.Operator)
	assert.Equal(t, "bar", constraint.Value)
	assert.NotZero(t, constraint.CreatedAt)
	assert.Equal(t, constraint.CreatedAt.Seconds, constraint.UpdatedAt.Seconds)

	// get the segment again
	segment, err = s.store.GetSegment(context.TODO(), segment.Key)

	require.NoError(t, err)
	assert.NotNil(t, segment)

	assert.Len(t, segment.Constraints, 1)
}

func (s *DBTestSuite) TestCreateConstraint_SegmentNotFound() {
	t := s.T()

	_, err := s.store.CreateConstraint(context.TODO(), &flipt.CreateConstraintRequest{
		SegmentKey: "foo",
		Type:       flipt.ComparisonType_STRING_COMPARISON_TYPE,
		Property:   "foo",
		Operator:   "NEQ",
		Value:      "baz",
	})

	assert.EqualError(t, err, "segment \"foo\" not found")
}

func (s *DBTestSuite) TestUpdateConstraint() {
	t := s.T()

	segment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	constraint, err := s.store.CreateConstraint(context.TODO(), &flipt.CreateConstraintRequest{
		SegmentKey: segment.Key,
		Type:       flipt.ComparisonType_STRING_COMPARISON_TYPE,
		Property:   "foo",
		Operator:   "EQ",
		Value:      "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, constraint)

	assert.NotZero(t, constraint.Id)
	assert.Equal(t, segment.Key, constraint.SegmentKey)
	assert.Equal(t, flipt.ComparisonType_STRING_COMPARISON_TYPE, constraint.Type)
	assert.Equal(t, "foo", constraint.Property)
	assert.Equal(t, flipt.OpEQ, constraint.Operator)
	assert.Equal(t, "bar", constraint.Value)
	assert.NotZero(t, constraint.CreatedAt)
	assert.Equal(t, constraint.CreatedAt.Seconds, constraint.UpdatedAt.Seconds)

	updated, err := s.store.UpdateConstraint(context.TODO(), &flipt.UpdateConstraintRequest{
		Id:         constraint.Id,
		SegmentKey: constraint.SegmentKey,
		Type:       flipt.ComparisonType_STRING_COMPARISON_TYPE,
		Property:   "foo",
		Operator:   "EMPTY",
		Value:      "bar",
	})

	require.NoError(t, err)

	assert.Equal(t, constraint.Id, updated.Id)
	assert.Equal(t, constraint.SegmentKey, updated.SegmentKey)
	assert.Equal(t, constraint.Type, updated.Type)
	assert.Equal(t, constraint.Property, updated.Property)
	assert.Equal(t, flipt.OpEmpty, updated.Operator)
	assert.Empty(t, updated.Value)
	assert.NotZero(t, updated.CreatedAt)
	assert.NotZero(t, updated.UpdatedAt)

	// get the segment again
	segment, err = s.store.GetSegment(context.TODO(), segment.Key)

	require.NoError(t, err)
	assert.NotNil(t, segment)

	assert.Len(t, segment.Constraints, 1)
}

func (s *DBTestSuite) TestUpdateConstraint_NotFound() {
	t := s.T()

	segment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	_, err = s.store.UpdateConstraint(context.TODO(), &flipt.UpdateConstraintRequest{
		Id:         "foo",
		SegmentKey: segment.Key,
		Type:       flipt.ComparisonType_STRING_COMPARISON_TYPE,
		Property:   "foo",
		Operator:   "NEQ",
		Value:      "baz",
	})

	assert.EqualError(t, err, "constraint \"foo\" not found")
}

func (s *DBTestSuite) TestDeleteConstraint() {
	t := s.T()

	segment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	constraint, err := s.store.CreateConstraint(context.TODO(), &flipt.CreateConstraintRequest{
		SegmentKey: segment.Key,
		Type:       flipt.ComparisonType_STRING_COMPARISON_TYPE,
		Property:   "foo",
		Operator:   "EQ",
		Value:      "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, constraint)

	err = s.store.DeleteConstraint(context.TODO(), &flipt.DeleteConstraintRequest{SegmentKey: constraint.SegmentKey, Id: constraint.Id})
	require.NoError(t, err)

	// get the segment again
	segment, err = s.store.GetSegment(context.TODO(), segment.Key)

	require.NoError(t, err)
	assert.NotNil(t, segment)

	assert.Empty(t, segment.Constraints)
}

func (s *DBTestSuite) TestDeleteConstraint_NotFound() {
	t := s.T()

	segment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	err = s.store.DeleteConstraint(context.TODO(), &flipt.DeleteConstraintRequest{
		Id:         "foo",
		SegmentKey: segment.Key,
	})

	require.NoError(t, err)
}

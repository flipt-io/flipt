package sql

import (
	"context"
	"testing"
	"time"

	flipt "github.com/markphelps/flipt/rpc/flipt"
	"github.com/markphelps/flipt/storage"
	"github.com/markphelps/flipt/storage/cache"
	"github.com/markphelps/flipt/storage/cache/memory"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSegment(t *testing.T) {
	segment, err := store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		MatchType:   flipt.MatchType_ALL_MATCH_TYPE,
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	got, err := store.GetSegment(context.TODO(), segment.Key)

	require.NoError(t, err)
	assert.NotNil(t, got)

	assert.Equal(t, segment.Key, got.Key)
	assert.Equal(t, segment.Name, got.Name)
	assert.Equal(t, segment.Description, got.Description)
	assert.NotZero(t, segment.CreatedAt)
	assert.NotZero(t, segment.UpdatedAt)
	assert.Equal(t, segment.MatchType, got.MatchType)
}

func TestGetSegmentNotFound(t *testing.T) {
	_, err := store.GetSegment(context.TODO(), "foo")
	assert.EqualError(t, err, "segment \"foo\" not found")
}

func TestListSegments(t *testing.T) {
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
		_, err := store.CreateSegment(context.TODO(), req)
		require.NoError(t, err)
	}

	got, err := store.ListSegments(context.TODO())
	require.NoError(t, err)
	assert.NotZero(t, len(got))
}

func TestListSegmentsPagination(t *testing.T) {
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
		_, err := store.CreateSegment(context.TODO(), req)
		require.NoError(t, err)
	}

	got, err := store.ListSegments(context.TODO(), storage.WithLimit(1), storage.WithOffset(1))

	require.NoError(t, err)
	assert.Len(t, got, 1)
}

func TestCreateSegment(t *testing.T) {
	segment, err := store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
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

func TestCreateSegment_DuplicateKey(t *testing.T) {
	_, err := store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)

	_, err = store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	assert.EqualError(t, err, "segment \"TestCreateSegment_DuplicateKey\" is not unique")
}

func TestUpdateSegment(t *testing.T) {
	segment, err := store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
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

	updated, err := store.UpdateSegment(context.TODO(), &flipt.UpdateSegmentRequest{
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

func TestUpdateSegment_NotFound(t *testing.T) {
	_, err := store.UpdateSegment(context.TODO(), &flipt.UpdateSegmentRequest{
		Key:         "foo",
		Name:        "foo",
		Description: "bar",
	})

	assert.EqualError(t, err, "segment \"foo\" not found")
}

func TestDeleteSegment(t *testing.T) {
	segment, err := store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	err = store.DeleteSegment(context.TODO(), &flipt.DeleteSegmentRequest{Key: segment.Key})
	require.NoError(t, err)
}

func TestDeleteSegment_ExistingRule(t *testing.T) {
	// TODO
	t.SkipNow()

	flag, err := store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variant, err := store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	segment, err := store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	rule, err := store.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
		FlagKey:    flag.Key,
		SegmentKey: segment.Key,
		Rank:       1,
	})

	require.NoError(t, err)
	assert.NotNil(t, rule)

	// try to delete segment with attached rule
	err = store.DeleteSegment(context.TODO(), &flipt.DeleteSegmentRequest{
		Key: segment.Key,
	})

	assert.EqualError(t, err, "atleast one rule exists that matches this segment")

	// delete the rule, then try to delete the segment again
	err = store.DeleteRule(context.TODO(), &flipt.DeleteRuleRequest{
		Id:      rule.Id,
		FlagKey: flag.Key,
	})

	require.NoError(t, err)

	err = store.DeleteSegment(context.TODO(), &flipt.DeleteSegmentRequest{
		Key: segment.Key,
	})

	require.NoError(t, err)
}

func TestDeleteSegment_NotFound(t *testing.T) {
	err := store.DeleteSegment(context.TODO(), &flipt.DeleteSegmentRequest{Key: "foo"})
	require.NoError(t, err)
}

func TestCreateConstraint(t *testing.T) {
	segment, err := store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	constraint, err := store.CreateConstraint(context.TODO(), &flipt.CreateConstraintRequest{
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
	segment, err = store.GetSegment(context.TODO(), segment.Key)

	require.NoError(t, err)
	assert.NotNil(t, segment)

	assert.Len(t, segment.Constraints, 1)
}

func TestCreateConstraint_SegmentNotFound(t *testing.T) {
	_, err := store.CreateConstraint(context.TODO(), &flipt.CreateConstraintRequest{
		SegmentKey: "foo",
		Type:       flipt.ComparisonType_STRING_COMPARISON_TYPE,
		Property:   "foo",
		Operator:   "NEQ",
		Value:      "baz",
	})

	assert.EqualError(t, err, "segment \"foo\" not found")
}

func TestUpdateConstraint(t *testing.T) {
	segment, err := store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	constraint, err := store.CreateConstraint(context.TODO(), &flipt.CreateConstraintRequest{
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

	updated, err := store.UpdateConstraint(context.TODO(), &flipt.UpdateConstraintRequest{
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
	segment, err = store.GetSegment(context.TODO(), segment.Key)

	require.NoError(t, err)
	assert.NotNil(t, segment)

	assert.Len(t, segment.Constraints, 1)
}

func TestUpdateConstraint_NotFound(t *testing.T) {
	segment, err := store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	_, err = store.UpdateConstraint(context.TODO(), &flipt.UpdateConstraintRequest{
		Id:         "foo",
		SegmentKey: segment.Key,
		Type:       flipt.ComparisonType_STRING_COMPARISON_TYPE,
		Property:   "foo",
		Operator:   "NEQ",
		Value:      "baz",
	})

	assert.EqualError(t, err, "constraint \"foo\" not found")
}

func TestDeleteConstraint(t *testing.T) {
	segment, err := store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	constraint, err := store.CreateConstraint(context.TODO(), &flipt.CreateConstraintRequest{
		SegmentKey: segment.Key,
		Type:       flipt.ComparisonType_STRING_COMPARISON_TYPE,
		Property:   "foo",
		Operator:   "EQ",
		Value:      "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, constraint)

	err = store.DeleteConstraint(context.TODO(), &flipt.DeleteConstraintRequest{SegmentKey: constraint.SegmentKey, Id: constraint.Id})
	require.NoError(t, err)

	// get the segment again
	segment, err = store.GetSegment(context.TODO(), segment.Key)

	require.NoError(t, err)
	assert.NotNil(t, segment)

	assert.Empty(t, segment.Constraints)
}

func TestDeleteConstraint_NotFound(t *testing.T) {
	segment, err := store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	err = store.DeleteConstraint(context.TODO(), &flipt.DeleteConstraintRequest{
		Id:         "foo",
		SegmentKey: segment.Key,
	})

	require.NoError(t, err)
}

var benchSegment *flipt.Segment

func BenchmarkGetSegment(b *testing.B) {
	var (
		ctx          = context.Background()
		segment, err = store.CreateSegment(ctx, &flipt.CreateSegmentRequest{
			Key:         b.Name(),
			Name:        "foo",
			Description: "bar",
		})
	)

	if err != nil {
		b.Fatal(err)
	}

	_, err = store.CreateConstraint(ctx, &flipt.CreateConstraintRequest{
		SegmentKey: segment.Key,
		Type:       flipt.ComparisonType_STRING_COMPARISON_TYPE,
		Property:   "foo",
		Operator:   "EQ",
		Value:      "bar",
	})

	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	b.Run("get-segment", func(b *testing.B) {
		var s *flipt.Segment

		for i := 0; i < b.N; i++ {
			s, _ = store.GetSegment(context.TODO(), segment.Key)
		}

		benchSegment = s
	})
}

func BenchmarkGetSegment_CacheMemory(b *testing.B) {
	var (
		l, _       = test.NewNullLogger()
		logger     = logrus.NewEntry(l)
		cacher     = memory.NewCache(5*time.Minute, 10*time.Minute, logger)
		storeCache = cache.NewStore(logger, cacher, store)

		ctx = context.Background()

		segment, err = storeCache.CreateSegment(ctx, &flipt.CreateSegmentRequest{
			Key:         b.Name(),
			Name:        "foo",
			Description: "bar",
		})
	)

	if err != nil {
		b.Fatal(err)
	}

	_, err = storeCache.CreateConstraint(ctx, &flipt.CreateConstraintRequest{
		SegmentKey: segment.Key,
		Type:       flipt.ComparisonType_STRING_COMPARISON_TYPE,
		Property:   "foo",
		Operator:   "EQ",
		Value:      "bar",
	})

	if err != nil {
		b.Fatal(err)
	}

	var s *flipt.Segment

	// warm the cache
	s, _ = storeCache.GetSegment(context.TODO(), segment.Key)

	b.ResetTimer()

	b.Run("get-segment-cache", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			s, _ = storeCache.GetSegment(context.TODO(), segment.Key)
		}

		benchSegment = s
	})
}

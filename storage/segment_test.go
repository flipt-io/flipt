package storage

import (
	"context"
	"testing"

	flipt "github.com/markphelps/flipt/proto"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSegment(t *testing.T) {
	segment, err := segmentRepo.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	got, err := segmentRepo.Segment(context.TODO(), &flipt.GetSegmentRequest{Key: segment.Key})

	require.NoError(t, err)
	assert.NotNil(t, got)

	assert.Equal(t, segment.Key, got.Key)
	assert.Equal(t, segment.Name, got.Name)
	assert.Equal(t, segment.Description, got.Description)
	assert.NotZero(t, segment.CreatedAt)
	assert.NotZero(t, segment.UpdatedAt)
}

func TestSegmentNotFound(t *testing.T) {
	_, err := segmentRepo.Segment(context.TODO(), &flipt.GetSegmentRequest{Key: "foo"})
	assert.EqualError(t, err, "segment \"foo\" not found")
}

func TestSegments(t *testing.T) {
	var (
		reqs = []*flipt.CreateSegmentRequest{
			{
				Key:         uuid.NewV4().String(),
				Name:        "foo",
				Description: "bar",
			},
			{
				Key:         uuid.NewV4().String(),
				Name:        "foo",
				Description: "bar",
			},
		}
	)

	for _, req := range reqs {
		_, err := segmentRepo.CreateSegment(context.TODO(), req)
		require.NoError(t, err)
	}

	got, err := segmentRepo.Segments(context.TODO(), &flipt.ListSegmentRequest{})
	require.NoError(t, err)
	assert.NotZero(t, len(got))
}

func TestSegmentsPagination(t *testing.T) {
	var (
		reqs = []*flipt.CreateSegmentRequest{
			{
				Key:         uuid.NewV4().String(),
				Name:        "foo",
				Description: "bar",
			},
			{
				Key:         uuid.NewV4().String(),
				Name:        "foo",
				Description: "bar",
			},
		}
	)

	for _, req := range reqs {
		_, err := segmentRepo.CreateSegment(context.TODO(), req)
		require.NoError(t, err)
	}

	got, err := segmentRepo.Segments(context.TODO(), &flipt.ListSegmentRequest{
		Limit:  1,
		Offset: 1,
	})
	require.NoError(t, err)
	assert.Len(t, got, 1)
}

func TestCreateSegment(t *testing.T) {
	segment, err := segmentRepo.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)

	assert.Equal(t, t.Name(), segment.Key)
	assert.Equal(t, "foo", segment.Name)
	assert.Equal(t, "bar", segment.Description)
	assert.NotZero(t, segment.CreatedAt)
	assert.Equal(t, segment.CreatedAt, segment.UpdatedAt)
}

func TestCreateSegment_DuplicateKey(t *testing.T) {
	_, err := segmentRepo.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)

	_, err = segmentRepo.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	assert.EqualError(t, err, "segment \"TestCreateSegment_DuplicateKey\" is not unique")
}

func TestUpdateSegment(t *testing.T) {
	segment, err := segmentRepo.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)

	assert.Equal(t, t.Name(), segment.Key)
	assert.Equal(t, "foo", segment.Name)
	assert.Equal(t, "bar", segment.Description)
	assert.NotZero(t, segment.CreatedAt)
	assert.Equal(t, segment.CreatedAt, segment.UpdatedAt)

	updated, err := segmentRepo.UpdateSegment(context.TODO(), &flipt.UpdateSegmentRequest{
		Key:         segment.Key,
		Name:        segment.Name,
		Description: "foobar",
	})

	require.NoError(t, err)

	assert.Equal(t, segment.Key, updated.Key)
	assert.Equal(t, segment.Name, updated.Name)
	assert.Equal(t, "foobar", updated.Description)
	assert.NotZero(t, updated.CreatedAt)
	assert.NotEqual(t, updated.CreatedAt, updated.UpdatedAt)
}

func TestUpdateSegment_NotFound(t *testing.T) {
	_, err := segmentRepo.UpdateSegment(context.TODO(), &flipt.UpdateSegmentRequest{
		Key:         "foo",
		Name:        "foo",
		Description: "bar",
	})

	assert.EqualError(t, err, "segment \"foo\" not found")
}

func TestDeleteSegment(t *testing.T) {
	segment, err := segmentRepo.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	err = segmentRepo.DeleteSegment(context.TODO(), &flipt.DeleteSegmentRequest{Key: segment.Key})
	require.NoError(t, err)
}

func TestDeleteSegment_ExistingRule(t *testing.T) {
	t.SkipNow()

	flag, err := flagRepo.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variant, err := flagRepo.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	segment, err := segmentRepo.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	rule, err := ruleRepo.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
		FlagKey:    flag.Key,
		SegmentKey: segment.Key,
		Rank:       1,
	})

	require.NoError(t, err)
	assert.NotNil(t, rule)

	// try to delete segment with attached rule
	err = segmentRepo.DeleteSegment(context.TODO(), &flipt.DeleteSegmentRequest{
		Key: segment.Key,
	})

	assert.EqualError(t, err, "atleast one rule exists that matches this segment")

	// delete the rule, then try to delete the segment again
	err = ruleRepo.DeleteRule(context.TODO(), &flipt.DeleteRuleRequest{
		Id:      rule.Id,
		FlagKey: flag.Key,
	})

	require.NoError(t, err)

	err = segmentRepo.DeleteSegment(context.TODO(), &flipt.DeleteSegmentRequest{
		Key: segment.Key,
	})

	require.NoError(t, err)
}

func TestDeleteSegment_NotFound(t *testing.T) {
	err := segmentRepo.DeleteSegment(context.TODO(), &flipt.DeleteSegmentRequest{Key: "foo"})
	require.NoError(t, err)
}

func TestCreateConstraint(t *testing.T) {
	segment, err := segmentRepo.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	constraint, err := segmentRepo.CreateConstraint(context.TODO(), &flipt.CreateConstraintRequest{
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
	assert.Equal(t, opEQ, constraint.Operator)
	assert.Equal(t, "bar", constraint.Value)
	assert.NotZero(t, constraint.CreatedAt)
	assert.Equal(t, constraint.CreatedAt, constraint.UpdatedAt)
}

func TestCreateConstraint_SegmentNotFound(t *testing.T) {
	_, err := segmentRepo.CreateConstraint(context.TODO(), &flipt.CreateConstraintRequest{
		SegmentKey: "foo",
		Type:       flipt.ComparisonType_STRING_COMPARISON_TYPE,
		Property:   "foo",
		Operator:   "NEQ",
		Value:      "baz",
	})

	assert.EqualError(t, err, "segment \"foo\" not found")
}

func TestUpdateConstraint(t *testing.T) {
	segment, err := segmentRepo.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	constraint, err := segmentRepo.CreateConstraint(context.TODO(), &flipt.CreateConstraintRequest{
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
	assert.Equal(t, opEQ, constraint.Operator)
	assert.Equal(t, "bar", constraint.Value)
	assert.NotZero(t, constraint.CreatedAt)
	assert.Equal(t, constraint.CreatedAt, constraint.UpdatedAt)

	updated, err := segmentRepo.UpdateConstraint(context.TODO(), &flipt.UpdateConstraintRequest{
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
	assert.Equal(t, opEmpty, updated.Operator)
	assert.Empty(t, updated.Value)
	assert.NotZero(t, updated.CreatedAt)
	assert.NotEqual(t, updated.CreatedAt, updated.UpdatedAt)
}

func TestUpdateConstraint_NotFound(t *testing.T) {
	segment, err := segmentRepo.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	_, err = segmentRepo.UpdateConstraint(context.TODO(), &flipt.UpdateConstraintRequest{
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
	segment, err := segmentRepo.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	constraint, err := segmentRepo.CreateConstraint(context.TODO(), &flipt.CreateConstraintRequest{
		SegmentKey: segment.Key,
		Type:       flipt.ComparisonType_STRING_COMPARISON_TYPE,
		Property:   "foo",
		Operator:   "EQ",
		Value:      "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, constraint)

	err = segmentRepo.DeleteConstraint(context.TODO(), &flipt.DeleteConstraintRequest{SegmentKey: constraint.SegmentKey, Id: constraint.Id})
	require.NoError(t, err)
}

func TestDeleteConstraint_NotFound(t *testing.T) {
	segment, err := segmentRepo.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	err = segmentRepo.DeleteConstraint(context.TODO(), &flipt.DeleteConstraintRequest{
		Id:         "foo",
		SegmentKey: segment.Key,
	})

	require.NoError(t, err)
}

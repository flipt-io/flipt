package sql_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"go.flipt.io/flipt/internal/storage"
	fliptsql "go.flipt.io/flipt/internal/storage/sql"
	"go.flipt.io/flipt/internal/storage/sql/common"
	flipt "go.flipt.io/flipt/rpc/flipt"

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

	got, err := s.store.GetSegment(context.TODO(), storage.NewResource(storage.DefaultNamespace, segment.Key))

	require.NoError(t, err)
	assert.NotNil(t, got)

	assert.Equal(t, storage.DefaultNamespace, got.NamespaceKey)
	assert.Equal(t, segment.Key, got.Key)
	assert.Equal(t, segment.Name, got.Name)
	assert.Equal(t, segment.Description, got.Description)
	assert.NotZero(t, got.CreatedAt)
	assert.NotZero(t, got.UpdatedAt)
	assert.Equal(t, segment.MatchType, got.MatchType)
}

func (s *DBTestSuite) TestGetSegmentNamespace() {
	t := s.T()

	segment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
		MatchType:    flipt.MatchType_ALL_MATCH_TYPE,
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	got, err := s.store.GetSegment(context.TODO(), storage.NewResource(s.namespace, segment.Key))

	require.NoError(t, err)
	assert.NotNil(t, got)

	assert.Equal(t, s.namespace, got.NamespaceKey)
	assert.Equal(t, segment.Key, got.Key)
	assert.Equal(t, segment.Name, got.Name)
	assert.Equal(t, segment.Description, got.Description)
	assert.NotZero(t, got.CreatedAt)
	assert.NotZero(t, got.UpdatedAt)
	assert.Equal(t, segment.MatchType, got.MatchType)
}

func (s *DBTestSuite) TestGetSegment_NotFound() {
	t := s.T()

	_, err := s.store.GetSegment(context.TODO(), storage.NewResource(storage.DefaultNamespace, "foo"))
	assert.EqualError(t, err, "segment \"default/foo\" not found")
}

func (s *DBTestSuite) TestGetSegmentNamespace_NotFound() {
	t := s.T()

	_, err := s.store.GetSegment(context.TODO(), storage.NewResource(s.namespace, "foo"))
	assert.EqualError(t, err, fmt.Sprintf("segment \"%s/foo\" not found", s.namespace))
}

func (s *DBTestSuite) TestGetSegment_WithConstraint() {
	t := s.T()

	segment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		MatchType:   flipt.MatchType_ALL_MATCH_TYPE,
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	// ensure we support older versions of Flipt where constraints have NULL descriptions.
	_, err = s.db.DB.Exec(fmt.Sprintf(`INSERT INTO constraints (id, segment_key, type, property, operator, value) VALUES ('%s', '%s', 1, 'foo', 'eq', 'bar');`,
		uuid.Must(uuid.NewV4()).String(),
		segment.Key))

	require.NoError(t, err)

	got, err := s.store.GetSegment(context.TODO(), storage.NewResource(storage.DefaultNamespace, segment.Key))

	require.NoError(t, err)
	assert.NotNil(t, got)

	assert.Equal(t, storage.DefaultNamespace, got.NamespaceKey)
	assert.Equal(t, segment.Key, got.Key)
	assert.Equal(t, segment.Name, got.Name)
	assert.Equal(t, segment.Description, got.Description)
	assert.NotZero(t, got.CreatedAt)
	assert.NotZero(t, got.UpdatedAt)
	assert.Equal(t, segment.MatchType, got.MatchType)

	require.Len(t, got.Constraints, 1)
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

	_, err := s.store.ListSegments(context.TODO(), storage.ListWithOptions(storage.NewNamespace(storage.DefaultNamespace), storage.ListWithQueryParamOptions[storage.NamespaceRequest](storage.WithPageToken("Hello World"))))
	require.EqualError(t, err, "pageToken is not valid: \"Hello World\"")

	res, err := s.store.ListSegments(context.TODO(), storage.ListWithOptions(storage.NewNamespace(storage.DefaultNamespace)))
	require.NoError(t, err)
	got := res.Results
	assert.NotEmpty(t, got)

	for _, segment := range got {
		assert.Equal(t, storage.DefaultNamespace, segment.NamespaceKey)
		assert.NotZero(t, segment.CreatedAt)
		assert.NotZero(t, segment.UpdatedAt)
	}
}

func (s *DBTestSuite) TestListSegmentsNamespace() {
	t := s.T()

	reqs := []*flipt.CreateSegmentRequest{
		{
			NamespaceKey: s.namespace,
			Key:          uuid.Must(uuid.NewV4()).String(),
			Name:         "foo",
			Description:  "bar",
		},
		{
			NamespaceKey: s.namespace,
			Key:          uuid.Must(uuid.NewV4()).String(),
			Name:         "foo",
			Description:  "bar",
		},
	}

	for _, req := range reqs {
		_, err := s.store.CreateSegment(context.TODO(), req)
		require.NoError(t, err)
	}

	res, err := s.store.ListSegments(context.TODO(), storage.ListWithOptions(storage.NewNamespace(s.namespace)))
	require.NoError(t, err)
	got := res.Results
	assert.NotEmpty(t, got)

	for _, segment := range got {
		assert.Equal(t, s.namespace, segment.NamespaceKey)
		assert.NotZero(t, segment.CreatedAt)
		assert.NotZero(t, segment.UpdatedAt)
	}
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
		{
			Key:         uuid.Must(uuid.NewV4()).String(),
			Name:        "foo",
			Description: "bar",
		},
	}

	for _, req := range reqs {
		if s.db.Driver == fliptsql.MySQL {
			// required for MySQL since it only s.stores timestamps to the second and not millisecond granularity
			time.Sleep(time.Second)
		}
		_, err := s.store.CreateSegment(context.TODO(), req)
		require.NoError(t, err)
	}

	oldest, middle, newest := reqs[0], reqs[1], reqs[2]

	// TODO: the ordering (DESC) is required because the default ordering is ASC and we are not clearing the DB between tests
	// get middle segment
	res, err := s.store.ListSegments(context.TODO(), storage.ListWithOptions(storage.NewNamespace(storage.DefaultNamespace), storage.ListWithQueryParamOptions[storage.NamespaceRequest](storage.WithOrder(storage.OrderDesc), storage.WithLimit(1), storage.WithOffset(1))))
	require.NoError(t, err)

	got := res.Results
	assert.Len(t, got, 1)

	assert.Equal(t, middle.Key, got[0].Key)

	// get first (newest) segment
	res, err = s.store.ListSegments(context.TODO(), storage.ListWithOptions(storage.NewNamespace(storage.DefaultNamespace), storage.ListWithQueryParamOptions[storage.NamespaceRequest](storage.WithOrder(storage.OrderDesc), storage.WithLimit(1))))
	require.NoError(t, err)

	got = res.Results
	assert.Len(t, got, 1)

	assert.Equal(t, newest.Key, got[0].Key)

	// get last (oldest) segment
	res, err = s.store.ListSegments(context.TODO(), storage.ListWithOptions(storage.NewNamespace(storage.DefaultNamespace), storage.ListWithQueryParamOptions[storage.NamespaceRequest](storage.WithOrder(storage.OrderDesc), storage.WithLimit(1), storage.WithOffset(2))))
	require.NoError(t, err)

	got = res.Results
	assert.Len(t, got, 1)

	assert.Equal(t, oldest.Key, got[0].Key)

	// get all segments
	res, err = s.store.ListSegments(context.TODO(), storage.ListWithOptions(storage.NewNamespace(storage.DefaultNamespace), storage.ListWithQueryParamOptions[storage.NamespaceRequest](storage.WithOrder(storage.OrderDesc))))
	require.NoError(t, err)

	got = res.Results

	assert.Equal(t, newest.Key, got[0].Key)
	assert.Equal(t, middle.Key, got[1].Key)
	assert.Equal(t, oldest.Key, got[2].Key)
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
		{
			Key:         uuid.Must(uuid.NewV4()).String(),
			Name:        "foo",
			Description: "bar",
		},
	}

	oldest, middle, newest := reqs[0], reqs[1], reqs[2]

	for _, req := range reqs {
		if s.db.Driver == fliptsql.MySQL {
			// required for MySQL since it only s.stores timestamps to the second and not millisecond granularity
			time.Sleep(time.Second)
		}
		_, err := s.store.CreateSegment(context.TODO(), req)
		require.NoError(t, err)
	}

	// TODO: the ordering (DESC) is required because the default ordering is ASC and we are not clearing the DB between tests
	// get newest segment
	opts := []storage.QueryOption{storage.WithOrder(storage.OrderDesc), storage.WithLimit(1)}

	res, err := s.store.ListSegments(context.TODO(), storage.ListWithOptions(storage.NewNamespace(storage.DefaultNamespace), storage.ListWithQueryParamOptions[storage.NamespaceRequest](opts...)))
	require.NoError(t, err)

	got := res.Results
	assert.Len(t, got, 1)
	assert.Equal(t, newest.Key, got[0].Key)
	assert.NotEmpty(t, res.NextPageToken)

	pTokenB, err := base64.StdEncoding.DecodeString(res.NextPageToken)
	require.NoError(t, err)

	pageToken := &common.PageToken{}
	err = json.Unmarshal(pTokenB, pageToken)
	require.NoError(t, err)
	// next page should be the middle segment
	assert.Equal(t, middle.Key, pageToken.Key)
	assert.NotZero(t, pageToken.Offset)

	opts = append(opts, storage.WithPageToken(res.NextPageToken))

	// get middle segment
	res, err = s.store.ListSegments(context.TODO(), storage.ListWithOptions(storage.NewNamespace(storage.DefaultNamespace), storage.ListWithQueryParamOptions[storage.NamespaceRequest](opts...)))
	require.NoError(t, err)

	got = res.Results
	assert.Len(t, got, 1)
	assert.Equal(t, middle.Key, got[0].Key)

	pTokenB, err = base64.StdEncoding.DecodeString(res.NextPageToken)
	require.NoError(t, err)

	err = json.Unmarshal(pTokenB, pageToken)
	require.NoError(t, err)
	// next page should be the oldest segment
	assert.Equal(t, oldest.Key, pageToken.Key)
	assert.NotZero(t, pageToken.Offset)

	opts = []storage.QueryOption{storage.WithOrder(storage.OrderDesc), storage.WithLimit(1), storage.WithPageToken(res.NextPageToken)}

	// get oldest segment
	res, err = s.store.ListSegments(context.TODO(), storage.ListWithOptions(storage.NewNamespace(storage.DefaultNamespace), storage.ListWithQueryParamOptions[storage.NamespaceRequest](opts...)))
	require.NoError(t, err)

	got = res.Results
	assert.Len(t, got, 1)
	assert.Equal(t, oldest.Key, got[0].Key)

	opts = []storage.QueryOption{storage.WithOrder(storage.OrderDesc), storage.WithLimit(3)}
	// get all segments
	res, err = s.store.ListSegments(context.TODO(), storage.ListWithOptions(storage.NewNamespace(storage.DefaultNamespace), storage.ListWithQueryParamOptions[storage.NamespaceRequest](opts...)))
	require.NoError(t, err)

	got = res.Results
	assert.Len(t, got, 3)
	assert.Equal(t, newest.Key, got[0].Key)
	assert.Equal(t, middle.Key, got[1].Key)
	assert.Equal(t, oldest.Key, got[2].Key)
}

func (s *DBTestSuite) TestListSegmentsPagination_FullWalk() {
	t := s.T()

	namespace := uuid.Must(uuid.NewV4()).String()

	ctx := context.Background()
	_, err := s.store.CreateNamespace(ctx, &flipt.CreateNamespaceRequest{
		Key: namespace,
	})
	require.NoError(t, err)

	var (
		totalSegments = 9
		pageSize      = uint64(3)
	)

	for i := 0; i < totalSegments; i++ {
		req := flipt.CreateSegmentRequest{
			NamespaceKey: namespace,
			Key:          fmt.Sprintf("segment_%03d", i),
			Name:         "foo",
			Description:  "bar",
		}

		_, err := s.store.CreateSegment(ctx, &req)
		require.NoError(t, err)

		for i := 0; i < 2; i++ {
			if i > 0 && s.db.Driver == fliptsql.MySQL {
				// required for MySQL since it only s.stores timestamps to the second and not millisecond granularity
				time.Sleep(time.Second)
			}

			_, err := s.store.CreateConstraint(ctx, &flipt.CreateConstraintRequest{
				NamespaceKey: namespace,
				SegmentKey:   req.Key,
				Type:         flipt.ComparisonType_STRING_COMPARISON_TYPE,
				Property:     "foo",
				Operator:     flipt.OpEQ,
				Value:        "bar",
			})
			require.NoError(t, err)
		}
	}

	req := storage.ListWithOptions(
		storage.NewNamespace(namespace),
		storage.ListWithQueryParamOptions[storage.NamespaceRequest](storage.WithLimit(pageSize)),
	)
	resp, err := s.store.ListSegments(ctx, req)

	require.NoError(t, err)

	found := resp.Results
	for token := resp.NextPageToken; token != ""; token = resp.NextPageToken {
		req.QueryParams.PageToken = token
		resp, err = s.store.ListSegments(ctx, req)
		require.NoError(t, err)

		found = append(found, resp.Results...)
	}

	require.Len(t, found, totalSegments)

	for i := 0; i < totalSegments; i++ {
		assert.Equal(t, namespace, found[i].NamespaceKey)

		expectedSegment := fmt.Sprintf("segment_%03d", i)
		assert.Equal(t, expectedSegment, found[i].Key)
		assert.Equal(t, "foo", found[i].Name)
		assert.Equal(t, "bar", found[i].Description)

		require.Len(t, found[i].Constraints, 2)
		assert.Equal(t, namespace, found[i].Constraints[0].NamespaceKey)
		assert.Equal(t, expectedSegment, found[i].Constraints[0].SegmentKey)
		assert.Equal(t, flipt.ComparisonType_STRING_COMPARISON_TYPE, found[i].Constraints[0].Type)
		assert.Equal(t, "foo", found[i].Constraints[0].Property)
		assert.Equal(t, flipt.OpEQ, found[i].Constraints[0].Operator)
		assert.Equal(t, "bar", found[i].Constraints[0].Value)

		assert.Equal(t, namespace, found[i].Constraints[1].NamespaceKey)
		assert.Equal(t, expectedSegment, found[i].Constraints[1].SegmentKey)
		assert.Equal(t, flipt.ComparisonType_STRING_COMPARISON_TYPE, found[i].Constraints[1].Type)
		assert.Equal(t, "foo", found[i].Constraints[1].Property)
		assert.Equal(t, flipt.OpEQ, found[i].Constraints[1].Operator)
		assert.Equal(t, "bar", found[i].Constraints[1].Value)
	}
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

func (s *DBTestSuite) TestCreateSegmentNamespace() {
	t := s.T()

	segment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
		MatchType:    flipt.MatchType_ANY_MATCH_TYPE,
	})

	require.NoError(t, err)

	assert.Equal(t, s.namespace, segment.NamespaceKey)
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

	assert.EqualError(t, err, "segment \"default/TestDBTestSuite/TestCreateSegment_DuplicateKey\" is not unique")
}

func (s *DBTestSuite) TestCreateSegmentNamespace_DuplicateKey() {
	t := s.T()

	_, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
	})

	require.NoError(t, err)

	_, err = s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
	})

	assert.EqualError(t, err, fmt.Sprintf("segment \"%s/%s\" is not unique", s.namespace, t.Name()))
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

	assert.Equal(t, storage.DefaultNamespace, segment.NamespaceKey)
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

	assert.Equal(t, storage.DefaultNamespace, updated.NamespaceKey)
	assert.Equal(t, segment.Key, updated.Key)
	assert.Equal(t, segment.Name, updated.Name)
	assert.Equal(t, "foobar", updated.Description)
	assert.Equal(t, flipt.MatchType_ANY_MATCH_TYPE, updated.MatchType)
	assert.NotZero(t, updated.CreatedAt)
	assert.NotZero(t, updated.UpdatedAt)
}
func (s *DBTestSuite) TestUpdateSegmentNamespace() {
	t := s.T()

	segment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
		MatchType:    flipt.MatchType_ALL_MATCH_TYPE,
	})

	require.NoError(t, err)

	assert.Equal(t, s.namespace, segment.NamespaceKey)
	assert.Equal(t, t.Name(), segment.Key)
	assert.Equal(t, "foo", segment.Name)
	assert.Equal(t, "bar", segment.Description)
	assert.Equal(t, flipt.MatchType_ALL_MATCH_TYPE, segment.MatchType)
	assert.NotZero(t, segment.CreatedAt)
	assert.Equal(t, segment.CreatedAt.Seconds, segment.UpdatedAt.Seconds)

	updated, err := s.store.UpdateSegment(context.TODO(), &flipt.UpdateSegmentRequest{
		NamespaceKey: s.namespace,
		Key:          segment.Key,
		Name:         segment.Name,
		Description:  "foobar",
		MatchType:    flipt.MatchType_ANY_MATCH_TYPE,
	})

	require.NoError(t, err)

	assert.Equal(t, s.namespace, updated.NamespaceKey)
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

	assert.EqualError(t, err, "segment \"default/foo\" not found")
}

func (s *DBTestSuite) TestUpdateSegmentNamespace_NotFound() {
	t := s.T()

	_, err := s.store.UpdateSegment(context.TODO(), &flipt.UpdateSegmentRequest{
		NamespaceKey: s.namespace,
		Key:          "foo",
		Name:         "foo",
		Description:  "bar",
	})

	assert.EqualError(t, err, fmt.Sprintf("segment \"%s/foo\" not found", s.namespace))
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

func (s *DBTestSuite) TestDeleteSegmentNamespace() {
	t := s.T()

	segment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	err = s.store.DeleteSegment(context.TODO(), &flipt.DeleteSegmentRequest{
		NamespaceKey: s.namespace,
		Key:          segment.Key,
	})
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

	require.EqualError(t, err, "atleast one rule exists that matches this segment")

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

func (s *DBTestSuite) TestDeleteSegmentNamespace_NotFound() {
	t := s.T()

	err := s.store.DeleteSegment(context.TODO(), &flipt.DeleteSegmentRequest{
		NamespaceKey: s.namespace,
		Key:          "foo",
	})
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
		SegmentKey:  segment.Key,
		Type:        flipt.ComparisonType_STRING_COMPARISON_TYPE,
		Property:    "foo",
		Operator:    "EQ",
		Value:       "bar",
		Description: "desc",
	})

	require.NoError(t, err)
	assert.NotNil(t, constraint)

	assert.NotZero(t, constraint.Id)
	assert.Equal(t, storage.DefaultNamespace, constraint.NamespaceKey)
	assert.Equal(t, segment.Key, constraint.SegmentKey)
	assert.Equal(t, flipt.ComparisonType_STRING_COMPARISON_TYPE, constraint.Type)
	assert.Equal(t, "foo", constraint.Property)
	assert.Equal(t, flipt.OpEQ, constraint.Operator)
	assert.Equal(t, "bar", constraint.Value)
	assert.NotZero(t, constraint.CreatedAt)
	assert.Equal(t, constraint.CreatedAt.Seconds, constraint.UpdatedAt.Seconds)
	assert.Equal(t, "desc", constraint.Description)

	// get the segment again
	segment, err = s.store.GetSegment(context.TODO(), storage.NewResource(storage.DefaultNamespace, segment.Key))

	require.NoError(t, err)
	assert.NotNil(t, segment)

	assert.Len(t, segment.Constraints, 1)
}

func (s *DBTestSuite) TestCreateConstraintNamespace() {
	t := s.T()

	segment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	constraint, err := s.store.CreateConstraint(context.TODO(), &flipt.CreateConstraintRequest{
		NamespaceKey: s.namespace,
		SegmentKey:   segment.Key,
		Type:         flipt.ComparisonType_STRING_COMPARISON_TYPE,
		Property:     "foo",
		Operator:     "EQ",
		Value:        "bar",
		Description:  "desc",
	})

	require.NoError(t, err)
	assert.NotNil(t, constraint)

	assert.NotZero(t, constraint.Id)
	assert.Equal(t, s.namespace, constraint.NamespaceKey)
	assert.Equal(t, segment.Key, constraint.SegmentKey)
	assert.Equal(t, flipt.ComparisonType_STRING_COMPARISON_TYPE, constraint.Type)
	assert.Equal(t, "foo", constraint.Property)
	assert.Equal(t, flipt.OpEQ, constraint.Operator)
	assert.Equal(t, "bar", constraint.Value)
	assert.NotZero(t, constraint.CreatedAt)
	assert.Equal(t, constraint.CreatedAt.Seconds, constraint.UpdatedAt.Seconds)
	assert.Equal(t, "desc", constraint.Description)

	// get the segment again
	segment, err = s.store.GetSegment(context.TODO(), storage.NewResource(s.namespace, segment.Key))

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

	assert.EqualError(t, err, "segment \"default/foo\" not found")
}

func (s *DBTestSuite) TestCreateConstraintNamespace_SegmentNotFound() {
	t := s.T()

	_, err := s.store.CreateConstraint(context.TODO(), &flipt.CreateConstraintRequest{
		NamespaceKey: s.namespace,
		SegmentKey:   "foo",
		Type:         flipt.ComparisonType_STRING_COMPARISON_TYPE,
		Property:     "foo",
		Operator:     "NEQ",
		Value:        "baz",
	})

	assert.EqualError(t, err, fmt.Sprintf("segment \"%s/foo\" not found", s.namespace))
}

// see: https://github.com/flipt-io/flipt/pull/1721/
func (s *DBTestSuite) TestGetSegmentWithConstraintMultiNamespace() {
	t := s.T()

	for _, namespace := range []string{storage.DefaultNamespace, s.namespace} {
		segment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
			NamespaceKey: namespace,
			Key:          t.Name(),
			Name:         "foo",
			Description:  "bar",
		})

		require.NoError(t, err)
		assert.NotNil(t, segment)

		constraint, err := s.store.CreateConstraint(context.TODO(), &flipt.CreateConstraintRequest{
			NamespaceKey: namespace,
			SegmentKey:   segment.Key,
			Type:         flipt.ComparisonType_STRING_COMPARISON_TYPE,
			Property:     "foo",
			Operator:     "EQ",
			Value:        "bar",
			Description:  "desc",
		})

		require.NoError(t, err)
		assert.NotNil(t, constraint)

		assert.NotZero(t, constraint.Id)
		assert.Equal(t, namespace, constraint.NamespaceKey)
		assert.Equal(t, segment.Key, constraint.SegmentKey)
		assert.Equal(t, flipt.ComparisonType_STRING_COMPARISON_TYPE, constraint.Type)
		assert.Equal(t, "foo", constraint.Property)
		assert.Equal(t, flipt.OpEQ, constraint.Operator)
		assert.Equal(t, "bar", constraint.Value)
		assert.NotZero(t, constraint.CreatedAt)
		assert.Equal(t, constraint.CreatedAt.Seconds, constraint.UpdatedAt.Seconds)
		assert.Equal(t, "desc", constraint.Description)
	}

	// get the default namespaced segment
	segment, err := s.store.GetSegment(context.TODO(), storage.NewResource(storage.DefaultNamespace, t.Name()))

	require.NoError(t, err)
	assert.NotNil(t, segment)

	// ensure we aren't crossing namespaces
	assert.Len(t, segment.Constraints, 1)

	constraint := segment.Constraints[0]
	assert.NotZero(t, constraint.Id)
	assert.Equal(t, storage.DefaultNamespace, constraint.NamespaceKey)
	assert.Equal(t, segment.Key, constraint.SegmentKey)
	assert.Equal(t, flipt.ComparisonType_STRING_COMPARISON_TYPE, constraint.Type)
	assert.Equal(t, "foo", constraint.Property)
	assert.Equal(t, flipt.OpEQ, constraint.Operator)
	assert.Equal(t, "bar", constraint.Value)
	assert.NotZero(t, constraint.CreatedAt)
	assert.Equal(t, constraint.CreatedAt.Seconds, constraint.UpdatedAt.Seconds)
	assert.Equal(t, "desc", constraint.Description)
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
	assert.Equal(t, storage.DefaultNamespace, constraint.NamespaceKey)
	assert.Equal(t, segment.Key, constraint.SegmentKey)
	assert.Equal(t, flipt.ComparisonType_STRING_COMPARISON_TYPE, constraint.Type)
	assert.Equal(t, "foo", constraint.Property)
	assert.Equal(t, flipt.OpEQ, constraint.Operator)
	assert.Equal(t, "bar", constraint.Value)
	assert.NotZero(t, constraint.CreatedAt)
	assert.Equal(t, constraint.CreatedAt.Seconds, constraint.UpdatedAt.Seconds)

	updated, err := s.store.UpdateConstraint(context.TODO(), &flipt.UpdateConstraintRequest{
		Id:          constraint.Id,
		SegmentKey:  constraint.SegmentKey,
		Type:        flipt.ComparisonType_STRING_COMPARISON_TYPE,
		Property:    "foo",
		Operator:    "EMPTY",
		Value:       "bar",
		Description: "desc",
	})

	require.NoError(t, err)

	assert.Equal(t, constraint.Id, updated.Id)
	assert.Equal(t, storage.DefaultNamespace, updated.NamespaceKey)
	assert.Equal(t, constraint.SegmentKey, updated.SegmentKey)
	assert.Equal(t, constraint.Type, updated.Type)
	assert.Equal(t, constraint.Property, updated.Property)
	assert.Equal(t, flipt.OpEmpty, updated.Operator)
	assert.Empty(t, updated.Value)
	assert.NotZero(t, updated.CreatedAt)
	assert.NotZero(t, updated.UpdatedAt)
	assert.Equal(t, "desc", updated.Description)

	// get the segment again
	segment, err = s.store.GetSegment(context.TODO(), storage.NewResource(storage.DefaultNamespace, segment.Key))

	require.NoError(t, err)
	assert.NotNil(t, segment)

	assert.Len(t, segment.Constraints, 1)
}

func (s *DBTestSuite) TestUpdateConstraintNamespace() {
	t := s.T()

	segment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	constraint, err := s.store.CreateConstraint(context.TODO(), &flipt.CreateConstraintRequest{
		NamespaceKey: s.namespace,
		SegmentKey:   segment.Key,
		Type:         flipt.ComparisonType_STRING_COMPARISON_TYPE,
		Property:     "foo",
		Operator:     "EQ",
		Value:        "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, constraint)

	assert.NotZero(t, constraint.Id)
	assert.Equal(t, s.namespace, constraint.NamespaceKey)
	assert.Equal(t, segment.Key, constraint.SegmentKey)
	assert.Equal(t, flipt.ComparisonType_STRING_COMPARISON_TYPE, constraint.Type)
	assert.Equal(t, "foo", constraint.Property)
	assert.Equal(t, flipt.OpEQ, constraint.Operator)
	assert.Equal(t, "bar", constraint.Value)
	assert.NotZero(t, constraint.CreatedAt)
	assert.Equal(t, constraint.CreatedAt.Seconds, constraint.UpdatedAt.Seconds)

	updated, err := s.store.UpdateConstraint(context.TODO(), &flipt.UpdateConstraintRequest{
		Id:           constraint.Id,
		NamespaceKey: s.namespace,
		SegmentKey:   constraint.SegmentKey,
		Type:         flipt.ComparisonType_STRING_COMPARISON_TYPE,
		Property:     "foo",
		Operator:     "EMPTY",
		Value:        "bar",
		Description:  "desc",
	})

	require.NoError(t, err)

	assert.Equal(t, constraint.Id, updated.Id)
	assert.Equal(t, s.namespace, updated.NamespaceKey)
	assert.Equal(t, constraint.SegmentKey, updated.SegmentKey)
	assert.Equal(t, constraint.Type, updated.Type)
	assert.Equal(t, constraint.Property, updated.Property)
	assert.Equal(t, flipt.OpEmpty, updated.Operator)
	assert.Empty(t, updated.Value)
	assert.NotZero(t, updated.CreatedAt)
	assert.NotZero(t, updated.UpdatedAt)
	assert.Equal(t, "desc", updated.Description)

	// get the segment again
	segment, err = s.store.GetSegment(context.TODO(), storage.NewResource(s.namespace, segment.Key))

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

func (s *DBTestSuite) TestUpdateConstraintNamespace_NotFound() {
	t := s.T()

	segment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	_, err = s.store.UpdateConstraint(context.TODO(), &flipt.UpdateConstraintRequest{
		Id:           "foo",
		NamespaceKey: s.namespace,
		SegmentKey:   segment.Key,
		Type:         flipt.ComparisonType_STRING_COMPARISON_TYPE,
		Property:     "foo",
		Operator:     "NEQ",
		Value:        "baz",
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
	segment, err = s.store.GetSegment(context.TODO(), storage.NewResource(storage.DefaultNamespace, segment.Key))

	require.NoError(t, err)
	assert.NotNil(t, segment)

	assert.Empty(t, segment.Constraints)
}

func (s *DBTestSuite) TestDeleteConstraintNamespace() {
	t := s.T()

	segment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	constraint, err := s.store.CreateConstraint(context.TODO(), &flipt.CreateConstraintRequest{
		NamespaceKey: s.namespace,
		SegmentKey:   segment.Key,
		Type:         flipt.ComparisonType_STRING_COMPARISON_TYPE,
		Property:     "foo",
		Operator:     "EQ",
		Value:        "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, constraint)

	err = s.store.DeleteConstraint(context.TODO(), &flipt.DeleteConstraintRequest{
		NamespaceKey: s.namespace,
		SegmentKey:   constraint.SegmentKey,
		Id:           constraint.Id,
	})
	require.NoError(t, err)

	// get the segment again
	segment, err = s.store.GetSegment(context.TODO(), storage.NewResource(s.namespace, segment.Key))

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

func (s *DBTestSuite) TestDeleteConstraintNamespace_NotFound() {
	t := s.T()

	segment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	err = s.store.DeleteConstraint(context.TODO(), &flipt.DeleteConstraintRequest{
		Id:           "foo",
		NamespaceKey: s.namespace,
		SegmentKey:   segment.Key,
	})

	require.NoError(t, err)
}

func BenchmarkListSegments(b *testing.B) {
	s := new(DBTestSuite)
	t := &testing.T{}
	s.SetT(t)
	s.SetupSuite()

	for i := 0; i < 1000; i++ {
		reqs := []*flipt.CreateSegmentRequest{
			{
				Key:  uuid.Must(uuid.NewV4()).String(),
				Name: fmt.Sprintf("foo_%d", i),
			},
		}

		for _, req := range reqs {
			ss, err := s.store.CreateSegment(context.TODO(), req)
			require.NoError(t, err)
			assert.NotNil(t, ss)

			for j := 0; j < 10; j++ {
				v, err := s.store.CreateConstraint(context.TODO(), &flipt.CreateConstraintRequest{
					SegmentKey: ss.Key,
					Type:       flipt.ComparisonType_STRING_COMPARISON_TYPE,
					Property:   fmt.Sprintf("foo_%d", j),
					Operator:   "EQ",
					Value:      fmt.Sprintf("bar_%d", j),
				})

				require.NoError(t, err)
				assert.NotNil(t, v)
			}
		}
	}

	b.ResetTimer()

	req := storage.ListWithOptions(storage.NewNamespace(storage.DefaultNamespace))
	b.Run("no-pagination", func(b *testing.B) {
		req := req
		for i := 0; i < b.N; i++ {
			segments, err := s.store.ListSegments(context.TODO(), req)
			require.NoError(t, err)
			assert.NotEmpty(t, segments)
		}
	})

	for _, pageSize := range []uint64{10, 25, 100, 500} {
		req := req
		req.QueryParams.Limit = pageSize
		b.Run(fmt.Sprintf("pagination-limit-%d", pageSize), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				segments, err := s.store.ListSegments(context.TODO(), req)
				require.NoError(t, err)
				assert.NotEmpty(t, segments)
			}
		})
	}

	b.Run("pagination", func(b *testing.B) {
		req := req
		req.QueryParams.Limit = 500
		req.QueryParams.Offset = 50
		req.QueryParams.Order = storage.OrderDesc
		for i := 0; i < b.N; i++ {
			segments, err := s.store.ListSegments(context.TODO(), req)
			require.NoError(t, err)
			assert.NotEmpty(t, segments)
		}
	})

	s.TearDownSuite()
}

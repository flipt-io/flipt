package sql_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	errs "go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/internal/storage/sql/common"
	flipt "go.flipt.io/flipt/rpc/flipt"
)

func (s *DBTestSuite) TestGetRollout() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
		Type:        flipt.FlagType_BOOLEAN_FLAG_TYPE,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	rollout, err := s.store.CreateRollout(context.TODO(), &flipt.CreateRolloutRequest{
		FlagKey: flag.Key,
		Rank:    1,
		Rule: &flipt.CreateRolloutRequest_Threshold{
			Threshold: &flipt.RolloutThreshold{
				Value:      true,
				Percentage: 40,
			},
		},
	})

	require.NoError(t, err)
	assert.NotNil(t, rollout)

	got, err := s.store.GetRollout(context.TODO(), storage.NewNamespace(storage.DefaultNamespace), rollout.Id)

	require.NoError(t, err)
	assert.NotNil(t, got)

	assert.Equal(t, rollout.Id, got.Id)
	assert.Equal(t, storage.DefaultNamespace, got.NamespaceKey)
	assert.Equal(t, rollout.FlagKey, got.FlagKey)
	assert.Equal(t, rollout.Rank, got.Rank)
	assert.Equal(t, rollout.Type, got.Type)
	assert.Equal(t, rollout.Rule, got.Rule)
	assert.NotZero(t, got.CreatedAt)
	assert.NotZero(t, got.UpdatedAt)
}

func (s *DBTestSuite) TestGetRolloutNamespace() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:          t.Name(),
		NamespaceKey: s.namespace,
		Name:         "foo",
		Description:  "bar",
		Enabled:      true,
		Type:         flipt.FlagType_BOOLEAN_FLAG_TYPE,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	rollout, err := s.store.CreateRollout(context.TODO(), &flipt.CreateRolloutRequest{
		FlagKey:      flag.Key,
		NamespaceKey: s.namespace,
		Rank:         1,
		Rule: &flipt.CreateRolloutRequest_Threshold{
			Threshold: &flipt.RolloutThreshold{
				Value:      true,
				Percentage: 40,
			},
		},
	})

	require.NoError(t, err)
	assert.NotNil(t, rollout)

	got, err := s.store.GetRollout(context.TODO(), storage.NewNamespace(s.namespace), rollout.Id)

	require.NoError(t, err)
	assert.NotNil(t, got)

	assert.Equal(t, rollout.Id, got.Id)
	assert.Equal(t, s.namespace, got.NamespaceKey)
	assert.Equal(t, rollout.FlagKey, got.FlagKey)
	assert.Equal(t, rollout.Rank, got.Rank)
	assert.Equal(t, rollout.Type, got.Type)
	assert.Equal(t, rollout.Rule, got.Rule)
	assert.NotZero(t, got.CreatedAt)
	assert.NotZero(t, got.UpdatedAt)
}

func (s *DBTestSuite) TestGetRollout_NotFound() {
	t := s.T()

	_, err := s.store.GetRollout(context.TODO(), storage.NewNamespace(storage.DefaultNamespace), "0")
	assert.EqualError(t, err, "rollout \"default/0\" not found")
}

func (s *DBTestSuite) TestGetRolloutNamespace_NotFound() {
	t := s.T()

	_, err := s.store.GetRollout(context.TODO(), storage.NewNamespace(s.namespace), "0")
	assert.EqualError(t, err, fmt.Sprintf("rollout \"%s/0\" not found", s.namespace))
}

func (s *DBTestSuite) TestListRollouts() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
		Type:        flipt.FlagType_BOOLEAN_FLAG_TYPE,
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

	reqs := []*flipt.CreateRolloutRequest{
		{
			FlagKey: flag.Key,
			Rank:    1,
			Rule: &flipt.CreateRolloutRequest_Threshold{
				Threshold: &flipt.RolloutThreshold{
					Value:      true,
					Percentage: 40,
				},
			},
		},
		{
			FlagKey: flag.Key,
			Rank:    2,
			Rule: &flipt.CreateRolloutRequest_Threshold{
				Threshold: &flipt.RolloutThreshold{
					Value:      true,
					Percentage: 80.2,
				},
			},
		},
	}

	for _, req := range reqs {
		_, err := s.store.CreateRollout(context.TODO(), req)
		require.NoError(t, err)
	}

	res, err := s.store.ListRollouts(context.TODO(), storage.ListWithOptions(storage.NewResource(storage.DefaultNamespace, flag.Key)))
	require.NoError(t, err)

	got := res.Results
	assert.NotEmpty(t, got)

	for _, rollout := range got {
		assert.Equal(t, storage.DefaultNamespace, rollout.NamespaceKey)
		assert.NotZero(t, rollout.CreatedAt)
		assert.NotZero(t, rollout.UpdatedAt)
	}
}

func (s *DBTestSuite) TestListRollouts_MultipleSegments() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
		Type:        flipt.FlagType_BOOLEAN_FLAG_TYPE,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	firstSegment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, firstSegment)

	secondSegment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         "another_segment_3",
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, secondSegment)

	reqs := []*flipt.CreateRolloutRequest{
		{
			FlagKey: flag.Key,
			Rank:    1,
			Rule: &flipt.CreateRolloutRequest_Segment{
				Segment: &flipt.RolloutSegment{
					Value:       true,
					SegmentKeys: []string{firstSegment.Key, secondSegment.Key},
				},
			},
		},
		{
			FlagKey: flag.Key,
			Rank:    2,
			Rule: &flipt.CreateRolloutRequest_Segment{
				Segment: &flipt.RolloutSegment{
					Value:       true,
					SegmentKeys: []string{firstSegment.Key, secondSegment.Key},
				},
			},
		},
	}

	for _, req := range reqs {
		_, err := s.store.CreateRollout(context.TODO(), req)
		require.NoError(t, err)
	}

	res, err := s.store.ListRollouts(context.TODO(), storage.ListWithOptions(storage.NewResource(storage.DefaultNamespace, flag.Key)))
	require.NoError(t, err)

	got := res.Results
	assert.NotEmpty(t, got)

	for _, rollout := range got {
		assert.Equal(t, storage.DefaultNamespace, rollout.NamespaceKey)

		rs, ok := rollout.Rule.(*flipt.Rollout_Segment)
		assert.True(t, ok, "rule should successfully assert to a rollout segment")
		assert.Len(t, rs.Segment.SegmentKeys, 2)
		assert.Contains(t, rs.Segment.SegmentKeys, firstSegment.Key)
		assert.Contains(t, rs.Segment.SegmentKeys, secondSegment.Key)
		assert.NotZero(t, rollout.CreatedAt)
		assert.NotZero(t, rollout.UpdatedAt)
	}
}

func (s *DBTestSuite) TestListRolloutsNamespace() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
		Enabled:      true,
		Type:         flipt.FlagType_BOOLEAN_FLAG_TYPE,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	reqs := []*flipt.CreateRolloutRequest{
		{
			NamespaceKey: s.namespace,
			FlagKey:      flag.Key,
			Rank:         1,
			Rule: &flipt.CreateRolloutRequest_Threshold{
				Threshold: &flipt.RolloutThreshold{
					Value:      true,
					Percentage: 40,
				},
			},
		},
		{
			NamespaceKey: s.namespace,
			FlagKey:      flag.Key,
			Rank:         2,
			Rule: &flipt.CreateRolloutRequest_Threshold{
				Threshold: &flipt.RolloutThreshold{
					Value:      true,
					Percentage: 80.2,
				},
			},
		},
	}

	for _, req := range reqs {
		_, err := s.store.CreateRollout(context.TODO(), req)
		require.NoError(t, err)
	}

	res, err := s.store.ListRollouts(context.TODO(), storage.ListWithOptions(storage.NewResource(s.namespace, flag.Key)))
	require.NoError(t, err)

	got := res.Results
	assert.NotEmpty(t, got)

	for _, rollout := range got {
		assert.Equal(t, s.namespace, rollout.NamespaceKey)
		assert.NotZero(t, rollout.CreatedAt)
		assert.NotZero(t, rollout.UpdatedAt)
	}
}

func (s *DBTestSuite) TestListRolloutsPagination_LimitOffset() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	reqs := []*flipt.CreateRolloutRequest{
		{
			FlagKey: flag.Key,
			Rank:    1,
			Rule: &flipt.CreateRolloutRequest_Threshold{
				Threshold: &flipt.RolloutThreshold{
					Value:      true,
					Percentage: 40,
				},
			},
		},
		{
			FlagKey: flag.Key,
			Rank:    2,
			Rule: &flipt.CreateRolloutRequest_Threshold{
				Threshold: &flipt.RolloutThreshold{
					Value:      true,
					Percentage: 80.2,
				},
			},
		},
	}

	for _, req := range reqs {
		_, err := s.store.CreateRollout(context.TODO(), req)
		require.NoError(t, err)
	}

	res, err := s.store.ListRollouts(context.TODO(), storage.ListWithOptions(
		storage.NewResource(storage.DefaultNamespace, flag.Key),
		storage.ListWithQueryParamOptions[storage.ResourceRequest](storage.WithLimit(1), storage.WithOffset(1)),
	))
	require.NoError(t, err)

	got := res.Results
	assert.Len(t, got, 1)
}

func (s *DBTestSuite) TestListRolloutsPagination_LimitWithNextPage() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
		Type:        flipt.FlagType_BOOLEAN_FLAG_TYPE,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	reqs := []*flipt.CreateRolloutRequest{
		{
			FlagKey: flag.Key,
			Rank:    1,
			Rule: &flipt.CreateRolloutRequest_Threshold{
				Threshold: &flipt.RolloutThreshold{
					Value:      true,
					Percentage: 40,
				},
			},
		},
		{
			FlagKey: flag.Key,
			Rank:    2,
			Rule: &flipt.CreateRolloutRequest_Threshold{
				Threshold: &flipt.RolloutThreshold{
					Value:      true,
					Percentage: 80.2,
				},
			},
		},
	}

	for _, req := range reqs {
		_, err := s.store.CreateRollout(context.TODO(), req)
		require.NoError(t, err)
	}

	// TODO: the ordering (DESC) is required because the default ordering is ASC and we are not clearing the DB between tests
	opts := []storage.QueryOption{storage.WithOrder(storage.OrderDesc), storage.WithLimit(1)}

	res, err := s.store.ListRollouts(context.TODO(),
		storage.ListWithOptions(
			storage.NewResource(storage.DefaultNamespace, flag.Key),
			storage.ListWithQueryParamOptions[storage.ResourceRequest](opts...),
		),
	)
	require.NoError(t, err)

	got := res.Results
	assert.Len(t, got, 1)
	assert.Equal(t, reqs[1].Rank, got[0].Rank)
	assert.NotEmpty(t, res.NextPageToken)

	pageToken := &common.PageToken{}
	pTokenB, err := base64.StdEncoding.DecodeString(res.NextPageToken)
	require.NoError(t, err)

	err = json.Unmarshal(pTokenB, pageToken)
	require.NoError(t, err)

	assert.NotEmpty(t, pageToken.Key)
	assert.Equal(t, uint64(1), pageToken.Offset)

	opts = append(opts, storage.WithPageToken(res.NextPageToken))

	res, err = s.store.ListRollouts(context.TODO(),
		storage.ListWithOptions(
			storage.NewResource(storage.DefaultNamespace, flag.Key),
			storage.ListWithQueryParamOptions[storage.ResourceRequest](opts...),
		),
	)
	require.NoError(t, err)

	got = res.Results
	assert.Len(t, got, 1)
	assert.Equal(t, reqs[0].Rank, got[0].Rank)
}

func (s *DBTestSuite) TestCreateRollout_InvalidRolloutType() {
	t := s.T()

	_, err := s.store.CreateRollout(context.TODO(), &flipt.CreateRolloutRequest{
		FlagKey: "foo",
		Rank:    1,
	})

	assert.Equal(t, err, errs.ErrInvalid("rollout rule is missing"))
}

func (s *DBTestSuite) TestCreateRollout_FlagNotFound() {
	t := s.T()

	_, err := s.store.CreateRollout(context.TODO(), &flipt.CreateRolloutRequest{
		FlagKey: "foo",
		Rank:    1,
		Rule: &flipt.CreateRolloutRequest_Threshold{
			Threshold: &flipt.RolloutThreshold{
				Percentage: 50.0,
				Value:      true,
			},
		},
	})

	assert.EqualError(t, err, "flag \"default/foo\" not found")
}

func (s *DBTestSuite) TestCreateRolloutNamespace_InvalidRolloutType() {
	t := s.T()

	_, err := s.store.CreateRollout(context.TODO(), &flipt.CreateRolloutRequest{
		NamespaceKey: s.namespace,
		FlagKey:      "foo",
		Rank:         1,
	})

	assert.Equal(t, err, errs.ErrInvalid("rollout rule is missing"))
}

func (s *DBTestSuite) TestCreateRolloutNamespace_FlagNotFound() {
	t := s.T()

	_, err := s.store.CreateRollout(context.TODO(), &flipt.CreateRolloutRequest{
		NamespaceKey: s.namespace,
		FlagKey:      "foo",
		Rank:         1,
		Rule: &flipt.CreateRolloutRequest_Threshold{
			Threshold: &flipt.RolloutThreshold{
				Percentage: 50.0,
				Value:      true,
			},
		},
	})

	assert.EqualError(t, err, fmt.Sprintf("flag \"%s/foo\" not found", s.namespace))
}

func (s *DBTestSuite) TestUpdateRollout() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	segmentOne, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:       "segment_one",
		Name:      "Segment One",
		MatchType: flipt.MatchType_ANY_MATCH_TYPE,
	})

	require.NoError(t, err)
	assert.NotNil(t, segmentOne)

	segmentTwo, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:       "segment_two",
		Name:      "Segment Two",
		MatchType: flipt.MatchType_ANY_MATCH_TYPE,
	})

	require.NoError(t, err)
	assert.NotNil(t, segmentTwo)

	rollout, err := s.store.CreateRollout(context.TODO(), &flipt.CreateRolloutRequest{
		FlagKey: flag.Key,
		Rank:    1,
		Rule: &flipt.CreateRolloutRequest_Segment{
			Segment: &flipt.RolloutSegment{
				Value:           true,
				SegmentKey:      "segment_one",
				SegmentOperator: flipt.SegmentOperator_AND_SEGMENT_OPERATOR,
			},
		},
	})

	require.NoError(t, err)
	assert.NotNil(t, rollout)

	assert.NotZero(t, rollout.Id)
	assert.Equal(t, storage.DefaultNamespace, rollout.NamespaceKey)
	assert.Equal(t, flag.Key, rollout.FlagKey)
	assert.Equal(t, int32(1), rollout.Rank)
	assert.Equal(t, flipt.RolloutType_SEGMENT_ROLLOUT_TYPE, rollout.Type)
	assert.Equal(t, segmentOne.Key, rollout.GetSegment().SegmentKey)
	assert.True(t, rollout.GetSegment().Value)
	assert.NotZero(t, rollout.CreatedAt)
	assert.Equal(t, rollout.CreatedAt.Seconds, rollout.UpdatedAt.Seconds)
	assert.Equal(t, flipt.SegmentOperator_OR_SEGMENT_OPERATOR, rollout.GetSegment().SegmentOperator)

	updated, err := s.store.UpdateRollout(context.TODO(), &flipt.UpdateRolloutRequest{
		Id:          rollout.Id,
		FlagKey:     rollout.FlagKey,
		Description: "foobar",
		Rule: &flipt.UpdateRolloutRequest_Segment{
			Segment: &flipt.RolloutSegment{
				Value:           false,
				SegmentKeys:     []string{segmentOne.Key, segmentTwo.Key},
				SegmentOperator: flipt.SegmentOperator_AND_SEGMENT_OPERATOR,
			},
		},
	})

	require.NoError(t, err)

	assert.Equal(t, rollout.Id, updated.Id)
	assert.Equal(t, storage.DefaultNamespace, updated.NamespaceKey)
	assert.Equal(t, rollout.FlagKey, updated.FlagKey)
	assert.Equal(t, "foobar", updated.Description)
	assert.Equal(t, int32(1), updated.Rank)
	assert.Equal(t, flipt.RolloutType_SEGMENT_ROLLOUT_TYPE, updated.Type)
	assert.Contains(t, updated.GetSegment().SegmentKeys, segmentOne.Key)
	assert.Contains(t, updated.GetSegment().SegmentKeys, segmentTwo.Key)
	assert.False(t, updated.GetSegment().Value)
	assert.Equal(t, flipt.SegmentOperator_AND_SEGMENT_OPERATOR, updated.GetSegment().SegmentOperator)
	assert.NotZero(t, updated.CreatedAt)
	assert.NotZero(t, updated.UpdatedAt)
}

func (s *DBTestSuite) TestUpdateRollout_OneSegment() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	segmentOne, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:       fmt.Sprintf("one_%s", t.Name()),
		Name:      "Segment One",
		MatchType: flipt.MatchType_ANY_MATCH_TYPE,
	})

	require.NoError(t, err)
	assert.NotNil(t, segmentOne)

	segmentTwo, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:       fmt.Sprintf("two_%s", t.Name()),
		Name:      "Segment Two",
		MatchType: flipt.MatchType_ANY_MATCH_TYPE,
	})

	require.NoError(t, err)
	assert.NotNil(t, segmentTwo)

	rollout, err := s.store.CreateRollout(context.TODO(), &flipt.CreateRolloutRequest{
		FlagKey: flag.Key,
		Rank:    1,
		Rule: &flipt.CreateRolloutRequest_Segment{
			Segment: &flipt.RolloutSegment{
				Value:           true,
				SegmentKeys:     []string{segmentOne.Key, segmentTwo.Key},
				SegmentOperator: flipt.SegmentOperator_AND_SEGMENT_OPERATOR,
			},
		},
	})

	require.NoError(t, err)
	assert.NotNil(t, rollout)

	assert.NotZero(t, rollout.Id)
	assert.Equal(t, storage.DefaultNamespace, rollout.NamespaceKey)
	assert.Equal(t, flag.Key, rollout.FlagKey)
	assert.Equal(t, int32(1), rollout.Rank)
	assert.Equal(t, flipt.RolloutType_SEGMENT_ROLLOUT_TYPE, rollout.Type)
	assert.Contains(t, rollout.GetSegment().SegmentKeys, segmentOne.Key)
	assert.Contains(t, rollout.GetSegment().SegmentKeys, segmentTwo.Key)
	assert.True(t, rollout.GetSegment().Value)
	assert.NotZero(t, rollout.CreatedAt)
	assert.Equal(t, rollout.CreatedAt.Seconds, rollout.UpdatedAt.Seconds)
	assert.Equal(t, flipt.SegmentOperator_AND_SEGMENT_OPERATOR, rollout.GetSegment().SegmentOperator)

	updated, err := s.store.UpdateRollout(context.TODO(), &flipt.UpdateRolloutRequest{
		Id:          rollout.Id,
		FlagKey:     rollout.FlagKey,
		Description: "foobar",
		Rule: &flipt.UpdateRolloutRequest_Segment{
			Segment: &flipt.RolloutSegment{
				Value:           false,
				SegmentKey:      segmentOne.Key,
				SegmentOperator: flipt.SegmentOperator_AND_SEGMENT_OPERATOR,
			},
		},
	})

	require.NoError(t, err)

	assert.Equal(t, rollout.Id, updated.Id)
	assert.Equal(t, storage.DefaultNamespace, updated.NamespaceKey)
	assert.Equal(t, rollout.FlagKey, updated.FlagKey)
	assert.Equal(t, "foobar", updated.Description)
	assert.Equal(t, int32(1), updated.Rank)
	assert.Equal(t, flipt.RolloutType_SEGMENT_ROLLOUT_TYPE, updated.Type)
	assert.Equal(t, segmentOne.Key, updated.GetSegment().SegmentKey)
	assert.False(t, updated.GetSegment().Value)
	assert.Equal(t, flipt.SegmentOperator_OR_SEGMENT_OPERATOR, updated.GetSegment().SegmentOperator)
	assert.NotZero(t, updated.CreatedAt)
	assert.NotZero(t, updated.UpdatedAt)
}

func (s *DBTestSuite) TestUpdateRolloutNamespace() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:          t.Name(),
		NamespaceKey: s.namespace,
		Name:         "foo",
		Description:  "bar",
		Enabled:      true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	rollout, err := s.store.CreateRollout(context.TODO(), &flipt.CreateRolloutRequest{
		FlagKey:      flag.Key,
		NamespaceKey: s.namespace,
		Rank:         1,
		Rule: &flipt.CreateRolloutRequest_Threshold{
			Threshold: &flipt.RolloutThreshold{
				Value:      true,
				Percentage: 40,
			},
		},
	})

	require.NoError(t, err)
	assert.NotNil(t, rollout)

	assert.NotZero(t, rollout.Id)
	assert.Equal(t, s.namespace, rollout.NamespaceKey)
	assert.Equal(t, flag.Key, rollout.FlagKey)
	assert.Equal(t, int32(1), rollout.Rank)
	assert.Equal(t, flipt.RolloutType_THRESHOLD_ROLLOUT_TYPE, rollout.Type)
	assert.InDelta(t, 40, rollout.GetThreshold().Percentage, 0)
	assert.True(t, rollout.GetThreshold().Value)
	assert.NotZero(t, rollout.CreatedAt)
	assert.Equal(t, rollout.CreatedAt.Seconds, rollout.UpdatedAt.Seconds)

	updated, err := s.store.UpdateRollout(context.TODO(), &flipt.UpdateRolloutRequest{
		Id:           rollout.Id,
		FlagKey:      rollout.FlagKey,
		NamespaceKey: s.namespace,
		Description:  "foobar",
		Rule: &flipt.UpdateRolloutRequest_Threshold{
			Threshold: &flipt.RolloutThreshold{
				Value:      false,
				Percentage: 80,
			},
		},
	})

	require.NoError(t, err)

	assert.Equal(t, rollout.Id, updated.Id)
	assert.Equal(t, s.namespace, updated.NamespaceKey)
	assert.Equal(t, rollout.FlagKey, updated.FlagKey)
	assert.Equal(t, "foobar", updated.Description)
	assert.Equal(t, int32(1), updated.Rank)
	assert.Equal(t, flipt.RolloutType_THRESHOLD_ROLLOUT_TYPE, updated.Type)
	assert.InDelta(t, 80, updated.GetThreshold().Percentage, 0)
	assert.False(t, updated.GetThreshold().Value)
	assert.NotZero(t, updated.CreatedAt)
	assert.NotZero(t, updated.UpdatedAt)
}

func (s *DBTestSuite) TestUpdateRollout_InvalidType() {
	t := s.T()

	ctx := context.TODO()
	flag, err := s.store.CreateFlag(ctx, &flipt.CreateFlagRequest{
		Key:          t.Name(),
		NamespaceKey: s.namespace,
		Name:         "foo",
		Description:  "bar",
		Enabled:      true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	_, err = s.store.CreateSegment(ctx, &flipt.CreateSegmentRequest{
		NamespaceKey: s.namespace,
		Key:          "segment_one",
		Name:         "Segment One",
		MatchType:    flipt.MatchType_ANY_MATCH_TYPE,
	})

	require.NoError(t, err)

	rollout, err := s.store.CreateRollout(context.TODO(), &flipt.CreateRolloutRequest{
		FlagKey:      flag.Key,
		NamespaceKey: s.namespace,
		Rank:         1,
		Rule: &flipt.CreateRolloutRequest_Segment{
			Segment: &flipt.RolloutSegment{
				SegmentKeys: []string{"segment_one"},
				Value:       true,
			},
		},
	})

	require.NoError(t, err)
	assert.NotNil(t, rollout)

	assert.NotZero(t, rollout.Id)
	assert.Equal(t, s.namespace, rollout.NamespaceKey)
	assert.Equal(t, flag.Key, rollout.FlagKey)
	assert.Equal(t, int32(1), rollout.Rank)
	assert.Equal(t, flipt.RolloutType_SEGMENT_ROLLOUT_TYPE, rollout.Type)
	assert.Equal(t, "segment_one", rollout.GetSegment().SegmentKey)
	assert.True(t, rollout.GetSegment().Value)
	assert.NotZero(t, rollout.CreatedAt)
	assert.Equal(t, rollout.CreatedAt.Seconds, rollout.UpdatedAt.Seconds)

	_, err = s.store.UpdateRollout(context.TODO(), &flipt.UpdateRolloutRequest{
		Id:           rollout.Id,
		FlagKey:      rollout.FlagKey,
		NamespaceKey: s.namespace,
		Description:  "foobar",
		Rule: &flipt.UpdateRolloutRequest_Threshold{
			Threshold: &flipt.RolloutThreshold{
				Value:      false,
				Percentage: 80,
			},
		},
	})

	require.EqualError(t, err, "cannot change type of rollout: have \"SEGMENT_ROLLOUT_TYPE\" attempted \"THRESHOLD_ROLLOUT_TYPE\"")
}

func (s *DBTestSuite) TestUpdateRollout_NotFound() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
		Type:        flipt.FlagType_BOOLEAN_FLAG_TYPE,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	_, err = s.store.UpdateRollout(context.TODO(), &flipt.UpdateRolloutRequest{
		Id:      "foo",
		FlagKey: flag.Key,
	})

	assert.EqualError(t, err, "rollout \"default/foo\" not found")
}

func (s *DBTestSuite) TestUpdateRolloutNamespace_NotFound() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
		Enabled:      true,
		Type:         flipt.FlagType_BOOLEAN_FLAG_TYPE,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	_, err = s.store.UpdateRollout(context.TODO(), &flipt.UpdateRolloutRequest{
		NamespaceKey: s.namespace,
		Id:           "foo",
		FlagKey:      flag.Key,
	})

	assert.EqualError(t, err, fmt.Sprintf("rollout \"%s/foo\" not found", s.namespace))
}

func (s *DBTestSuite) TestDeleteRollout() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
		Type:        flipt.FlagType_BOOLEAN_FLAG_TYPE,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	var rollouts []*flipt.Rollout

	// create 3 rollouts
	for i := 0; i < 3; i++ {
		rollout, err := s.store.CreateRollout(context.TODO(), &flipt.CreateRolloutRequest{
			FlagKey: flag.Key,
			Rank:    int32(i + 1),
			Rule: &flipt.CreateRolloutRequest_Threshold{
				Threshold: &flipt.RolloutThreshold{
					Value:      true,
					Percentage: 40 + float32(i),
				},
			},
		})

		require.NoError(t, err)
		assert.NotNil(t, rollout)
		rollouts = append(rollouts, rollout)
	}

	// delete second rollout
	err = s.store.DeleteRollout(context.TODO(), &flipt.DeleteRolloutRequest{
		FlagKey: flag.Key,
		Id:      rollouts[1].Id,
	})

	require.NoError(t, err)

	res, err := s.store.ListRollouts(context.TODO(), storage.ListWithOptions(storage.NewResource(storage.DefaultNamespace, flag.Key)))
	// ensure rollouts are in correct order
	require.NoError(t, err)

	got := res.Results
	assert.NotNil(t, got)
	assert.Len(t, got, 2)
	assert.Equal(t, rollouts[0].Id, got[0].Id)
	assert.Equal(t, int32(1), got[0].Rank)
	assert.Equal(t, storage.DefaultNamespace, got[0].NamespaceKey)
	assert.Equal(t, rollouts[2].Id, got[1].Id)
	assert.Equal(t, int32(2), got[1].Rank)
	assert.Equal(t, storage.DefaultNamespace, got[1].NamespaceKey)
}

func (s *DBTestSuite) TestDeleteRolloutNamespace() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
		Enabled:      true,
		Type:         flipt.FlagType_BOOLEAN_FLAG_TYPE,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	var rollouts []*flipt.Rollout

	// create 3 rollouts
	for i := 0; i < 3; i++ {
		rollout, err := s.store.CreateRollout(context.TODO(), &flipt.CreateRolloutRequest{
			NamespaceKey: s.namespace,
			FlagKey:      flag.Key,
			Rank:         int32(i + 1),
			Rule: &flipt.CreateRolloutRequest_Threshold{
				Threshold: &flipt.RolloutThreshold{
					Value:      true,
					Percentage: 40 + float32(i),
				},
			},
		})

		require.NoError(t, err)
		assert.NotNil(t, rollout)
		rollouts = append(rollouts, rollout)
	}

	// delete second rollout
	err = s.store.DeleteRollout(context.TODO(), &flipt.DeleteRolloutRequest{
		NamespaceKey: s.namespace,
		FlagKey:      flag.Key,
		Id:           rollouts[1].Id,
	})

	require.NoError(t, err)

	res, err := s.store.ListRollouts(context.TODO(), storage.ListWithOptions(storage.NewResource(s.namespace, flag.Key)))
	// ensure rollouts are in correct order
	require.NoError(t, err)

	got := res.Results
	assert.NotNil(t, got)
	assert.Len(t, got, 2)
	assert.Equal(t, rollouts[0].Id, got[0].Id)
	assert.Equal(t, int32(1), got[0].Rank)
	assert.Equal(t, s.namespace, got[0].NamespaceKey)
	assert.Equal(t, rollouts[2].Id, got[1].Id)
	assert.Equal(t, int32(2), got[1].Rank)
	assert.Equal(t, s.namespace, got[1].NamespaceKey)
}

func (s *DBTestSuite) TestDeleteRollout_NotFound() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
		Type:        flipt.FlagType_BOOLEAN_FLAG_TYPE,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	err = s.store.DeleteRollout(context.TODO(), &flipt.DeleteRolloutRequest{
		Id:      "foo",
		FlagKey: flag.Key,
	})

	require.NoError(t, err)
}

func (s *DBTestSuite) TestDeleteRolloutNamespace_NotFound() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
		Enabled:      true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	err = s.store.DeleteRollout(context.TODO(), &flipt.DeleteRolloutRequest{
		NamespaceKey: s.namespace,
		Id:           "foo",
		FlagKey:      flag.Key,
	})

	require.NoError(t, err)
}

func (s *DBTestSuite) TestOrderRollouts() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
		Type:        flipt.FlagType_BOOLEAN_FLAG_TYPE,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	var rollouts []*flipt.Rollout

	// create 3 rollouts
	for i := 0; i < 3; i++ {
		rollout, err := s.store.CreateRollout(context.TODO(), &flipt.CreateRolloutRequest{
			FlagKey: flag.Key,
			Rank:    int32(i + 1),
			Rule: &flipt.CreateRolloutRequest_Threshold{
				Threshold: &flipt.RolloutThreshold{
					Value:      true,
					Percentage: 40 + float32(i),
				},
			},
		})

		require.NoError(t, err)
		assert.NotNil(t, rollout)
		rollouts = append(rollouts, rollout)
	}

	// order rollouts in reverse order
	sort.Slice(rollouts, func(i, j int) bool { return rollouts[i].Rank > rollouts[j].Rank })

	var rolloutIds []string
	for _, rollout := range rollouts {
		rolloutIds = append(rolloutIds, rollout.Id)
	}

	// re-order rollouts
	err = s.store.OrderRollouts(context.TODO(), &flipt.OrderRolloutsRequest{
		FlagKey:    flag.Key,
		RolloutIds: rolloutIds,
	})

	require.NoError(t, err)

	res, err := s.store.ListRollouts(context.TODO(), storage.ListWithOptions(storage.NewResource(storage.DefaultNamespace, flag.Key)))

	// ensure rollouts are in correct order
	require.NoError(t, err)
	got := res.Results
	assert.NotNil(t, got)
	assert.Len(t, got, 3)

	assert.Equal(t, rollouts[0].Id, got[0].Id)
	assert.Equal(t, int32(1), got[0].Rank)
	assert.Equal(t, storage.DefaultNamespace, got[0].NamespaceKey)

	assert.Equal(t, rollouts[1].Id, got[1].Id)
	assert.Equal(t, int32(2), got[1].Rank)
	assert.Equal(t, storage.DefaultNamespace, got[1].NamespaceKey)

	assert.Equal(t, rollouts[2].Id, got[2].Id)
	assert.Equal(t, int32(3), got[2].Rank)
	assert.Equal(t, storage.DefaultNamespace, got[2].NamespaceKey)
}

func (s *DBTestSuite) TestOrderRolloutsNamespace() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
		Enabled:      true,
		Type:         flipt.FlagType_BOOLEAN_FLAG_TYPE,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	var rollouts []*flipt.Rollout

	// create 3 rollouts
	for i := 0; i < 3; i++ {
		rollout, err := s.store.CreateRollout(context.TODO(), &flipt.CreateRolloutRequest{
			NamespaceKey: s.namespace,
			FlagKey:      flag.Key,
			Rank:         int32(i + 1),
			Rule: &flipt.CreateRolloutRequest_Threshold{
				Threshold: &flipt.RolloutThreshold{
					Value:      true,
					Percentage: 40 + float32(i),
				},
			},
		})

		require.NoError(t, err)
		assert.NotNil(t, rollout)
		rollouts = append(rollouts, rollout)
	}

	// order rollouts in reverse order
	sort.Slice(rollouts, func(i, j int) bool { return rollouts[i].Rank > rollouts[j].Rank })

	var rolloutIds []string
	for _, rollout := range rollouts {
		rolloutIds = append(rolloutIds, rollout.Id)
	}

	// re-order rollouts
	err = s.store.OrderRollouts(context.TODO(), &flipt.OrderRolloutsRequest{
		NamespaceKey: s.namespace,
		FlagKey:      flag.Key,
		RolloutIds:   rolloutIds,
	})

	require.NoError(t, err)

	res, err := s.store.ListRollouts(context.TODO(), storage.ListWithOptions(storage.NewResource(s.namespace, flag.Key)))

	// ensure rollouts are in correct order
	require.NoError(t, err)
	got := res.Results
	assert.NotNil(t, got)
	assert.Len(t, got, 3)

	assert.Equal(t, rollouts[0].Id, got[0].Id)
	assert.Equal(t, int32(1), got[0].Rank)
	assert.Equal(t, s.namespace, got[0].NamespaceKey)

	assert.Equal(t, rollouts[1].Id, got[1].Id)
	assert.Equal(t, int32(2), got[1].Rank)
	assert.Equal(t, s.namespace, got[1].NamespaceKey)

	assert.Equal(t, rollouts[2].Id, got[2].Id)
	assert.Equal(t, int32(3), got[2].Rank)
	assert.Equal(t, s.namespace, got[2].NamespaceKey)
}

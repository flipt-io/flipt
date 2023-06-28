package sql_test

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		Type:    flipt.RolloutType_PERCENTAGE_ROLLOUT_TYPE,
		Rule: &flipt.CreateRolloutRequest_Percentage{
			Percentage: &flipt.RolloutPercentage{
				Value:      true,
				Percentage: 40,
			},
		},
	})

	require.NoError(t, err)
	assert.NotNil(t, rollout)

	got, err := s.store.GetRollout(context.TODO(), storage.DefaultNamespace, rollout.Id)

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
		Type:         flipt.RolloutType_PERCENTAGE_ROLLOUT_TYPE,
		Rule: &flipt.CreateRolloutRequest_Percentage{
			Percentage: &flipt.RolloutPercentage{
				Value:      true,
				Percentage: 40,
			},
		},
	})

	require.NoError(t, err)
	assert.NotNil(t, rollout)

	got, err := s.store.GetRollout(context.TODO(), s.namespace, rollout.Id)

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

	_, err := s.store.GetRollout(context.TODO(), storage.DefaultNamespace, "0")
	assert.EqualError(t, err, "rollout \"default/0\" not found")
}

func (s *DBTestSuite) TestGetRolloutNamespace_NotFound() {
	t := s.T()

	_, err := s.store.GetRollout(context.TODO(), s.namespace, "0")
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

	segment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	reqs := []*flipt.CreateRolloutRequest{
		{
			FlagKey: flag.Key,
			Rank:    1,
			Type:    flipt.RolloutType_PERCENTAGE_ROLLOUT_TYPE,
			Rule: &flipt.CreateRolloutRequest_Percentage{
				Percentage: &flipt.RolloutPercentage{
					Value:      true,
					Percentage: 40,
				},
			},
		},
		{
			FlagKey: flag.Key,
			Rank:    2,
			Type:    flipt.RolloutType_PERCENTAGE_ROLLOUT_TYPE,
			Rule: &flipt.CreateRolloutRequest_Percentage{
				Percentage: &flipt.RolloutPercentage{
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

	res, err := s.store.ListRollouts(context.TODO(), storage.DefaultNamespace, flag.Key)
	require.NoError(t, err)

	got := res.Results
	assert.NotZero(t, len(got))

	for _, rollout := range got {
		assert.Equal(t, storage.DefaultNamespace, rollout.NamespaceKey)
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
			Type:         flipt.RolloutType_PERCENTAGE_ROLLOUT_TYPE,
			Rule: &flipt.CreateRolloutRequest_Percentage{
				Percentage: &flipt.RolloutPercentage{
					Value:      true,
					Percentage: 40,
				},
			},
		},
		{
			NamespaceKey: s.namespace,
			FlagKey:      flag.Key,
			Rank:         2,
			Type:         flipt.RolloutType_PERCENTAGE_ROLLOUT_TYPE,
			Rule: &flipt.CreateRolloutRequest_Percentage{
				Percentage: &flipt.RolloutPercentage{
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

	res, err := s.store.ListRollouts(context.TODO(), s.namespace, flag.Key)
	require.NoError(t, err)

	got := res.Results
	assert.NotZero(t, len(got))

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
			Type:    flipt.RolloutType_PERCENTAGE_ROLLOUT_TYPE,
			Rule: &flipt.CreateRolloutRequest_Percentage{
				Percentage: &flipt.RolloutPercentage{
					Value:      true,
					Percentage: 40,
				},
			},
		},
		{
			FlagKey: flag.Key,
			Rank:    2,
			Type:    flipt.RolloutType_PERCENTAGE_ROLLOUT_TYPE,
			Rule: &flipt.CreateRolloutRequest_Percentage{
				Percentage: &flipt.RolloutPercentage{
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

	res, err := s.store.ListRollouts(context.TODO(), storage.DefaultNamespace, flag.Key, storage.WithLimit(1), storage.WithOffset(1))
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
			Type:    flipt.RolloutType_PERCENTAGE_ROLLOUT_TYPE,
			Rule: &flipt.CreateRolloutRequest_Percentage{
				Percentage: &flipt.RolloutPercentage{
					Value:      true,
					Percentage: 40,
				},
			},
		},
		{
			FlagKey: flag.Key,
			Rank:    2,
			Type:    flipt.RolloutType_PERCENTAGE_ROLLOUT_TYPE,
			Rule: &flipt.CreateRolloutRequest_Percentage{
				Percentage: &flipt.RolloutPercentage{
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

	res, err := s.store.ListRollouts(context.TODO(), storage.DefaultNamespace, flag.Key, opts...)
	require.NoError(t, err)

	got := res.Results
	assert.Len(t, got, 1)
	assert.Equal(t, reqs[1].Rank, got[0].Rank)
	assert.NotEmpty(t, res.NextPageToken)

	pageToken := &common.PageToken{}
	err = json.Unmarshal([]byte(res.NextPageToken), pageToken)
	require.NoError(t, err)
	assert.NotEmpty(t, pageToken.Key)
	assert.Equal(t, uint64(1), pageToken.Offset)

	opts = append(opts, storage.WithPageToken(res.NextPageToken))

	res, err = s.store.ListRollouts(context.TODO(), storage.DefaultNamespace, flag.Key, opts...)
	require.NoError(t, err)

	got = res.Results
	assert.Len(t, got, 1)
	assert.Equal(t, reqs[0].Rank, got[0].Rank)
}

func (s *DBTestSuite) TestCreateRollout_FlagNotFound() {
	t := s.T()

	_, err := s.store.CreateRollout(context.TODO(), &flipt.CreateRolloutRequest{
		FlagKey: "foo",
		Rank:    1,
	})

	assert.EqualError(t, err, "flag \"default/foo\" not found")
}

func (s *DBTestSuite) TestCreateRolloutNamespace_FlagNotFound() {
	t := s.T()

	_, err := s.store.CreateRollout(context.TODO(), &flipt.CreateRolloutRequest{
		NamespaceKey: s.namespace,
		FlagKey:      "foo",
		Rank:         1,
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

	rollout, err := s.store.CreateRollout(context.TODO(), &flipt.CreateRolloutRequest{
		FlagKey: flag.Key,
		Rank:    1,
		Type:    flipt.RolloutType_PERCENTAGE_ROLLOUT_TYPE,
		Rule: &flipt.CreateRolloutRequest_Percentage{
			Percentage: &flipt.RolloutPercentage{
				Value:      true,
				Percentage: 40,
			},
		},
	})

	require.NoError(t, err)
	assert.NotNil(t, rollout)

	assert.NotZero(t, rollout.Id)
	assert.Equal(t, storage.DefaultNamespace, rollout.NamespaceKey)
	assert.Equal(t, flag.Key, rollout.FlagKey)
	assert.Equal(t, int32(1), rollout.Rank)
	assert.Equal(t, flipt.RolloutType_PERCENTAGE_ROLLOUT_TYPE, rollout.Type)
	assert.Equal(t, float32(40), rollout.GetPercentage().Percentage)
	assert.Equal(t, true, rollout.GetPercentage().Value)
	assert.NotZero(t, rollout.CreatedAt)
	assert.Equal(t, rollout.CreatedAt.Seconds, rollout.UpdatedAt.Seconds)

	updated, err := s.store.UpdateRollout(context.TODO(), &flipt.UpdateRolloutRequest{
		Id:          rollout.Id,
		FlagKey:     rollout.FlagKey,
		Description: "foobar",
		Type:        flipt.RolloutType_PERCENTAGE_ROLLOUT_TYPE,
		Rule: &flipt.UpdateRolloutRequest_Percentage{
			Percentage: &flipt.RolloutPercentage{
				Value:      false,
				Percentage: 80,
			},
		},
	})

	require.NoError(t, err)

	assert.Equal(t, rollout.Id, updated.Id)
	assert.Equal(t, storage.DefaultNamespace, updated.NamespaceKey)
	assert.Equal(t, rollout.FlagKey, updated.FlagKey)
	assert.Equal(t, "foobar", updated.Description)
	assert.Equal(t, int32(1), updated.Rank)
	assert.Equal(t, flipt.RolloutType_PERCENTAGE_ROLLOUT_TYPE, updated.Type)
	assert.Equal(t, float32(80), updated.GetPercentage().Percentage)
	assert.Equal(t, false, updated.GetPercentage().Value)
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
		Type:         flipt.RolloutType_PERCENTAGE_ROLLOUT_TYPE,
		Rule: &flipt.CreateRolloutRequest_Percentage{
			Percentage: &flipt.RolloutPercentage{
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
	assert.Equal(t, flipt.RolloutType_PERCENTAGE_ROLLOUT_TYPE, rollout.Type)
	assert.Equal(t, float32(40), rollout.GetPercentage().Percentage)
	assert.Equal(t, true, rollout.GetPercentage().Value)
	assert.NotZero(t, rollout.CreatedAt)
	assert.Equal(t, rollout.CreatedAt.Seconds, rollout.UpdatedAt.Seconds)

	updated, err := s.store.UpdateRollout(context.TODO(), &flipt.UpdateRolloutRequest{
		Id:           rollout.Id,
		FlagKey:      rollout.FlagKey,
		NamespaceKey: s.namespace,
		Description:  "foobar",
		Type:         flipt.RolloutType_PERCENTAGE_ROLLOUT_TYPE,
		Rule: &flipt.UpdateRolloutRequest_Percentage{
			Percentage: &flipt.RolloutPercentage{
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
	assert.Equal(t, flipt.RolloutType_PERCENTAGE_ROLLOUT_TYPE, updated.Type)
	assert.Equal(t, float32(80), updated.GetPercentage().Percentage)
	assert.Equal(t, false, updated.GetPercentage().Value)
	assert.NotZero(t, updated.CreatedAt)
	assert.NotZero(t, updated.UpdatedAt)
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
			Type:    flipt.RolloutType_PERCENTAGE_ROLLOUT_TYPE,
			Rule: &flipt.CreateRolloutRequest_Percentage{
				Percentage: &flipt.RolloutPercentage{
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

	res, err := s.store.ListRollouts(context.TODO(), storage.DefaultNamespace, flag.Key)
	// ensure rollouts are in correct order
	require.NoError(t, err)

	got := res.Results
	assert.NotNil(t, got)
	assert.Equal(t, 2, len(got))
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
			Type:         flipt.RolloutType_PERCENTAGE_ROLLOUT_TYPE,
			Rule: &flipt.CreateRolloutRequest_Percentage{
				Percentage: &flipt.RolloutPercentage{
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

	res, err := s.store.ListRollouts(context.TODO(), s.namespace, flag.Key)
	// ensure rollouts are in correct order
	require.NoError(t, err)

	got := res.Results
	assert.NotNil(t, got)
	assert.Equal(t, 2, len(got))
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
	t.Skip("TODO: implement")
}

func (s *DBTestSuite) TestOrderRolloutsNamespace() {
	t := s.T()
	t.Skip("TODO: implement")
}

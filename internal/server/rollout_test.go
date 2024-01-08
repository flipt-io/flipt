package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/common"
	"go.flipt.io/flipt/internal/storage"
	flipt "go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap/zaptest"
)

func TestGetRollout(t *testing.T) {
	var (
		store  = &common.StoreMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
		req = &flipt.GetRolloutRequest{Id: "id", FlagKey: "flagKey"}
	)

	store.On("GetRollout", mock.Anything, storage.NewNamespace(""), "id").Return(&flipt.Rollout{
		Id: "1",
	}, nil)

	got, err := s.GetRollout(context.TODO(), req)
	require.NoError(t, err)

	assert.NotNil(t, got)
}

func TestListRollouts_PaginationPageToken(t *testing.T) {
	var (
		store  = &common.StoreMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
	)

	defer store.AssertExpectations(t)

	store.On("ListRollouts", mock.Anything, storage.ListWithOptions(storage.NewResource("", "flagKey"),
		storage.ListWithQueryParamOptions[storage.ResourceRequest](
			storage.WithPageToken("Zm9v"),
		),
	)).Return(
		storage.ResultSet[*flipt.Rollout]{
			Results: []*flipt.Rollout{
				{
					FlagKey: "flagKey",
				},
			},
			NextPageToken: "YmFy",
		}, nil)

	store.On("CountRollouts", mock.Anything, storage.NewResource("", "flagKey")).Return(uint64(1), nil)

	got, err := s.ListRollouts(context.TODO(), &flipt.ListRolloutRequest{FlagKey: "flagKey",
		PageToken: "Zm9v",
	})

	require.NoError(t, err)

	assert.NotEmpty(t, got.Rules)
	assert.Equal(t, "YmFy", got.NextPageToken)
	assert.Equal(t, int32(1), got.TotalCount)
}

func TestCreateRollout(t *testing.T) {
	var (
		store  = &common.StoreMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
	)

	for _, test := range []struct {
		name string
		req  *flipt.CreateRolloutRequest
		resp *flipt.Rollout
	}{
		{
			name: "segment",
			req: &flipt.CreateRolloutRequest{
				FlagKey: "flagKey",
				Rank:    1,
				Rule: &flipt.CreateRolloutRequest_Segment{
					Segment: &flipt.RolloutSegment{
						SegmentKey: "segmentKey",
						Value:      true,
					},
				},
			},
			resp: &flipt.Rollout{
				Id:      "1",
				FlagKey: "flagKey",
				Rank:    1,
				Rule: &flipt.Rollout_Segment{
					Segment: &flipt.RolloutSegment{
						SegmentOperator: flipt.SegmentOperator_OR_SEGMENT_OPERATOR,
						SegmentKey:      "segmentKey",
						Value:           true,
					},
				},
			},
		},
		{
			name: "segments",
			req: &flipt.CreateRolloutRequest{
				FlagKey: "flagKey",
				Rank:    1,
				Rule: &flipt.CreateRolloutRequest_Segment{
					Segment: &flipt.RolloutSegment{
						SegmentOperator: flipt.SegmentOperator_AND_SEGMENT_OPERATOR,
						SegmentKeys:     []string{"segmentKey1", "segmentKey2"},
						Value:           true,
					},
				},
			},
			resp: &flipt.Rollout{
				Id:      "1",
				FlagKey: "flagKey",
				Rank:    1,
				Rule: &flipt.Rollout_Segment{
					Segment: &flipt.RolloutSegment{
						SegmentOperator: flipt.SegmentOperator_AND_SEGMENT_OPERATOR,
						SegmentKeys:     []string{"segmentKey1", "segmentKey2"},
						Value:           true,
					},
				},
			},
		},
		{
			name: "threshold",
			req: &flipt.CreateRolloutRequest{
				FlagKey: "flagKey",
				Rank:    1,
				Rule: &flipt.CreateRolloutRequest_Threshold{
					Threshold: &flipt.RolloutThreshold{
						Percentage: 50.0,
						Value:      true,
					},
				},
			},
			resp: &flipt.Rollout{
				Id:      "1",
				FlagKey: "flagKey",
				Rank:    1,
				Rule: &flipt.Rollout_Threshold{
					Threshold: &flipt.RolloutThreshold{
						Percentage: 50.0,
						Value:      true,
					},
				},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			store.On("CreateRollout", mock.Anything, test.req).Return(test.resp, nil)

			got, err := s.CreateRollout(context.TODO(), test.req)
			require.NoError(t, err)
			assert.Equal(t, test.resp, got)
		})
	}
}

func TestUpdateRollout(t *testing.T) {
	var (
		store  = &common.StoreMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
		req = &flipt.UpdateRolloutRequest{
			Id:      "1",
			FlagKey: "flagKey",
			Rule: &flipt.UpdateRolloutRequest_Segment{
				Segment: &flipt.RolloutSegment{
					SegmentKey: "segmentKey",
					Value:      false,
				},
			},
		}
	)

	store.On("UpdateRollout", mock.Anything, req).Return(&flipt.Rollout{
		Id:      "1",
		FlagKey: req.FlagKey,
		Rule: &flipt.Rollout_Segment{
			Segment: &flipt.RolloutSegment{
				SegmentKey: "segmentKey",
				Value:      false,
			},
		},
	}, nil)

	got, err := s.UpdateRollout(context.TODO(), req)
	require.NoError(t, err)

	assert.NotNil(t, got)
}

func TestDeleteRollout(t *testing.T) {
	var (
		store  = &common.StoreMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
		req = &flipt.DeleteRolloutRequest{
			Id: "id",
		}
	)

	store.On("DeleteRollout", mock.Anything, req).Return(nil)

	got, err := s.DeleteRollout(context.TODO(), req)
	require.NoError(t, err)

	assert.NotNil(t, got)
}

func TestOrderRollouts(t *testing.T) {
	var (
		store  = &common.StoreMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
		req = &flipt.OrderRolloutsRequest{FlagKey: "flagKey", RolloutIds: []string{"1", "2"}}
	)

	store.On("OrderRollouts", mock.Anything, req).Return(nil)

	got, err := s.OrderRollouts(context.TODO(), req)
	require.NoError(t, err)

	assert.NotNil(t, got)
}

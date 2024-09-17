package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	errs "go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/common"
	"go.flipt.io/flipt/internal/storage"
	flipt "go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap/zaptest"
)

var (
	enabledFlag = &flipt.Flag{
		Key:     "foo",
		Enabled: true,
	}
)

func TestBatchEvaluate(t *testing.T) {
	var (
		store  = &common.StoreMock{}
		logger = zaptest.NewLogger(t)
		s      = New(logger, store)
	)

	disabled := &flipt.Flag{
		Key:     "bar",
		Enabled: false,
	}
	store.On("GetFlag", mock.Anything, storage.NewResource("default", "foo")).Return(enabledFlag, nil)
	store.On("GetFlag", mock.Anything, storage.NewResource("default", "bar")).Return(disabled, nil)

	store.On("GetEvaluationRules", mock.Anything, storage.NewResource("default", "foo")).Return([]*storage.EvaluationRule{}, nil)

	resp, err := s.BatchEvaluate(context.TODO(), &flipt.BatchEvaluationRequest{
		RequestId: "12345",
		Requests: []*flipt.EvaluationRequest{
			{
				EntityId: "1",
				FlagKey:  "foo",
				Context: map[string]string{
					"bar": "boz",
				},
			},
			{
				EntityId: "1",
				FlagKey:  "bar",
			},
		},
	})

	require.NoError(t, err)
	assert.NotNil(t, resp.Responses)
	assert.Len(t, resp.Responses, 2)
	assert.False(t, resp.Responses[0].Match)
}

func TestBatchEvaluate_NamespaceMismatch(t *testing.T) {
	var (
		store  = &common.StoreMock{}
		logger = zaptest.NewLogger(t)
		s      = New(logger, store)
	)

	_, err := s.BatchEvaluate(context.TODO(), &flipt.BatchEvaluationRequest{
		NamespaceKey:    "foo",
		RequestId:       "12345",
		ExcludeNotFound: true,
		Requests: []*flipt.EvaluationRequest{
			{
				NamespaceKey: "bar",
				EntityId:     "1",
				FlagKey:      "foo",
				Context: map[string]string{
					"bar": "boz",
				},
			},
		},
	})

	assert.EqualError(t, err, "invalid field namespace_key: must be the same for all requests if specified")
}

func TestBatchEvaluate_FlagNotFoundExcluded(t *testing.T) {
	var (
		store  = &common.StoreMock{}
		logger = zaptest.NewLogger(t)
		s      = New(logger, store)
	)

	disabled := &flipt.Flag{
		Key:     "bar",
		Enabled: false,
	}
	store.On("GetFlag", mock.Anything, storage.NewResource("default", "foo")).Return(enabledFlag, nil)
	store.On("GetFlag", mock.Anything, storage.NewResource("default", "bar")).Return(disabled, nil)
	store.On("GetFlag", mock.Anything, storage.NewResource("default", "NotFoundFlag")).Return(&flipt.Flag{}, errs.ErrNotFoundf("flag %q", "NotFoundFlag"))

	store.On("GetEvaluationRules", mock.Anything, storage.NewResource("default", "foo")).Return([]*storage.EvaluationRule{}, nil)

	resp, err := s.BatchEvaluate(context.TODO(), &flipt.BatchEvaluationRequest{
		RequestId:       "12345",
		ExcludeNotFound: true,
		Requests: []*flipt.EvaluationRequest{
			{
				EntityId: "1",
				FlagKey:  "foo",
				Context: map[string]string{
					"bar": "boz",
				},
			},
			{
				EntityId: "1",
				FlagKey:  "bar",
			},
			{
				EntityId: "1",
				FlagKey:  "NotFoundFlag",
			},
		},
	})

	require.NoError(t, err)
	assert.NotNil(t, resp.Responses)
	assert.Len(t, resp.Responses, 2)
	assert.False(t, resp.Responses[0].Match)
}

func TestBatchEvaluate_FlagNotFound(t *testing.T) {
	var (
		store  = &common.StoreMock{}
		logger = zaptest.NewLogger(t)
		s      = New(logger, store)
	)

	disabled := &flipt.Flag{
		Key:     "bar",
		Enabled: false,
	}
	store.On("GetFlag", mock.Anything, storage.NewResource("default", "foo")).Return(enabledFlag, nil)
	store.On("GetFlag", mock.Anything, storage.NewResource("default", "bar")).Return(disabled, nil)
	store.On("GetFlag", mock.Anything, storage.NewResource("default", "NotFoundFlag")).Return(&flipt.Flag{}, errs.ErrNotFoundf("flag %q", "NotFoundFlag"))

	store.On("GetEvaluationRules", mock.Anything, storage.NewResource("default", "foo")).Return([]*storage.EvaluationRule{}, nil)

	_, err := s.BatchEvaluate(context.TODO(), &flipt.BatchEvaluationRequest{
		RequestId:       "12345",
		ExcludeNotFound: false,
		Requests: []*flipt.EvaluationRequest{
			{
				EntityId: "1",
				FlagKey:  "foo",
				Context: map[string]string{
					"bar": "boz",
				},
			},
			{
				EntityId: "1",
				FlagKey:  "bar",
			},
			{
				EntityId: "1",
				FlagKey:  "NotFoundFlag",
			},
		},
	})

	require.Error(t, err)
	assert.EqualError(t, err, "flag \"NotFoundFlag\" not found")
}

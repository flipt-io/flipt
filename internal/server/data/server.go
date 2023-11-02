package data

import (
	"context"
	"fmt"

	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/rpc/flipt/data"
	"go.uber.org/zap"
)

type Server struct {
	logger *zap.Logger
	store  storage.Store

	data.UnimplementedDataServer
}

func NewServer(logger *zap.Logger, store storage.Store) *Server {
	return &Server{
		logger: logger,
		store:  store,
	}
}

func (s *Server) SnapshotNamespace(ctx context.Context, r *data.SnapshotNamespaceRequest) (*data.SnapshotNamespaceResponse, error) {
	var (
		namespaceKey = r.Key
		resp         = &data.SnapshotNamespaceResponse{
			Rules:    make(map[string]*flipt.RuleList),
			Rollouts: make(map[string]*flipt.RolloutList),
		}
		remaining = true
		nextPage  string
	)

	//  flags/variants in batches
	for remaining {
		res, err := s.store.ListFlags(
			ctx,
			namespaceKey,
			storage.WithPageToken(nextPage),
		)
		if err != nil {
			return nil, fmt.Errorf("getting flags: %w", err)
		}

		flags := res.Results
		nextPage = res.NextPageToken
		remaining = nextPage != ""

		resp.Flags = flags

		for _, f := range flags {
			//  rules for flag
			rules, err := s.store.ListRules(
				ctx,
				namespaceKey,
				f.Key,
			)
			if err != nil {
				return nil, fmt.Errorf("getting rules for flag %q: %w", f.Key, err)
			}

			resp.Rules[f.Key] = &flipt.RuleList{
				Rules:      rules.Results,
				TotalCount: int32(len(rules.Results)),
			}

			rollouts, err := s.store.ListRollouts(ctx, namespaceKey, f.Key)
			if err != nil {
				return nil, fmt.Errorf("getting rollout rules for flag %q: %w", f.Key, err)
			}

			resp.Rollouts[f.Key] = &flipt.RolloutList{
				Rules:      rollouts.Results,
				TotalCount: int32(len(rollouts.Results)),
			}
		}
	}

	remaining = true
	nextPage = ""

	//  segments/constraints in batches
	for remaining {
		res, err := s.store.ListSegments(
			ctx,
			namespaceKey,
			storage.WithPageToken(nextPage),
		)
		if err != nil {
			return nil, fmt.Errorf("getting segments: %w", err)
		}

		segments := res.Results
		nextPage = res.NextPageToken
		remaining = nextPage != ""

		resp.Segments = append(resp.Segments, segments...)
	}

	return resp, nil
}

package server

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	flipt "github.com/markphelps/flipt/rpc"
	"github.com/markphelps/flipt/storage"
)

// GetFlag gets a flag
func (s *Server) GetFlag(ctx context.Context, r *flipt.GetFlagRequest) (*flipt.Flag, error) {
	s.logger.WithField("request", r).Debug("get flag")
	flag, err := s.FlagStore.GetFlag(ctx, r.Key)
	s.logger.WithField("response", flag).Debug("get flag")
	return flag, err
}

// ListFlags lists all flags
func (s *Server) ListFlags(ctx context.Context, r *flipt.ListFlagRequest) (*flipt.FlagList, error) {
	s.logger.WithField("request", r).Debug("list flags")

	flags, err := s.FlagStore.ListFlags(ctx, storage.WithLimit(uint64(r.Limit)), storage.WithOffset(uint64(r.Offset)))
	if err != nil {
		return nil, err
	}

	var resp flipt.FlagList

	for i := range flags {
		resp.Flags = append(resp.Flags, flags[i])
	}

	s.logger.WithField("response", &resp).Debug("list flags")
	return &resp, nil
}

// CreateFlag creates a flag
func (s *Server) CreateFlag(ctx context.Context, r *flipt.CreateFlagRequest) (*flipt.Flag, error) {
	s.logger.WithField("request", r).Debug("create flag")
	flag, err := s.FlagStore.CreateFlag(ctx, r)
	s.logger.WithField("response", flag).Debug("create flag")
	return flag, err
}

// UpdateFlag updates an existing flag
func (s *Server) UpdateFlag(ctx context.Context, r *flipt.UpdateFlagRequest) (*flipt.Flag, error) {
	s.logger.WithField("request", r).Debug("update flag")
	flag, err := s.FlagStore.UpdateFlag(ctx, r)
	s.logger.WithField("response", flag).Debug("update flag")
	return flag, err
}

// DeleteFlag deletes a flag
func (s *Server) DeleteFlag(ctx context.Context, r *flipt.DeleteFlagRequest) (*empty.Empty, error) {
	s.logger.WithField("request", r).Debug("delete flag")
	if err := s.FlagStore.DeleteFlag(ctx, r); err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

// CreateVariant creates a variant
func (s *Server) CreateVariant(ctx context.Context, r *flipt.CreateVariantRequest) (*flipt.Variant, error) {
	s.logger.WithField("request", r).Debug("create variant")
	variant, err := s.FlagStore.CreateVariant(ctx, r)
	s.logger.WithField("response", variant).Debug("create variant")
	return variant, err
}

// UpdateVariant updates an existing variant
func (s *Server) UpdateVariant(ctx context.Context, r *flipt.UpdateVariantRequest) (*flipt.Variant, error) {
	s.logger.WithField("request", r).Debug("update variant")
	variant, err := s.FlagStore.UpdateVariant(ctx, r)
	s.logger.WithField("response", variant).Debug("update variant")
	return variant, err
}

// DeleteVariant deletes a variant
func (s *Server) DeleteVariant(ctx context.Context, r *flipt.DeleteVariantRequest) (*empty.Empty, error) {
	s.logger.WithField("request", r).Debug("delete variant")
	if err := s.FlagStore.DeleteVariant(ctx, r); err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

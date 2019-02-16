package server

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/markphelps/flipt"
)

func (s *Server) GetFlag(ctx context.Context, req *flipt.GetFlagRequest) (*flipt.Flag, error) {
	if req.Key == "" {
		return nil, flipt.EmptyFieldError("key")
	}
	return s.FlagService.Flag(ctx, req)
}

func (s *Server) ListFlags(ctx context.Context, req *flipt.ListFlagRequest) (*flipt.FlagList, error) {
	flags, err := s.FlagService.Flags(ctx, req)
	if err != nil {
		return nil, err
	}

	var resp flipt.FlagList

	for i := range flags {
		resp.Flags = append(resp.Flags, flags[i])
	}

	return &resp, nil
}

func (s *Server) CreateFlag(ctx context.Context, req *flipt.CreateFlagRequest) (*flipt.Flag, error) {
	if req.Key == "" {
		return nil, flipt.EmptyFieldError("key")
	}
	if req.Name == "" {
		return nil, flipt.EmptyFieldError("name")
	}
	return s.FlagService.CreateFlag(ctx, req)
}

func (s *Server) UpdateFlag(ctx context.Context, req *flipt.UpdateFlagRequest) (*flipt.Flag, error) {
	if req.Key == "" {
		return nil, flipt.EmptyFieldError("key")
	}
	if req.Name == "" {
		return nil, flipt.EmptyFieldError("name")
	}
	return s.FlagService.UpdateFlag(ctx, req)
}

func (s *Server) DeleteFlag(ctx context.Context, req *flipt.DeleteFlagRequest) (*empty.Empty, error) {
	if req.Key == "" {
		return nil, flipt.EmptyFieldError("key")
	}

	if err := s.FlagService.DeleteFlag(ctx, req); err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (s *Server) CreateVariant(ctx context.Context, req *flipt.CreateVariantRequest) (*flipt.Variant, error) {
	if req.FlagKey == "" {
		return nil, flipt.EmptyFieldError("flagKey")
	}
	if req.Key == "" {
		return nil, flipt.EmptyFieldError("key")
	}
	return s.FlagService.CreateVariant(ctx, req)
}

func (s *Server) UpdateVariant(ctx context.Context, req *flipt.UpdateVariantRequest) (*flipt.Variant, error) {
	if req.Id == "" {
		return nil, flipt.EmptyFieldError("id")
	}
	if req.FlagKey == "" {
		return nil, flipt.EmptyFieldError("flagKey")
	}
	if req.Key == "" {
		return nil, flipt.EmptyFieldError("key")
	}
	return s.FlagService.UpdateVariant(ctx, req)
}

func (s *Server) DeleteVariant(ctx context.Context, req *flipt.DeleteVariantRequest) (*empty.Empty, error) {
	if req.Id == "" {
		return nil, flipt.EmptyFieldError("id")
	}
	if req.FlagKey == "" {
		return nil, flipt.EmptyFieldError("flagKey")
	}

	if err := s.FlagService.DeleteVariant(ctx, req); err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

package server

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	flipt "github.com/markphelps/flipt/proto"
)

func (s *Server) GetFlag(ctx context.Context, req *flipt.GetFlagRequest) (*flipt.Flag, error) {
	if req.Key == "" {
		return nil, EmptyFieldError("key")
	}
	return s.FlagRepository.Flag(ctx, req)
}

func (s *Server) ListFlags(ctx context.Context, req *flipt.ListFlagRequest) (*flipt.FlagList, error) {
	flags, err := s.FlagRepository.Flags(ctx, req)
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
		return nil, EmptyFieldError("key")
	}
	if req.Name == "" {
		return nil, EmptyFieldError("name")
	}
	return s.FlagRepository.CreateFlag(ctx, req)
}

func (s *Server) UpdateFlag(ctx context.Context, req *flipt.UpdateFlagRequest) (*flipt.Flag, error) {
	if req.Key == "" {
		return nil, EmptyFieldError("key")
	}
	if req.Name == "" {
		return nil, EmptyFieldError("name")
	}
	return s.FlagRepository.UpdateFlag(ctx, req)
}

func (s *Server) DeleteFlag(ctx context.Context, req *flipt.DeleteFlagRequest) (*empty.Empty, error) {
	if req.Key == "" {
		return nil, EmptyFieldError("key")
	}

	if err := s.FlagRepository.DeleteFlag(ctx, req); err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (s *Server) CreateVariant(ctx context.Context, req *flipt.CreateVariantRequest) (*flipt.Variant, error) {
	if req.FlagKey == "" {
		return nil, EmptyFieldError("flagKey")
	}
	if req.Key == "" {
		return nil, EmptyFieldError("key")
	}
	return s.FlagRepository.CreateVariant(ctx, req)
}

func (s *Server) UpdateVariant(ctx context.Context, req *flipt.UpdateVariantRequest) (*flipt.Variant, error) {
	if req.Id == "" {
		return nil, EmptyFieldError("id")
	}
	if req.FlagKey == "" {
		return nil, EmptyFieldError("flagKey")
	}
	if req.Key == "" {
		return nil, EmptyFieldError("key")
	}
	return s.FlagRepository.UpdateVariant(ctx, req)
}

func (s *Server) DeleteVariant(ctx context.Context, req *flipt.DeleteVariantRequest) (*empty.Empty, error) {
	if req.Id == "" {
		return nil, EmptyFieldError("id")
	}
	if req.FlagKey == "" {
		return nil, EmptyFieldError("flagKey")
	}

	if err := s.FlagRepository.DeleteVariant(ctx, req); err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

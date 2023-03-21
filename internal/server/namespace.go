package server

import (
	"context"
	"encoding/base64"

	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/storage"
	flipt "go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap"
	empty "google.golang.org/protobuf/types/known/emptypb"
)

// GetNamespace gets a namespace
func (s *Server) GetNamespace(ctx context.Context, r *flipt.GetNamespaceRequest) (*flipt.Namespace, error) {
	s.logger.Debug("get namespace", zap.Stringer("request", r))
	namespace, err := s.store.GetNamespace(ctx, r.Key)
	s.logger.Debug("get namespace", zap.Stringer("response", namespace))
	return namespace, err
}

// ListNamespaces lists all namespaces
func (s *Server) ListNamespaces(ctx context.Context, r *flipt.ListNamespaceRequest) (*flipt.NamespaceList, error) {
	s.logger.Debug("list namespaces", zap.Stringer("request", r))

	if r.Offset < 0 {
		r.Offset = 0
	}

	opts := []storage.QueryOption{storage.WithLimit(uint64(r.Limit))}

	if r.PageToken != "" {
		tok, err := base64.StdEncoding.DecodeString(r.PageToken)
		if err != nil {
			return nil, errors.ErrInvalidf("pageToken is not valid: %q", r.PageToken)
		}

		opts = append(opts, storage.WithPageToken(string(tok)))
	} else if r.Offset >= 0 {
		// TODO: deprecate
		opts = append(opts, storage.WithOffset(uint64(r.Offset)))
	}

	results, err := s.store.ListNamespaces(ctx, opts...)
	if err != nil {
		return nil, err
	}

	resp := flipt.NamespaceList{
		Namespaces: results.Results,
	}

	total, err := s.store.CountNamespaces(ctx)
	if err != nil {
		return nil, err
	}

	resp.TotalCount = int32(total)
	resp.NextPageToken = base64.StdEncoding.EncodeToString([]byte(results.NextPageToken))

	s.logger.Debug("list namespaces", zap.Stringer("response", &resp))
	return &resp, nil
}

// CreateNamespace creates a namespace
func (s *Server) CreateNamespace(ctx context.Context, r *flipt.CreateNamespaceRequest) (*flipt.Namespace, error) {
	s.logger.Debug("create namespace", zap.Stringer("request", r))
	namespace, err := s.store.CreateNamespace(ctx, r)
	s.logger.Debug("create namespace", zap.Stringer("response", namespace))
	return namespace, err
}

// UpdateNamespace updates an existing namespace
func (s *Server) UpdateNamespace(ctx context.Context, r *flipt.UpdateNamespaceRequest) (*flipt.Namespace, error) {
	s.logger.Debug("update namespace", zap.Stringer("request", r))
	namespace, err := s.store.UpdateNamespace(ctx, r)
	s.logger.Debug("update namespace", zap.Stringer("response", namespace))
	return namespace, err
}

// DeleteNamespace deletes a namespace
func (s *Server) DeleteNamespace(ctx context.Context, r *flipt.DeleteNamespaceRequest) (*empty.Empty, error) {
	s.logger.Debug("delete namespace", zap.Stringer("request", r))
	namespace, err := s.store.GetNamespace(ctx, r.Key)
	if err != nil {
		return nil, err
	}

	// if namespace is not found then nothing to do
	if namespace == nil {
		return &empty.Empty{}, nil
	}

	if namespace.Protected {
		return nil, errors.ErrInvalidf("namespace %q is protected", r.Key)
	}

	// if any flags exist under the namespace then it cannot be deleted
	count, err := s.store.CountFlags(ctx, r.Key)
	if err != nil {
		return nil, err
	}

	if count > 0 {
		return nil, errors.ErrInvalidf("namespace %q cannot be deleted; flags must be deleted first", r.Key)
	}

	if err := s.store.DeleteNamespace(ctx, r); err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

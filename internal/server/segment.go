package server

import (
	"context"

	"go.flipt.io/flipt/internal/storage"
	flipt "go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap"
	empty "google.golang.org/protobuf/types/known/emptypb"
)

// GetSegment gets a segment
func (s *Server) GetSegment(ctx context.Context, r *flipt.GetSegmentRequest) (*flipt.Segment, error) {
	s.logger.Debug("get segment", zap.Stringer("request", r))
	segment, err := s.store.GetSegment(ctx, storage.NewResource(r.NamespaceKey, r.Key, storage.WithReference(r.Reference)))
	s.logger.Debug("get segment", zap.Stringer("response", segment))
	return segment, err
}

// ListSegments lists all segments
func (s *Server) ListSegments(ctx context.Context, r *flipt.ListSegmentRequest) (*flipt.SegmentList, error) {
	s.logger.Debug("list segments", zap.Stringer("request", r))

	ns := storage.NewNamespace(r.NamespaceKey, storage.WithReference(r.Reference))
	results, err := s.store.ListSegments(ctx, storage.ListWithParameters(ns, r))
	if err != nil {
		return nil, err
	}

	resp := flipt.SegmentList{
		Segments: results.Results,
	}

	total, err := s.store.CountSegments(ctx, ns)
	if err != nil {
		return nil, err
	}

	resp.TotalCount = int32(total)
	resp.NextPageToken = results.NextPageToken

	s.logger.Debug("list segments", zap.Stringer("response", &resp))
	return &resp, nil
}

// CreateSegment creates a segment
func (s *Server) CreateSegment(ctx context.Context, r *flipt.CreateSegmentRequest) (*flipt.Segment, error) {
	s.logger.Debug("create segment", zap.Stringer("request", r))
	segment, err := s.store.CreateSegment(ctx, r)
	s.logger.Debug("create segment", zap.Stringer("response", segment))
	return segment, err
}

// UpdateSegment updates an existing segment
func (s *Server) UpdateSegment(ctx context.Context, r *flipt.UpdateSegmentRequest) (*flipt.Segment, error) {
	s.logger.Debug("update segment", zap.Stringer("request", r))
	segment, err := s.store.UpdateSegment(ctx, r)
	s.logger.Debug("update segment", zap.Stringer("response", segment))
	return segment, err
}

// DeleteSegment deletes a segment
func (s *Server) DeleteSegment(ctx context.Context, r *flipt.DeleteSegmentRequest) (*empty.Empty, error) {
	s.logger.Debug("delete segment", zap.Stringer("request", r))
	if err := s.store.DeleteSegment(ctx, r); err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

// CreateConstraint creates a constraint
func (s *Server) CreateConstraint(ctx context.Context, r *flipt.CreateConstraintRequest) (*flipt.Constraint, error) {
	s.logger.Debug("create constraint", zap.Stringer("request", r))
	constraint, err := s.store.CreateConstraint(ctx, r)
	s.logger.Debug("create constraint", zap.Stringer("response", constraint))
	return constraint, err
}

// UpdateConstraint updates an existing constraint
func (s *Server) UpdateConstraint(ctx context.Context, r *flipt.UpdateConstraintRequest) (*flipt.Constraint, error) {
	s.logger.Debug("update constraint", zap.Stringer("request", r))
	constraint, err := s.store.UpdateConstraint(ctx, r)
	s.logger.Debug("update constraint", zap.Stringer("response", constraint))
	return constraint, err
}

// DeleteConstraint deletes a constraint
func (s *Server) DeleteConstraint(ctx context.Context, r *flipt.DeleteConstraintRequest) (*empty.Empty, error) {
	s.logger.Debug("delete constraint", zap.Stringer("request", r))
	if err := s.store.DeleteConstraint(ctx, r); err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

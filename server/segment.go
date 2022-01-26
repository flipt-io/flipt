package server

import (
	"context"

	flipt "github.com/markphelps/flipt/rpc/flipt"
	"github.com/markphelps/flipt/storage"
	empty "google.golang.org/protobuf/types/known/emptypb"
)

// GetSegment gets a segment
func (s *Server) GetSegment(ctx context.Context, r *flipt.GetSegmentRequest) (*flipt.Segment, error) {
	s.logger.WithField("request", r).Debug("get segment")
	segment, err := s.store.GetSegment(ctx, r.Key)
	s.logger.WithField("response", segment).Debug("get segment")
	return segment, err
}

// ListSegments lists all segments
func (s *Server) ListSegments(ctx context.Context, r *flipt.ListSegmentRequest) (*flipt.SegmentList, error) {
	s.logger.WithField("request", r).Debug("list segments")
	segments, err := s.store.ListSegments(ctx, storage.WithLimit(uint64(r.Limit)), storage.WithOffset(uint64(r.Offset)))
	if err != nil {
		return nil, err
	}

	var resp flipt.SegmentList

	for i := range segments {
		resp.Segments = append(resp.Segments, segments[i])
	}

	s.logger.WithField("response", &resp).Debug("list segments")
	return &resp, nil
}

// CreateSegment creates a segment
func (s *Server) CreateSegment(ctx context.Context, r *flipt.CreateSegmentRequest) (*flipt.Segment, error) {
	s.logger.WithField("request", r).Debug("create segment")
	segment, err := s.store.CreateSegment(ctx, r)
	s.logger.WithField("response", segment).Debug("create segment")
	return segment, err
}

// UpdateSegment updates an existing segment
func (s *Server) UpdateSegment(ctx context.Context, r *flipt.UpdateSegmentRequest) (*flipt.Segment, error) {
	s.logger.WithField("request", r).Debug("update segment")
	segment, err := s.store.UpdateSegment(ctx, r)
	s.logger.WithField("response", segment).Debug("update segment")
	return segment, err
}

// DeleteSegment deletes a segment
func (s *Server) DeleteSegment(ctx context.Context, r *flipt.DeleteSegmentRequest) (*empty.Empty, error) {
	s.logger.WithField("request", r).Debug("delete segment")
	if err := s.store.DeleteSegment(ctx, r); err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

// CreateConstraint creates a constraint
func (s *Server) CreateConstraint(ctx context.Context, r *flipt.CreateConstraintRequest) (*flipt.Constraint, error) {
	s.logger.WithField("request", r).Debug("create constraint")
	constraint, err := s.store.CreateConstraint(ctx, r)
	s.logger.WithField("response", constraint).Debug("create constraint")
	return constraint, err
}

// UpdateConstraint updates an existing constraint
func (s *Server) UpdateConstraint(ctx context.Context, r *flipt.UpdateConstraintRequest) (*flipt.Constraint, error) {
	s.logger.WithField("request", r).Debug("update constraint")
	constraint, err := s.store.UpdateConstraint(ctx, r)
	s.logger.WithField("response", constraint).Debug("update constraint")
	return constraint, err
}

// DeleteConstraint deletes a constraint
func (s *Server) DeleteConstraint(ctx context.Context, r *flipt.DeleteConstraintRequest) (*empty.Empty, error) {
	s.logger.WithField("request", r).Debug("delete constraint")
	if err := s.store.DeleteConstraint(ctx, r); err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

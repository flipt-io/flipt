package server

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	flipt "github.com/markphelps/flipt/rpc"
)

// GetSegment gets a segment
func (s *Server) GetSegment(ctx context.Context, req *flipt.GetSegmentRequest) (*flipt.Segment, error) {
	return s.SegmentStore.GetSegment(ctx, req)
}

// ListSegments lists all segments
func (s *Server) ListSegments(ctx context.Context, req *flipt.ListSegmentRequest) (*flipt.SegmentList, error) {
	segments, err := s.SegmentStore.ListSegments(ctx, req)
	if err != nil {
		return nil, err
	}

	var resp flipt.SegmentList

	for i := range segments {
		resp.Segments = append(resp.Segments, segments[i])
	}

	return &resp, nil
}

// CreateSegment creates a segment
func (s *Server) CreateSegment(ctx context.Context, req *flipt.CreateSegmentRequest) (*flipt.Segment, error) {
	return s.SegmentStore.CreateSegment(ctx, req)
}

// UpdateSegment updates an existing segment
func (s *Server) UpdateSegment(ctx context.Context, req *flipt.UpdateSegmentRequest) (*flipt.Segment, error) {
	return s.SegmentStore.UpdateSegment(ctx, req)
}

// DeleteSegment deletes a segment
func (s *Server) DeleteSegment(ctx context.Context, req *flipt.DeleteSegmentRequest) (*empty.Empty, error) {
	if err := s.SegmentStore.DeleteSegment(ctx, req); err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

// CreateConstraint creates a constraint
func (s *Server) CreateConstraint(ctx context.Context, req *flipt.CreateConstraintRequest) (*flipt.Constraint, error) {
	return s.SegmentStore.CreateConstraint(ctx, req)
}

// UpdateConstraint updates an existing constraint
func (s *Server) UpdateConstraint(ctx context.Context, req *flipt.UpdateConstraintRequest) (*flipt.Constraint, error) {
	return s.SegmentStore.UpdateConstraint(ctx, req)
}

// DeleteConstraint deletes a constraint
func (s *Server) DeleteConstraint(ctx context.Context, req *flipt.DeleteConstraintRequest) (*empty.Empty, error) {
	if err := s.SegmentStore.DeleteConstraint(ctx, req); err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

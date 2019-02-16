package server

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/markphelps/flipt"
)

func (s *Server) GetSegment(ctx context.Context, req *flipt.GetSegmentRequest) (*flipt.Segment, error) {
	if req.Key == "" {
		return nil, flipt.EmptyFieldError("key")
	}
	return s.SegmentService.Segment(ctx, req)
}

func (s *Server) ListSegments(ctx context.Context, req *flipt.ListSegmentRequest) (*flipt.SegmentList, error) {
	segments, err := s.SegmentService.Segments(ctx, req)
	if err != nil {
		return nil, err
	}

	var resp flipt.SegmentList

	for i := range segments {
		resp.Segments = append(resp.Segments, segments[i])
	}

	return &resp, nil
}

func (s *Server) CreateSegment(ctx context.Context, req *flipt.CreateSegmentRequest) (*flipt.Segment, error) {
	if req.Key == "" {
		return nil, flipt.EmptyFieldError("key")
	}
	if req.Name == "" {
		return nil, flipt.EmptyFieldError("name")
	}
	return s.SegmentService.CreateSegment(ctx, req)
}

func (s *Server) UpdateSegment(ctx context.Context, req *flipt.UpdateSegmentRequest) (*flipt.Segment, error) {
	if req.Key == "" {
		return nil, flipt.EmptyFieldError("key")
	}
	if req.Name == "" {
		return nil, flipt.EmptyFieldError("name")
	}
	return s.SegmentService.UpdateSegment(ctx, req)
}

func (s *Server) DeleteSegment(ctx context.Context, req *flipt.DeleteSegmentRequest) (*empty.Empty, error) {
	if req.Key == "" {
		return nil, flipt.EmptyFieldError("key")
	}

	if err := s.SegmentService.DeleteSegment(ctx, req); err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (s *Server) CreateConstraint(ctx context.Context, req *flipt.CreateConstraintRequest) (*flipt.Constraint, error) {
	if req.SegmentKey == "" {
		return nil, flipt.EmptyFieldError("segmentKey")
	}
	if req.Property == "" {
		return nil, flipt.EmptyFieldError("property")
	}
	if req.Operator == "" {
		return nil, flipt.EmptyFieldError("operator")
	}
	return s.SegmentService.CreateConstraint(ctx, req)
}

func (s *Server) UpdateConstraint(ctx context.Context, req *flipt.UpdateConstraintRequest) (*flipt.Constraint, error) {
	if req.Id == "" {
		return nil, flipt.EmptyFieldError("id")
	}
	if req.SegmentKey == "" {
		return nil, flipt.EmptyFieldError("segmentKey")
	}
	if req.Property == "" {
		return nil, flipt.EmptyFieldError("property")
	}
	if req.Operator == "" {
		return nil, flipt.EmptyFieldError("operator")
	}
	return s.SegmentService.UpdateConstraint(ctx, req)
}

func (s *Server) DeleteConstraint(ctx context.Context, req *flipt.DeleteConstraintRequest) (*empty.Empty, error) {
	if req.Id == "" {
		return nil, flipt.EmptyFieldError("id")
	}
	if req.SegmentKey == "" {
		return nil, flipt.EmptyFieldError("segmentKey")
	}

	if err := s.SegmentService.DeleteConstraint(ctx, req); err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

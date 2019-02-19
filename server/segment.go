package server

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	flipt "github.com/markphelps/flipt/proto"
)

func (s *Server) GetSegment(ctx context.Context, req *flipt.GetSegmentRequest) (*flipt.Segment, error) {
	if req.Key == "" {
		return nil, EmptyFieldError("key")
	}
	return s.SegmentRepository.Segment(ctx, req)
}

func (s *Server) ListSegments(ctx context.Context, req *flipt.ListSegmentRequest) (*flipt.SegmentList, error) {
	segments, err := s.SegmentRepository.Segments(ctx, req)
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
		return nil, EmptyFieldError("key")
	}
	if req.Name == "" {
		return nil, EmptyFieldError("name")
	}
	return s.SegmentRepository.CreateSegment(ctx, req)
}

func (s *Server) UpdateSegment(ctx context.Context, req *flipt.UpdateSegmentRequest) (*flipt.Segment, error) {
	if req.Key == "" {
		return nil, EmptyFieldError("key")
	}
	if req.Name == "" {
		return nil, EmptyFieldError("name")
	}
	return s.SegmentRepository.UpdateSegment(ctx, req)
}

func (s *Server) DeleteSegment(ctx context.Context, req *flipt.DeleteSegmentRequest) (*empty.Empty, error) {
	if req.Key == "" {
		return nil, EmptyFieldError("key")
	}

	if err := s.SegmentRepository.DeleteSegment(ctx, req); err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (s *Server) CreateConstraint(ctx context.Context, req *flipt.CreateConstraintRequest) (*flipt.Constraint, error) {
	if req.SegmentKey == "" {
		return nil, EmptyFieldError("segmentKey")
	}
	if req.Property == "" {
		return nil, EmptyFieldError("property")
	}
	if req.Operator == "" {
		return nil, EmptyFieldError("operator")
	}
	return s.SegmentRepository.CreateConstraint(ctx, req)
}

func (s *Server) UpdateConstraint(ctx context.Context, req *flipt.UpdateConstraintRequest) (*flipt.Constraint, error) {
	if req.Id == "" {
		return nil, EmptyFieldError("id")
	}
	if req.SegmentKey == "" {
		return nil, EmptyFieldError("segmentKey")
	}
	if req.Property == "" {
		return nil, EmptyFieldError("property")
	}
	if req.Operator == "" {
		return nil, EmptyFieldError("operator")
	}
	return s.SegmentRepository.UpdateConstraint(ctx, req)
}

func (s *Server) DeleteConstraint(ctx context.Context, req *flipt.DeleteConstraintRequest) (*empty.Empty, error) {
	if req.Id == "" {
		return nil, EmptyFieldError("id")
	}
	if req.SegmentKey == "" {
		return nil, EmptyFieldError("segmentKey")
	}

	if err := s.SegmentRepository.DeleteConstraint(ctx, req); err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

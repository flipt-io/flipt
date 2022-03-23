// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             (unknown)
// source: flipt.proto

package flipt

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// FliptClient is the client API for Flipt service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type FliptClient interface {
	Evaluate(ctx context.Context, in *EvaluationRequest, opts ...grpc.CallOption) (*EvaluationResponse, error)
	BatchEvaluate(ctx context.Context, in *BatchEvaluationRequest, opts ...grpc.CallOption) (*BatchEvaluationResponse, error)
	GetFlag(ctx context.Context, in *GetFlagRequest, opts ...grpc.CallOption) (*Flag, error)
	ListFlags(ctx context.Context, in *ListFlagRequest, opts ...grpc.CallOption) (*FlagList, error)
	CreateFlag(ctx context.Context, in *CreateFlagRequest, opts ...grpc.CallOption) (*Flag, error)
	UpdateFlag(ctx context.Context, in *UpdateFlagRequest, opts ...grpc.CallOption) (*Flag, error)
	DeleteFlag(ctx context.Context, in *DeleteFlagRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	CreateVariant(ctx context.Context, in *CreateVariantRequest, opts ...grpc.CallOption) (*Variant, error)
	UpdateVariant(ctx context.Context, in *UpdateVariantRequest, opts ...grpc.CallOption) (*Variant, error)
	DeleteVariant(ctx context.Context, in *DeleteVariantRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	GetRule(ctx context.Context, in *GetRuleRequest, opts ...grpc.CallOption) (*Rule, error)
	ListRules(ctx context.Context, in *ListRuleRequest, opts ...grpc.CallOption) (*RuleList, error)
	OrderRules(ctx context.Context, in *OrderRulesRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	CreateRule(ctx context.Context, in *CreateRuleRequest, opts ...grpc.CallOption) (*Rule, error)
	UpdateRule(ctx context.Context, in *UpdateRuleRequest, opts ...grpc.CallOption) (*Rule, error)
	DeleteRule(ctx context.Context, in *DeleteRuleRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	CreateDistribution(ctx context.Context, in *CreateDistributionRequest, opts ...grpc.CallOption) (*Distribution, error)
	UpdateDistribution(ctx context.Context, in *UpdateDistributionRequest, opts ...grpc.CallOption) (*Distribution, error)
	DeleteDistribution(ctx context.Context, in *DeleteDistributionRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	GetSegment(ctx context.Context, in *GetSegmentRequest, opts ...grpc.CallOption) (*Segment, error)
	ListSegments(ctx context.Context, in *ListSegmentRequest, opts ...grpc.CallOption) (*SegmentList, error)
	CreateSegment(ctx context.Context, in *CreateSegmentRequest, opts ...grpc.CallOption) (*Segment, error)
	UpdateSegment(ctx context.Context, in *UpdateSegmentRequest, opts ...grpc.CallOption) (*Segment, error)
	DeleteSegment(ctx context.Context, in *DeleteSegmentRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	CreateConstraint(ctx context.Context, in *CreateConstraintRequest, opts ...grpc.CallOption) (*Constraint, error)
	UpdateConstraint(ctx context.Context, in *UpdateConstraintRequest, opts ...grpc.CallOption) (*Constraint, error)
	DeleteConstraint(ctx context.Context, in *DeleteConstraintRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type fliptClient struct {
	cc grpc.ClientConnInterface
}

func NewFliptClient(cc grpc.ClientConnInterface) FliptClient {
	return &fliptClient{cc}
}

func (c *fliptClient) Evaluate(ctx context.Context, in *EvaluationRequest, opts ...grpc.CallOption) (*EvaluationResponse, error) {
	out := new(EvaluationResponse)
	err := c.cc.Invoke(ctx, "/flipt.Flipt/Evaluate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fliptClient) BatchEvaluate(ctx context.Context, in *BatchEvaluationRequest, opts ...grpc.CallOption) (*BatchEvaluationResponse, error) {
	out := new(BatchEvaluationResponse)
	err := c.cc.Invoke(ctx, "/flipt.Flipt/BatchEvaluate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fliptClient) GetFlag(ctx context.Context, in *GetFlagRequest, opts ...grpc.CallOption) (*Flag, error) {
	out := new(Flag)
	err := c.cc.Invoke(ctx, "/flipt.Flipt/GetFlag", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fliptClient) ListFlags(ctx context.Context, in *ListFlagRequest, opts ...grpc.CallOption) (*FlagList, error) {
	out := new(FlagList)
	err := c.cc.Invoke(ctx, "/flipt.Flipt/ListFlags", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fliptClient) CreateFlag(ctx context.Context, in *CreateFlagRequest, opts ...grpc.CallOption) (*Flag, error) {
	out := new(Flag)
	err := c.cc.Invoke(ctx, "/flipt.Flipt/CreateFlag", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fliptClient) UpdateFlag(ctx context.Context, in *UpdateFlagRequest, opts ...grpc.CallOption) (*Flag, error) {
	out := new(Flag)
	err := c.cc.Invoke(ctx, "/flipt.Flipt/UpdateFlag", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fliptClient) DeleteFlag(ctx context.Context, in *DeleteFlagRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/flipt.Flipt/DeleteFlag", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fliptClient) CreateVariant(ctx context.Context, in *CreateVariantRequest, opts ...grpc.CallOption) (*Variant, error) {
	out := new(Variant)
	err := c.cc.Invoke(ctx, "/flipt.Flipt/CreateVariant", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fliptClient) UpdateVariant(ctx context.Context, in *UpdateVariantRequest, opts ...grpc.CallOption) (*Variant, error) {
	out := new(Variant)
	err := c.cc.Invoke(ctx, "/flipt.Flipt/UpdateVariant", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fliptClient) DeleteVariant(ctx context.Context, in *DeleteVariantRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/flipt.Flipt/DeleteVariant", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fliptClient) GetRule(ctx context.Context, in *GetRuleRequest, opts ...grpc.CallOption) (*Rule, error) {
	out := new(Rule)
	err := c.cc.Invoke(ctx, "/flipt.Flipt/GetRule", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fliptClient) ListRules(ctx context.Context, in *ListRuleRequest, opts ...grpc.CallOption) (*RuleList, error) {
	out := new(RuleList)
	err := c.cc.Invoke(ctx, "/flipt.Flipt/ListRules", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fliptClient) OrderRules(ctx context.Context, in *OrderRulesRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/flipt.Flipt/OrderRules", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fliptClient) CreateRule(ctx context.Context, in *CreateRuleRequest, opts ...grpc.CallOption) (*Rule, error) {
	out := new(Rule)
	err := c.cc.Invoke(ctx, "/flipt.Flipt/CreateRule", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fliptClient) UpdateRule(ctx context.Context, in *UpdateRuleRequest, opts ...grpc.CallOption) (*Rule, error) {
	out := new(Rule)
	err := c.cc.Invoke(ctx, "/flipt.Flipt/UpdateRule", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fliptClient) DeleteRule(ctx context.Context, in *DeleteRuleRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/flipt.Flipt/DeleteRule", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fliptClient) CreateDistribution(ctx context.Context, in *CreateDistributionRequest, opts ...grpc.CallOption) (*Distribution, error) {
	out := new(Distribution)
	err := c.cc.Invoke(ctx, "/flipt.Flipt/CreateDistribution", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fliptClient) UpdateDistribution(ctx context.Context, in *UpdateDistributionRequest, opts ...grpc.CallOption) (*Distribution, error) {
	out := new(Distribution)
	err := c.cc.Invoke(ctx, "/flipt.Flipt/UpdateDistribution", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fliptClient) DeleteDistribution(ctx context.Context, in *DeleteDistributionRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/flipt.Flipt/DeleteDistribution", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fliptClient) GetSegment(ctx context.Context, in *GetSegmentRequest, opts ...grpc.CallOption) (*Segment, error) {
	out := new(Segment)
	err := c.cc.Invoke(ctx, "/flipt.Flipt/GetSegment", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fliptClient) ListSegments(ctx context.Context, in *ListSegmentRequest, opts ...grpc.CallOption) (*SegmentList, error) {
	out := new(SegmentList)
	err := c.cc.Invoke(ctx, "/flipt.Flipt/ListSegments", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fliptClient) CreateSegment(ctx context.Context, in *CreateSegmentRequest, opts ...grpc.CallOption) (*Segment, error) {
	out := new(Segment)
	err := c.cc.Invoke(ctx, "/flipt.Flipt/CreateSegment", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fliptClient) UpdateSegment(ctx context.Context, in *UpdateSegmentRequest, opts ...grpc.CallOption) (*Segment, error) {
	out := new(Segment)
	err := c.cc.Invoke(ctx, "/flipt.Flipt/UpdateSegment", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fliptClient) DeleteSegment(ctx context.Context, in *DeleteSegmentRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/flipt.Flipt/DeleteSegment", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fliptClient) CreateConstraint(ctx context.Context, in *CreateConstraintRequest, opts ...grpc.CallOption) (*Constraint, error) {
	out := new(Constraint)
	err := c.cc.Invoke(ctx, "/flipt.Flipt/CreateConstraint", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fliptClient) UpdateConstraint(ctx context.Context, in *UpdateConstraintRequest, opts ...grpc.CallOption) (*Constraint, error) {
	out := new(Constraint)
	err := c.cc.Invoke(ctx, "/flipt.Flipt/UpdateConstraint", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fliptClient) DeleteConstraint(ctx context.Context, in *DeleteConstraintRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/flipt.Flipt/DeleteConstraint", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// FliptServer is the server API for Flipt service.
// All implementations must embed UnimplementedFliptServer
// for forward compatibility
type FliptServer interface {
	Evaluate(context.Context, *EvaluationRequest) (*EvaluationResponse, error)
	BatchEvaluate(context.Context, *BatchEvaluationRequest) (*BatchEvaluationResponse, error)
	GetFlag(context.Context, *GetFlagRequest) (*Flag, error)
	ListFlags(context.Context, *ListFlagRequest) (*FlagList, error)
	CreateFlag(context.Context, *CreateFlagRequest) (*Flag, error)
	UpdateFlag(context.Context, *UpdateFlagRequest) (*Flag, error)
	DeleteFlag(context.Context, *DeleteFlagRequest) (*emptypb.Empty, error)
	CreateVariant(context.Context, *CreateVariantRequest) (*Variant, error)
	UpdateVariant(context.Context, *UpdateVariantRequest) (*Variant, error)
	DeleteVariant(context.Context, *DeleteVariantRequest) (*emptypb.Empty, error)
	GetRule(context.Context, *GetRuleRequest) (*Rule, error)
	ListRules(context.Context, *ListRuleRequest) (*RuleList, error)
	OrderRules(context.Context, *OrderRulesRequest) (*emptypb.Empty, error)
	CreateRule(context.Context, *CreateRuleRequest) (*Rule, error)
	UpdateRule(context.Context, *UpdateRuleRequest) (*Rule, error)
	DeleteRule(context.Context, *DeleteRuleRequest) (*emptypb.Empty, error)
	CreateDistribution(context.Context, *CreateDistributionRequest) (*Distribution, error)
	UpdateDistribution(context.Context, *UpdateDistributionRequest) (*Distribution, error)
	DeleteDistribution(context.Context, *DeleteDistributionRequest) (*emptypb.Empty, error)
	GetSegment(context.Context, *GetSegmentRequest) (*Segment, error)
	ListSegments(context.Context, *ListSegmentRequest) (*SegmentList, error)
	CreateSegment(context.Context, *CreateSegmentRequest) (*Segment, error)
	UpdateSegment(context.Context, *UpdateSegmentRequest) (*Segment, error)
	DeleteSegment(context.Context, *DeleteSegmentRequest) (*emptypb.Empty, error)
	CreateConstraint(context.Context, *CreateConstraintRequest) (*Constraint, error)
	UpdateConstraint(context.Context, *UpdateConstraintRequest) (*Constraint, error)
	DeleteConstraint(context.Context, *DeleteConstraintRequest) (*emptypb.Empty, error)
	mustEmbedUnimplementedFliptServer()
}

// UnimplementedFliptServer must be embedded to have forward compatible implementations.
type UnimplementedFliptServer struct {
}

func (UnimplementedFliptServer) Evaluate(context.Context, *EvaluationRequest) (*EvaluationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Evaluate not implemented")
}
func (UnimplementedFliptServer) BatchEvaluate(context.Context, *BatchEvaluationRequest) (*BatchEvaluationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BatchEvaluate not implemented")
}
func (UnimplementedFliptServer) GetFlag(context.Context, *GetFlagRequest) (*Flag, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetFlag not implemented")
}
func (UnimplementedFliptServer) ListFlags(context.Context, *ListFlagRequest) (*FlagList, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListFlags not implemented")
}
func (UnimplementedFliptServer) CreateFlag(context.Context, *CreateFlagRequest) (*Flag, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateFlag not implemented")
}
func (UnimplementedFliptServer) UpdateFlag(context.Context, *UpdateFlagRequest) (*Flag, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateFlag not implemented")
}
func (UnimplementedFliptServer) DeleteFlag(context.Context, *DeleteFlagRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteFlag not implemented")
}
func (UnimplementedFliptServer) CreateVariant(context.Context, *CreateVariantRequest) (*Variant, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateVariant not implemented")
}
func (UnimplementedFliptServer) UpdateVariant(context.Context, *UpdateVariantRequest) (*Variant, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateVariant not implemented")
}
func (UnimplementedFliptServer) DeleteVariant(context.Context, *DeleteVariantRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteVariant not implemented")
}
func (UnimplementedFliptServer) GetRule(context.Context, *GetRuleRequest) (*Rule, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRule not implemented")
}
func (UnimplementedFliptServer) ListRules(context.Context, *ListRuleRequest) (*RuleList, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListRules not implemented")
}
func (UnimplementedFliptServer) OrderRules(context.Context, *OrderRulesRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method OrderRules not implemented")
}
func (UnimplementedFliptServer) CreateRule(context.Context, *CreateRuleRequest) (*Rule, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateRule not implemented")
}
func (UnimplementedFliptServer) UpdateRule(context.Context, *UpdateRuleRequest) (*Rule, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateRule not implemented")
}
func (UnimplementedFliptServer) DeleteRule(context.Context, *DeleteRuleRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteRule not implemented")
}
func (UnimplementedFliptServer) CreateDistribution(context.Context, *CreateDistributionRequest) (*Distribution, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateDistribution not implemented")
}
func (UnimplementedFliptServer) UpdateDistribution(context.Context, *UpdateDistributionRequest) (*Distribution, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateDistribution not implemented")
}
func (UnimplementedFliptServer) DeleteDistribution(context.Context, *DeleteDistributionRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteDistribution not implemented")
}
func (UnimplementedFliptServer) GetSegment(context.Context, *GetSegmentRequest) (*Segment, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetSegment not implemented")
}
func (UnimplementedFliptServer) ListSegments(context.Context, *ListSegmentRequest) (*SegmentList, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListSegments not implemented")
}
func (UnimplementedFliptServer) CreateSegment(context.Context, *CreateSegmentRequest) (*Segment, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateSegment not implemented")
}
func (UnimplementedFliptServer) UpdateSegment(context.Context, *UpdateSegmentRequest) (*Segment, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateSegment not implemented")
}
func (UnimplementedFliptServer) DeleteSegment(context.Context, *DeleteSegmentRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteSegment not implemented")
}
func (UnimplementedFliptServer) CreateConstraint(context.Context, *CreateConstraintRequest) (*Constraint, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateConstraint not implemented")
}
func (UnimplementedFliptServer) UpdateConstraint(context.Context, *UpdateConstraintRequest) (*Constraint, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateConstraint not implemented")
}
func (UnimplementedFliptServer) DeleteConstraint(context.Context, *DeleteConstraintRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteConstraint not implemented")
}
func (UnimplementedFliptServer) mustEmbedUnimplementedFliptServer() {}

// UnsafeFliptServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to FliptServer will
// result in compilation errors.
type UnsafeFliptServer interface {
	mustEmbedUnimplementedFliptServer()
}

func RegisterFliptServer(s grpc.ServiceRegistrar, srv FliptServer) {
	s.RegisterService(&Flipt_ServiceDesc, srv)
}

func _Flipt_Evaluate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EvaluationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FliptServer).Evaluate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/flipt.Flipt/Evaluate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FliptServer).Evaluate(ctx, req.(*EvaluationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Flipt_BatchEvaluate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BatchEvaluationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FliptServer).BatchEvaluate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/flipt.Flipt/BatchEvaluate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FliptServer).BatchEvaluate(ctx, req.(*BatchEvaluationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Flipt_GetFlag_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetFlagRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FliptServer).GetFlag(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/flipt.Flipt/GetFlag",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FliptServer).GetFlag(ctx, req.(*GetFlagRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Flipt_ListFlags_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListFlagRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FliptServer).ListFlags(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/flipt.Flipt/ListFlags",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FliptServer).ListFlags(ctx, req.(*ListFlagRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Flipt_CreateFlag_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateFlagRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FliptServer).CreateFlag(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/flipt.Flipt/CreateFlag",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FliptServer).CreateFlag(ctx, req.(*CreateFlagRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Flipt_UpdateFlag_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateFlagRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FliptServer).UpdateFlag(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/flipt.Flipt/UpdateFlag",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FliptServer).UpdateFlag(ctx, req.(*UpdateFlagRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Flipt_DeleteFlag_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteFlagRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FliptServer).DeleteFlag(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/flipt.Flipt/DeleteFlag",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FliptServer).DeleteFlag(ctx, req.(*DeleteFlagRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Flipt_CreateVariant_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateVariantRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FliptServer).CreateVariant(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/flipt.Flipt/CreateVariant",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FliptServer).CreateVariant(ctx, req.(*CreateVariantRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Flipt_UpdateVariant_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateVariantRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FliptServer).UpdateVariant(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/flipt.Flipt/UpdateVariant",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FliptServer).UpdateVariant(ctx, req.(*UpdateVariantRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Flipt_DeleteVariant_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteVariantRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FliptServer).DeleteVariant(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/flipt.Flipt/DeleteVariant",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FliptServer).DeleteVariant(ctx, req.(*DeleteVariantRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Flipt_GetRule_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRuleRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FliptServer).GetRule(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/flipt.Flipt/GetRule",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FliptServer).GetRule(ctx, req.(*GetRuleRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Flipt_ListRules_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListRuleRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FliptServer).ListRules(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/flipt.Flipt/ListRules",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FliptServer).ListRules(ctx, req.(*ListRuleRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Flipt_OrderRules_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(OrderRulesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FliptServer).OrderRules(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/flipt.Flipt/OrderRules",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FliptServer).OrderRules(ctx, req.(*OrderRulesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Flipt_CreateRule_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateRuleRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FliptServer).CreateRule(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/flipt.Flipt/CreateRule",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FliptServer).CreateRule(ctx, req.(*CreateRuleRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Flipt_UpdateRule_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateRuleRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FliptServer).UpdateRule(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/flipt.Flipt/UpdateRule",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FliptServer).UpdateRule(ctx, req.(*UpdateRuleRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Flipt_DeleteRule_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteRuleRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FliptServer).DeleteRule(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/flipt.Flipt/DeleteRule",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FliptServer).DeleteRule(ctx, req.(*DeleteRuleRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Flipt_CreateDistribution_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateDistributionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FliptServer).CreateDistribution(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/flipt.Flipt/CreateDistribution",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FliptServer).CreateDistribution(ctx, req.(*CreateDistributionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Flipt_UpdateDistribution_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateDistributionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FliptServer).UpdateDistribution(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/flipt.Flipt/UpdateDistribution",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FliptServer).UpdateDistribution(ctx, req.(*UpdateDistributionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Flipt_DeleteDistribution_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteDistributionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FliptServer).DeleteDistribution(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/flipt.Flipt/DeleteDistribution",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FliptServer).DeleteDistribution(ctx, req.(*DeleteDistributionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Flipt_GetSegment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetSegmentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FliptServer).GetSegment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/flipt.Flipt/GetSegment",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FliptServer).GetSegment(ctx, req.(*GetSegmentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Flipt_ListSegments_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListSegmentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FliptServer).ListSegments(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/flipt.Flipt/ListSegments",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FliptServer).ListSegments(ctx, req.(*ListSegmentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Flipt_CreateSegment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateSegmentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FliptServer).CreateSegment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/flipt.Flipt/CreateSegment",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FliptServer).CreateSegment(ctx, req.(*CreateSegmentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Flipt_UpdateSegment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateSegmentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FliptServer).UpdateSegment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/flipt.Flipt/UpdateSegment",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FliptServer).UpdateSegment(ctx, req.(*UpdateSegmentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Flipt_DeleteSegment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteSegmentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FliptServer).DeleteSegment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/flipt.Flipt/DeleteSegment",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FliptServer).DeleteSegment(ctx, req.(*DeleteSegmentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Flipt_CreateConstraint_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateConstraintRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FliptServer).CreateConstraint(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/flipt.Flipt/CreateConstraint",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FliptServer).CreateConstraint(ctx, req.(*CreateConstraintRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Flipt_UpdateConstraint_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateConstraintRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FliptServer).UpdateConstraint(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/flipt.Flipt/UpdateConstraint",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FliptServer).UpdateConstraint(ctx, req.(*UpdateConstraintRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Flipt_DeleteConstraint_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteConstraintRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FliptServer).DeleteConstraint(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/flipt.Flipt/DeleteConstraint",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FliptServer).DeleteConstraint(ctx, req.(*DeleteConstraintRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Flipt_ServiceDesc is the grpc.ServiceDesc for Flipt service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Flipt_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "flipt.Flipt",
	HandlerType: (*FliptServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Evaluate",
			Handler:    _Flipt_Evaluate_Handler,
		},
		{
			MethodName: "BatchEvaluate",
			Handler:    _Flipt_BatchEvaluate_Handler,
		},
		{
			MethodName: "GetFlag",
			Handler:    _Flipt_GetFlag_Handler,
		},
		{
			MethodName: "ListFlags",
			Handler:    _Flipt_ListFlags_Handler,
		},
		{
			MethodName: "CreateFlag",
			Handler:    _Flipt_CreateFlag_Handler,
		},
		{
			MethodName: "UpdateFlag",
			Handler:    _Flipt_UpdateFlag_Handler,
		},
		{
			MethodName: "DeleteFlag",
			Handler:    _Flipt_DeleteFlag_Handler,
		},
		{
			MethodName: "CreateVariant",
			Handler:    _Flipt_CreateVariant_Handler,
		},
		{
			MethodName: "UpdateVariant",
			Handler:    _Flipt_UpdateVariant_Handler,
		},
		{
			MethodName: "DeleteVariant",
			Handler:    _Flipt_DeleteVariant_Handler,
		},
		{
			MethodName: "GetRule",
			Handler:    _Flipt_GetRule_Handler,
		},
		{
			MethodName: "ListRules",
			Handler:    _Flipt_ListRules_Handler,
		},
		{
			MethodName: "OrderRules",
			Handler:    _Flipt_OrderRules_Handler,
		},
		{
			MethodName: "CreateRule",
			Handler:    _Flipt_CreateRule_Handler,
		},
		{
			MethodName: "UpdateRule",
			Handler:    _Flipt_UpdateRule_Handler,
		},
		{
			MethodName: "DeleteRule",
			Handler:    _Flipt_DeleteRule_Handler,
		},
		{
			MethodName: "CreateDistribution",
			Handler:    _Flipt_CreateDistribution_Handler,
		},
		{
			MethodName: "UpdateDistribution",
			Handler:    _Flipt_UpdateDistribution_Handler,
		},
		{
			MethodName: "DeleteDistribution",
			Handler:    _Flipt_DeleteDistribution_Handler,
		},
		{
			MethodName: "GetSegment",
			Handler:    _Flipt_GetSegment_Handler,
		},
		{
			MethodName: "ListSegments",
			Handler:    _Flipt_ListSegments_Handler,
		},
		{
			MethodName: "CreateSegment",
			Handler:    _Flipt_CreateSegment_Handler,
		},
		{
			MethodName: "UpdateSegment",
			Handler:    _Flipt_UpdateSegment_Handler,
		},
		{
			MethodName: "DeleteSegment",
			Handler:    _Flipt_DeleteSegment_Handler,
		},
		{
			MethodName: "CreateConstraint",
			Handler:    _Flipt_CreateConstraint_Handler,
		},
		{
			MethodName: "UpdateConstraint",
			Handler:    _Flipt_UpdateConstraint_Handler,
		},
		{
			MethodName: "DeleteConstraint",
			Handler:    _Flipt_DeleteConstraint_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "flipt.proto",
}

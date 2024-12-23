// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.0
// 	protoc        (unknown)
// source: analytics/analytics.proto

package analytics

import (
	_ "google.golang.org/genproto/googleapis/api/visibility"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type GetFlagEvaluationsCountRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	NamespaceKey  string                 `protobuf:"bytes,1,opt,name=namespace_key,json=namespaceKey,proto3" json:"namespace_key,omitempty"`
	FlagKey       string                 `protobuf:"bytes,2,opt,name=flag_key,json=flagKey,proto3" json:"flag_key,omitempty"`
	From          string                 `protobuf:"bytes,3,opt,name=from,proto3" json:"from,omitempty"`
	To            string                 `protobuf:"bytes,4,opt,name=to,proto3" json:"to,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetFlagEvaluationsCountRequest) Reset() {
	*x = GetFlagEvaluationsCountRequest{}
	mi := &file_analytics_analytics_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetFlagEvaluationsCountRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetFlagEvaluationsCountRequest) ProtoMessage() {}

func (x *GetFlagEvaluationsCountRequest) ProtoReflect() protoreflect.Message {
	mi := &file_analytics_analytics_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetFlagEvaluationsCountRequest.ProtoReflect.Descriptor instead.
func (*GetFlagEvaluationsCountRequest) Descriptor() ([]byte, []int) {
	return file_analytics_analytics_proto_rawDescGZIP(), []int{0}
}

func (x *GetFlagEvaluationsCountRequest) GetNamespaceKey() string {
	if x != nil {
		return x.NamespaceKey
	}
	return ""
}

func (x *GetFlagEvaluationsCountRequest) GetFlagKey() string {
	if x != nil {
		return x.FlagKey
	}
	return ""
}

func (x *GetFlagEvaluationsCountRequest) GetFrom() string {
	if x != nil {
		return x.From
	}
	return ""
}

func (x *GetFlagEvaluationsCountRequest) GetTo() string {
	if x != nil {
		return x.To
	}
	return ""
}

type GetFlagEvaluationsCountResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Timestamps    []string               `protobuf:"bytes,1,rep,name=timestamps,proto3" json:"timestamps,omitempty"`
	Values        []float32              `protobuf:"fixed32,2,rep,packed,name=values,proto3" json:"values,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetFlagEvaluationsCountResponse) Reset() {
	*x = GetFlagEvaluationsCountResponse{}
	mi := &file_analytics_analytics_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetFlagEvaluationsCountResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetFlagEvaluationsCountResponse) ProtoMessage() {}

func (x *GetFlagEvaluationsCountResponse) ProtoReflect() protoreflect.Message {
	mi := &file_analytics_analytics_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetFlagEvaluationsCountResponse.ProtoReflect.Descriptor instead.
func (*GetFlagEvaluationsCountResponse) Descriptor() ([]byte, []int) {
	return file_analytics_analytics_proto_rawDescGZIP(), []int{1}
}

func (x *GetFlagEvaluationsCountResponse) GetTimestamps() []string {
	if x != nil {
		return x.Timestamps
	}
	return nil
}

func (x *GetFlagEvaluationsCountResponse) GetValues() []float32 {
	if x != nil {
		return x.Values
	}
	return nil
}

var File_analytics_analytics_proto protoreflect.FileDescriptor

var file_analytics_analytics_proto_rawDesc = []byte{
	0x0a, 0x19, 0x61, 0x6e, 0x61, 0x6c, 0x79, 0x74, 0x69, 0x63, 0x73, 0x2f, 0x61, 0x6e, 0x61, 0x6c,
	0x79, 0x74, 0x69, 0x63, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0f, 0x66, 0x6c, 0x69,
	0x70, 0x74, 0x2e, 0x61, 0x6e, 0x61, 0x6c, 0x79, 0x74, 0x69, 0x63, 0x73, 0x1a, 0x1b, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x69, 0x73, 0x69, 0x62, 0x69, 0x6c,
	0x69, 0x74, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x84, 0x01, 0x0a, 0x1e, 0x47, 0x65,
	0x74, 0x46, 0x6c, 0x61, 0x67, 0x45, 0x76, 0x61, 0x6c, 0x75, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x43, 0x6f, 0x75, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x23, 0x0a, 0x0d,
	0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0c, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x4b, 0x65,
	0x79, 0x12, 0x19, 0x0a, 0x08, 0x66, 0x6c, 0x61, 0x67, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x07, 0x66, 0x6c, 0x61, 0x67, 0x4b, 0x65, 0x79, 0x12, 0x12, 0x0a, 0x04,
	0x66, 0x72, 0x6f, 0x6d, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x66, 0x72, 0x6f, 0x6d,
	0x12, 0x0e, 0x0a, 0x02, 0x74, 0x6f, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x74, 0x6f,
	0x22, 0x59, 0x0a, 0x1f, 0x47, 0x65, 0x74, 0x46, 0x6c, 0x61, 0x67, 0x45, 0x76, 0x61, 0x6c, 0x75,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x1e, 0x0a, 0x0a, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70,
	0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0a, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61,
	0x6d, 0x70, 0x73, 0x12, 0x16, 0x0a, 0x06, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x18, 0x02, 0x20,
	0x03, 0x28, 0x02, 0x52, 0x06, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x32, 0xac, 0x01, 0x0a, 0x10,
	0x41, 0x6e, 0x61, 0x6c, 0x79, 0x74, 0x69, 0x63, 0x73, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x12, 0x7e, 0x0a, 0x17, 0x47, 0x65, 0x74, 0x46, 0x6c, 0x61, 0x67, 0x45, 0x76, 0x61, 0x6c, 0x75,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x2f, 0x2e, 0x66, 0x6c,
	0x69, 0x70, 0x74, 0x2e, 0x61, 0x6e, 0x61, 0x6c, 0x79, 0x74, 0x69, 0x63, 0x73, 0x2e, 0x47, 0x65,
	0x74, 0x46, 0x6c, 0x61, 0x67, 0x45, 0x76, 0x61, 0x6c, 0x75, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x43, 0x6f, 0x75, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x30, 0x2e, 0x66,
	0x6c, 0x69, 0x70, 0x74, 0x2e, 0x61, 0x6e, 0x61, 0x6c, 0x79, 0x74, 0x69, 0x63, 0x73, 0x2e, 0x47,
	0x65, 0x74, 0x46, 0x6c, 0x61, 0x67, 0x45, 0x76, 0x61, 0x6c, 0x75, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00,
	0x1a, 0x18, 0xfa, 0xd2, 0xe4, 0x93, 0x02, 0x12, 0x12, 0x10, 0x66, 0x6c, 0x69, 0x70, 0x74, 0x3a,
	0x73, 0x64, 0x6b, 0x3a, 0x69, 0x67, 0x6e, 0x6f, 0x72, 0x65, 0x42, 0x27, 0x5a, 0x25, 0x67, 0x6f,
	0x2e, 0x66, 0x6c, 0x69, 0x70, 0x74, 0x2e, 0x69, 0x6f, 0x2f, 0x66, 0x6c, 0x69, 0x70, 0x74, 0x2f,
	0x72, 0x70, 0x63, 0x2f, 0x66, 0x6c, 0x69, 0x70, 0x74, 0x2f, 0x61, 0x6e, 0x61, 0x6c, 0x79, 0x74,
	0x69, 0x63, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_analytics_analytics_proto_rawDescOnce sync.Once
	file_analytics_analytics_proto_rawDescData = file_analytics_analytics_proto_rawDesc
)

func file_analytics_analytics_proto_rawDescGZIP() []byte {
	file_analytics_analytics_proto_rawDescOnce.Do(func() {
		file_analytics_analytics_proto_rawDescData = protoimpl.X.CompressGZIP(file_analytics_analytics_proto_rawDescData)
	})
	return file_analytics_analytics_proto_rawDescData
}

var file_analytics_analytics_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_analytics_analytics_proto_goTypes = []any{
	(*GetFlagEvaluationsCountRequest)(nil),  // 0: flipt.analytics.GetFlagEvaluationsCountRequest
	(*GetFlagEvaluationsCountResponse)(nil), // 1: flipt.analytics.GetFlagEvaluationsCountResponse
}
var file_analytics_analytics_proto_depIdxs = []int32{
	0, // 0: flipt.analytics.AnalyticsService.GetFlagEvaluationsCount:input_type -> flipt.analytics.GetFlagEvaluationsCountRequest
	1, // 1: flipt.analytics.AnalyticsService.GetFlagEvaluationsCount:output_type -> flipt.analytics.GetFlagEvaluationsCountResponse
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_analytics_analytics_proto_init() }
func file_analytics_analytics_proto_init() {
	if File_analytics_analytics_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_analytics_analytics_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_analytics_analytics_proto_goTypes,
		DependencyIndexes: file_analytics_analytics_proto_depIdxs,
		MessageInfos:      file_analytics_analytics_proto_msgTypes,
	}.Build()
	File_analytics_analytics_proto = out.File
	file_analytics_analytics_proto_rawDesc = nil
	file_analytics_analytics_proto_goTypes = nil
	file_analytics_analytics_proto_depIdxs = nil
}

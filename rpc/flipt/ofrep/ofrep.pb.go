// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.4
// 	protoc        (unknown)
// source: ofrep/ofrep.proto

package ofrep

import (
	_ "github.com/google/gnostic/openapiv3"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	_ "google.golang.org/genproto/googleapis/api/visibility"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	structpb "google.golang.org/protobuf/types/known/structpb"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type EvaluateReason int32

const (
	EvaluateReason_UNKNOWN         EvaluateReason = 0
	EvaluateReason_DISABLED        EvaluateReason = 1
	EvaluateReason_TARGETING_MATCH EvaluateReason = 2
	EvaluateReason_DEFAULT         EvaluateReason = 3
)

// Enum value maps for EvaluateReason.
var (
	EvaluateReason_name = map[int32]string{
		0: "UNKNOWN",
		1: "DISABLED",
		2: "TARGETING_MATCH",
		3: "DEFAULT",
	}
	EvaluateReason_value = map[string]int32{
		"UNKNOWN":         0,
		"DISABLED":        1,
		"TARGETING_MATCH": 2,
		"DEFAULT":         3,
	}
)

func (x EvaluateReason) Enum() *EvaluateReason {
	p := new(EvaluateReason)
	*p = x
	return p
}

func (x EvaluateReason) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (EvaluateReason) Descriptor() protoreflect.EnumDescriptor {
	return file_ofrep_ofrep_proto_enumTypes[0].Descriptor()
}

func (EvaluateReason) Type() protoreflect.EnumType {
	return &file_ofrep_ofrep_proto_enumTypes[0]
}

func (x EvaluateReason) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use EvaluateReason.Descriptor instead.
func (EvaluateReason) EnumDescriptor() ([]byte, []int) {
	return file_ofrep_ofrep_proto_rawDescGZIP(), []int{0}
}

type GetProviderConfigurationRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetProviderConfigurationRequest) Reset() {
	*x = GetProviderConfigurationRequest{}
	mi := &file_ofrep_ofrep_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetProviderConfigurationRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetProviderConfigurationRequest) ProtoMessage() {}

func (x *GetProviderConfigurationRequest) ProtoReflect() protoreflect.Message {
	mi := &file_ofrep_ofrep_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetProviderConfigurationRequest.ProtoReflect.Descriptor instead.
func (*GetProviderConfigurationRequest) Descriptor() ([]byte, []int) {
	return file_ofrep_ofrep_proto_rawDescGZIP(), []int{0}
}

type GetProviderConfigurationResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Name          string                 `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Capabilities  *Capabilities          `protobuf:"bytes,2,opt,name=capabilities,proto3" json:"capabilities,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetProviderConfigurationResponse) Reset() {
	*x = GetProviderConfigurationResponse{}
	mi := &file_ofrep_ofrep_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetProviderConfigurationResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetProviderConfigurationResponse) ProtoMessage() {}

func (x *GetProviderConfigurationResponse) ProtoReflect() protoreflect.Message {
	mi := &file_ofrep_ofrep_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetProviderConfigurationResponse.ProtoReflect.Descriptor instead.
func (*GetProviderConfigurationResponse) Descriptor() ([]byte, []int) {
	return file_ofrep_ofrep_proto_rawDescGZIP(), []int{1}
}

func (x *GetProviderConfigurationResponse) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *GetProviderConfigurationResponse) GetCapabilities() *Capabilities {
	if x != nil {
		return x.Capabilities
	}
	return nil
}

type Capabilities struct {
	state             protoimpl.MessageState `protogen:"open.v1"`
	CacheInvalidation *CacheInvalidation     `protobuf:"bytes,1,opt,name=cache_invalidation,json=cacheInvalidation,proto3" json:"cache_invalidation,omitempty"`
	FlagEvaluation    *FlagEvaluation        `protobuf:"bytes,2,opt,name=flag_evaluation,json=flagEvaluation,proto3" json:"flag_evaluation,omitempty"`
	unknownFields     protoimpl.UnknownFields
	sizeCache         protoimpl.SizeCache
}

func (x *Capabilities) Reset() {
	*x = Capabilities{}
	mi := &file_ofrep_ofrep_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Capabilities) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Capabilities) ProtoMessage() {}

func (x *Capabilities) ProtoReflect() protoreflect.Message {
	mi := &file_ofrep_ofrep_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Capabilities.ProtoReflect.Descriptor instead.
func (*Capabilities) Descriptor() ([]byte, []int) {
	return file_ofrep_ofrep_proto_rawDescGZIP(), []int{2}
}

func (x *Capabilities) GetCacheInvalidation() *CacheInvalidation {
	if x != nil {
		return x.CacheInvalidation
	}
	return nil
}

func (x *Capabilities) GetFlagEvaluation() *FlagEvaluation {
	if x != nil {
		return x.FlagEvaluation
	}
	return nil
}

type CacheInvalidation struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Polling       *Polling               `protobuf:"bytes,1,opt,name=polling,proto3" json:"polling,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CacheInvalidation) Reset() {
	*x = CacheInvalidation{}
	mi := &file_ofrep_ofrep_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CacheInvalidation) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CacheInvalidation) ProtoMessage() {}

func (x *CacheInvalidation) ProtoReflect() protoreflect.Message {
	mi := &file_ofrep_ofrep_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CacheInvalidation.ProtoReflect.Descriptor instead.
func (*CacheInvalidation) Descriptor() ([]byte, []int) {
	return file_ofrep_ofrep_proto_rawDescGZIP(), []int{3}
}

func (x *CacheInvalidation) GetPolling() *Polling {
	if x != nil {
		return x.Polling
	}
	return nil
}

type Polling struct {
	state                protoimpl.MessageState `protogen:"open.v1"`
	Enabled              bool                   `protobuf:"varint,1,opt,name=enabled,proto3" json:"enabled,omitempty"`
	MinPollingIntervalMs uint32                 `protobuf:"varint,2,opt,name=min_polling_interval_ms,json=minPollingIntervalMs,proto3" json:"min_polling_interval_ms,omitempty"`
	unknownFields        protoimpl.UnknownFields
	sizeCache            protoimpl.SizeCache
}

func (x *Polling) Reset() {
	*x = Polling{}
	mi := &file_ofrep_ofrep_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Polling) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Polling) ProtoMessage() {}

func (x *Polling) ProtoReflect() protoreflect.Message {
	mi := &file_ofrep_ofrep_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Polling.ProtoReflect.Descriptor instead.
func (*Polling) Descriptor() ([]byte, []int) {
	return file_ofrep_ofrep_proto_rawDescGZIP(), []int{4}
}

func (x *Polling) GetEnabled() bool {
	if x != nil {
		return x.Enabled
	}
	return false
}

func (x *Polling) GetMinPollingIntervalMs() uint32 {
	if x != nil {
		return x.MinPollingIntervalMs
	}
	return 0
}

type FlagEvaluation struct {
	state          protoimpl.MessageState `protogen:"open.v1"`
	SupportedTypes []string               `protobuf:"bytes,1,rep,name=supported_types,json=supportedTypes,proto3" json:"supported_types,omitempty"`
	unknownFields  protoimpl.UnknownFields
	sizeCache      protoimpl.SizeCache
}

func (x *FlagEvaluation) Reset() {
	*x = FlagEvaluation{}
	mi := &file_ofrep_ofrep_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *FlagEvaluation) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FlagEvaluation) ProtoMessage() {}

func (x *FlagEvaluation) ProtoReflect() protoreflect.Message {
	mi := &file_ofrep_ofrep_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FlagEvaluation.ProtoReflect.Descriptor instead.
func (*FlagEvaluation) Descriptor() ([]byte, []int) {
	return file_ofrep_ofrep_proto_rawDescGZIP(), []int{5}
}

func (x *FlagEvaluation) GetSupportedTypes() []string {
	if x != nil {
		return x.SupportedTypes
	}
	return nil
}

type EvaluateFlagRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Key           string                 `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Context       map[string]string      `protobuf:"bytes,2,rep,name=context,proto3" json:"context,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *EvaluateFlagRequest) Reset() {
	*x = EvaluateFlagRequest{}
	mi := &file_ofrep_ofrep_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *EvaluateFlagRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EvaluateFlagRequest) ProtoMessage() {}

func (x *EvaluateFlagRequest) ProtoReflect() protoreflect.Message {
	mi := &file_ofrep_ofrep_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EvaluateFlagRequest.ProtoReflect.Descriptor instead.
func (*EvaluateFlagRequest) Descriptor() ([]byte, []int) {
	return file_ofrep_ofrep_proto_rawDescGZIP(), []int{6}
}

func (x *EvaluateFlagRequest) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *EvaluateFlagRequest) GetContext() map[string]string {
	if x != nil {
		return x.Context
	}
	return nil
}

type EvaluatedFlag struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Key           string                 `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Reason        EvaluateReason         `protobuf:"varint,2,opt,name=reason,proto3,enum=flipt.ofrep.EvaluateReason" json:"reason,omitempty"`
	Variant       string                 `protobuf:"bytes,3,opt,name=variant,proto3" json:"variant,omitempty"`
	Metadata      *structpb.Struct       `protobuf:"bytes,4,opt,name=metadata,proto3" json:"metadata,omitempty"`
	Value         *structpb.Value        `protobuf:"bytes,5,opt,name=value,proto3" json:"value,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *EvaluatedFlag) Reset() {
	*x = EvaluatedFlag{}
	mi := &file_ofrep_ofrep_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *EvaluatedFlag) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EvaluatedFlag) ProtoMessage() {}

func (x *EvaluatedFlag) ProtoReflect() protoreflect.Message {
	mi := &file_ofrep_ofrep_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EvaluatedFlag.ProtoReflect.Descriptor instead.
func (*EvaluatedFlag) Descriptor() ([]byte, []int) {
	return file_ofrep_ofrep_proto_rawDescGZIP(), []int{7}
}

func (x *EvaluatedFlag) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *EvaluatedFlag) GetReason() EvaluateReason {
	if x != nil {
		return x.Reason
	}
	return EvaluateReason_UNKNOWN
}

func (x *EvaluatedFlag) GetVariant() string {
	if x != nil {
		return x.Variant
	}
	return ""
}

func (x *EvaluatedFlag) GetMetadata() *structpb.Struct {
	if x != nil {
		return x.Metadata
	}
	return nil
}

func (x *EvaluatedFlag) GetValue() *structpb.Value {
	if x != nil {
		return x.Value
	}
	return nil
}

type EvaluateBulkRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Context       map[string]string      `protobuf:"bytes,2,rep,name=context,proto3" json:"context,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *EvaluateBulkRequest) Reset() {
	*x = EvaluateBulkRequest{}
	mi := &file_ofrep_ofrep_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *EvaluateBulkRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EvaluateBulkRequest) ProtoMessage() {}

func (x *EvaluateBulkRequest) ProtoReflect() protoreflect.Message {
	mi := &file_ofrep_ofrep_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EvaluateBulkRequest.ProtoReflect.Descriptor instead.
func (*EvaluateBulkRequest) Descriptor() ([]byte, []int) {
	return file_ofrep_ofrep_proto_rawDescGZIP(), []int{8}
}

func (x *EvaluateBulkRequest) GetContext() map[string]string {
	if x != nil {
		return x.Context
	}
	return nil
}

type BulkEvaluationResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Flags         []*EvaluatedFlag       `protobuf:"bytes,1,rep,name=flags,proto3" json:"flags,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *BulkEvaluationResponse) Reset() {
	*x = BulkEvaluationResponse{}
	mi := &file_ofrep_ofrep_proto_msgTypes[9]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *BulkEvaluationResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BulkEvaluationResponse) ProtoMessage() {}

func (x *BulkEvaluationResponse) ProtoReflect() protoreflect.Message {
	mi := &file_ofrep_ofrep_proto_msgTypes[9]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BulkEvaluationResponse.ProtoReflect.Descriptor instead.
func (*BulkEvaluationResponse) Descriptor() ([]byte, []int) {
	return file_ofrep_ofrep_proto_rawDescGZIP(), []int{9}
}

func (x *BulkEvaluationResponse) GetFlags() []*EvaluatedFlag {
	if x != nil {
		return x.Flags
	}
	return nil
}

var File_ofrep_ofrep_proto protoreflect.FileDescriptor

var file_ofrep_ofrep_proto_rawDesc = string([]byte{
	0x0a, 0x11, 0x6f, 0x66, 0x72, 0x65, 0x70, 0x2f, 0x6f, 0x66, 0x72, 0x65, 0x70, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x0b, 0x66, 0x6c, 0x69, 0x70, 0x74, 0x2e, 0x6f, 0x66, 0x72, 0x65, 0x70,
	0x1a, 0x24, 0x67, 0x6e, 0x6f, 0x73, 0x74, 0x69, 0x63, 0x2f, 0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70,
	0x69, 0x2f, 0x76, 0x33, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61,
	0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69,
	0x2f, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x5f, 0x62, 0x65, 0x68, 0x61, 0x76, 0x69, 0x6f, 0x72, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70,
	0x69, 0x2f, 0x76, 0x69, 0x73, 0x69, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x79, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2f, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x22, 0x21, 0x0a, 0x1f, 0x47, 0x65, 0x74, 0x50, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x43,
	0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x22, 0x75, 0x0a, 0x20, 0x47, 0x65, 0x74, 0x50, 0x72, 0x6f, 0x76, 0x69, 0x64,
	0x65, 0x72, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x3d, 0x0a, 0x0c, 0x63,
	0x61, 0x70, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x69, 0x65, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x19, 0x2e, 0x66, 0x6c, 0x69, 0x70, 0x74, 0x2e, 0x6f, 0x66, 0x72, 0x65, 0x70, 0x2e,
	0x43, 0x61, 0x70, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x69, 0x65, 0x73, 0x52, 0x0c, 0x63, 0x61,
	0x70, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x69, 0x65, 0x73, 0x22, 0xa3, 0x01, 0x0a, 0x0c, 0x43,
	0x61, 0x70, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x69, 0x65, 0x73, 0x12, 0x4d, 0x0a, 0x12, 0x63,
	0x61, 0x63, 0x68, 0x65, 0x5f, 0x69, 0x6e, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1e, 0x2e, 0x66, 0x6c, 0x69, 0x70, 0x74, 0x2e,
	0x6f, 0x66, 0x72, 0x65, 0x70, 0x2e, 0x43, 0x61, 0x63, 0x68, 0x65, 0x49, 0x6e, 0x76, 0x61, 0x6c,
	0x69, 0x64, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x11, 0x63, 0x61, 0x63, 0x68, 0x65, 0x49, 0x6e,
	0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x44, 0x0a, 0x0f, 0x66, 0x6c,
	0x61, 0x67, 0x5f, 0x65, 0x76, 0x61, 0x6c, 0x75, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x66, 0x6c, 0x69, 0x70, 0x74, 0x2e, 0x6f, 0x66, 0x72, 0x65,
	0x70, 0x2e, 0x46, 0x6c, 0x61, 0x67, 0x45, 0x76, 0x61, 0x6c, 0x75, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x52, 0x0e, 0x66, 0x6c, 0x61, 0x67, 0x45, 0x76, 0x61, 0x6c, 0x75, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x22, 0x43, 0x0a, 0x11, 0x43, 0x61, 0x63, 0x68, 0x65, 0x49, 0x6e, 0x76, 0x61, 0x6c, 0x69, 0x64,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x2e, 0x0a, 0x07, 0x70, 0x6f, 0x6c, 0x6c, 0x69, 0x6e, 0x67,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x66, 0x6c, 0x69, 0x70, 0x74, 0x2e, 0x6f,
	0x66, 0x72, 0x65, 0x70, 0x2e, 0x50, 0x6f, 0x6c, 0x6c, 0x69, 0x6e, 0x67, 0x52, 0x07, 0x70, 0x6f,
	0x6c, 0x6c, 0x69, 0x6e, 0x67, 0x22, 0x5a, 0x0a, 0x07, 0x50, 0x6f, 0x6c, 0x6c, 0x69, 0x6e, 0x67,
	0x12, 0x18, 0x0a, 0x07, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x08, 0x52, 0x07, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, 0x12, 0x35, 0x0a, 0x17, 0x6d, 0x69,
	0x6e, 0x5f, 0x70, 0x6f, 0x6c, 0x6c, 0x69, 0x6e, 0x67, 0x5f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x76,
	0x61, 0x6c, 0x5f, 0x6d, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x14, 0x6d, 0x69, 0x6e,
	0x50, 0x6f, 0x6c, 0x6c, 0x69, 0x6e, 0x67, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x76, 0x61, 0x6c, 0x4d,
	0x73, 0x22, 0x39, 0x0a, 0x0e, 0x46, 0x6c, 0x61, 0x67, 0x45, 0x76, 0x61, 0x6c, 0x75, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x12, 0x27, 0x0a, 0x0f, 0x73, 0x75, 0x70, 0x70, 0x6f, 0x72, 0x74, 0x65, 0x64,
	0x5f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0e, 0x73, 0x75,
	0x70, 0x70, 0x6f, 0x72, 0x74, 0x65, 0x64, 0x54, 0x79, 0x70, 0x65, 0x73, 0x22, 0xac, 0x01, 0x0a,
	0x13, 0x45, 0x76, 0x61, 0x6c, 0x75, 0x61, 0x74, 0x65, 0x46, 0x6c, 0x61, 0x67, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x47, 0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x78,
	0x74, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x2d, 0x2e, 0x66, 0x6c, 0x69, 0x70, 0x74, 0x2e,
	0x6f, 0x66, 0x72, 0x65, 0x70, 0x2e, 0x45, 0x76, 0x61, 0x6c, 0x75, 0x61, 0x74, 0x65, 0x46, 0x6c,
	0x61, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78,
	0x74, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x1a,
	0x3a, 0x0a, 0x0c, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12,
	0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65,
	0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0xd3, 0x01, 0x0a, 0x0d,
	0x45, 0x76, 0x61, 0x6c, 0x75, 0x61, 0x74, 0x65, 0x64, 0x46, 0x6c, 0x61, 0x67, 0x12, 0x10, 0x0a,
	0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12,
	0x33, 0x0a, 0x06, 0x72, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32,
	0x1b, 0x2e, 0x66, 0x6c, 0x69, 0x70, 0x74, 0x2e, 0x6f, 0x66, 0x72, 0x65, 0x70, 0x2e, 0x45, 0x76,
	0x61, 0x6c, 0x75, 0x61, 0x74, 0x65, 0x52, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x52, 0x06, 0x72, 0x65,
	0x61, 0x73, 0x6f, 0x6e, 0x12, 0x18, 0x0a, 0x07, 0x76, 0x61, 0x72, 0x69, 0x61, 0x6e, 0x74, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x76, 0x61, 0x72, 0x69, 0x61, 0x6e, 0x74, 0x12, 0x33,
	0x0a, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x17, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2e, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x52, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64,
	0x61, 0x74, 0x61, 0x12, 0x2c, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x05, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2e, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x22, 0x9a, 0x01, 0x0a, 0x13, 0x45, 0x76, 0x61, 0x6c, 0x75, 0x61, 0x74, 0x65, 0x42, 0x75,
	0x6c, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x47, 0x0a, 0x07, 0x63, 0x6f, 0x6e,
	0x74, 0x65, 0x78, 0x74, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x2d, 0x2e, 0x66, 0x6c, 0x69,
	0x70, 0x74, 0x2e, 0x6f, 0x66, 0x72, 0x65, 0x70, 0x2e, 0x45, 0x76, 0x61, 0x6c, 0x75, 0x61, 0x74,
	0x65, 0x42, 0x75, 0x6c, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x43, 0x6f, 0x6e,
	0x74, 0x65, 0x78, 0x74, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65,
	0x78, 0x74, 0x1a, 0x3a, 0x0a, 0x0c, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x45, 0x6e, 0x74,
	0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x4f,
	0x0a, 0x16, 0x42, 0x75, 0x6c, 0x6b, 0x45, 0x76, 0x61, 0x6c, 0x75, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x35, 0x0a, 0x05, 0x66, 0x6c, 0x61, 0x67,
	0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x66, 0x6c, 0x69, 0x70, 0x74, 0x2e,
	0x6f, 0x66, 0x72, 0x65, 0x70, 0x2e, 0x45, 0x76, 0x61, 0x6c, 0x75, 0x61, 0x74, 0x65, 0x64, 0x46,
	0x6c, 0x61, 0x67, 0x42, 0x03, 0xe0, 0x41, 0x02, 0x52, 0x05, 0x66, 0x6c, 0x61, 0x67, 0x73, 0x2a,
	0x4d, 0x0a, 0x0e, 0x45, 0x76, 0x61, 0x6c, 0x75, 0x61, 0x74, 0x65, 0x52, 0x65, 0x61, 0x73, 0x6f,
	0x6e, 0x12, 0x0b, 0x0a, 0x07, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x10, 0x00, 0x12, 0x0c,
	0x0a, 0x08, 0x44, 0x49, 0x53, 0x41, 0x42, 0x4c, 0x45, 0x44, 0x10, 0x01, 0x12, 0x13, 0x0a, 0x0f,
	0x54, 0x41, 0x52, 0x47, 0x45, 0x54, 0x49, 0x4e, 0x47, 0x5f, 0x4d, 0x41, 0x54, 0x43, 0x48, 0x10,
	0x02, 0x12, 0x0b, 0x0a, 0x07, 0x44, 0x45, 0x46, 0x41, 0x55, 0x4c, 0x54, 0x10, 0x03, 0x32, 0x80,
	0x04, 0x0a, 0x0c, 0x4f, 0x46, 0x52, 0x45, 0x50, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12,
	0xb0, 0x01, 0x0a, 0x18, 0x47, 0x65, 0x74, 0x50, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x43,
	0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x2c, 0x2e, 0x66,
	0x6c, 0x69, 0x70, 0x74, 0x2e, 0x6f, 0x66, 0x72, 0x65, 0x70, 0x2e, 0x47, 0x65, 0x74, 0x50, 0x72,
	0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x2d, 0x2e, 0x66, 0x6c, 0x69,
	0x70, 0x74, 0x2e, 0x6f, 0x66, 0x72, 0x65, 0x70, 0x2e, 0x47, 0x65, 0x74, 0x50, 0x72, 0x6f, 0x76,
	0x69, 0x64, 0x65, 0x72, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x37, 0xba, 0x47, 0x15, 0x2a, 0x13,
	0x6f, 0x66, 0x72, 0x65, 0x70, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x19, 0x12, 0x17, 0x2f, 0x6f, 0x66, 0x72, 0x65,
	0x70, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x12, 0x8e, 0x01, 0x0a, 0x0c, 0x45, 0x76, 0x61, 0x6c, 0x75, 0x61, 0x74, 0x65, 0x46,
	0x6c, 0x61, 0x67, 0x12, 0x20, 0x2e, 0x66, 0x6c, 0x69, 0x70, 0x74, 0x2e, 0x6f, 0x66, 0x72, 0x65,
	0x70, 0x2e, 0x45, 0x76, 0x61, 0x6c, 0x75, 0x61, 0x74, 0x65, 0x46, 0x6c, 0x61, 0x67, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1a, 0x2e, 0x66, 0x6c, 0x69, 0x70, 0x74, 0x2e, 0x6f, 0x66,
	0x72, 0x65, 0x70, 0x2e, 0x45, 0x76, 0x61, 0x6c, 0x75, 0x61, 0x74, 0x65, 0x64, 0x46, 0x6c, 0x61,
	0x67, 0x22, 0x40, 0xba, 0x47, 0x14, 0x2a, 0x12, 0x6f, 0x66, 0x72, 0x65, 0x70, 0x2e, 0x65, 0x76,
	0x61, 0x6c, 0x75, 0x61, 0x74, 0x65, 0x46, 0x6c, 0x61, 0x67, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x23,
	0x3a, 0x01, 0x2a, 0x22, 0x1e, 0x2f, 0x6f, 0x66, 0x72, 0x65, 0x70, 0x2f, 0x76, 0x31, 0x2f, 0x65,
	0x76, 0x61, 0x6c, 0x75, 0x61, 0x74, 0x65, 0x2f, 0x66, 0x6c, 0x61, 0x67, 0x73, 0x2f, 0x7b, 0x6b,
	0x65, 0x79, 0x7d, 0x12, 0x91, 0x01, 0x0a, 0x0c, 0x45, 0x76, 0x61, 0x6c, 0x75, 0x61, 0x74, 0x65,
	0x42, 0x75, 0x6c, 0x6b, 0x12, 0x20, 0x2e, 0x66, 0x6c, 0x69, 0x70, 0x74, 0x2e, 0x6f, 0x66, 0x72,
	0x65, 0x70, 0x2e, 0x45, 0x76, 0x61, 0x6c, 0x75, 0x61, 0x74, 0x65, 0x42, 0x75, 0x6c, 0x6b, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x23, 0x2e, 0x66, 0x6c, 0x69, 0x70, 0x74, 0x2e, 0x6f,
	0x66, 0x72, 0x65, 0x70, 0x2e, 0x42, 0x75, 0x6c, 0x6b, 0x45, 0x76, 0x61, 0x6c, 0x75, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x3a, 0xba, 0x47, 0x14,
	0x2a, 0x12, 0x6f, 0x66, 0x72, 0x65, 0x70, 0x2e, 0x65, 0x76, 0x61, 0x6c, 0x75, 0x61, 0x74, 0x65,
	0x42, 0x75, 0x6c, 0x6b, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x1d, 0x3a, 0x01, 0x2a, 0x22, 0x18, 0x2f,
	0x6f, 0x66, 0x72, 0x65, 0x70, 0x2f, 0x76, 0x31, 0x2f, 0x65, 0x76, 0x61, 0x6c, 0x75, 0x61, 0x74,
	0x65, 0x2f, 0x66, 0x6c, 0x61, 0x67, 0x73, 0x1a, 0x18, 0xfa, 0xd2, 0xe4, 0x93, 0x02, 0x12, 0x12,
	0x10, 0x66, 0x6c, 0x69, 0x70, 0x74, 0x3a, 0x73, 0x64, 0x6b, 0x3a, 0x69, 0x67, 0x6e, 0x6f, 0x72,
	0x65, 0x42, 0x23, 0x5a, 0x21, 0x67, 0x6f, 0x2e, 0x66, 0x6c, 0x69, 0x70, 0x74, 0x2e, 0x69, 0x6f,
	0x2f, 0x66, 0x6c, 0x69, 0x70, 0x74, 0x2f, 0x72, 0x70, 0x63, 0x2f, 0x66, 0x6c, 0x69, 0x70, 0x74,
	0x2f, 0x6f, 0x66, 0x72, 0x65, 0x70, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
})

var (
	file_ofrep_ofrep_proto_rawDescOnce sync.Once
	file_ofrep_ofrep_proto_rawDescData []byte
)

func file_ofrep_ofrep_proto_rawDescGZIP() []byte {
	file_ofrep_ofrep_proto_rawDescOnce.Do(func() {
		file_ofrep_ofrep_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_ofrep_ofrep_proto_rawDesc), len(file_ofrep_ofrep_proto_rawDesc)))
	})
	return file_ofrep_ofrep_proto_rawDescData
}

var file_ofrep_ofrep_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_ofrep_ofrep_proto_msgTypes = make([]protoimpl.MessageInfo, 12)
var file_ofrep_ofrep_proto_goTypes = []any{
	(EvaluateReason)(0),                      // 0: flipt.ofrep.EvaluateReason
	(*GetProviderConfigurationRequest)(nil),  // 1: flipt.ofrep.GetProviderConfigurationRequest
	(*GetProviderConfigurationResponse)(nil), // 2: flipt.ofrep.GetProviderConfigurationResponse
	(*Capabilities)(nil),                     // 3: flipt.ofrep.Capabilities
	(*CacheInvalidation)(nil),                // 4: flipt.ofrep.CacheInvalidation
	(*Polling)(nil),                          // 5: flipt.ofrep.Polling
	(*FlagEvaluation)(nil),                   // 6: flipt.ofrep.FlagEvaluation
	(*EvaluateFlagRequest)(nil),              // 7: flipt.ofrep.EvaluateFlagRequest
	(*EvaluatedFlag)(nil),                    // 8: flipt.ofrep.EvaluatedFlag
	(*EvaluateBulkRequest)(nil),              // 9: flipt.ofrep.EvaluateBulkRequest
	(*BulkEvaluationResponse)(nil),           // 10: flipt.ofrep.BulkEvaluationResponse
	nil,                                      // 11: flipt.ofrep.EvaluateFlagRequest.ContextEntry
	nil,                                      // 12: flipt.ofrep.EvaluateBulkRequest.ContextEntry
	(*structpb.Struct)(nil),                  // 13: google.protobuf.Struct
	(*structpb.Value)(nil),                   // 14: google.protobuf.Value
}
var file_ofrep_ofrep_proto_depIdxs = []int32{
	3,  // 0: flipt.ofrep.GetProviderConfigurationResponse.capabilities:type_name -> flipt.ofrep.Capabilities
	4,  // 1: flipt.ofrep.Capabilities.cache_invalidation:type_name -> flipt.ofrep.CacheInvalidation
	6,  // 2: flipt.ofrep.Capabilities.flag_evaluation:type_name -> flipt.ofrep.FlagEvaluation
	5,  // 3: flipt.ofrep.CacheInvalidation.polling:type_name -> flipt.ofrep.Polling
	11, // 4: flipt.ofrep.EvaluateFlagRequest.context:type_name -> flipt.ofrep.EvaluateFlagRequest.ContextEntry
	0,  // 5: flipt.ofrep.EvaluatedFlag.reason:type_name -> flipt.ofrep.EvaluateReason
	13, // 6: flipt.ofrep.EvaluatedFlag.metadata:type_name -> google.protobuf.Struct
	14, // 7: flipt.ofrep.EvaluatedFlag.value:type_name -> google.protobuf.Value
	12, // 8: flipt.ofrep.EvaluateBulkRequest.context:type_name -> flipt.ofrep.EvaluateBulkRequest.ContextEntry
	8,  // 9: flipt.ofrep.BulkEvaluationResponse.flags:type_name -> flipt.ofrep.EvaluatedFlag
	1,  // 10: flipt.ofrep.OFREPService.GetProviderConfiguration:input_type -> flipt.ofrep.GetProviderConfigurationRequest
	7,  // 11: flipt.ofrep.OFREPService.EvaluateFlag:input_type -> flipt.ofrep.EvaluateFlagRequest
	9,  // 12: flipt.ofrep.OFREPService.EvaluateBulk:input_type -> flipt.ofrep.EvaluateBulkRequest
	2,  // 13: flipt.ofrep.OFREPService.GetProviderConfiguration:output_type -> flipt.ofrep.GetProviderConfigurationResponse
	8,  // 14: flipt.ofrep.OFREPService.EvaluateFlag:output_type -> flipt.ofrep.EvaluatedFlag
	10, // 15: flipt.ofrep.OFREPService.EvaluateBulk:output_type -> flipt.ofrep.BulkEvaluationResponse
	13, // [13:16] is the sub-list for method output_type
	10, // [10:13] is the sub-list for method input_type
	10, // [10:10] is the sub-list for extension type_name
	10, // [10:10] is the sub-list for extension extendee
	0,  // [0:10] is the sub-list for field type_name
}

func init() { file_ofrep_ofrep_proto_init() }
func file_ofrep_ofrep_proto_init() {
	if File_ofrep_ofrep_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_ofrep_ofrep_proto_rawDesc), len(file_ofrep_ofrep_proto_rawDesc)),
			NumEnums:      1,
			NumMessages:   12,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_ofrep_ofrep_proto_goTypes,
		DependencyIndexes: file_ofrep_ofrep_proto_depIdxs,
		EnumInfos:         file_ofrep_ofrep_proto_enumTypes,
		MessageInfos:      file_ofrep_ofrep_proto_msgTypes,
	}.Build()
	File_ofrep_ofrep_proto = out.File
	file_ofrep_ofrep_proto_goTypes = nil
	file_ofrep_ofrep_proto_depIdxs = nil
}

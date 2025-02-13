// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.4
// 	protoc        (unknown)
// source: flipt.proto

package flipt

import (
	_ "github.com/google/gnostic/openapiv3"
	_ "go.flipt.io/flipt/rpc/flipt/options"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	structpb "google.golang.org/protobuf/types/known/structpb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
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

type FlagType int32

const (
	FlagType_VARIANT_FLAG_TYPE FlagType = 0
	FlagType_BOOLEAN_FLAG_TYPE FlagType = 1
)

// Enum value maps for FlagType.
var (
	FlagType_name = map[int32]string{
		0: "VARIANT_FLAG_TYPE",
		1: "BOOLEAN_FLAG_TYPE",
	}
	FlagType_value = map[string]int32{
		"VARIANT_FLAG_TYPE": 0,
		"BOOLEAN_FLAG_TYPE": 1,
	}
)

func (x FlagType) Enum() *FlagType {
	p := new(FlagType)
	*p = x
	return p
}

func (x FlagType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (FlagType) Descriptor() protoreflect.EnumDescriptor {
	return file_flipt_proto_enumTypes[0].Descriptor()
}

func (FlagType) Type() protoreflect.EnumType {
	return &file_flipt_proto_enumTypes[0]
}

func (x FlagType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use FlagType.Descriptor instead.
func (FlagType) EnumDescriptor() ([]byte, []int) {
	return file_flipt_proto_rawDescGZIP(), []int{0}
}

type Flag struct {
	state          protoimpl.MessageState `protogen:"open.v1"`
	Key            string                 `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Name           string                 `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Description    string                 `protobuf:"bytes,3,opt,name=description,proto3" json:"description,omitempty"`
	Enabled        bool                   `protobuf:"varint,4,opt,name=enabled,proto3" json:"enabled,omitempty"`
	CreatedAt      *timestamppb.Timestamp `protobuf:"bytes,5,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	UpdatedAt      *timestamppb.Timestamp `protobuf:"bytes,6,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"`
	Variants       []*Variant             `protobuf:"bytes,7,rep,name=variants,proto3" json:"variants,omitempty"`
	NamespaceKey   string                 `protobuf:"bytes,8,opt,name=namespace_key,json=namespaceKey,proto3" json:"namespace_key,omitempty"`
	Type           FlagType               `protobuf:"varint,9,opt,name=type,proto3,enum=flipt.FlagType" json:"type,omitempty"`
	DefaultVariant *Variant               `protobuf:"bytes,10,opt,name=default_variant,json=defaultVariant,proto3,oneof" json:"default_variant,omitempty"`
	Metadata       *structpb.Struct       `protobuf:"bytes,11,opt,name=metadata,proto3,oneof" json:"metadata,omitempty"`
	unknownFields  protoimpl.UnknownFields
	sizeCache      protoimpl.SizeCache
}

func (x *Flag) Reset() {
	*x = Flag{}
	mi := &file_flipt_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Flag) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Flag) ProtoMessage() {}

func (x *Flag) ProtoReflect() protoreflect.Message {
	mi := &file_flipt_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Flag.ProtoReflect.Descriptor instead.
func (*Flag) Descriptor() ([]byte, []int) {
	return file_flipt_proto_rawDescGZIP(), []int{0}
}

func (x *Flag) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *Flag) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Flag) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *Flag) GetEnabled() bool {
	if x != nil {
		return x.Enabled
	}
	return false
}

func (x *Flag) GetCreatedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.CreatedAt
	}
	return nil
}

func (x *Flag) GetUpdatedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.UpdatedAt
	}
	return nil
}

func (x *Flag) GetVariants() []*Variant {
	if x != nil {
		return x.Variants
	}
	return nil
}

func (x *Flag) GetNamespaceKey() string {
	if x != nil {
		return x.NamespaceKey
	}
	return ""
}

func (x *Flag) GetType() FlagType {
	if x != nil {
		return x.Type
	}
	return FlagType_VARIANT_FLAG_TYPE
}

func (x *Flag) GetDefaultVariant() *Variant {
	if x != nil {
		return x.DefaultVariant
	}
	return nil
}

func (x *Flag) GetMetadata() *structpb.Struct {
	if x != nil {
		return x.Metadata
	}
	return nil
}

type FlagList struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Flags         []*Flag                `protobuf:"bytes,1,rep,name=flags,proto3" json:"flags,omitempty"`
	NextPageToken string                 `protobuf:"bytes,2,opt,name=next_page_token,json=nextPageToken,proto3" json:"next_page_token,omitempty"`
	TotalCount    int32                  `protobuf:"varint,3,opt,name=total_count,json=totalCount,proto3" json:"total_count,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *FlagList) Reset() {
	*x = FlagList{}
	mi := &file_flipt_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *FlagList) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FlagList) ProtoMessage() {}

func (x *FlagList) ProtoReflect() protoreflect.Message {
	mi := &file_flipt_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FlagList.ProtoReflect.Descriptor instead.
func (*FlagList) Descriptor() ([]byte, []int) {
	return file_flipt_proto_rawDescGZIP(), []int{1}
}

func (x *FlagList) GetFlags() []*Flag {
	if x != nil {
		return x.Flags
	}
	return nil
}

func (x *FlagList) GetNextPageToken() string {
	if x != nil {
		return x.NextPageToken
	}
	return ""
}

func (x *FlagList) GetTotalCount() int32 {
	if x != nil {
		return x.TotalCount
	}
	return 0
}

type ListFlagRequest struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	Limit int32                  `protobuf:"varint,1,opt,name=limit,proto3" json:"limit,omitempty"`
	// Deprecated: Marked as deprecated in flipt.proto.
	Offset        int32  `protobuf:"varint,2,opt,name=offset,proto3" json:"offset,omitempty"`
	PageToken     string `protobuf:"bytes,3,opt,name=page_token,json=pageToken,proto3" json:"page_token,omitempty"`
	NamespaceKey  string `protobuf:"bytes,4,opt,name=namespace_key,json=namespaceKey,proto3" json:"namespace_key,omitempty"`
	Reference     string `protobuf:"bytes,5,opt,name=reference,proto3" json:"reference,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ListFlagRequest) Reset() {
	*x = ListFlagRequest{}
	mi := &file_flipt_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ListFlagRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListFlagRequest) ProtoMessage() {}

func (x *ListFlagRequest) ProtoReflect() protoreflect.Message {
	mi := &file_flipt_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListFlagRequest.ProtoReflect.Descriptor instead.
func (*ListFlagRequest) Descriptor() ([]byte, []int) {
	return file_flipt_proto_rawDescGZIP(), []int{2}
}

func (x *ListFlagRequest) GetLimit() int32 {
	if x != nil {
		return x.Limit
	}
	return 0
}

// Deprecated: Marked as deprecated in flipt.proto.
func (x *ListFlagRequest) GetOffset() int32 {
	if x != nil {
		return x.Offset
	}
	return 0
}

func (x *ListFlagRequest) GetPageToken() string {
	if x != nil {
		return x.PageToken
	}
	return ""
}

func (x *ListFlagRequest) GetNamespaceKey() string {
	if x != nil {
		return x.NamespaceKey
	}
	return ""
}

func (x *ListFlagRequest) GetReference() string {
	if x != nil {
		return x.Reference
	}
	return ""
}

type Variant struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	FlagKey       string                 `protobuf:"bytes,2,opt,name=flag_key,json=flagKey,proto3" json:"flag_key,omitempty"`
	Key           string                 `protobuf:"bytes,3,opt,name=key,proto3" json:"key,omitempty"`
	Name          string                 `protobuf:"bytes,4,opt,name=name,proto3" json:"name,omitempty"`
	Description   string                 `protobuf:"bytes,5,opt,name=description,proto3" json:"description,omitempty"`
	CreatedAt     *timestamppb.Timestamp `protobuf:"bytes,6,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	UpdatedAt     *timestamppb.Timestamp `protobuf:"bytes,7,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"`
	Attachment    string                 `protobuf:"bytes,8,opt,name=attachment,proto3" json:"attachment,omitempty"`
	NamespaceKey  string                 `protobuf:"bytes,9,opt,name=namespace_key,json=namespaceKey,proto3" json:"namespace_key,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Variant) Reset() {
	*x = Variant{}
	mi := &file_flipt_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Variant) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Variant) ProtoMessage() {}

func (x *Variant) ProtoReflect() protoreflect.Message {
	mi := &file_flipt_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Variant.ProtoReflect.Descriptor instead.
func (*Variant) Descriptor() ([]byte, []int) {
	return file_flipt_proto_rawDescGZIP(), []int{3}
}

func (x *Variant) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Variant) GetFlagKey() string {
	if x != nil {
		return x.FlagKey
	}
	return ""
}

func (x *Variant) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *Variant) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Variant) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *Variant) GetCreatedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.CreatedAt
	}
	return nil
}

func (x *Variant) GetUpdatedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.UpdatedAt
	}
	return nil
}

func (x *Variant) GetAttachment() string {
	if x != nil {
		return x.Attachment
	}
	return ""
}

func (x *Variant) GetNamespaceKey() string {
	if x != nil {
		return x.NamespaceKey
	}
	return ""
}

var File_flipt_proto protoreflect.FileDescriptor

var file_flipt_proto_rawDesc = string([]byte{
	0x0a, 0x0b, 0x66, 0x6c, 0x69, 0x70, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05, 0x66,
	0x6c, 0x69, 0x70, 0x74, 0x1a, 0x24, 0x67, 0x6e, 0x6f, 0x73, 0x74, 0x69, 0x63, 0x2f, 0x6f, 0x70,
	0x65, 0x6e, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x33, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d,
	0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x15, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x2f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xed,
	0x03, 0x0a, 0x04, 0x46, 0x6c, 0x61, 0x67, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x20, 0x0a,
	0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x12,
	0x18, 0x0a, 0x07, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x08,
	0x52, 0x07, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, 0x12, 0x39, 0x0a, 0x0a, 0x63, 0x72, 0x65,
	0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e,
	0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x09, 0x63, 0x72, 0x65, 0x61, 0x74,
	0x65, 0x64, 0x41, 0x74, 0x12, 0x39, 0x0a, 0x0a, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x5f,
	0x61, 0x74, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73,
	0x74, 0x61, 0x6d, 0x70, 0x52, 0x09, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x12,
	0x2a, 0x0a, 0x08, 0x76, 0x61, 0x72, 0x69, 0x61, 0x6e, 0x74, 0x73, 0x18, 0x07, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x0e, 0x2e, 0x66, 0x6c, 0x69, 0x70, 0x74, 0x2e, 0x56, 0x61, 0x72, 0x69, 0x61, 0x6e,
	0x74, 0x52, 0x08, 0x76, 0x61, 0x72, 0x69, 0x61, 0x6e, 0x74, 0x73, 0x12, 0x23, 0x0a, 0x0d, 0x6e,
	0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x08, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0c, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x4b, 0x65, 0x79,
	0x12, 0x23, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x09, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0f,
	0x2e, 0x66, 0x6c, 0x69, 0x70, 0x74, 0x2e, 0x46, 0x6c, 0x61, 0x67, 0x54, 0x79, 0x70, 0x65, 0x52,
	0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x3c, 0x0a, 0x0f, 0x64, 0x65, 0x66, 0x61, 0x75, 0x6c, 0x74,
	0x5f, 0x76, 0x61, 0x72, 0x69, 0x61, 0x6e, 0x74, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e,
	0x2e, 0x66, 0x6c, 0x69, 0x70, 0x74, 0x2e, 0x56, 0x61, 0x72, 0x69, 0x61, 0x6e, 0x74, 0x48, 0x00,
	0x52, 0x0e, 0x64, 0x65, 0x66, 0x61, 0x75, 0x6c, 0x74, 0x56, 0x61, 0x72, 0x69, 0x61, 0x6e, 0x74,
	0x88, 0x01, 0x01, 0x12, 0x38, 0x0a, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x18,
	0x0b, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x48, 0x01,
	0x52, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x88, 0x01, 0x01, 0x42, 0x12, 0x0a,
	0x10, 0x5f, 0x64, 0x65, 0x66, 0x61, 0x75, 0x6c, 0x74, 0x5f, 0x76, 0x61, 0x72, 0x69, 0x61, 0x6e,
	0x74, 0x42, 0x0b, 0x0a, 0x09, 0x5f, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x22, 0x76,
	0x0a, 0x08, 0x46, 0x6c, 0x61, 0x67, 0x4c, 0x69, 0x73, 0x74, 0x12, 0x21, 0x0a, 0x05, 0x66, 0x6c,
	0x61, 0x67, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0b, 0x2e, 0x66, 0x6c, 0x69, 0x70,
	0x74, 0x2e, 0x46, 0x6c, 0x61, 0x67, 0x52, 0x05, 0x66, 0x6c, 0x61, 0x67, 0x73, 0x12, 0x26, 0x0a,
	0x0f, 0x6e, 0x65, 0x78, 0x74, 0x5f, 0x70, 0x61, 0x67, 0x65, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x6e, 0x65, 0x78, 0x74, 0x50, 0x61, 0x67, 0x65,
	0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x12, 0x1f, 0x0a, 0x0b, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x5f, 0x63,
	0x6f, 0x75, 0x6e, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0a, 0x74, 0x6f, 0x74, 0x61,
	0x6c, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x22, 0xa5, 0x01, 0x0a, 0x0f, 0x4c, 0x69, 0x73, 0x74, 0x46,
	0x6c, 0x61, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x6c, 0x69,
	0x6d, 0x69, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x6c, 0x69, 0x6d, 0x69, 0x74,
	0x12, 0x1a, 0x0a, 0x06, 0x6f, 0x66, 0x66, 0x73, 0x65, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05,
	0x42, 0x02, 0x18, 0x01, 0x52, 0x06, 0x6f, 0x66, 0x66, 0x73, 0x65, 0x74, 0x12, 0x1d, 0x0a, 0x0a,
	0x70, 0x61, 0x67, 0x65, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x09, 0x70, 0x61, 0x67, 0x65, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x12, 0x23, 0x0a, 0x0d, 0x6e,
	0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0c, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x4b, 0x65, 0x79,
	0x12, 0x1c, 0x0a, 0x09, 0x72, 0x65, 0x66, 0x65, 0x72, 0x65, 0x6e, 0x63, 0x65, 0x18, 0x05, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x09, 0x72, 0x65, 0x66, 0x65, 0x72, 0x65, 0x6e, 0x63, 0x65, 0x22, 0xb7,
	0x02, 0x0a, 0x07, 0x56, 0x61, 0x72, 0x69, 0x61, 0x6e, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x19, 0x0a, 0x08, 0x66, 0x6c,
	0x61, 0x67, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x66, 0x6c,
	0x61, 0x67, 0x4b, 0x65, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x20, 0x0a, 0x0b, 0x64,
	0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x39, 0x0a,
	0x0a, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x06, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x09, 0x63,
	0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x12, 0x39, 0x0a, 0x0a, 0x75, 0x70, 0x64, 0x61,
	0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54,
	0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x09, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65,
	0x64, 0x41, 0x74, 0x12, 0x1e, 0x0a, 0x0a, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e,
	0x74, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d,
	0x65, 0x6e, 0x74, 0x12, 0x23, 0x0a, 0x0d, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65,
	0x5f, 0x6b, 0x65, 0x79, 0x18, 0x09, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x6e, 0x61, 0x6d, 0x65,
	0x73, 0x70, 0x61, 0x63, 0x65, 0x4b, 0x65, 0x79, 0x2a, 0x38, 0x0a, 0x08, 0x46, 0x6c, 0x61, 0x67,
	0x54, 0x79, 0x70, 0x65, 0x12, 0x15, 0x0a, 0x11, 0x56, 0x41, 0x52, 0x49, 0x41, 0x4e, 0x54, 0x5f,
	0x46, 0x4c, 0x41, 0x47, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x10, 0x00, 0x12, 0x15, 0x0a, 0x11, 0x42,
	0x4f, 0x4f, 0x4c, 0x45, 0x41, 0x4e, 0x5f, 0x46, 0x4c, 0x41, 0x47, 0x5f, 0x54, 0x59, 0x50, 0x45,
	0x10, 0x01, 0x32, 0xac, 0x01, 0x0a, 0x05, 0x46, 0x6c, 0x69, 0x70, 0x74, 0x12, 0xa2, 0x01, 0x0a,
	0x09, 0x4c, 0x69, 0x73, 0x74, 0x46, 0x6c, 0x61, 0x67, 0x73, 0x12, 0x16, 0x2e, 0x66, 0x6c, 0x69,
	0x70, 0x74, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x46, 0x6c, 0x61, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x0f, 0x2e, 0x66, 0x6c, 0x69, 0x70, 0x74, 0x2e, 0x46, 0x6c, 0x61, 0x67, 0x4c,
	0x69, 0x73, 0x74, 0x22, 0x6c, 0xba, 0x47, 0x19, 0x0a, 0x0c, 0x46, 0x6c, 0x61, 0x67, 0x73, 0x53,
	0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2a, 0x09, 0x6c, 0x69, 0x73, 0x74, 0x46, 0x6c, 0x61, 0x67,
	0x73, 0x8a, 0xb3, 0x97, 0x8b, 0x01, 0x1a, 0x0a, 0x18, 0x0a, 0x0d, 0x6e, 0x61, 0x6d, 0x65, 0x73,
	0x70, 0x61, 0x63, 0x65, 0x5f, 0x6b, 0x65, 0x79, 0x12, 0x07, 0x64, 0x65, 0x66, 0x61, 0x75, 0x6c,
	0x74, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x2a, 0x12, 0x28, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31,
	0x2f, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x73, 0x2f, 0x7b, 0x6e, 0x61, 0x6d,
	0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x5f, 0x6b, 0x65, 0x79, 0x7d, 0x2f, 0x66, 0x6c, 0x61, 0x67,
	0x73, 0x42, 0x1d, 0x5a, 0x1b, 0x67, 0x6f, 0x2e, 0x66, 0x6c, 0x69, 0x70, 0x74, 0x2e, 0x69, 0x6f,
	0x2f, 0x66, 0x6c, 0x69, 0x70, 0x74, 0x2f, 0x72, 0x70, 0x63, 0x2f, 0x66, 0x6c, 0x69, 0x70, 0x74,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
})

var (
	file_flipt_proto_rawDescOnce sync.Once
	file_flipt_proto_rawDescData []byte
)

func file_flipt_proto_rawDescGZIP() []byte {
	file_flipt_proto_rawDescOnce.Do(func() {
		file_flipt_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_flipt_proto_rawDesc), len(file_flipt_proto_rawDesc)))
	})
	return file_flipt_proto_rawDescData
}

var file_flipt_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_flipt_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_flipt_proto_goTypes = []any{
	(FlagType)(0),                 // 0: flipt.FlagType
	(*Flag)(nil),                  // 1: flipt.Flag
	(*FlagList)(nil),              // 2: flipt.FlagList
	(*ListFlagRequest)(nil),       // 3: flipt.ListFlagRequest
	(*Variant)(nil),               // 4: flipt.Variant
	(*timestamppb.Timestamp)(nil), // 5: google.protobuf.Timestamp
	(*structpb.Struct)(nil),       // 6: google.protobuf.Struct
}
var file_flipt_proto_depIdxs = []int32{
	5,  // 0: flipt.Flag.created_at:type_name -> google.protobuf.Timestamp
	5,  // 1: flipt.Flag.updated_at:type_name -> google.protobuf.Timestamp
	4,  // 2: flipt.Flag.variants:type_name -> flipt.Variant
	0,  // 3: flipt.Flag.type:type_name -> flipt.FlagType
	4,  // 4: flipt.Flag.default_variant:type_name -> flipt.Variant
	6,  // 5: flipt.Flag.metadata:type_name -> google.protobuf.Struct
	1,  // 6: flipt.FlagList.flags:type_name -> flipt.Flag
	5,  // 7: flipt.Variant.created_at:type_name -> google.protobuf.Timestamp
	5,  // 8: flipt.Variant.updated_at:type_name -> google.protobuf.Timestamp
	3,  // 9: flipt.Flipt.ListFlags:input_type -> flipt.ListFlagRequest
	2,  // 10: flipt.Flipt.ListFlags:output_type -> flipt.FlagList
	10, // [10:11] is the sub-list for method output_type
	9,  // [9:10] is the sub-list for method input_type
	9,  // [9:9] is the sub-list for extension type_name
	9,  // [9:9] is the sub-list for extension extendee
	0,  // [0:9] is the sub-list for field type_name
}

func init() { file_flipt_proto_init() }
func file_flipt_proto_init() {
	if File_flipt_proto != nil {
		return
	}
	file_flipt_proto_msgTypes[0].OneofWrappers = []any{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_flipt_proto_rawDesc), len(file_flipt_proto_rawDesc)),
			NumEnums:      1,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_flipt_proto_goTypes,
		DependencyIndexes: file_flipt_proto_depIdxs,
		EnumInfos:         file_flipt_proto_enumTypes,
		MessageInfos:      file_flipt_proto_msgTypes,
	}.Build()
	File_flipt_proto = out.File
	file_flipt_proto_goTypes = nil
	file_flipt_proto_depIdxs = nil
}

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        (unknown)
// source: meta/meta.proto

package meta

import (
	httpbody "google.golang.org/genproto/googleapis/api/httpbody"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	reflect "reflect"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

var File_meta_meta_proto protoreflect.FileDescriptor

const file_meta_meta_proto_rawDesc = "" +
	"\n" +
	"\x0fmeta/meta.proto\x12\n" +
	"flipt.meta\x1a\x19google/api/httpbody.proto\x1a\x1bgoogle/protobuf/empty.proto2\x90\x01\n" +
	"\x0fMetadataService\x12B\n" +
	"\x10GetConfiguration\x12\x16.google.protobuf.Empty\x1a\x14.google.api.HttpBody\"\x00\x129\n" +
	"\aGetInfo\x12\x16.google.protobuf.Empty\x1a\x14.google.api.HttpBody\"\x00B\"Z go.flipt.io/flipt/rpc/flipt/metab\x06proto3"

var file_meta_meta_proto_goTypes = []any{
	(*emptypb.Empty)(nil),     // 0: google.protobuf.Empty
	(*httpbody.HttpBody)(nil), // 1: google.api.HttpBody
}
var file_meta_meta_proto_depIdxs = []int32{
	0, // 0: flipt.meta.MetadataService.GetConfiguration:input_type -> google.protobuf.Empty
	0, // 1: flipt.meta.MetadataService.GetInfo:input_type -> google.protobuf.Empty
	1, // 2: flipt.meta.MetadataService.GetConfiguration:output_type -> google.api.HttpBody
	1, // 3: flipt.meta.MetadataService.GetInfo:output_type -> google.api.HttpBody
	2, // [2:4] is the sub-list for method output_type
	0, // [0:2] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_meta_meta_proto_init() }
func file_meta_meta_proto_init() {
	if File_meta_meta_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_meta_meta_proto_rawDesc), len(file_meta_meta_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_meta_meta_proto_goTypes,
		DependencyIndexes: file_meta_meta_proto_depIdxs,
	}.Build()
	File_meta_meta_proto = out.File
	file_meta_meta_proto_goTypes = nil
	file_meta_meta_proto_depIdxs = nil
}

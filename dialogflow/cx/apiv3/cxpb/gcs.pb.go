// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.21.12
// source: google/cloud/dialogflow/cx/v3/gcs.proto

package cxpb

import (
	reflect "reflect"
	sync "sync"

	_ "google.golang.org/genproto/googleapis/api/annotations"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Google Cloud Storage location for a Dialogflow operation that writes or
// exports objects (e.g. exported agent or transcripts) outside of Dialogflow.
type GcsDestination struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Required. The Google Cloud Storage URI for the exported objects. A URI is
	// of the form:
	//   gs://bucket/object-name-or-prefix
	// Whether a full object name, or just a prefix, its usage depends on the
	// Dialogflow operation.
	Uri string `protobuf:"bytes,1,opt,name=uri,proto3" json:"uri,omitempty"`
}

func (x *GcsDestination) Reset() {
	*x = GcsDestination{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_cloud_dialogflow_cx_v3_gcs_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GcsDestination) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GcsDestination) ProtoMessage() {}

func (x *GcsDestination) ProtoReflect() protoreflect.Message {
	mi := &file_google_cloud_dialogflow_cx_v3_gcs_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GcsDestination.ProtoReflect.Descriptor instead.
func (*GcsDestination) Descriptor() ([]byte, []int) {
	return file_google_cloud_dialogflow_cx_v3_gcs_proto_rawDescGZIP(), []int{0}
}

func (x *GcsDestination) GetUri() string {
	if x != nil {
		return x.Uri
	}
	return ""
}

var File_google_cloud_dialogflow_cx_v3_gcs_proto protoreflect.FileDescriptor

var file_google_cloud_dialogflow_cx_v3_gcs_proto_rawDesc = []byte{
	0x0a, 0x27, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2f, 0x64,
	0x69, 0x61, 0x6c, 0x6f, 0x67, 0x66, 0x6c, 0x6f, 0x77, 0x2f, 0x63, 0x78, 0x2f, 0x76, 0x33, 0x2f,
	0x67, 0x63, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x1d, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x64, 0x69, 0x61, 0x6c, 0x6f, 0x67, 0x66, 0x6c,
	0x6f, 0x77, 0x2e, 0x63, 0x78, 0x2e, 0x76, 0x33, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2f, 0x61, 0x70, 0x69, 0x2f, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x5f, 0x62, 0x65, 0x68, 0x61, 0x76,
	0x69, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x27, 0x0a, 0x0e, 0x47, 0x63, 0x73,
	0x44, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x15, 0x0a, 0x03, 0x75,
	0x72, 0x69, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x03, 0xe0, 0x41, 0x02, 0x52, 0x03, 0x75,
	0x72, 0x69, 0x42, 0xae, 0x01, 0x0a, 0x21, 0x63, 0x6f, 0x6d, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x64, 0x69, 0x61, 0x6c, 0x6f, 0x67, 0x66, 0x6c,
	0x6f, 0x77, 0x2e, 0x63, 0x78, 0x2e, 0x76, 0x33, 0x42, 0x08, 0x47, 0x63, 0x73, 0x50, 0x72, 0x6f,
	0x74, 0x6f, 0x50, 0x01, 0x5a, 0x31, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x67, 0x6f, 0x2f, 0x64, 0x69, 0x61, 0x6c, 0x6f, 0x67,
	0x66, 0x6c, 0x6f, 0x77, 0x2f, 0x63, 0x78, 0x2f, 0x61, 0x70, 0x69, 0x76, 0x33, 0x2f, 0x63, 0x78,
	0x70, 0x62, 0x3b, 0x63, 0x78, 0x70, 0x62, 0xf8, 0x01, 0x01, 0xa2, 0x02, 0x02, 0x44, 0x46, 0xaa,
	0x02, 0x1d, 0x47, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x43, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x44,
	0x69, 0x61, 0x6c, 0x6f, 0x67, 0x66, 0x6c, 0x6f, 0x77, 0x2e, 0x43, 0x78, 0x2e, 0x56, 0x33, 0xea,
	0x02, 0x21, 0x47, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x3a, 0x3a, 0x43, 0x6c, 0x6f, 0x75, 0x64, 0x3a,
	0x3a, 0x44, 0x69, 0x61, 0x6c, 0x6f, 0x67, 0x66, 0x6c, 0x6f, 0x77, 0x3a, 0x3a, 0x43, 0x58, 0x3a,
	0x3a, 0x56, 0x33, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_google_cloud_dialogflow_cx_v3_gcs_proto_rawDescOnce sync.Once
	file_google_cloud_dialogflow_cx_v3_gcs_proto_rawDescData = file_google_cloud_dialogflow_cx_v3_gcs_proto_rawDesc
)

func file_google_cloud_dialogflow_cx_v3_gcs_proto_rawDescGZIP() []byte {
	file_google_cloud_dialogflow_cx_v3_gcs_proto_rawDescOnce.Do(func() {
		file_google_cloud_dialogflow_cx_v3_gcs_proto_rawDescData = protoimpl.X.CompressGZIP(file_google_cloud_dialogflow_cx_v3_gcs_proto_rawDescData)
	})
	return file_google_cloud_dialogflow_cx_v3_gcs_proto_rawDescData
}

var file_google_cloud_dialogflow_cx_v3_gcs_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_google_cloud_dialogflow_cx_v3_gcs_proto_goTypes = []interface{}{
	(*GcsDestination)(nil), // 0: google.cloud.dialogflow.cx.v3.GcsDestination
}
var file_google_cloud_dialogflow_cx_v3_gcs_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_google_cloud_dialogflow_cx_v3_gcs_proto_init() }
func file_google_cloud_dialogflow_cx_v3_gcs_proto_init() {
	if File_google_cloud_dialogflow_cx_v3_gcs_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_google_cloud_dialogflow_cx_v3_gcs_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GcsDestination); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_google_cloud_dialogflow_cx_v3_gcs_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_google_cloud_dialogflow_cx_v3_gcs_proto_goTypes,
		DependencyIndexes: file_google_cloud_dialogflow_cx_v3_gcs_proto_depIdxs,
		MessageInfos:      file_google_cloud_dialogflow_cx_v3_gcs_proto_msgTypes,
	}.Build()
	File_google_cloud_dialogflow_cx_v3_gcs_proto = out.File
	file_google_cloud_dialogflow_cx_v3_gcs_proto_rawDesc = nil
	file_google_cloud_dialogflow_cx_v3_gcs_proto_goTypes = nil
	file_google_cloud_dialogflow_cx_v3_gcs_proto_depIdxs = nil
}
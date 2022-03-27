// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.6.1
// source: examplepb.proto

package examplepb

import (
	proto "github.com/golang/protobuf/proto"
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

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

// Updates a resource of Key to version provided.
type UpdateVersionRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key     *Key     `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Version *Version `protobuf:"bytes,2,opt,name=version,proto3" json:"version,omitempty"`
	Data    *Data    `protobuf:"bytes,3,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *UpdateVersionRequest) Reset() {
	*x = UpdateVersionRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_examplepb_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateVersionRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateVersionRequest) ProtoMessage() {}

func (x *UpdateVersionRequest) ProtoReflect() protoreflect.Message {
	mi := &file_examplepb_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateVersionRequest.ProtoReflect.Descriptor instead.
func (*UpdateVersionRequest) Descriptor() ([]byte, []int) {
	return file_examplepb_proto_rawDescGZIP(), []int{0}
}

func (x *UpdateVersionRequest) GetKey() *Key {
	if x != nil {
		return x.Key
	}
	return nil
}

func (x *UpdateVersionRequest) GetVersion() *Version {
	if x != nil {
		return x.Version
	}
	return nil
}

func (x *UpdateVersionRequest) GetData() *Data {
	if x != nil {
		return x.Data
	}
	return nil
}

// Requests the current version of the provided keyed resource.
type GetVersionRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key *Key `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
}

func (x *GetVersionRequest) Reset() {
	*x = GetVersionRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_examplepb_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetVersionRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetVersionRequest) ProtoMessage() {}

func (x *GetVersionRequest) ProtoReflect() protoreflect.Message {
	mi := &file_examplepb_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetVersionRequest.ProtoReflect.Descriptor instead.
func (*GetVersionRequest) Descriptor() ([]byte, []int) {
	return file_examplepb_proto_rawDescGZIP(), []int{1}
}

func (x *GetVersionRequest) GetKey() *Key {
	if x != nil {
		return x.Key
	}
	return nil
}

// Response containing the current version of the keyed resource.
type VersionResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key     *Key     `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Version *Version `protobuf:"bytes,2,opt,name=version,proto3" json:"version,omitempty"`
}

func (x *VersionResponse) Reset() {
	*x = VersionResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_examplepb_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *VersionResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*VersionResponse) ProtoMessage() {}

func (x *VersionResponse) ProtoReflect() protoreflect.Message {
	mi := &file_examplepb_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use VersionResponse.ProtoReflect.Descriptor instead.
func (*VersionResponse) Descriptor() ([]byte, []int) {
	return file_examplepb_proto_rawDescGZIP(), []int{2}
}

func (x *VersionResponse) GetKey() *Key {
	if x != nil {
		return x.Key
	}
	return nil
}

func (x *VersionResponse) GetVersion() *Version {
	if x != nil {
		return x.Version
	}
	return nil
}

type Key struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id int64 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *Key) Reset() {
	*x = Key{}
	if protoimpl.UnsafeEnabled {
		mi := &file_examplepb_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Key) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Key) ProtoMessage() {}

func (x *Key) ProtoReflect() protoreflect.Message {
	mi := &file_examplepb_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Key.ProtoReflect.Descriptor instead.
func (*Key) Descriptor() ([]byte, []int) {
	return file_examplepb_proto_rawDescGZIP(), []int{3}
}

func (x *Key) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

type Version struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id int64 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *Version) Reset() {
	*x = Version{}
	if protoimpl.UnsafeEnabled {
		mi := &file_examplepb_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Version) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Version) ProtoMessage() {}

func (x *Version) ProtoReflect() protoreflect.Message {
	mi := &file_examplepb_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Version.ProtoReflect.Descriptor instead.
func (*Version) Descriptor() ([]byte, []int) {
	return file_examplepb_proto_rawDescGZIP(), []int{4}
}

func (x *Version) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

type Data struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *Data) Reset() {
	*x = Data{}
	if protoimpl.UnsafeEnabled {
		mi := &file_examplepb_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Data) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Data) ProtoMessage() {}

func (x *Data) ProtoReflect() protoreflect.Message {
	mi := &file_examplepb_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Data.ProtoReflect.Descriptor instead.
func (*Data) Descriptor() ([]byte, []int) {
	return file_examplepb_proto_rawDescGZIP(), []int{5}
}

func (x *Data) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

var File_examplepb_proto protoreflect.FileDescriptor

var file_examplepb_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x70, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x09, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x70, 0x62, 0x22, 0x8b, 0x01, 0x0a,
	0x14, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x20, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x70, 0x62, 0x2e, 0x4b,
	0x65, 0x79, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x2c, 0x0a, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69,
	0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x65, 0x78, 0x61, 0x6d, 0x70,
	0x6c, 0x65, 0x70, 0x62, 0x2e, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x07, 0x76, 0x65,
	0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x23, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x70, 0x62, 0x2e,
	0x44, 0x61, 0x74, 0x61, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x22, 0x35, 0x0a, 0x11, 0x47, 0x65,
	0x74, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x20, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x65,
	0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x70, 0x62, 0x2e, 0x4b, 0x65, 0x79, 0x52, 0x03, 0x6b, 0x65,
	0x79, 0x22, 0x61, 0x0a, 0x0f, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x20, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x0e, 0x2e, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x70, 0x62, 0x2e, 0x4b, 0x65,
	0x79, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x2c, 0x0a, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f,
	0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c,
	0x65, 0x70, 0x62, 0x2e, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x07, 0x76, 0x65, 0x72,
	0x73, 0x69, 0x6f, 0x6e, 0x22, 0x15, 0x0a, 0x03, 0x4b, 0x65, 0x79, 0x12, 0x0e, 0x0a, 0x02, 0x69,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x02, 0x69, 0x64, 0x22, 0x19, 0x0a, 0x07, 0x56,
	0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x03, 0x52, 0x02, 0x69, 0x64, 0x22, 0x16, 0x0a, 0x04, 0x44, 0x61, 0x74, 0x61, 0x12, 0x0e,
	0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x32, 0xa1,
	0x01, 0x0a, 0x09, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x65, 0x72, 0x12, 0x4c, 0x0a, 0x0d,
	0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x1f, 0x2e,
	0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x70, 0x62, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65,
	0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1a,
	0x2e, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x70, 0x62, 0x2e, 0x56, 0x65, 0x72, 0x73, 0x69,
	0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x46, 0x0a, 0x0a, 0x47, 0x65,
	0x74, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x1c, 0x2e, 0x65, 0x78, 0x61, 0x6d, 0x70,
	0x6c, 0x65, 0x70, 0x62, 0x2e, 0x47, 0x65, 0x74, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1a, 0x2e, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65,
	0x70, 0x62, 0x2e, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x42, 0x30, 0x5a, 0x2e, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x62, 0x74, 0x73, 0x6f, 0x6d, 0x6f, 0x67, 0x79, 0x69, 0x2f, 0x61, 0x72, 0x62, 0x69, 0x74,
	0x65, 0x72, 0x2f, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2f, 0x65, 0x78, 0x61, 0x6d, 0x70,
	0x6c, 0x65, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_examplepb_proto_rawDescOnce sync.Once
	file_examplepb_proto_rawDescData = file_examplepb_proto_rawDesc
)

func file_examplepb_proto_rawDescGZIP() []byte {
	file_examplepb_proto_rawDescOnce.Do(func() {
		file_examplepb_proto_rawDescData = protoimpl.X.CompressGZIP(file_examplepb_proto_rawDescData)
	})
	return file_examplepb_proto_rawDescData
}

var file_examplepb_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_examplepb_proto_goTypes = []interface{}{
	(*UpdateVersionRequest)(nil), // 0: examplepb.UpdateVersionRequest
	(*GetVersionRequest)(nil),    // 1: examplepb.GetVersionRequest
	(*VersionResponse)(nil),      // 2: examplepb.VersionResponse
	(*Key)(nil),                  // 3: examplepb.Key
	(*Version)(nil),              // 4: examplepb.Version
	(*Data)(nil),                 // 5: examplepb.Data
}
var file_examplepb_proto_depIdxs = []int32{
	3, // 0: examplepb.UpdateVersionRequest.key:type_name -> examplepb.Key
	4, // 1: examplepb.UpdateVersionRequest.version:type_name -> examplepb.Version
	5, // 2: examplepb.UpdateVersionRequest.data:type_name -> examplepb.Data
	3, // 3: examplepb.GetVersionRequest.key:type_name -> examplepb.Key
	3, // 4: examplepb.VersionResponse.key:type_name -> examplepb.Key
	4, // 5: examplepb.VersionResponse.version:type_name -> examplepb.Version
	0, // 6: examplepb.Versioner.UpdateVersion:input_type -> examplepb.UpdateVersionRequest
	1, // 7: examplepb.Versioner.GetVersion:input_type -> examplepb.GetVersionRequest
	2, // 8: examplepb.Versioner.UpdateVersion:output_type -> examplepb.VersionResponse
	2, // 9: examplepb.Versioner.GetVersion:output_type -> examplepb.VersionResponse
	8, // [8:10] is the sub-list for method output_type
	6, // [6:8] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_name
	6, // [6:6] is the sub-list for extension extendee
	0, // [0:6] is the sub-list for field type_name
}

func init() { file_examplepb_proto_init() }
func file_examplepb_proto_init() {
	if File_examplepb_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_examplepb_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpdateVersionRequest); i {
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
		file_examplepb_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetVersionRequest); i {
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
		file_examplepb_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*VersionResponse); i {
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
		file_examplepb_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Key); i {
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
		file_examplepb_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Version); i {
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
		file_examplepb_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Data); i {
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
			RawDescriptor: file_examplepb_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_examplepb_proto_goTypes,
		DependencyIndexes: file_examplepb_proto_depIdxs,
		MessageInfos:      file_examplepb_proto_msgTypes,
	}.Build()
	File_examplepb_proto = out.File
	file_examplepb_proto_rawDesc = nil
	file_examplepb_proto_goTypes = nil
	file_examplepb_proto_depIdxs = nil
}

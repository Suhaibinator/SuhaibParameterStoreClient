// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v5.29.3
// source: proto/parameter_store_interface.proto

package proto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
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

type StoreRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Key           string                 `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Value         string                 `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
	Password      string                 `protobuf:"bytes,3,opt,name=password,proto3" json:"password,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *StoreRequest) Reset() {
	*x = StoreRequest{}
	mi := &file_proto_parameter_store_interface_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *StoreRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StoreRequest) ProtoMessage() {}

func (x *StoreRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_parameter_store_interface_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StoreRequest.ProtoReflect.Descriptor instead.
func (*StoreRequest) Descriptor() ([]byte, []int) {
	return file_proto_parameter_store_interface_proto_rawDescGZIP(), []int{0}
}

func (x *StoreRequest) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *StoreRequest) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

func (x *StoreRequest) GetPassword() string {
	if x != nil {
		return x.Password
	}
	return ""
}

type StoreResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Message       string                 `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *StoreResponse) Reset() {
	*x = StoreResponse{}
	mi := &file_proto_parameter_store_interface_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *StoreResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StoreResponse) ProtoMessage() {}

func (x *StoreResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_parameter_store_interface_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StoreResponse.ProtoReflect.Descriptor instead.
func (*StoreResponse) Descriptor() ([]byte, []int) {
	return file_proto_parameter_store_interface_proto_rawDescGZIP(), []int{1}
}

func (x *StoreResponse) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

type RetrieveRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Key           string                 `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Password      string                 `protobuf:"bytes,2,opt,name=password,proto3" json:"password,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RetrieveRequest) Reset() {
	*x = RetrieveRequest{}
	mi := &file_proto_parameter_store_interface_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RetrieveRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RetrieveRequest) ProtoMessage() {}

func (x *RetrieveRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_parameter_store_interface_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RetrieveRequest.ProtoReflect.Descriptor instead.
func (*RetrieveRequest) Descriptor() ([]byte, []int) {
	return file_proto_parameter_store_interface_proto_rawDescGZIP(), []int{2}
}

func (x *RetrieveRequest) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *RetrieveRequest) GetPassword() string {
	if x != nil {
		return x.Password
	}
	return ""
}

type RetrieveResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Value         string                 `protobuf:"bytes,1,opt,name=value,proto3" json:"value,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RetrieveResponse) Reset() {
	*x = RetrieveResponse{}
	mi := &file_proto_parameter_store_interface_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RetrieveResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RetrieveResponse) ProtoMessage() {}

func (x *RetrieveResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_parameter_store_interface_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RetrieveResponse.ProtoReflect.Descriptor instead.
func (*RetrieveResponse) Descriptor() ([]byte, []int) {
	return file_proto_parameter_store_interface_proto_rawDescGZIP(), []int{3}
}

func (x *RetrieveResponse) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

type AddAccessRequest struct {
	state          protoimpl.MessageState `protogen:"open.v1"`
	Password       string                 `protobuf:"bytes,1,opt,name=password,proto3" json:"password,omitempty"`
	Key            string                 `protobuf:"bytes,2,opt,name=key,proto3" json:"key,omitempty"`
	MasterPassword string                 `protobuf:"bytes,3,opt,name=masterPassword,proto3" json:"masterPassword,omitempty"`
	unknownFields  protoimpl.UnknownFields
	sizeCache      protoimpl.SizeCache
}

func (x *AddAccessRequest) Reset() {
	*x = AddAccessRequest{}
	mi := &file_proto_parameter_store_interface_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AddAccessRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddAccessRequest) ProtoMessage() {}

func (x *AddAccessRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_parameter_store_interface_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AddAccessRequest.ProtoReflect.Descriptor instead.
func (*AddAccessRequest) Descriptor() ([]byte, []int) {
	return file_proto_parameter_store_interface_proto_rawDescGZIP(), []int{4}
}

func (x *AddAccessRequest) GetPassword() string {
	if x != nil {
		return x.Password
	}
	return ""
}

func (x *AddAccessRequest) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *AddAccessRequest) GetMasterPassword() string {
	if x != nil {
		return x.MasterPassword
	}
	return ""
}

type AddAccessResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Message       string                 `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *AddAccessResponse) Reset() {
	*x = AddAccessResponse{}
	mi := &file_proto_parameter_store_interface_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AddAccessResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddAccessResponse) ProtoMessage() {}

func (x *AddAccessResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_parameter_store_interface_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AddAccessResponse.ProtoReflect.Descriptor instead.
func (*AddAccessResponse) Descriptor() ([]byte, []int) {
	return file_proto_parameter_store_interface_proto_rawDescGZIP(), []int{5}
}

func (x *AddAccessResponse) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

var File_proto_parameter_store_interface_proto protoreflect.FileDescriptor

const file_proto_parameter_store_interface_proto_rawDesc = "" +
	"\n" +
	"%proto/parameter_store_interface.proto\x12\x0eparameterstore\"R\n" +
	"\fStoreRequest\x12\x10\n" +
	"\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n" +
	"\x05value\x18\x02 \x01(\tR\x05value\x12\x1a\n" +
	"\bpassword\x18\x03 \x01(\tR\bpassword\")\n" +
	"\rStoreResponse\x12\x18\n" +
	"\amessage\x18\x01 \x01(\tR\amessage\"?\n" +
	"\x0fRetrieveRequest\x12\x10\n" +
	"\x03key\x18\x01 \x01(\tR\x03key\x12\x1a\n" +
	"\bpassword\x18\x02 \x01(\tR\bpassword\"(\n" +
	"\x10RetrieveResponse\x12\x14\n" +
	"\x05value\x18\x01 \x01(\tR\x05value\"h\n" +
	"\x10AddAccessRequest\x12\x1a\n" +
	"\bpassword\x18\x01 \x01(\tR\bpassword\x12\x10\n" +
	"\x03key\x18\x02 \x01(\tR\x03key\x12&\n" +
	"\x0emasterPassword\x18\x03 \x01(\tR\x0emasterPassword\"-\n" +
	"\x11AddAccessResponse\x12\x18\n" +
	"\amessage\x18\x01 \x01(\tR\amessage2\xfd\x01\n" +
	"\x0eParameterStore\x12F\n" +
	"\x05Store\x12\x1c.parameterstore.StoreRequest\x1a\x1d.parameterstore.StoreResponse\"\x00\x12O\n" +
	"\bRetrieve\x12\x1f.parameterstore.RetrieveRequest\x1a .parameterstore.RetrieveResponse\"\x00\x12R\n" +
	"\tAddAccess\x12 .parameterstore.AddAccessRequest\x1a!.parameterstore.AddAccessResponse\"\x00B:Z8github.com/Suhaibinator/SuhaibParameterStoreClient/protob\x06proto3"

var (
	file_proto_parameter_store_interface_proto_rawDescOnce sync.Once
	file_proto_parameter_store_interface_proto_rawDescData []byte
)

func file_proto_parameter_store_interface_proto_rawDescGZIP() []byte {
	file_proto_parameter_store_interface_proto_rawDescOnce.Do(func() {
		file_proto_parameter_store_interface_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_proto_parameter_store_interface_proto_rawDesc), len(file_proto_parameter_store_interface_proto_rawDesc)))
	})
	return file_proto_parameter_store_interface_proto_rawDescData
}

var file_proto_parameter_store_interface_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_proto_parameter_store_interface_proto_goTypes = []any{
	(*StoreRequest)(nil),      // 0: parameterstore.StoreRequest
	(*StoreResponse)(nil),     // 1: parameterstore.StoreResponse
	(*RetrieveRequest)(nil),   // 2: parameterstore.RetrieveRequest
	(*RetrieveResponse)(nil),  // 3: parameterstore.RetrieveResponse
	(*AddAccessRequest)(nil),  // 4: parameterstore.AddAccessRequest
	(*AddAccessResponse)(nil), // 5: parameterstore.AddAccessResponse
}
var file_proto_parameter_store_interface_proto_depIdxs = []int32{
	0, // 0: parameterstore.ParameterStore.Store:input_type -> parameterstore.StoreRequest
	2, // 1: parameterstore.ParameterStore.Retrieve:input_type -> parameterstore.RetrieveRequest
	4, // 2: parameterstore.ParameterStore.AddAccess:input_type -> parameterstore.AddAccessRequest
	1, // 3: parameterstore.ParameterStore.Store:output_type -> parameterstore.StoreResponse
	3, // 4: parameterstore.ParameterStore.Retrieve:output_type -> parameterstore.RetrieveResponse
	5, // 5: parameterstore.ParameterStore.AddAccess:output_type -> parameterstore.AddAccessResponse
	3, // [3:6] is the sub-list for method output_type
	0, // [0:3] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_proto_parameter_store_interface_proto_init() }
func file_proto_parameter_store_interface_proto_init() {
	if File_proto_parameter_store_interface_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_proto_parameter_store_interface_proto_rawDesc), len(file_proto_parameter_store_interface_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_proto_parameter_store_interface_proto_goTypes,
		DependencyIndexes: file_proto_parameter_store_interface_proto_depIdxs,
		MessageInfos:      file_proto_parameter_store_interface_proto_msgTypes,
	}.Build()
	File_proto_parameter_store_interface_proto = out.File
	file_proto_parameter_store_interface_proto_goTypes = nil
	file_proto_parameter_store_interface_proto_depIdxs = nil
}

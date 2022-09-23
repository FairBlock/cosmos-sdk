// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.19.4
// source: listener.proto

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
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

// PutRequest is used for storing ABCI request and response
// and Store KV data for streaming to external grpc service.
type PutRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	BlockHeight int64  `protobuf:"varint,1,opt,name=block_height,json=blockHeight,proto3" json:"block_height,omitempty"`
	Req         []byte `protobuf:"bytes,2,opt,name=req,proto3" json:"req,omitempty"`
	Res         []byte `protobuf:"bytes,3,opt,name=res,proto3" json:"res,omitempty"`
	StoreKvPair []byte `protobuf:"bytes,4,opt,name=store_kv_pair,json=storeKvPair,proto3" json:"store_kv_pair,omitempty"`
}

func (x *PutRequest) Reset() {
	*x = PutRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_listener_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PutRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PutRequest) ProtoMessage() {}

func (x *PutRequest) ProtoReflect() protoreflect.Message {
	mi := &file_listener_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PutRequest.ProtoReflect.Descriptor instead.
func (*PutRequest) Descriptor() ([]byte, []int) {
	return file_listener_proto_rawDescGZIP(), []int{0}
}

func (x *PutRequest) GetBlockHeight() int64 {
	if x != nil {
		return x.BlockHeight
	}
	return 0
}

func (x *PutRequest) GetReq() []byte {
	if x != nil {
		return x.Req
	}
	return nil
}

func (x *PutRequest) GetRes() []byte {
	if x != nil {
		return x.Res
	}
	return nil
}

func (x *PutRequest) GetStoreKvPair() []byte {
	if x != nil {
		return x.StoreKvPair
	}
	return nil
}

type Empty struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Empty) Reset() {
	*x = Empty{}
	if protoimpl.UnsafeEnabled {
		mi := &file_listener_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Empty) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Empty) ProtoMessage() {}

func (x *Empty) ProtoReflect() protoreflect.Message {
	mi := &file_listener_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Empty.ProtoReflect.Descriptor instead.
func (*Empty) Descriptor() ([]byte, []int) {
	return file_listener_proto_rawDescGZIP(), []int{1}
}

var File_listener_proto protoreflect.FileDescriptor

var file_listener_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x6c, 0x69, 0x73, 0x74, 0x65, 0x6e, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x18, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x62, 0x61, 0x73, 0x65, 0x2e, 0x61, 0x62,
	0x63, 0x69, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x22, 0x77, 0x0a, 0x0a, 0x50, 0x75,
	0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x21, 0x0a, 0x0c, 0x62, 0x6c, 0x6f, 0x63,
	0x6b, 0x5f, 0x68, 0x65, 0x69, 0x67, 0x68, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0b,
	0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x48, 0x65, 0x69, 0x67, 0x68, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x72,
	0x65, 0x71, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x03, 0x72, 0x65, 0x71, 0x12, 0x10, 0x0a,
	0x03, 0x72, 0x65, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x03, 0x72, 0x65, 0x73, 0x12,
	0x22, 0x0a, 0x0d, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x5f, 0x6b, 0x76, 0x5f, 0x70, 0x61, 0x69, 0x72,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0b, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x4b, 0x76, 0x50,
	0x61, 0x69, 0x72, 0x22, 0x07, 0x0a, 0x05, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x32, 0xff, 0x02, 0x0a,
	0x13, 0x41, 0x42, 0x43, 0x49, 0x4c, 0x69, 0x73, 0x74, 0x65, 0x6e, 0x65, 0x72, 0x53, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x12, 0x59, 0x0a, 0x10, 0x4c, 0x69, 0x73, 0x74, 0x65, 0x6e, 0x42, 0x65,
	0x67, 0x69, 0x6e, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x12, 0x24, 0x2e, 0x63, 0x6f, 0x73, 0x6d, 0x6f,
	0x73, 0x2e, 0x62, 0x61, 0x73, 0x65, 0x2e, 0x61, 0x62, 0x63, 0x69, 0x2e, 0x76, 0x31, 0x62, 0x65,
	0x74, 0x61, 0x31, 0x2e, 0x50, 0x75, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1f,
	0x2e, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x62, 0x61, 0x73, 0x65, 0x2e, 0x61, 0x62, 0x63,
	0x69, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x12,
	0x57, 0x0a, 0x0e, 0x4c, 0x69, 0x73, 0x74, 0x65, 0x6e, 0x45, 0x6e, 0x64, 0x42, 0x6c, 0x6f, 0x63,
	0x6b, 0x12, 0x24, 0x2e, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x62, 0x61, 0x73, 0x65, 0x2e,
	0x61, 0x62, 0x63, 0x69, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x50, 0x75, 0x74,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1f, 0x2e, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73,
	0x2e, 0x62, 0x61, 0x73, 0x65, 0x2e, 0x61, 0x62, 0x63, 0x69, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74,
	0x61, 0x31, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x12, 0x58, 0x0a, 0x0f, 0x4c, 0x69, 0x73, 0x74,
	0x65, 0x6e, 0x44, 0x65, 0x6c, 0x69, 0x76, 0x65, 0x72, 0x54, 0x78, 0x12, 0x24, 0x2e, 0x63, 0x6f,
	0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x62, 0x61, 0x73, 0x65, 0x2e, 0x61, 0x62, 0x63, 0x69, 0x2e, 0x76,
	0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x50, 0x75, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x1f, 0x2e, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x62, 0x61, 0x73, 0x65, 0x2e,
	0x61, 0x62, 0x63, 0x69, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x45, 0x6d, 0x70,
	0x74, 0x79, 0x12, 0x5a, 0x0a, 0x11, 0x4c, 0x69, 0x73, 0x74, 0x65, 0x6e, 0x53, 0x74, 0x6f, 0x72,
	0x65, 0x4b, 0x56, 0x50, 0x61, 0x69, 0x72, 0x12, 0x24, 0x2e, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73,
	0x2e, 0x62, 0x61, 0x73, 0x65, 0x2e, 0x61, 0x62, 0x63, 0x69, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74,
	0x61, 0x31, 0x2e, 0x50, 0x75, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1f, 0x2e,
	0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x62, 0x61, 0x73, 0x65, 0x2e, 0x61, 0x62, 0x63, 0x69,
	0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x42, 0x68,
	0x0a, 0x29, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x2e, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73,
	0x2e, 0x73, 0x64, 0x6b, 0x2e, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x69, 0x6e, 0x67, 0x2e, 0x70,
	0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x2e, 0x61, 0x62, 0x63, 0x69, 0x50, 0x01, 0x5a, 0x39, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73,
	0x2f, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2d, 0x73, 0x64, 0x6b, 0x2f, 0x73, 0x74, 0x72, 0x65,
	0x61, 0x6d, 0x69, 0x6e, 0x67, 0x2f, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x2f, 0x61, 0x62,
	0x63, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_listener_proto_rawDescOnce sync.Once
	file_listener_proto_rawDescData = file_listener_proto_rawDesc
)

func file_listener_proto_rawDescGZIP() []byte {
	file_listener_proto_rawDescOnce.Do(func() {
		file_listener_proto_rawDescData = protoimpl.X.CompressGZIP(file_listener_proto_rawDescData)
	})
	return file_listener_proto_rawDescData
}

var file_listener_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_listener_proto_goTypes = []interface{}{
	(*PutRequest)(nil), // 0: cosmos.base.abci.v1beta1.PutRequest
	(*Empty)(nil),      // 1: cosmos.base.abci.v1beta1.Empty
}
var file_listener_proto_depIdxs = []int32{
	0, // 0: cosmos.base.abci.v1beta1.ABCIListenerService.ListenBeginBlock:input_type -> cosmos.base.abci.v1beta1.PutRequest
	0, // 1: cosmos.base.abci.v1beta1.ABCIListenerService.ListenEndBlock:input_type -> cosmos.base.abci.v1beta1.PutRequest
	0, // 2: cosmos.base.abci.v1beta1.ABCIListenerService.ListenDeliverTx:input_type -> cosmos.base.abci.v1beta1.PutRequest
	0, // 3: cosmos.base.abci.v1beta1.ABCIListenerService.ListenStoreKVPair:input_type -> cosmos.base.abci.v1beta1.PutRequest
	1, // 4: cosmos.base.abci.v1beta1.ABCIListenerService.ListenBeginBlock:output_type -> cosmos.base.abci.v1beta1.Empty
	1, // 5: cosmos.base.abci.v1beta1.ABCIListenerService.ListenEndBlock:output_type -> cosmos.base.abci.v1beta1.Empty
	1, // 6: cosmos.base.abci.v1beta1.ABCIListenerService.ListenDeliverTx:output_type -> cosmos.base.abci.v1beta1.Empty
	1, // 7: cosmos.base.abci.v1beta1.ABCIListenerService.ListenStoreKVPair:output_type -> cosmos.base.abci.v1beta1.Empty
	4, // [4:8] is the sub-list for method output_type
	0, // [0:4] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_listener_proto_init() }
func file_listener_proto_init() {
	if File_listener_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_listener_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PutRequest); i {
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
		file_listener_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Empty); i {
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
			RawDescriptor: file_listener_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_listener_proto_goTypes,
		DependencyIndexes: file_listener_proto_depIdxs,
		MessageInfos:      file_listener_proto_msgTypes,
	}.Build()
	File_listener_proto = out.File
	file_listener_proto_rawDesc = nil
	file_listener_proto_goTypes = nil
	file_listener_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// ABCIListenerServiceClient is the client API for ABCIListenerService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type ABCIListenerServiceClient interface {
	ListenBeginBlock(ctx context.Context, in *PutRequest, opts ...grpc.CallOption) (*Empty, error)
	ListenEndBlock(ctx context.Context, in *PutRequest, opts ...grpc.CallOption) (*Empty, error)
	ListenDeliverTx(ctx context.Context, in *PutRequest, opts ...grpc.CallOption) (*Empty, error)
	ListenStoreKVPair(ctx context.Context, in *PutRequest, opts ...grpc.CallOption) (*Empty, error)
}

type aBCIListenerServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewABCIListenerServiceClient(cc grpc.ClientConnInterface) ABCIListenerServiceClient {
	return &aBCIListenerServiceClient{cc}
}

func (c *aBCIListenerServiceClient) ListenBeginBlock(ctx context.Context, in *PutRequest, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/cosmos.base.abci.v1beta1.ABCIListenerService/ListenBeginBlock", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *aBCIListenerServiceClient) ListenEndBlock(ctx context.Context, in *PutRequest, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/cosmos.base.abci.v1beta1.ABCIListenerService/ListenEndBlock", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *aBCIListenerServiceClient) ListenDeliverTx(ctx context.Context, in *PutRequest, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/cosmos.base.abci.v1beta1.ABCIListenerService/ListenDeliverTx", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *aBCIListenerServiceClient) ListenStoreKVPair(ctx context.Context, in *PutRequest, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/cosmos.base.abci.v1beta1.ABCIListenerService/ListenStoreKVPair", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ABCIListenerServiceServer is the server API for ABCIListenerService service.
type ABCIListenerServiceServer interface {
	ListenBeginBlock(context.Context, *PutRequest) (*Empty, error)
	ListenEndBlock(context.Context, *PutRequest) (*Empty, error)
	ListenDeliverTx(context.Context, *PutRequest) (*Empty, error)
	ListenStoreKVPair(context.Context, *PutRequest) (*Empty, error)
}

// UnimplementedABCIListenerServiceServer can be embedded to have forward compatible implementations.
type UnimplementedABCIListenerServiceServer struct {
}

func (*UnimplementedABCIListenerServiceServer) ListenBeginBlock(context.Context, *PutRequest) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListenBeginBlock not implemented")
}
func (*UnimplementedABCIListenerServiceServer) ListenEndBlock(context.Context, *PutRequest) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListenEndBlock not implemented")
}
func (*UnimplementedABCIListenerServiceServer) ListenDeliverTx(context.Context, *PutRequest) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListenDeliverTx not implemented")
}
func (*UnimplementedABCIListenerServiceServer) ListenStoreKVPair(context.Context, *PutRequest) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListenStoreKVPair not implemented")
}

func RegisterABCIListenerServiceServer(s *grpc.Server, srv ABCIListenerServiceServer) {
	s.RegisterService(&_ABCIListenerService_serviceDesc, srv)
}

func _ABCIListenerService_ListenBeginBlock_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PutRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ABCIListenerServiceServer).ListenBeginBlock(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cosmos.base.abci.v1beta1.ABCIListenerService/ListenBeginBlock",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ABCIListenerServiceServer).ListenBeginBlock(ctx, req.(*PutRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ABCIListenerService_ListenEndBlock_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PutRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ABCIListenerServiceServer).ListenEndBlock(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cosmos.base.abci.v1beta1.ABCIListenerService/ListenEndBlock",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ABCIListenerServiceServer).ListenEndBlock(ctx, req.(*PutRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ABCIListenerService_ListenDeliverTx_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PutRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ABCIListenerServiceServer).ListenDeliverTx(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cosmos.base.abci.v1beta1.ABCIListenerService/ListenDeliverTx",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ABCIListenerServiceServer).ListenDeliverTx(ctx, req.(*PutRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ABCIListenerService_ListenStoreKVPair_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PutRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ABCIListenerServiceServer).ListenStoreKVPair(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cosmos.base.abci.v1beta1.ABCIListenerService/ListenStoreKVPair",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ABCIListenerServiceServer).ListenStoreKVPair(ctx, req.(*PutRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _ABCIListenerService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "cosmos.base.abci.v1beta1.ABCIListenerService",
	HandlerType: (*ABCIListenerServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ListenBeginBlock",
			Handler:    _ABCIListenerService_ListenBeginBlock_Handler,
		},
		{
			MethodName: "ListenEndBlock",
			Handler:    _ABCIListenerService_ListenEndBlock_Handler,
		},
		{
			MethodName: "ListenDeliverTx",
			Handler:    _ABCIListenerService_ListenDeliverTx_Handler,
		},
		{
			MethodName: "ListenStoreKVPair",
			Handler:    _ABCIListenerService_ListenStoreKVPair_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "listener.proto",
}

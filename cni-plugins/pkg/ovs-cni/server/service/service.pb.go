// Code generated by protoc-gen-go. DO NOT EDIT.
// source: service.proto

package service

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type PortRequest struct {
	BrName               string   `protobuf:"bytes,1,opt,name=brName,proto3" json:"brName,omitempty"`
	PortName             string   `protobuf:"bytes,2,opt,name=portName,proto3" json:"portName,omitempty"`
	PortMac              string   `protobuf:"bytes,3,opt,name=portMac,proto3" json:"portMac,omitempty"`
	VlanTag              uint32   `protobuf:"varint,4,opt,name=vlanTag,proto3" json:"vlanTag,omitempty"`
	Mtu                  uint32   `protobuf:"varint,5,opt,name=mtu,proto3" json:"mtu,omitempty"`
	DoSetUp              bool     `protobuf:"varint,6,opt,name=doSetUp,proto3" json:"doSetUp,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PortRequest) Reset()         { *m = PortRequest{} }
func (m *PortRequest) String() string { return proto.CompactTextString(m) }
func (*PortRequest) ProtoMessage()    {}
func (*PortRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_a0b84a42fa06f626, []int{0}
}

func (m *PortRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PortRequest.Unmarshal(m, b)
}
func (m *PortRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PortRequest.Marshal(b, m, deterministic)
}
func (m *PortRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PortRequest.Merge(m, src)
}
func (m *PortRequest) XXX_Size() int {
	return xxx_messageInfo_PortRequest.Size(m)
}
func (m *PortRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_PortRequest.DiscardUnknown(m)
}

var xxx_messageInfo_PortRequest proto.InternalMessageInfo

func (m *PortRequest) GetBrName() string {
	if m != nil {
		return m.BrName
	}
	return ""
}

func (m *PortRequest) GetPortName() string {
	if m != nil {
		return m.PortName
	}
	return ""
}

func (m *PortRequest) GetPortMac() string {
	if m != nil {
		return m.PortMac
	}
	return ""
}

func (m *PortRequest) GetVlanTag() uint32 {
	if m != nil {
		return m.VlanTag
	}
	return 0
}

func (m *PortRequest) GetMtu() uint32 {
	if m != nil {
		return m.Mtu
	}
	return 0
}

func (m *PortRequest) GetDoSetUp() bool {
	if m != nil {
		return m.DoSetUp
	}
	return false
}

type PortResponse struct {
	Success              bool     `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PortResponse) Reset()         { *m = PortResponse{} }
func (m *PortResponse) String() string { return proto.CompactTextString(m) }
func (*PortResponse) ProtoMessage()    {}
func (*PortResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_a0b84a42fa06f626, []int{1}
}

func (m *PortResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PortResponse.Unmarshal(m, b)
}
func (m *PortResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PortResponse.Marshal(b, m, deterministic)
}
func (m *PortResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PortResponse.Merge(m, src)
}
func (m *PortResponse) XXX_Size() int {
	return xxx_messageInfo_PortResponse.Size(m)
}
func (m *PortResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_PortResponse.DiscardUnknown(m)
}

var xxx_messageInfo_PortResponse proto.InternalMessageInfo

func (m *PortResponse) GetSuccess() bool {
	if m != nil {
		return m.Success
	}
	return false
}

func init() {
	proto.RegisterType((*PortRequest)(nil), "service.PortRequest")
	proto.RegisterType((*PortResponse)(nil), "service.PortResponse")
}

func init() { proto.RegisterFile("service.proto", fileDescriptor_a0b84a42fa06f626) }

var fileDescriptor_a0b84a42fa06f626 = []byte{
	// 229 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x5c, 0x90, 0xcf, 0x4e, 0xc3, 0x30,
	0x0c, 0x87, 0x09, 0x83, 0xae, 0x18, 0x86, 0x90, 0x05, 0x28, 0xda, 0xa9, 0xea, 0xa9, 0xa7, 0x1d,
	0xe0, 0x11, 0x76, 0xe2, 0xc0, 0x1f, 0xa5, 0x8c, 0x7b, 0x16, 0x2c, 0x84, 0xb4, 0x25, 0x21, 0x71,
	0xf7, 0x3e, 0xbc, 0x29, 0x4a, 0xda, 0x20, 0xc4, 0xcd, 0x9f, 0x3f, 0x4b, 0xb6, 0x7f, 0xb0, 0x88,
	0x14, 0x0e, 0x9f, 0x86, 0x56, 0x3e, 0x38, 0x76, 0x38, 0x9f, 0xb0, 0xfd, 0x16, 0x70, 0xfe, 0xe2,
	0x02, 0x2b, 0xfa, 0x1a, 0x28, 0x32, 0xde, 0x42, 0xb5, 0x0d, 0x4f, 0x7a, 0x4f, 0x52, 0x34, 0xa2,
	0x3b, 0x53, 0x13, 0xe1, 0x12, 0x6a, 0xef, 0x02, 0x67, 0x73, 0x9c, 0xcd, 0x2f, 0xa3, 0x84, 0x79,
	0xaa, 0x1f, 0xb5, 0x91, 0xb3, 0xac, 0x0a, 0x26, 0x73, 0xd8, 0x69, 0xfb, 0xaa, 0x3f, 0xe4, 0x49,
	0x23, 0xba, 0x85, 0x2a, 0x88, 0x57, 0x30, 0xdb, 0xf3, 0x20, 0x4f, 0x73, 0x37, 0x95, 0x69, 0xf6,
	0xdd, 0xf5, 0xc4, 0x1b, 0x2f, 0xab, 0x46, 0x74, 0xb5, 0x2a, 0xd8, 0x76, 0x70, 0x31, 0x9e, 0x18,
	0xbd, 0xb3, 0x31, 0xef, 0x8b, 0x83, 0x31, 0x14, 0x63, 0x3e, 0xb2, 0x56, 0x05, 0xef, 0x36, 0x70,
	0xf9, 0xfc, 0xd6, 0xa7, 0xe1, 0x7e, 0xfc, 0x0f, 0xd7, 0x80, 0xeb, 0x40, 0x9a, 0xe9, 0xc1, 0x32,
	0x05, 0xab, 0x77, 0x49, 0xe2, 0xf5, 0xaa, 0xc4, 0xf1, 0xe7, 0xf7, 0xe5, 0xcd, 0xbf, 0xee, 0xb8,
	0xae, 0x3d, 0xda, 0x56, 0x39, 0xb4, 0xfb, 0x9f, 0x00, 0x00, 0x00, 0xff, 0xff, 0xcb, 0xa2, 0x5c,
	0x92, 0x45, 0x01, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// OVSPortServiceClient is the client API for OVSPortService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type OVSPortServiceClient interface {
	CreateInternalPort(ctx context.Context, in *PortRequest, opts ...grpc.CallOption) (*PortResponse, error)
}

type oVSPortServiceClient struct {
	cc *grpc.ClientConn
}

func NewOVSPortServiceClient(cc *grpc.ClientConn) OVSPortServiceClient {
	return &oVSPortServiceClient{cc}
}

func (c *oVSPortServiceClient) CreateInternalPort(ctx context.Context, in *PortRequest, opts ...grpc.CallOption) (*PortResponse, error) {
	out := new(PortResponse)
	err := c.cc.Invoke(ctx, "/service.OVSPortService/CreateInternalPort", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// OVSPortServiceServer is the server API for OVSPortService service.
type OVSPortServiceServer interface {
	CreateInternalPort(context.Context, *PortRequest) (*PortResponse, error)
}

// UnimplementedOVSPortServiceServer can be embedded to have forward compatible implementations.
type UnimplementedOVSPortServiceServer struct {
}

func (*UnimplementedOVSPortServiceServer) CreateInternalPort(ctx context.Context, req *PortRequest) (*PortResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateInternalPort not implemented")
}

func RegisterOVSPortServiceServer(s *grpc.Server, srv OVSPortServiceServer) {
	s.RegisterService(&_OVSPortService_serviceDesc, srv)
}

func _OVSPortService_CreateInternalPort_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PortRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OVSPortServiceServer).CreateInternalPort(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/service.OVSPortService/CreateInternalPort",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OVSPortServiceServer).CreateInternalPort(ctx, req.(*PortRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _OVSPortService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "service.OVSPortService",
	HandlerType: (*OVSPortServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateInternalPort",
			Handler:    _OVSPortService_CreateInternalPort_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "service.proto",
}

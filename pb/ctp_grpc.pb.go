// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.4
// source: ctp.proto

package pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// CtpClient is the client API for Ctp service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type CtpClient interface {
	GetKline(ctx context.Context, in *KlineReq, opts ...grpc.CallOption) (*KlineResp, error)
	GetTick(ctx context.Context, in *RangeReq, opts ...grpc.CallOption) (*MarketDatas, error)
}

type ctpClient struct {
	cc grpc.ClientConnInterface
}

func NewCtpClient(cc grpc.ClientConnInterface) CtpClient {
	return &ctpClient{cc}
}

func (c *ctpClient) GetKline(ctx context.Context, in *KlineReq, opts ...grpc.CallOption) (*KlineResp, error) {
	out := new(KlineResp)
	err := c.cc.Invoke(ctx, "/ctpmonitor.pb.Ctp/GetKline", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ctpClient) GetTick(ctx context.Context, in *RangeReq, opts ...grpc.CallOption) (*MarketDatas, error) {
	out := new(MarketDatas)
	err := c.cc.Invoke(ctx, "/ctpmonitor.pb.Ctp/GetTick", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CtpServer is the server API for Ctp service.
// All implementations must embed UnimplementedCtpServer
// for forward compatibility
type CtpServer interface {
	GetKline(context.Context, *KlineReq) (*KlineResp, error)
	GetTick(context.Context, *RangeReq) (*MarketDatas, error)
	mustEmbedUnimplementedCtpServer()
}

// UnimplementedCtpServer must be embedded to have forward compatible implementations.
type UnimplementedCtpServer struct {
}

func (UnimplementedCtpServer) GetKline(context.Context, *KlineReq) (*KlineResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetKline not implemented")
}
func (UnimplementedCtpServer) GetTick(context.Context, *RangeReq) (*MarketDatas, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetTick not implemented")
}
func (UnimplementedCtpServer) mustEmbedUnimplementedCtpServer() {}

// UnsafeCtpServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to CtpServer will
// result in compilation errors.
type UnsafeCtpServer interface {
	mustEmbedUnimplementedCtpServer()
}

func RegisterCtpServer(s grpc.ServiceRegistrar, srv CtpServer) {
	s.RegisterService(&Ctp_ServiceDesc, srv)
}

func _Ctp_GetKline_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(KlineReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CtpServer).GetKline(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ctpmonitor.pb.Ctp/GetKline",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CtpServer).GetKline(ctx, req.(*KlineReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Ctp_GetTick_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RangeReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CtpServer).GetTick(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ctpmonitor.pb.Ctp/GetTick",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CtpServer).GetTick(ctx, req.(*RangeReq))
	}
	return interceptor(ctx, in, info, handler)
}

// Ctp_ServiceDesc is the grpc.ServiceDesc for Ctp service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Ctp_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "ctpmonitor.pb.Ctp",
	HandlerType: (*CtpServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetKline",
			Handler:    _Ctp_GetKline_Handler,
		},
		{
			MethodName: "GetTick",
			Handler:    _Ctp_GetTick_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "ctp.proto",
}
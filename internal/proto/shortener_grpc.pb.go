// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.4.0
// - protoc             v5.29.3
// source: shortener.proto

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.62.0 or later.
const _ = grpc.SupportPackageIsVersion8

const (
	Shortener_Shorten_FullMethodName        = "/shortener.Shortener/Shorten"
	Shortener_GetOriginal_FullMethodName    = "/shortener.Shortener/GetOriginal"
	Shortener_BatchShorten_FullMethodName   = "/shortener.Shortener/BatchShorten"
	Shortener_GetUserURLs_FullMethodName    = "/shortener.Shortener/GetUserURLs"
	Shortener_DeleteUserURLs_FullMethodName = "/shortener.Shortener/DeleteUserURLs"
	Shortener_Ping_FullMethodName           = "/shortener.Shortener/Ping"
	Shortener_InternalStats_FullMethodName  = "/shortener.Shortener/InternalStats"
)

// ShortenerClient is the client API for Shortener service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ShortenerClient interface {
	Shorten(ctx context.Context, in *ShortenRequest, opts ...grpc.CallOption) (*ShortenResponse, error)
	GetOriginal(ctx context.Context, in *GetOriginalRequest, opts ...grpc.CallOption) (*GetOriginalResponse, error)
	BatchShorten(ctx context.Context, in *BatchShortenRequest, opts ...grpc.CallOption) (*BatchShortenResponse, error)
	GetUserURLs(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*GetUserURLsResponse, error)
	DeleteUserURLs(ctx context.Context, in *DeleteUserURLsRequest, opts ...grpc.CallOption) (*Empty, error)
	Ping(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Empty, error)
	InternalStats(ctx context.Context, in *InternalStatsRequest, opts ...grpc.CallOption) (*InternalStatsResponse, error)
}

type shortenerClient struct {
	cc grpc.ClientConnInterface
}

func NewShortenerClient(cc grpc.ClientConnInterface) ShortenerClient {
	return &shortenerClient{cc}
}

func (c *shortenerClient) Shorten(ctx context.Context, in *ShortenRequest, opts ...grpc.CallOption) (*ShortenResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ShortenResponse)
	err := c.cc.Invoke(ctx, Shortener_Shorten_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shortenerClient) GetOriginal(ctx context.Context, in *GetOriginalRequest, opts ...grpc.CallOption) (*GetOriginalResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetOriginalResponse)
	err := c.cc.Invoke(ctx, Shortener_GetOriginal_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shortenerClient) BatchShorten(ctx context.Context, in *BatchShortenRequest, opts ...grpc.CallOption) (*BatchShortenResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(BatchShortenResponse)
	err := c.cc.Invoke(ctx, Shortener_BatchShorten_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shortenerClient) GetUserURLs(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*GetUserURLsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetUserURLsResponse)
	err := c.cc.Invoke(ctx, Shortener_GetUserURLs_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shortenerClient) DeleteUserURLs(ctx context.Context, in *DeleteUserURLsRequest, opts ...grpc.CallOption) (*Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Empty)
	err := c.cc.Invoke(ctx, Shortener_DeleteUserURLs_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shortenerClient) Ping(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Empty)
	err := c.cc.Invoke(ctx, Shortener_Ping_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shortenerClient) InternalStats(ctx context.Context, in *InternalStatsRequest, opts ...grpc.CallOption) (*InternalStatsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(InternalStatsResponse)
	err := c.cc.Invoke(ctx, Shortener_InternalStats_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ShortenerServer is the server API for Shortener service.
// All implementations must embed UnimplementedShortenerServer
// for forward compatibility
type ShortenerServer interface {
	Shorten(context.Context, *ShortenRequest) (*ShortenResponse, error)
	GetOriginal(context.Context, *GetOriginalRequest) (*GetOriginalResponse, error)
	BatchShorten(context.Context, *BatchShortenRequest) (*BatchShortenResponse, error)
	GetUserURLs(context.Context, *Empty) (*GetUserURLsResponse, error)
	DeleteUserURLs(context.Context, *DeleteUserURLsRequest) (*Empty, error)
	Ping(context.Context, *Empty) (*Empty, error)
	InternalStats(context.Context, *InternalStatsRequest) (*InternalStatsResponse, error)
	mustEmbedUnimplementedShortenerServer()
}

// UnimplementedShortenerServer must be embedded to have forward compatible implementations.
type UnimplementedShortenerServer struct {
}

func (UnimplementedShortenerServer) Shorten(context.Context, *ShortenRequest) (*ShortenResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Shorten not implemented")
}
func (UnimplementedShortenerServer) GetOriginal(context.Context, *GetOriginalRequest) (*GetOriginalResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetOriginal not implemented")
}
func (UnimplementedShortenerServer) BatchShorten(context.Context, *BatchShortenRequest) (*BatchShortenResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BatchShorten not implemented")
}
func (UnimplementedShortenerServer) GetUserURLs(context.Context, *Empty) (*GetUserURLsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUserURLs not implemented")
}
func (UnimplementedShortenerServer) DeleteUserURLs(context.Context, *DeleteUserURLsRequest) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteUserURLs not implemented")
}
func (UnimplementedShortenerServer) Ping(context.Context, *Empty) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}
func (UnimplementedShortenerServer) InternalStats(context.Context, *InternalStatsRequest) (*InternalStatsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method InternalStats not implemented")
}
func (UnimplementedShortenerServer) mustEmbedUnimplementedShortenerServer() {}

// UnsafeShortenerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ShortenerServer will
// result in compilation errors.
type UnsafeShortenerServer interface {
	mustEmbedUnimplementedShortenerServer()
}

func RegisterShortenerServer(s grpc.ServiceRegistrar, srv ShortenerServer) {
	s.RegisterService(&Shortener_ServiceDesc, srv)
}

func _Shortener_Shorten_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ShortenRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortenerServer).Shorten(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Shortener_Shorten_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortenerServer).Shorten(ctx, req.(*ShortenRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Shortener_GetOriginal_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetOriginalRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortenerServer).GetOriginal(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Shortener_GetOriginal_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortenerServer).GetOriginal(ctx, req.(*GetOriginalRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Shortener_BatchShorten_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BatchShortenRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortenerServer).BatchShorten(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Shortener_BatchShorten_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortenerServer).BatchShorten(ctx, req.(*BatchShortenRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Shortener_GetUserURLs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortenerServer).GetUserURLs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Shortener_GetUserURLs_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortenerServer).GetUserURLs(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _Shortener_DeleteUserURLs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteUserURLsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortenerServer).DeleteUserURLs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Shortener_DeleteUserURLs_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortenerServer).DeleteUserURLs(ctx, req.(*DeleteUserURLsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Shortener_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortenerServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Shortener_Ping_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortenerServer).Ping(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _Shortener_InternalStats_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(InternalStatsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortenerServer).InternalStats(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Shortener_InternalStats_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortenerServer).InternalStats(ctx, req.(*InternalStatsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Shortener_ServiceDesc is the grpc.ServiceDesc for Shortener service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Shortener_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "shortener.Shortener",
	HandlerType: (*ShortenerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Shorten",
			Handler:    _Shortener_Shorten_Handler,
		},
		{
			MethodName: "GetOriginal",
			Handler:    _Shortener_GetOriginal_Handler,
		},
		{
			MethodName: "BatchShorten",
			Handler:    _Shortener_BatchShorten_Handler,
		},
		{
			MethodName: "GetUserURLs",
			Handler:    _Shortener_GetUserURLs_Handler,
		},
		{
			MethodName: "DeleteUserURLs",
			Handler:    _Shortener_DeleteUserURLs_Handler,
		},
		{
			MethodName: "Ping",
			Handler:    _Shortener_Ping_Handler,
		},
		{
			MethodName: "InternalStats",
			Handler:    _Shortener_InternalStats_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "shortener.proto",
}

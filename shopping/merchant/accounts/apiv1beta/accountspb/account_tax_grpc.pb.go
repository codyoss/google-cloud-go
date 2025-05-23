// Copyright 2025 Google LLC
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

// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.25.7
// source: google/shopping/merchant/accounts/v1beta/account_tax.proto

package accountspb

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

const (
	AccountTaxService_GetAccountTax_FullMethodName    = "/google.shopping.merchant.accounts.v1beta.AccountTaxService/GetAccountTax"
	AccountTaxService_ListAccountTax_FullMethodName   = "/google.shopping.merchant.accounts.v1beta.AccountTaxService/ListAccountTax"
	AccountTaxService_UpdateAccountTax_FullMethodName = "/google.shopping.merchant.accounts.v1beta.AccountTaxService/UpdateAccountTax"
)

// AccountTaxServiceClient is the client API for AccountTaxService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AccountTaxServiceClient interface {
	// Returns the tax rules that match the conditions of GetAccountTaxRequest
	GetAccountTax(ctx context.Context, in *GetAccountTaxRequest, opts ...grpc.CallOption) (*AccountTax, error)
	// Lists the tax settings of the sub-accounts only in your
	// Merchant Center account.
	// This method can only be called on a multi-client account, otherwise it'll
	// return an error.
	ListAccountTax(ctx context.Context, in *ListAccountTaxRequest, opts ...grpc.CallOption) (*ListAccountTaxResponse, error)
	// Updates the tax settings of the account.
	UpdateAccountTax(ctx context.Context, in *UpdateAccountTaxRequest, opts ...grpc.CallOption) (*AccountTax, error)
}

type accountTaxServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewAccountTaxServiceClient(cc grpc.ClientConnInterface) AccountTaxServiceClient {
	return &accountTaxServiceClient{cc}
}

func (c *accountTaxServiceClient) GetAccountTax(ctx context.Context, in *GetAccountTaxRequest, opts ...grpc.CallOption) (*AccountTax, error) {
	out := new(AccountTax)
	err := c.cc.Invoke(ctx, AccountTaxService_GetAccountTax_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *accountTaxServiceClient) ListAccountTax(ctx context.Context, in *ListAccountTaxRequest, opts ...grpc.CallOption) (*ListAccountTaxResponse, error) {
	out := new(ListAccountTaxResponse)
	err := c.cc.Invoke(ctx, AccountTaxService_ListAccountTax_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *accountTaxServiceClient) UpdateAccountTax(ctx context.Context, in *UpdateAccountTaxRequest, opts ...grpc.CallOption) (*AccountTax, error) {
	out := new(AccountTax)
	err := c.cc.Invoke(ctx, AccountTaxService_UpdateAccountTax_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AccountTaxServiceServer is the server API for AccountTaxService service.
// All implementations should embed UnimplementedAccountTaxServiceServer
// for forward compatibility
type AccountTaxServiceServer interface {
	// Returns the tax rules that match the conditions of GetAccountTaxRequest
	GetAccountTax(context.Context, *GetAccountTaxRequest) (*AccountTax, error)
	// Lists the tax settings of the sub-accounts only in your
	// Merchant Center account.
	// This method can only be called on a multi-client account, otherwise it'll
	// return an error.
	ListAccountTax(context.Context, *ListAccountTaxRequest) (*ListAccountTaxResponse, error)
	// Updates the tax settings of the account.
	UpdateAccountTax(context.Context, *UpdateAccountTaxRequest) (*AccountTax, error)
}

// UnimplementedAccountTaxServiceServer should be embedded to have forward compatible implementations.
type UnimplementedAccountTaxServiceServer struct {
}

func (UnimplementedAccountTaxServiceServer) GetAccountTax(context.Context, *GetAccountTaxRequest) (*AccountTax, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAccountTax not implemented")
}
func (UnimplementedAccountTaxServiceServer) ListAccountTax(context.Context, *ListAccountTaxRequest) (*ListAccountTaxResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListAccountTax not implemented")
}
func (UnimplementedAccountTaxServiceServer) UpdateAccountTax(context.Context, *UpdateAccountTaxRequest) (*AccountTax, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateAccountTax not implemented")
}

// UnsafeAccountTaxServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AccountTaxServiceServer will
// result in compilation errors.
type UnsafeAccountTaxServiceServer interface {
	mustEmbedUnimplementedAccountTaxServiceServer()
}

func RegisterAccountTaxServiceServer(s grpc.ServiceRegistrar, srv AccountTaxServiceServer) {
	s.RegisterService(&AccountTaxService_ServiceDesc, srv)
}

func _AccountTaxService_GetAccountTax_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetAccountTaxRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccountTaxServiceServer).GetAccountTax(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AccountTaxService_GetAccountTax_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccountTaxServiceServer).GetAccountTax(ctx, req.(*GetAccountTaxRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AccountTaxService_ListAccountTax_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListAccountTaxRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccountTaxServiceServer).ListAccountTax(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AccountTaxService_ListAccountTax_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccountTaxServiceServer).ListAccountTax(ctx, req.(*ListAccountTaxRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AccountTaxService_UpdateAccountTax_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateAccountTaxRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccountTaxServiceServer).UpdateAccountTax(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AccountTaxService_UpdateAccountTax_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccountTaxServiceServer).UpdateAccountTax(ctx, req.(*UpdateAccountTaxRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// AccountTaxService_ServiceDesc is the grpc.ServiceDesc for AccountTaxService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var AccountTaxService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "google.shopping.merchant.accounts.v1beta.AccountTaxService",
	HandlerType: (*AccountTaxServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetAccountTax",
			Handler:    _AccountTaxService_GetAccountTax_Handler,
		},
		{
			MethodName: "ListAccountTax",
			Handler:    _AccountTaxService_ListAccountTax_Handler,
		},
		{
			MethodName: "UpdateAccountTax",
			Handler:    _AccountTaxService_UpdateAccountTax_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "google/shopping/merchant/accounts/v1beta/account_tax.proto",
}

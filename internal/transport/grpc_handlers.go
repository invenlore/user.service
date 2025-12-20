package transport

import (
	"context"

	"github.com/invenlore/proto/pkg/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func (s *GRPCUserServer) HealthCheck(
	ctx context.Context,
	req *user.HealthRequest,
) (
	*user.HealthResponse,
	error,
) {
	resp := &user.HealthResponse{Status: "up"}

	return resp, nil
}

func (s *GRPCUserServer) AddUser(
	ctx context.Context,
	req *user.AddUserRequest,
) (
	*user.AddUserResponse,
	error,
) {
	ptrUser, code, err := s.svc.AddUser(ctx)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}

	resp := &user.AddUserResponse{User: ptrUser}
	return resp, err
}

func (s *GRPCUserServer) GetUser(
	ctx context.Context,
	req *user.GetUserRequest,
) (
	*user.GetUserResponse,
	error,
) {
	ptrUser, code, err := s.svc.GetUser(ctx, req.Id)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}

	resp := &user.GetUserResponse{User: ptrUser}
	return resp, err
}

func (s *GRPCUserServer) DeleteUser(
	ctx context.Context,
	req *user.DeleteUserRequest,
) (
	*user.DeleteUserResponse,
	error,
) {
	code, err := s.svc.DeleteUser(ctx, req.Id)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}

	resp := &user.DeleteUserResponse{}
	return resp, err
}

func (s *GRPCUserServer) ListUsers(
	req *user.ListUsersRequest,
	srv grpc.ServerStreamingServer[user.ListUsersResponse],
) error {
	ctx := srv.Context()

	ptrsUsers, code, err := s.svc.ListUsers(ctx)
	if err != nil {
		return status.Error(code, err.Error())
	}

	for _, ptrUser := range ptrsUsers {
		err := srv.Send(&user.ListUsersResponse{User: ptrUser})

		if err != nil {
			return err
		}
	}

	return err
}

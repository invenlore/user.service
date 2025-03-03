package transport

import (
	"context"
	"net"

	"github.com/google/uuid"
	user "github.com/invenlore/proto/proto/user"
	"github.com/invenlore/user.service/internal/service"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func StartGRPCServer(listenAddr string, svc service.UserService) error {
	logrus.Info("starting gRPC server on ", listenAddr)

	grpcUserServer := NewGRPCUserServer(svc)

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		logrus.Fatalln(err)

		return err
	}

	opts := []grpc.ServerOption{}
	server := grpc.NewServer(opts...)
	user.RegisterUserServiceServer(server, grpcUserServer)

	err = server.Serve(ln)
	if err != nil {
		logrus.Fatalln(err)

		return err
	}

	return err
}

type GRPCUserServer struct {
	svc service.UserService
	user.UnimplementedUserServiceServer
}

func NewGRPCUserServer(svc service.UserService) *GRPCUserServer {
	return &GRPCUserServer{
		svc: svc,
	}
}

func (s *GRPCUserServer) AddUser(ctx context.Context, req *user.AddUserRequest) (*user.AddUserResponse, error) {
	ctx = context.WithValue(ctx, "requestID", uuid.New().String())

	ptrUser, code, err := s.svc.AddUser(ctx)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}

	resp := &user.AddUserResponse{User: ptrUser}
	return resp, err
}

func (s *GRPCUserServer) GetUser(ctx context.Context, req *user.GetUserRequest) (*user.GetUserResponse, error) {
	ctx = context.WithValue(ctx, "requestID", uuid.New().String())

	ptrUser, code, err := s.svc.GetUser(ctx, req.Id)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}

	resp := &user.GetUserResponse{User: ptrUser}
	return resp, err
}

func (s *GRPCUserServer) ListUsers(req *user.ListUsersRequest, srv grpc.ServerStreamingServer[user.ListUsersResponse]) error {
	ctx := context.WithValue(srv.Context(), "requestID", uuid.New().String())

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

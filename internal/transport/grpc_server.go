package transport

import (
	"context"
	"fmt"
	"net"

	"github.com/invenlore/proto/pkg/user"
	"github.com/invenlore/user.service/internal/service"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

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
	ptrUser, code, err := s.svc.AddUser(ctx)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}

	resp := &user.AddUserResponse{User: ptrUser}
	return resp, err
}

func (s *GRPCUserServer) GetUser(ctx context.Context, req *user.GetUserRequest) (*user.GetUserResponse, error) {
	ptrUser, code, err := s.svc.GetUser(ctx, req.Id)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}

	resp := &user.GetUserResponse{User: ptrUser}
	return resp, err
}

func (s *GRPCUserServer) DeleteUser(ctx context.Context, req *user.DeleteUserRequest) (*user.DeleteUserResponse, error) {
	code, err := s.svc.DeleteUser(ctx, req.Id)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}

	resp := &user.DeleteUserResponse{}
	return resp, err
}

func (s *GRPCUserServer) ListUsers(req *user.ListUsersRequest, srv grpc.ServerStreamingServer[user.ListUsersResponse]) error {
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

func StartGRPCServer(listenAddr string, svc service.UserService, errChan chan error) (*grpc.Server, error) {
	logrus.Info("starting gRPC server on ", listenAddr)

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return nil, fmt.Errorf("gRPC server failed to listen on %s: %w", listenAddr, err)
	}

	grpcUserServer := NewGRPCUserServer(svc)

	unaryInterceptors := []grpc.UnaryServerInterceptor{
		RequestIDInterceptor,
		LoggingInterceptor,
	}

	streamInterceptors := []grpc.StreamServerInterceptor{
		StreamRequestIDInterceptor,
		StreamLoggingInterceptor,
	}

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(unaryInterceptors...),
		grpc.ChainStreamInterceptor(streamInterceptors...),
	)

	user.RegisterUserServiceServer(server, grpcUserServer)

	go func() {
		logrus.Printf("gRPC server serving on %s", listenAddr)

		if serveErr := server.Serve(ln); serveErr != nil && serveErr != grpc.ErrServerStopped {
			errChan <- fmt.Errorf("gRPC server failed to serve: %w", serveErr)
		}
	}()

	return server, nil
}

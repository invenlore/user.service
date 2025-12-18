package transport

import (
	"context"
	"fmt"
	"net"

	"github.com/invenlore/core/pkg/config"
	"github.com/invenlore/core/pkg/logger"
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

func (s *GRPCUserServer) HealthCheck(ctx context.Context, req *user.HealthRequest) (*user.HealthResponse, error) {
	resp := &user.HealthResponse{Status: "up"}

	return resp, nil
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

func StartGRPCServer(ctx context.Context, cfg *config.ServerConfig, svc service.UserService, errChan chan error) (*grpc.Server, net.Listener, error) {
	listenAddr := net.JoinHostPort(cfg.GRPC.Host, cfg.GRPC.Port)
	logrus.Info("starting gRPC server on ", listenAddr)

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return nil, nil, fmt.Errorf("gRPC server failed to listen on %s: %w", listenAddr, err)
	}

	grpcUserServer := NewGRPCUserServer(svc)

	unaryInterceptors := []grpc.UnaryServerInterceptor{
		logger.ServerRequestIDInterceptor,
		logger.ServerLoggingInterceptor,
	}

	streamInterceptors := []grpc.StreamServerInterceptor{
		logger.ServerStreamRequestIDInterceptor,
		logger.ServerStreamLoggingInterceptor,
	}

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(unaryInterceptors...),
		grpc.ChainStreamInterceptor(streamInterceptors...),
	)

	user.RegisterUserServiceServer(server, grpcUserServer)

	go func() {
		logrus.Infof("gRPC server serving on %s", listenAddr)

		if serveErr := server.Serve(ln); serveErr != nil && serveErr != grpc.ErrServerStopped {
			errChan <- fmt.Errorf("gRPC server failed to serve: %w", serveErr)
		}
	}()

	return server, ln, nil
}

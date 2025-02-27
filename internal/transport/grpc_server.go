package transport

import (
	"context"
	"net"

	"github.com/google/uuid"
	proto "github.com/invenlore/proto/user/gen/go/user"
	"github.com/invenlore/user.service/internal/service"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func MakeGRPCServerAndRun(listenAddr string, svc service.UserService) error {
	logrus.Info("Starting gRPC server on ", listenAddr)

	grpcUserServer := NewGRPCUserServer(svc)

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}

	opts := []grpc.ServerOption{}
	server := grpc.NewServer(opts...)
	proto.RegisterUserServiceServer(server, grpcUserServer)

	return server.Serve(ln)
}

type GRPCUserServer struct {
	svc service.UserService
	proto.UnimplementedUserServiceServer
}

func NewGRPCUserServer(svc service.UserService) *GRPCUserServer {
	return &GRPCUserServer{
		svc: svc,
	}
}

func (s *GRPCUserServer) GetUser(ctx context.Context, req *proto.UserRequest) (*proto.UserResponse, error) {
	ctx = context.WithValue(ctx, "requestID", uuid.New().String())

	email, err := s.svc.GetUser(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	resp := &proto.UserResponse{
		Id:    req.Id,
		Email: email,
	}

	return resp, err
}

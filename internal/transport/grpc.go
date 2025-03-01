package transport

import (
	"context"
	"net"

	"github.com/google/uuid"
	proto "github.com/invenlore/proto/user/gen/go/user"
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
	proto.RegisterUserServiceServer(server, grpcUserServer)

	err = server.Serve(ln)
	if err != nil {
		logrus.Fatalln(err)

		return err
	}

	return err
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

	email, code, err := s.svc.GetUser(ctx, req.Id)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}

	resp := &proto.UserResponse{
		Id:    req.Id,
		Email: email,
	}

	return resp, err
}

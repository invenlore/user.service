package server

import (
	"context"
	"math/rand"
	"net"

	"github.com/invenlore/proto/user/gen/go/user"
	"github.com/invenlore/user.service/internal/service"
	"google.golang.org/grpc"
)

func MakeGRPCServerAndRun(listenAddr string, svc service.UserService) error {
	grpcUserFetcher := NewGRPCUserFetcherServer(svc)

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}

	opts := []grpc.ServerOption{}
	server := grpc.NewServer(opts...)
	user.RegisterUserServiceServer(server, grpcUserFetcher)

	return server.Serve(ln)
}

type GRPCUserFetcherServer struct {
	svc service.UserService
	user.UnimplementedUserServiceServer
}

func NewGRPCUserFetcherServer(svc service.UserService) *GRPCUserFetcherServer {
	return &GRPCUserFetcherServer{
		svc: svc,
	}
}

func (s *GRPCUserFetcherServer) GetUser(ctx context.Context, req *user.UserRequest) (*user.UserResponse, error) {
	reqid := rand.Intn(10000)
	ctx = context.WithValue(ctx, "requestID", reqid)

	email, err := s.svc.GetUser(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	resp := &user.UserResponse{
		Id:    req.Id,
		Email: email,
	}

	return resp, err
}

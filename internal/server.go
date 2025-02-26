package main

import (
	"context"
	"math/rand"
	"net"

	"github.com/invenlore/proto"
	"google.golang.org/grpc"
)

func makeGRPCServerAndRun(listenAddr string, svc proto.UserService) error {
	grpcUserFetcher := NewGRPCUserFetcherServer(svc)

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}

	opts := []grpc.ServerOption{}
	server := grpc.NewServer(opts...)
	proto.RegisterUserFetcherServer(server, grpcUserFetcher)

	return server.Serve(ln)
}

type GRPCUserFetcherServer struct {
	svc UserService
	proto.UnimplementedUserFetcherServer
}

func NewGRPCUserFetcherServer(svc UserService) *GRPCUserFetcherServer {
	return &GRPCUserFetcherServer{
		svc: svc,
	}
}

func (s *GRPCUserFetcherServer) GetUser(ctx context.Context, req *proto.UserRequest) (*proto.UserResponse, error) {
	reqid := rand.Intn(10000)
	ctx = context.WithValue(ctx, "requestID", reqid)

	email, err := s.svc.GetUser(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	resp := &proto.PriceResponse{
		Id: req.Id,
		Email: email,
	}

	return resp, err
}

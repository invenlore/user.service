package transport

import (
	"context"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/invenlore/proto/pkg/user"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var v = validator.New(validator.WithRequiredStructEnabled())

type addUserInput struct {
	Name  string `validate:"required,min=1,max=100"`
	Email string `validate:"required,email,max=254"`
}

func (s *GRPCUserServer) HealthCheck(
	ctx context.Context,
	req *user.HealthRequest,
) (*user.HealthResponse, error) {
	if s.mongoClient == nil {
		return nil, status.Error(codes.Internal, "MongoDB client is not initialized")
	}

	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := s.mongoClient.Ping(pingCtx, readpref.Primary()); err != nil {
		return nil, status.Error(codes.Unavailable, "MongoDB unavailable: "+err.Error())
	}

	return &user.HealthResponse{Status: "up"}, nil
}

func (s *GRPCUserServer) AddUser(
	ctx context.Context,
	req *user.AddUserRequest,
) (*user.AddUserResponse, error) {
	if req == nil || req.User == nil {
		return nil, status.Error(codes.InvalidArgument, "user is required")
	}

	in := addUserInput{
		Name:  strings.TrimSpace(req.User.Name),
		Email: strings.TrimSpace(req.User.Email),
	}
	if err := v.Struct(in); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	req.User.Name = in.Name
	req.User.Email = in.Email

	lastInsertId, code, err := s.svc.AddUser(ctx, req.User)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}

	return &user.AddUserResponse{Id: lastInsertId}, nil
}

func (s *GRPCUserServer) GetUser(
	ctx context.Context,
	req *user.GetUserRequest,
) (*user.GetUserResponse, error) {
	if req == nil || strings.TrimSpace(req.Id) == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	ptrUser, code, err := s.svc.GetUser(ctx, strings.TrimSpace(req.Id))
	if err != nil {
		return nil, status.Error(code, err.Error())
	}

	return &user.GetUserResponse{User: ptrUser}, nil
}

func (s *GRPCUserServer) DeleteUser(
	ctx context.Context,
	req *user.DeleteUserRequest,
) (*user.DeleteUserResponse, error) {
	if req == nil || strings.TrimSpace(req.Id) == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	code, err := s.svc.DeleteUser(ctx, strings.TrimSpace(req.Id))
	if err != nil {
		return nil, status.Error(code, err.Error())
	}

	return &user.DeleteUserResponse{}, nil
}

// TODO: ListUsers -> StreamUsers
func (s *GRPCUserServer) ListUsers(
	req *user.ListUsersRequest,
	srv grpc.ServerStreamingServer[user.ListUsersResponse],
) error {
	ctx := srv.Context()

	code, err := s.svc.StreamUsers(ctx, func(u *user.User) error {
		return srv.Send(&user.ListUsersResponse{User: u})
	})

	if err != nil {
		return status.Error(code, err.Error())
	}

	return nil
}

package transport

import (
	"context"
	"strings"

	"github.com/go-playground/validator/v10"
	identity_v1 "github.com/invenlore/proto/pkg/identity/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var v = validator.New(validator.WithRequiredStructEnabled())

type addUserInput struct {
	Name  string `validate:"required,min=1,max=100"`
	Email string `validate:"required,email,max=254"`
}

func (s *GRPCUserServer) HealthCheck(ctx context.Context, req *identity_v1.HealthRequest) (*identity_v1.HealthResponse, error) {
	if !s.mongoReadiness.Ready() {
		return &identity_v1.HealthResponse{Status: "down"}, status.Error(codes.Unavailable, s.mongoReadiness.LastError())
	}

	return &identity_v1.HealthResponse{Status: "up"}, nil
}

func (s *GRPCUserServer) AddUser(ctx context.Context, req *identity_v1.AddUserRequest) (*identity_v1.AddUserResponse, error) {
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

	return &identity_v1.AddUserResponse{Id: lastInsertId}, nil
}

func (s *GRPCUserServer) GetUser(ctx context.Context, req *identity_v1.GetUserRequest) (*identity_v1.GetUserResponse, error) {
	if req == nil || strings.TrimSpace(req.Id) == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	ptrUser, code, err := s.svc.GetUser(ctx, strings.TrimSpace(req.Id))
	if err != nil {
		return nil, status.Error(code, err.Error())
	}

	return &identity_v1.GetUserResponse{User: ptrUser}, nil
}

func (s *GRPCUserServer) DeleteUser(ctx context.Context, req *identity_v1.DeleteUserRequest) (*identity_v1.DeleteUserResponse, error) {
	if req == nil || strings.TrimSpace(req.Id) == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	code, err := s.svc.DeleteUser(ctx, strings.TrimSpace(req.Id))
	if err != nil {
		return nil, status.Error(code, err.Error())
	}

	return &identity_v1.DeleteUserResponse{}, nil
}

// TODO: ListUsers -> StreamUsers
func (s *GRPCUserServer) ListUsers(req *identity_v1.ListUsersRequest, srv grpc.ServerStreamingServer[identity_v1.ListUsersResponse]) error {
	ctx := srv.Context()

	code, err := s.svc.StreamUsers(ctx, func(u *identity_v1.User) error {
		return srv.Send(&identity_v1.ListUsersResponse{User: u})
	})

	if err != nil {
		return status.Error(code, err.Error())
	}

	return nil
}

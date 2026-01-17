package transport

import (
	"context"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/invenlore/proto/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var v = validator.New(validator.WithRequiredStructEnabled())

type addUserInput struct {
	Name  string `validate:"required,min=1,max=100"`
	Email string `validate:"required,email,max=254"`
}

func (s *GRPCUserServer) HealthCheck(ctx context.Context, req *identity.HealthRequest) (*identity.HealthResponse, error) {
	if !s.mongoReadiness.Ready() {
		return &identity.HealthResponse{Status: "down"}, status.Error(codes.Unavailable, s.mongoReadiness.LastError())
	}

	return &identity.HealthResponse{Status: "up"}, nil
}

func (s *GRPCUserServer) AddUser(ctx context.Context, req *identity.AddUserRequest) (*identity.AddUserResponse, error) {
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

	return &identity.AddUserResponse{Id: lastInsertId}, nil
}

func (s *GRPCUserServer) GetUser(ctx context.Context, req *identity.GetUserRequest) (*identity.GetUserResponse, error) {
	if req == nil || strings.TrimSpace(req.Id) == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	ptrUser, code, err := s.svc.GetUser(ctx, strings.TrimSpace(req.Id))
	if err != nil {
		return nil, status.Error(code, err.Error())
	}

	return &identity.GetUserResponse{User: ptrUser}, nil
}

func (s *GRPCUserServer) DeleteUser(ctx context.Context, req *identity.DeleteUserRequest) (*identity.DeleteUserResponse, error) {
	if req == nil || strings.TrimSpace(req.Id) == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	code, err := s.svc.DeleteUser(ctx, strings.TrimSpace(req.Id))
	if err != nil {
		return nil, status.Error(code, err.Error())
	}

	return &identity.DeleteUserResponse{}, nil
}

// TODO: ListUsers -> StreamUsers
func (s *GRPCUserServer) ListUsers(req *identity.ListUsersRequest, srv grpc.ServerStreamingServer[identity.ListUsersResponse]) error {
	ctx := srv.Context()

	code, err := s.svc.StreamUsers(ctx, func(u *identity.User) error {
		return srv.Send(&identity.ListUsersResponse{User: u})
	})

	if err != nil {
		return status.Error(code, err.Error())
	}

	return nil
}

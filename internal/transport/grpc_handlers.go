package transport

import (
	"context"
	"strings"

	"github.com/go-playground/validator/v10"
	common_v1 "github.com/invenlore/proto/pkg/common/v1"
	identity_v1 "github.com/invenlore/proto/pkg/identity/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var v = validator.New(validator.WithRequiredStructEnabled())

type addUserInput struct {
	Name  string `validate:"required,min=1,max=100"`
	Email string `validate:"required,email,max=254"`
}

// PUBLIC SCOPE
func (s *GRPCIdentityServer) Register(ctx context.Context, req *identity_v1.RegisterRequest) (*identity_v1.RegisterResponse, error) {
	return nil, status.Error(codes.Unimplemented, "register is not implemented")
}

// PUBLIC SCOPE
func (s *GRPCIdentityServer) Login(ctx context.Context, req *identity_v1.LoginRequest) (*identity_v1.LoginResponse, error) {
	return nil, status.Error(codes.Unimplemented, "login is not implemented")
}

// PUBLIC SCOPE
func (s *GRPCIdentityServer) Refresh(ctx context.Context, req *identity_v1.RefreshRequest) (*identity_v1.RefreshResponse, error) {
	return nil, status.Error(codes.Unimplemented, "refresh is not implemented")
}

// PUBLIC SCOPE
func (s *GRPCIdentityServer) Logout(ctx context.Context, req *identity_v1.LogoutRequest) (*identity_v1.LogoutResponse, error) {
	return nil, status.Error(codes.Unimplemented, "logout is not implemented")
}

// PUBLIC SCOPE
func (s *GRPCIdentityServer) GetProfile(ctx context.Context, req *identity_v1.GetProfileRequest) (*identity_v1.GetProfileResponse, error) {
	return nil, status.Error(codes.Unimplemented, "get profile is not implemented")
}

// PUBLIC SCOPE
func (s *GRPCIdentityServer) UpdateProfile(ctx context.Context, req *identity_v1.UpdateProfileRequest) (*identity_v1.UpdateProfileResponse, error) {
	return nil, status.Error(codes.Unimplemented, "update profile is not implemented")
}

// INTERNAL SCOPE
func (s *GRPCIdentityServer) HealthCheck(ctx context.Context, req *common_v1.ServiceHealthRequest) (*common_v1.ServiceHealthResponse, error) {
	if !s.mongoReadiness.Ready() {
		return &common_v1.ServiceHealthResponse{Status: "down"}, status.Error(codes.Unavailable, s.mongoReadiness.LastError())
	}

	return &common_v1.ServiceHealthResponse{Status: "up"}, nil
}

// INTERNAL SCOPE
func (s *GRPCIdentityServer) GetJWKS(ctx context.Context, req *identity_v1.GetJWKSRequest) (*identity_v1.GetJWKSResponse, error) {
	return nil, status.Error(codes.Unimplemented, "get jwks is not implemented")
}

// INTERNAL SCOPE
func (s *GRPCIdentityServer) ValidateToken(ctx context.Context, req *identity_v1.ValidateTokenRequest) (*identity_v1.ValidateTokenResponse, error) {
	return nil, status.Error(codes.Unimplemented, "validate token is not implemented")
}

// INTERNAL SCOPE
func (s *GRPCIdentityServer) Authorize(ctx context.Context, req *identity_v1.AuthorizeRequest) (*identity_v1.AuthorizeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "authorize is not implemented")
}

// INTERNAL SCOPE
func (s *GRPCIdentityServer) GetUserBrief(ctx context.Context, req *identity_v1.GetUserBriefRequest) (*identity_v1.GetUserBriefResponse, error) {
	return nil, status.Error(codes.Unimplemented, "get user brief is not implemented")
}

// ADMIN SCOPE
func (s *GRPCIdentityServer) AddUser(ctx context.Context, req *identity_v1.AddUserRequest) (*identity_v1.AddUserResponse, error) {
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

	lastInsertId, code, err := s.adminSvc.AddUser(ctx, req.User)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}

	return &identity_v1.AddUserResponse{Id: lastInsertId}, nil
}

// ADMIN SCOPE
func (s *GRPCIdentityServer) GetUser(ctx context.Context, req *identity_v1.GetUserRequest) (*identity_v1.GetUserResponse, error) {
	if req == nil || strings.TrimSpace(req.Id) == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	ptrUser, code, err := s.adminSvc.GetUser(ctx, strings.TrimSpace(req.Id))
	if err != nil {
		return nil, status.Error(code, err.Error())
	}

	return &identity_v1.GetUserResponse{User: ptrUser}, nil
}

// ADMIN SCOPE
func (s *GRPCIdentityServer) DeleteUser(ctx context.Context, req *identity_v1.DeleteUserRequest) (*identity_v1.DeleteUserResponse, error) {
	if req == nil || strings.TrimSpace(req.Id) == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	code, err := s.adminSvc.DeleteUser(ctx, strings.TrimSpace(req.Id))
	if err != nil {
		return nil, status.Error(code, err.Error())
	}

	return &identity_v1.DeleteUserResponse{}, nil
}

// ADMIN SCOPE
func (s *GRPCIdentityServer) ListUsers(ctx context.Context, req *identity_v1.ListUsersRequest) (*identity_v1.ListUsersResponse, error) {
	users, nextToken, code, err := s.adminSvc.ListUsers(ctx)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}

	return &identity_v1.ListUsersResponse{Users: users, NextPageToken: nextToken}, nil
}

package service

import (
	"context"
	"fmt"

	"github.com/invenlore/identity.service/internal/domain"
	"github.com/invenlore/identity.service/internal/repository"
	identity_v1 "github.com/invenlore/proto/pkg/identity/v1"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/codes"
)

type identityService struct {
	Repository repository.IdentityRepository
}

type IdentityService interface {
	AddUser(context.Context, *identity_v1.User) (string, codes.Code, error)
	GetUser(context.Context, string) (*identity_v1.User, codes.Code, error)
	DeleteUser(context.Context, string) (codes.Code, error)
	StreamUsers(ctx context.Context, send func(*identity_v1.User) error) (codes.Code, error)
}

func NewIdentityService(repository repository.IdentityRepository) IdentityService {
	return &identityService{Repository: repository}
}

func (s *identityService) AddUser(ctx context.Context, u *identity_v1.User) (string, codes.Code, error) {
	lastInsertId, err := s.Repository.InsertUser(ctx, &domain.User{
		Name:  u.Name,
		Email: u.Email,
	})

	if err != nil {
		return "", codes.Internal, err
	}

	return lastInsertId.Hex(), codes.OK, nil
}

func (s *identityService) GetUser(ctx context.Context, id string) (*identity_v1.User, codes.Code, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, codes.InvalidArgument, fmt.Errorf("converting string to ObjectID failed: %v", err)
	}

	ptrUser, err := s.Repository.FindOneUser(ctx, objID)

	if err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			return nil, codes.NotFound, fmt.Errorf("user for id (%s) is not found", id)
		default:
			return nil, codes.Internal, err
		}
	}

	return &identity_v1.User{
		Id:    ptrUser.Id.Hex(),
		Name:  ptrUser.Name,
		Email: ptrUser.Email,
	}, codes.OK, nil
}

func (s *identityService) DeleteUser(ctx context.Context, id string) (codes.Code, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return codes.InvalidArgument, fmt.Errorf("converting string to ObjectID failed: %v", err)
	}

	deletedCount, err := s.Repository.DeleteOneUser(ctx, objID)
	if err != nil {
		return codes.Internal, err
	}

	if deletedCount == 0 {
		return codes.NotFound, fmt.Errorf("user for id (%s) is not found", id)
	}

	return codes.OK, nil
}

func (s *identityService) StreamUsers(ctx context.Context, send func(*identity_v1.User) error) (codes.Code, error) {
	err := s.Repository.StreamAllUsers(ctx, func(u *domain.User) error {
		return send(&identity_v1.User{
			Id:    u.Id.Hex(),
			Name:  u.Name,
			Email: u.Email,
		})
	})

	if err != nil {
		// TODO: Check for mongo/stream error - context cancel
		return codes.Internal, err
	}

	return codes.OK, nil
}

package service

import (
	"context"
	"fmt"

	"github.com/invenlore/proto/pkg/user"
	"github.com/invenlore/user.service/internal/domain"
	"github.com/invenlore/user.service/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/codes"
)

type userService struct {
	Repository repository.UserRepository
}

type UserService interface {
	AddUser(context.Context, *user.User) (string, codes.Code, error)
	GetUser(context.Context, string) (*user.User, codes.Code, error)
	DeleteUser(context.Context, string) (codes.Code, error)
	StreamUsers(ctx context.Context, send func(*user.User) error) (codes.Code, error)
}

func NewUserService(repository repository.UserRepository) UserService {
	return &userService{Repository: repository}
}

func (s *userService) AddUser(ctx context.Context, u *user.User) (string, codes.Code, error) {
	lastInsertId, err := s.Repository.Insert(ctx, &domain.User{
		Name:  u.Name,
		Email: u.Email,
	})

	if err != nil {
		return "", codes.Internal, err
	}

	return lastInsertId.Hex(), codes.OK, nil
}

func (s *userService) GetUser(ctx context.Context, id string) (*user.User, codes.Code, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, codes.InvalidArgument, fmt.Errorf("converting string to ObjectID failed: %v", err)
	}

	ptrUser, err := s.Repository.FindOne(ctx, objID)

	if err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			return nil, codes.NotFound, fmt.Errorf("user for id (%s) is not found", id)
		default:
			return nil, codes.Internal, err
		}
	}

	return &user.User{
		Id:    ptrUser.Id.Hex(),
		Name:  ptrUser.Name,
		Email: ptrUser.Email,
	}, codes.OK, nil
}

func (s *userService) DeleteUser(ctx context.Context, id string) (codes.Code, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return codes.InvalidArgument, fmt.Errorf("converting string to ObjectID failed: %v", err)
	}

	deletedCount, err := s.Repository.DeleteOne(ctx, objID)
	if err != nil {
		return codes.Internal, err
	}

	if deletedCount == 0 {
		return codes.NotFound, fmt.Errorf("user for id (%s) is not found", id)
	}

	return codes.OK, nil
}

func (s *userService) StreamUsers(ctx context.Context, send func(*user.User) error) (codes.Code, error) {
	err := s.Repository.StreamAll(ctx, func(u *domain.User) error {
		return send(&user.User{
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

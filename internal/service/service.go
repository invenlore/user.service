package service

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"sync"

	"github.com/google/uuid"
	"github.com/invenlore/proto/pkg/user"
	"google.golang.org/grpc/codes"
)

var (
	// MOCK
	users = map[string]*user.User{
		"1": {
			Id:    "1",
			Email: "emailnew@email.com",
		},
	}

	mu sync.RWMutex
)

type UserService interface {
	AddUser(context.Context) (*user.User, codes.Code, error)
	GetUser(context.Context, string) (*user.User, codes.Code, error)
	DeleteUser(context.Context, string) (codes.Code, error)
	ListUsers(context.Context) ([]*user.User, codes.Code, error)
}

type UserServiceStruct struct{}

func (s *UserServiceStruct) AddUser(ctx context.Context) (*user.User, codes.Code, error) {
	mu.Lock()
	defer mu.Unlock()

	newId := uuid.NewString()

	users[newId] = &user.User{
		Id:    newId,
		Email: "added@email.com",
	}

	return users[newId], codes.OK, nil
}

func (s *UserServiceStruct) GetUser(ctx context.Context, id string) (*user.User, codes.Code, error) {
	mu.RLock()
	defer mu.RUnlock()

	ptrUser, ok := users[id]
	if !ok {
		return nil, codes.NotFound, fmt.Errorf("user for id (%s) is not found", id)
	}

	return ptrUser, codes.OK, nil
}

func (s *UserServiceStruct) DeleteUser(ctx context.Context, id string) (codes.Code, error) {
	mu.Lock()
	defer mu.Unlock()

	if _, ok := users[id]; ok {
		delete(users, id)

		return codes.OK, nil
	}

	return codes.NotFound, fmt.Errorf("user for id (%s) is not found", id)
}

func (s *UserServiceStruct) ListUsers(ctx context.Context) ([]*user.User, codes.Code, error) {
	mu.RLock()
	defer mu.RUnlock()

	return slices.Collect(maps.Values(users)), codes.OK, nil
}

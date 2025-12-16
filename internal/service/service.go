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

	mutex sync.RWMutex
)

type UserService interface {
	AddUser(context.Context) (*user.User, codes.Code, error)
	GetUser(context.Context, string) (*user.User, codes.Code, error)
	DeleteUser(context.Context, string) (codes.Code, error)
	ListUsers(context.Context) ([]*user.User, codes.Code, error)
}

type UserServiceStruct struct{}

func (s *UserServiceStruct) AddUser(_ context.Context) (*user.User, codes.Code, error) {
	mutex.Lock()
	defer mutex.Unlock()

	newId := uuid.NewString()

	users[newId] = &user.User{
		Id:    newId,
		Email: "added@email.com",
	}

	return users[newId], codes.OK, nil
}

func (s *UserServiceStruct) GetUser(_ context.Context, id string) (*user.User, codes.Code, error) {
	mutex.RLock()
	defer mutex.RUnlock()

	ptrUser, ok := users[id]
	if !ok {
		return nil, codes.NotFound, fmt.Errorf("user for id (%s) is not found", id)
	}

	return ptrUser, codes.OK, nil
}

func (s *UserServiceStruct) DeleteUser(_ context.Context, id string) (codes.Code, error) {
	mutex.Lock()
	defer mutex.Unlock()

	if _, ok := users[id]; ok {
		delete(users, id)

		return codes.OK, nil

	}

	return codes.NotFound, fmt.Errorf("user for id (%s) is not found", id)
}

func (s *UserServiceStruct) ListUsers(_ context.Context) ([]*user.User, codes.Code, error) {
	mutex.RLock()
	defer mutex.RUnlock()

	return slices.Collect(maps.Values(users)), codes.OK, nil
}

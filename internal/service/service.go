package service

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"time"

	user "github.com/invenlore/proto/proto/user"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
)

// MOCK
var users = map[string]*user.User{
	"1": {
		Id:    "1",
		Email: "emailnew@email.com",
	},
}

type UserService interface {
	AddUser(context.Context) (*user.User, codes.Code, error)
	GetUser(context.Context, string) (*user.User, codes.Code, error)
	ListUsers(context.Context) ([]*user.User, codes.Code, error)
}

type UserServiceStruct struct{}

func (s *UserServiceStruct) AddUser(_ context.Context) (*user.User, codes.Code, error) {
	var newId = strconv.Itoa(len(users) + 1)

	users[newId] = &user.User{
		Id:    newId,
		Email: "added@email.com",
	}

	ptrUser, ok := users[newId]
	if !ok {
		return nil, codes.NotFound, fmt.Errorf("Can't create User for id (%s)", newId)
	}

	return ptrUser, codes.OK, nil
}

func (s *UserServiceStruct) GetUser(_ context.Context, id string) (*user.User, codes.Code, error) {
	ptrUser, ok := users[id]
	if !ok {
		return nil, codes.NotFound, fmt.Errorf("User for id (%s) is not found", id)
	}

	return ptrUser, codes.OK, nil
}

func (s *UserServiceStruct) ListUsers(_ context.Context) ([]*user.User, codes.Code, error) {
	return slices.Collect(maps.Values(users)), codes.OK, nil
}

type LoggingServiceStruct struct {
	Next UserService
}

func (s LoggingServiceStruct) AddUser(ctx context.Context) (user *user.User, code codes.Code, err error) {
	defer func(begin time.Time) {
		reqID := ctx.Value("requestID")

		logrus.WithFields(logrus.Fields{
			"requestID":     reqID,
			"took":          time.Since(begin),
			"err":           err,
			"errCode":       code,
			"createdUserID": user.Id,
		}).Info("AddUser")
	}(time.Now())

	return s.Next.AddUser(ctx)
}

func (s LoggingServiceStruct) GetUser(ctx context.Context, id string) (user *user.User, code codes.Code, err error) {
	defer func(begin time.Time) {
		reqID := ctx.Value("requestID")

		logrus.WithFields(logrus.Fields{
			"requestID": reqID,
			"took":      time.Since(begin),
			"err":       err,
			"errCode":   code,
			"userID":    id,
		}).Info("GetUser")
	}(time.Now())

	return s.Next.GetUser(ctx, id)
}

func (s LoggingServiceStruct) ListUsers(ctx context.Context) (users []*user.User, code codes.Code, err error) {
	defer func(begin time.Time) {
		reqID := ctx.Value("requestID")

		logrus.WithFields(logrus.Fields{
			"requestID":  reqID,
			"took":       time.Since(begin),
			"err":        err,
			"errCode":    code,
			"usersCount": len(users),
		}).Info("ListUsers")
	}(time.Now())

	return s.Next.ListUsers(ctx)
}

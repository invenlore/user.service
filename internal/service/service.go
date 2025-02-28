package service

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
)

// MOCK
var users = map[string]string{
	"1": "email@email.com",
}

type UserService interface {
	GetUser(context.Context, string) (string, codes.Code, error)
}

type UserServiceStruct struct{}

func (s *UserServiceStruct) GetUser(_ context.Context, id string) (string, codes.Code, error) {
	email, ok := users[id]
	if !ok {
		return "", codes.NotFound, fmt.Errorf("User for id (%s) is not found", id)
	}

	return email, codes.OK, nil
}

type LoggingServiceStruct struct {
	Next UserService
}

func (s LoggingServiceStruct) GetUser(ctx context.Context, id string) (email string, code codes.Code, err error) {
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

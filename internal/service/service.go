package service

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// MOCK
var users = map[string]string{
	"1": "email@email.com",
}

type UserService interface {
	GetUser(context.Context, string) (string, error)
}

type UserServiceStruct struct{}

func (s *UserServiceStruct) GetUser(_ context.Context, id string) (string, error) {
	email, ok := users[id]
	if !ok {
		return "404", fmt.Errorf("User for id (%s) is not found", id)
	}

	return email, nil
}

type LoggingServiceStruct struct {
	Next UserService
}

func (s LoggingServiceStruct) GetUser(ctx context.Context, id string) (email string, err error) {
	defer func(begin time.Time) {
		reqID := ctx.Value("requestID")

		logrus.WithFields(logrus.Fields{
			"requestID": reqID,
			"took":      time.Since(begin),
			"err":       err,
			"userID":    id,
		}).Info("GetUser")
	}(time.Now())

	return s.Next.GetUser(ctx, id)
}

package service

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

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
		return "404", fmt.Errorf("price for ticker (%s) is not available", id)
	}

	return email, nil
}

type LoggingService struct {
	Next UserService
}

func (s LoggingService) GetUser(ctx context.Context, id string) (price string, err error) {
	defer func(begin time.Time) {
		reqID := ctx.Value("requestID")

		logrus.WithFields(logrus.Fields{
			"requestID": reqID,
			"took":      time.Since(begin),
			"err":       err,
			"id":        id,
		}).Info("GetUser")
	}(time.Now())

	return s.Next.GetUser(ctx, id)
}

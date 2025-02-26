package main

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

var user = map[string]string{
	"1": "email@email.com",
}

// PriceService is an interface that can fetch the price for any given ticker.
type UserService interface {
	GetUser(context.Context, string) (string, error)
}

type userService struct{}

// is the business logic
func (s *userService) GetUser(_ context.Context, id string) (string, error) {
	email, ok := user[id]
	if !ok {
		return "404", fmt.Errorf("price for ticker (%s) is not available", id)
	}

	return email, nil
}

type loggingService struct {
	next UserService
}

func (s loggingService) GetUser(ctx context.Context, id string) (price string, err error) {
	defer func(begin time.Time) {
		reqID := ctx.Value("requestID")

		logrus.WithFields(logrus.Fields{
			"requestID": reqID,
			"took":      time.Since(begin),
			"err":       err,
			"id":        id,
		}).Info("GetUser")
	}(time.Now())

	return s.next.GetUser(ctx, id)
}

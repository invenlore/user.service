package config

import (
	"sync"

	"github.com/caarlos0/env/v11"
	"github.com/sirupsen/logrus"
)

type config struct {
	APP_ENV     string `env:"APP_ENV" envDefault:"dev"`
	GRPC_PORT   string `env:"APP_GRPC_PORT" envDefault:"8081"`
	HEALTH_PORT string `env:"APP_HEALTH_PORT" envDefault:"8082"`
}

type Config interface {
	GetAppEnv() string
	GetGRPCPort() string
	GetHealthPort() string
}

var (
	once     sync.Once
	instance *config = nil
)

func (c *config) GetAppEnv() string {
	return c.APP_ENV
}

func (c *config) GetGRPCPort() string {
	return c.GRPC_PORT
}

func (c *config) GetHealthPort() string {
	return c.HEALTH_PORT
}

func GetConfig() Config {
	once.Do(func() {
		cfg, err := env.ParseAs[config]()
		if err != nil {
			logrus.Error("parsing environment variables failed")

			instance = nil
		}

		instance = &cfg
	})

	return instance
}

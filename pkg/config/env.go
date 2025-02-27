package config

import (
	"sync"

	"github.com/caarlos0/env/v11"
	"github.com/sirupsen/logrus"
)

type config struct {
	APP_ENV   string `env:"APP_ENV" envDefault:"dev"`
	GRPC_PORT string `env:"CONTAINER_GRPC_PORT" envDefault:"8080"`
	JSON_PORT string `env:"CONTAINER_JSON_PORT" envDefault:"8081"`
}

type Config interface {
	GetAppEnv() string
	GetGRPCPort() string
	GetJSONPort() string
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

func (c *config) GetJSONPort() string {
	return c.JSON_PORT
}

func GetConfig() Config {
	once.Do(func() {
		cfg, err := env.ParseAs[config]()
		if err != nil {
			logrus.Error("Parsing environment variables failed")

			instance = nil
		}

		instance = &cfg
	})

	return instance
}

package config

import (
	"sync"

	"github.com/caarlos0/env/v11"
	"github.com/sirupsen/logrus"
)

type config struct {
	APP_ENV  string `env:"APP_ENV" envDefault:"dev"`
	APP_HOST string `env:"APP_HOST" envDefault:""`
}

type Config interface {
	GetAppEnv() string
	GetAppHost() string
}

var (
	once     sync.Once
	instance *config = nil
)

func (c *config) GetAppEnv() string {
	return c.APP_ENV
}

func (c *config) GetAppHost() string {
	return c.APP_HOST
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

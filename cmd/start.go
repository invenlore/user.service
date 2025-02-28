package cmd

import (
	"github.com/invenlore/user.service/internal/service"
	"github.com/invenlore/user.service/internal/transport"
	"github.com/invenlore/user.service/pkg/config"
)

func Start() {
	var (
		svc = service.LoggingServiceStruct{Next: &service.UserServiceStruct{}}
		cfg = config.GetConfig()
	)

	go transport.StartGRPCServer(":"+cfg.GetGRPCPort(), svc)

	healthServer := transport.NewHealthServer(":" + cfg.GetHealthPort())
	healthServer.Run()
}

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

	go transport.StartGRPCServer(cfg.GetAppHost()+":8081", svc)

	healthServer := transport.NewHealthServer(cfg.GetAppHost() + ":80")
	healthServer.Run()
}

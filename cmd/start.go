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

	go transport.MakeGRPCServerAndRun(":"+cfg.GetGRPCPort(), svc)

	jsonServer := transport.NewJSONAPIServer(":"+cfg.GetJSONPort(), svc)
	jsonServer.Run()
}

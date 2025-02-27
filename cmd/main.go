package main

import (
	"flag"

	"github.com/invenlore/user.service/internal/server"
	"github.com/invenlore/user.service/internal/service"
)

func main() {
	var (
		grpcAddr = flag.String("grpc", ":4000", "listen address of the grpc transport")
		svc      = service.LoggingService{Next: &service.UserServiceStruct{}}
	)

	flag.Parse()

	go server.MakeGRPCServerAndRun(*grpcAddr, svc)
}

package main

import (
	"os"

	"github.com/invenlore/user.service/internal/server"
	"github.com/invenlore/user.service/internal/service"
)

func main() {
	var svc = service.LoggingServiceStruct{Next: &service.UserServiceStruct{}}

	grpcPort, exists := os.LookupEnv("CONTAINER_GRPC_PORT")
	if !exists {
		grpcPort = "8080"
	}

	jsonPort, exists := os.LookupEnv("CONTAINER_JSON_PORT")
	if !exists {
		jsonPort = "8081"
	}

	go server.MakeGRPCServerAndRun(":"+grpcPort, svc)

	jsonServer := server.NewJSONAPIServer(":"+jsonPort, svc)
	jsonServer.Run()
}

package main

import (
	"flag"
)

func main() {
	var (
		grpcAddr = flag.String("grpc", ":4000", "listen address of the grpc transport")
		svc      = loggingService{&userService{}}
	)

	flag.Parse()

	go makeGRPCServerAndRun(*grpcAddr, svc)
}

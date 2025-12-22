package transport

import (
	"fmt"
	"net"

	"github.com/invenlore/core/pkg/config"
	"github.com/invenlore/core/pkg/db"
	"github.com/invenlore/core/pkg/logger"
	"github.com/invenlore/core/pkg/recovery"
	"github.com/invenlore/proto/pkg/user"
	"github.com/invenlore/user.service/internal/service"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type GRPCUserServer struct {
	svc            service.UserService
	mongoReadiness *db.MongoReadiness
	user.UnimplementedUserServiceServer
}

func NewGRPCUserServer(svc service.UserService, mongoReadiness *db.MongoReadiness) *GRPCUserServer {
	return &GRPCUserServer{
		svc:            svc,
		mongoReadiness: mongoReadiness,
	}
}

func NewGRPCServer(
	cfg *config.GRPCServerConfig,
	svc service.UserService,
	mongoReadiness *db.MongoReadiness,
) (*grpc.Server, net.Listener, error) {
	var (
		loggerEntry = logrus.WithField("scope", "grpcServer")
		listenAddr  = net.JoinHostPort(cfg.Host, cfg.Port)
	)

	loggerEntry.Info("starting gRPC server...")

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return nil, nil, fmt.Errorf("gRPC server failed to listen on %s: %w", listenAddr, err)
	}

	unaryInterceptors := []grpc.UnaryServerInterceptor{
		recovery.RecoveryUnaryInterceptor,
		db.MongoGateUnary(mongoReadiness),
		logger.ServerRequestIDInterceptor,
		logger.ServerLoggingInterceptor,
	}

	streamInterceptors := []grpc.StreamServerInterceptor{
		recovery.RecoveryStreamInterceptor,
		db.MongoGateStream(mongoReadiness),
		logger.ServerStreamRequestIDInterceptor,
		logger.ServerStreamLoggingInterceptor,
	}

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(unaryInterceptors...),
		grpc.ChainStreamInterceptor(streamInterceptors...),
	)

	user.RegisterUserServiceServer(server, NewGRPCUserServer(svc, mongoReadiness))

	return server, ln, nil
}

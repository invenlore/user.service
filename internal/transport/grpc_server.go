package transport

import (
	"fmt"
	"net"

	"github.com/invenlore/core/pkg/config"
	"github.com/invenlore/core/pkg/db"
	"github.com/invenlore/core/pkg/logger"
	"github.com/invenlore/core/pkg/recovery"
	"github.com/invenlore/identity.service/internal/service"
	"github.com/invenlore/proto/pkg/identity"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type GRPCUserServer struct {
	svc            service.IdentityService
	mongoReadiness *db.MongoReadiness
	identity.UnimplementedIdentityServiceServer
}

func NewGRPCUserServer(svc service.IdentityService, mongoReadiness *db.MongoReadiness) *GRPCUserServer {
	return &GRPCUserServer{
		svc:            svc,
		mongoReadiness: mongoReadiness,
	}
}

func NewGRPCServer(cfg *config.GRPCServerConfig, svc service.IdentityService, mongoReadiness *db.MongoReadiness) (*grpc.Server, net.Listener, error) {
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
		logger.ServerRequestIDInterceptor,
		logger.ServerLoggingInterceptor,
		db.MongoGateUnary(mongoReadiness, identity.IdentityService_HealthCheck_FullMethodName),
	}

	streamInterceptors := []grpc.StreamServerInterceptor{
		recovery.RecoveryStreamInterceptor,
		logger.ServerStreamRequestIDInterceptor,
		logger.ServerStreamLoggingInterceptor,
		db.MongoGateStream(mongoReadiness, identity.IdentityService_HealthCheck_FullMethodName),
	}

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(unaryInterceptors...),
		grpc.ChainStreamInterceptor(streamInterceptors...),
	)

	identity.RegisterIdentityServiceServer(server, NewGRPCUserServer(svc, mongoReadiness))

	return server, ln, nil
}

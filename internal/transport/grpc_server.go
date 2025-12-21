package transport

import (
	"fmt"
	"net"

	"github.com/invenlore/core/pkg/config"
	"github.com/invenlore/core/pkg/logger"
	"github.com/invenlore/core/pkg/recovery"
	"github.com/invenlore/proto/pkg/user"
	"github.com/invenlore/user.service/internal/service"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
)

type GRPCUserServer struct {
	svc         service.UserService
	mongoClient *mongo.Client
	user.UnimplementedUserServiceServer
}

func NewGRPCUserServer(svc service.UserService, mongoClient *mongo.Client) *GRPCUserServer {
	return &GRPCUserServer{
		svc:         svc,
		mongoClient: mongoClient,
	}
}

func NewGRPCServer(
	cfg *config.GRPCServerConfig,
	svc service.UserService,
	mongoClient *mongo.Client,
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
		logger.ServerRequestIDInterceptor,
		logger.ServerLoggingInterceptor,
	}

	streamInterceptors := []grpc.StreamServerInterceptor{
		recovery.RecoveryStreamInterceptor,
		logger.ServerStreamRequestIDInterceptor,
		logger.ServerStreamLoggingInterceptor,
	}

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(unaryInterceptors...),
		grpc.ChainStreamInterceptor(streamInterceptors...),
	)

	user.RegisterUserServiceServer(server, NewGRPCUserServer(svc, mongoClient))
	return server, ln, nil
}

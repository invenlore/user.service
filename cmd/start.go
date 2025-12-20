package cmd

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/invenlore/core/pkg/config"
	"github.com/invenlore/core/pkg/db"
	"github.com/invenlore/user.service/internal/repository"
	"github.com/invenlore/user.service/internal/service"
	"github.com/invenlore/user.service/internal/transport"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func Start() {
	var (
		errChan  = make(chan error, 2)
		stopChan = make(chan os.Signal, 1)

		grpcServer           *grpc.Server = nil
		grpcServerListener   net.Listener = nil
		healthServer         *http.Server = nil
		healthServerListener net.Listener = nil

		serviceErr error = nil
	)

	logrus.Info("service starting...")

	cfg, err := config.Config()
	if err != nil {
		logrus.Fatalf("failed to load service configuration: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db := db.MongoDBConnect(cfg.GetMongoConfig())
	repo := repository.NewUserRepository(db, cfg.GetMongoConfig())
	svc := service.NewUserService(repo)

	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		var err error

		grpcServer, grpcServerListener, err = transport.StartGRPCServer(ctx, cfg.GetGRPCConfig(), svc, errChan)
		if err != nil {
			if grpcServerListener != nil {
				grpcServerListener.Close()
			}

			errChan <- fmt.Errorf("gRPC server failed to start: %w", err)
		}
	}()

	go func() {
		var err error

		healthServer, healthServerListener, err = transport.StartHealthServer(ctx, cfg.GetConfig(), errChan)
		if err != nil {
			if healthServerListener != nil {
				healthServerListener.Close()
			}

			errChan <- fmt.Errorf("health server failed to start: %w", err)
		}
	}()

	select {
	case err := <-errChan:
		serviceErr = err
		logrus.Errorf("service startup error: %v", serviceErr)

	case <-stopChan:
		logrus.Debug("received stop signal")
	}

	defer func() {
		logrus.Debug("attempting service graceful shutdown...")

		if db != nil {
			logrus.Info("closing MongoDB connection...")

			stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := db.Disconnect(stopCtx); err != nil {
				logrus.Errorf("MongoDB connection closing error: %v", err)
			} else {
				logrus.Info("MongoDB connection closed gracefully")
			}
		}

		if healthServer != nil {
			logrus.Info("stopping health server...")

			stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := healthServer.Shutdown(stopCtx); err != nil {
				logrus.Errorf("health server shutdown error: %v", err)
			} else {
				logrus.Info("health server stopped gracefully")
			}
		} else {
			logrus.Warn("health server was not started")
		}

		if grpcServer != nil {
			logrus.Info("stopping gRPC server...")
			grpcServer.GracefulStop()

			if grpcServerListener != nil {
				grpcServerListener.Close()
			}

			logrus.Info("gRPC server stopped gracefully")
		} else {
			logrus.Warn("gRPC server was not started")
		}

		logrus.Info("clean service shutdown complete")

		if serviceErr != nil {
			os.Exit(1)
		}
	}()
}

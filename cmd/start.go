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

	"github.com/invenlore/user.service/internal/service"
	"github.com/invenlore/user.service/internal/transport"
	"github.com/invenlore/user.service/pkg/config"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func Start() {
	var (
		cfg = config.GetConfig()

		svc = &service.UserServiceStruct{}

		errChan  = make(chan error, 2)
		stopChan = make(chan os.Signal, 1)

		grpcServer           *grpc.Server = nil
		healthServer         *http.Server = nil
		healthServerListener net.Listener = nil

		serviceErr error = nil
	)

	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	logrus.Info("service starting...")

	go func() {
		var err error

		grpcServer, err = transport.StartGRPCServer(cfg.GetAppHost()+":8081", svc, errChan)
		if err != nil {
			errChan <- fmt.Errorf("gRPC server failed to start: %w", err)
		}
	}()

	go func() {
		var err error

		healthServer, healthServerListener, err = transport.StartHealthServer(cfg.GetAppHost()+":80", errChan)
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
		logrus.Info("received stop signal")
	}

	defer func() {
		logrus.Info("attempting service graceful shutdown...")

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
			logrus.Info("health server was not started")
		}

		if grpcServer != nil {
			logrus.Info("stopping gRPC server...")
			grpcServer.GracefulStop()
			logrus.Info("gRPC server stopped gracefully")
		} else {
			logrus.Info("gRPC server was not started")
		}

		logrus.Info("clean service shutdown complete")

		if serviceErr != nil {
			os.Exit(1)
		}
	}()
}
